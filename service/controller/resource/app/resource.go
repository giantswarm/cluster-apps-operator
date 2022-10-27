package app

import (
	"github.com/giantswarm/apiextensions-application/api/v1alpha1"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	// Name is the identifier of the resource.
	Name = "app"

	uniqueOperatorVersion = "0.0.0"
)

// Config represents the configuration used to create a new app resource.
type Config struct {
	CtrlClient client.Client
	Logger     micrologger.Logger

	AppOperatorCatalog   string
	AppOperatorVersion   string
	ChartOperatorCatalog string
	ChartOperatorVersion string
}

// Resource implements the app resource.
type Resource struct {
	ctrlClient client.Client
	logger     micrologger.Logger

	appOperatorCatalog   string
	appOperatorVersion   string
	chartOperatorCatalog string
	chartOperatorVersion string
}

// New creates a new configured app state getter resource managing
// app CRs.
//
//	https://pkg.go.dev/github.com/giantswarm/resource/v2/appresource#StateGetter
func New(config Config) (*Resource, error) {
	if config.CtrlClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.CtrlClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	if config.AppOperatorCatalog == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.AppOperatorCatalog must not be empty", config)
	}
	if config.AppOperatorVersion == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.AppOperatorVersion must not be empty", config)
	}
	if config.ChartOperatorCatalog == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.ChartOperatorCatalog must not be empty", config)
	}
	if config.ChartOperatorVersion == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.ChartOperatorVersion must not be empty", config)
	}

	r := &Resource{
		ctrlClient: config.CtrlClient,
		logger:     config.Logger,

		appOperatorCatalog:   config.AppOperatorCatalog,
		appOperatorVersion:   config.AppOperatorVersion,
		chartOperatorCatalog: config.ChartOperatorCatalog,
		chartOperatorVersion: config.ChartOperatorVersion,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}

func findAppByName(apps []*v1alpha1.App, name, namespace string) *v1alpha1.App {
	for _, a := range apps {
		if name == a.Name && namespace == a.Namespace {
			return a
		}
	}

	return nil
}
