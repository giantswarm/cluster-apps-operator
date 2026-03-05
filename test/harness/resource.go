package harness

import (
	"context"

	appresource "github.com/giantswarm/cluster-apps-operator/v3/service/controller/resource/app"
	configmapresource "github.com/giantswarm/cluster-apps-operator/v3/service/controller/resource/clusterconfigmap"
	secretresource "github.com/giantswarm/cluster-apps-operator/v3/service/controller/resource/clustersecret"
	"github.com/giantswarm/cluster-apps-operator/v3/service/collector"

	"github.com/giantswarm/cluster-apps-operator/v3/flag/service/proxy"
)

// staticPodCIDR satisfies the podcidr.Interface from the internal package
// using Go's structural typing.
type staticPodCIDR struct {
	cidr string
}

func (s *staticPodCIDR) PodCIDR(_ context.Context, _ interface{}) (string, error) {
	return s.cidr, nil
}

// ConfigMapResourceOpts configures the clusterconfigmap resource for testing.
type ConfigMapResourceOpts struct {
	BaseDomain          string
	ClusterIPRange      string
	DNSIP               string
	ManagementClusterID string
	RegistryDomain      string
	InstallationCIDR    string
	Proxy               proxy.Proxy
}

// DefaultConfigMapResourceOpts returns sensible defaults for testing.
func DefaultConfigMapResourceOpts() ConfigMapResourceOpts {
	return ConfigMapResourceOpts{
		BaseDomain:       "test.gigantic.io",
		ClusterIPRange:   "10.96.0.0/12",
		DNSIP:            "10.96.0.10",
		RegistryDomain:   "gsoci.azurecr.io/giantswarm",
		InstallationCIDR: "10.0.0.0/16",
	}
}

// NewConfigMapResource creates a clusterconfigmap.Resource with test defaults.
func (e *TestEnv) NewConfigMapResource(opts ConfigMapResourceOpts) *configmapresource.Resource {
	e.T.Helper()

	r, err := configmapresource.New(configmapresource.Config{
		K8sClient:           e.k8sClient,
		Logger:              e.logger,
		PodCIDR:             &staticPodCIDR{cidr: opts.InstallationCIDR},
		BaseDomain:          opts.BaseDomain,
		ClusterIPRange:      opts.ClusterIPRange,
		DNSIP:               opts.DNSIP,
		ManagementClusterID: opts.ManagementClusterID,
		RegistryDomain:      opts.RegistryDomain,
		Proxy:               opts.Proxy,
	})
	if err != nil {
		e.T.Fatalf("harness: create configmap resource: %v", err)
	}

	return r
}

// AppResourceOpts configures the app resource for testing.
type AppResourceOpts struct {
	AppOperatorCatalog   string
	AppOperatorVersion   string
	ChartOperatorCatalog string
	ChartOperatorVersion string
}

// DefaultAppResourceOpts returns sensible defaults for testing.
func DefaultAppResourceOpts() AppResourceOpts {
	return AppResourceOpts{
		AppOperatorCatalog:   "control-plane-catalog",
		AppOperatorVersion:   "1.0.0",
		ChartOperatorCatalog: "default",
		ChartOperatorVersion: "1.0.0",
	}
}

// NewAppResource creates an app.Resource with test defaults.
func (e *TestEnv) NewAppResource(opts AppResourceOpts) *appresource.Resource {
	e.T.Helper()

	r, err := appresource.New(appresource.Config{
		CtrlClient:           e.ctrlClient,
		Logger:               e.logger,
		AppOperatorCatalog:   opts.AppOperatorCatalog,
		AppOperatorVersion:   opts.AppOperatorVersion,
		ChartOperatorCatalog: opts.ChartOperatorCatalog,
		ChartOperatorVersion: opts.ChartOperatorVersion,
	})
	if err != nil {
		e.T.Fatalf("harness: create app resource: %v", err)
	}

	return r
}

// SecretResourceOpts configures the clustersecret resource for testing.
type SecretResourceOpts struct {
	Proxy proxy.Proxy
}

// NewSecretResource creates a clustersecret.Resource with test defaults.
func (e *TestEnv) NewSecretResource(opts SecretResourceOpts) *secretresource.Resource {
	e.T.Helper()

	r, err := secretresource.New(secretresource.Config{
		K8sClient: e.k8sClient,
		Logger:    e.logger,
		Proxy:     opts.Proxy,
	})
	if err != nil {
		e.T.Fatalf("harness: create secret resource: %v", err)
	}

	return r
}

// NewCollector creates a collector.Cluster for testing.
func (e *TestEnv) NewCollector() *collector.Cluster {
	e.T.Helper()

	c, err := collector.NewCluster(collector.ClusterConfig{
		K8sClient: e.k8sClient,
		Logger:    e.logger,
	})
	if err != nil {
		e.T.Fatalf("harness: create collector: %v", err)
	}

	return c
}
