package chartname

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	applicationv1alpha1 "github.com/giantswarm/apiextensions-application/api/v1alpha1"
	"github.com/giantswarm/backoff"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"

	"github.com/giantswarm/cluster-apps-operator/service/internal/chartname/internal/cache"
)

const (
	httpClientTimeout = 10 * time.Second
)

type Config struct {
	G8sClient client.Client
	Logger    micrologger.Logger
}

type ChartName struct {
	g8sClient  client.Client
	httpClient *http.Client
	indexCache *cache.Index
	logger     micrologger.Logger
}

func New(c Config) (*ChartName, error) {
	if c.G8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.G8sClient must not be empty", c)
	}
	if c.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", c)
	}

	// Set client timeout to prevent leakages.
	httpClient := &http.Client{
		Timeout: httpClientTimeout,
	}

	cn := &ChartName{
		g8sClient:  c.G8sClient,
		httpClient: httpClient,
		indexCache: cache.NewIndex(),
		logger:     c.Logger,
	}

	return cn, nil
}

// ChartName returns the name of the chart as it appears in the index.yaml
// for the catalog.
func (cn *ChartName) ChartName(ctx context.Context, catalogName, appName, appVersion string) (string, error) {
	var index catalogIndex

	index, err := cn.cachedCatalogIndex(ctx, catalogName)
	if err != nil {
		return "", microerror.Mask(err)
	}

	appNameWithoutAppSuffix := strings.TrimSuffix(appName, "-app")
	appNameWithAppSuffix := fmt.Sprintf("%s-app", appNameWithoutAppSuffix)
	chartName := ""

	entries, ok := index.Entries[appNameWithAppSuffix]
	if !ok || len(entries) == 0 {
		entries, ok = index.Entries[appNameWithoutAppSuffix]
		if !ok || len(entries) == 0 {
			return "", microerror.Maskf(notFoundError, "could not find chart %s in %s catalog", appName, catalogName)
		}
		chartName = appNameWithoutAppSuffix
	} else {
		chartName = appNameWithAppSuffix
	}

	for _, entry := range entries {
		if entry.Version == appVersion && entry.Name == chartName {
			return entry.Name, nil
		}
	}

	return "", microerror.Maskf(notFoundError, "could not find chart %s in %s catalog", appName, catalogName)
}

func (cn *ChartName) cachedCatalogIndex(ctx context.Context, catalogName string) (catalogIndex, error) {
	var index catalogIndex
	var err error

	var catalog applicationv1alpha1.AppCatalog
	{
		err = cn.g8sClient.Get(ctx, client.ObjectKey{
			Name: catalogName,
		}, &catalog)
		if err != nil {
			return index, microerror.Mask(err)
		}
	}

	{
		indexYaml, ok := cn.indexCache.Get(ctx, catalogName)
		if !ok {
			indexYaml, err = cn.fetchCatalogIndex(ctx, catalogName, catalog.Spec.Storage.URL)
			if err != nil {
				return index, microerror.Mask(err)
			}

			cn.indexCache.Set(ctx, catalogName, indexYaml)
		}

		err = yaml.Unmarshal(indexYaml, &index)
		if err != nil {
			return index, microerror.Mask(err)
		}
	}

	return index, nil
}

func (cn *ChartName) fetchCatalogIndex(ctx context.Context, catalogName, catalogURL string) ([]byte, error) {
	var err error

	url := strings.TrimRight(catalogURL, "/") + "/index.yaml"
	body := []byte{}

	o := func() error {
		request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, &bytes.Buffer{}) // nolint: gosec
		if err != nil {
			return microerror.Mask(err)
		}
		response, err := cn.httpClient.Do(request)
		if err != nil {
			return microerror.Mask(err)
		}
		defer response.Body.Close()

		body, err = ioutil.ReadAll(response.Body)
		if err != nil {
			return microerror.Mask(err)
		}

		return nil
	}
	b := backoff.NewExponential(30*time.Second, 5*time.Second)
	n := backoff.NewNotifier(cn.logger, ctx)

	err = backoff.RetryNotify(o, b, n)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return body, nil
}
