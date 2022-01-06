package app

import (
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"

	"github.com/giantswarm/cluster-apps-operator/service/internal/chartname"
	"github.com/giantswarm/cluster-apps-operator/service/internal/releaseversion"
)

const (
	// Name is the identifier of the resource.
	Name = "app"

	uniqueOperatorVersion = "0.0.0"
)

// Config represents the configuration used to create a new app resource.
type Config struct {
	ChartName      chartname.Interface
	G8sClient      client.Client
	K8sClient      kubernetes.Interface
	Logger         micrologger.Logger
	ReleaseVersion releaseversion.Interface

	RawAppDefaultConfig  string
	RawAppOverrideConfig string
}

// Resource implements the app resource.
type Resource struct {
	chartName      chartname.Interface
	g8sClient      client.Client
	k8sClient      kubernetes.Interface
	logger         micrologger.Logger
	releaseVersion releaseversion.Interface

	defaultConfig  defaultConfig
	overrideConfig overrideConfig
}

type defaultConfig struct {
	Catalog         string `json:"catalog"`
	Namespace       string `json:"namespace"`
	UseUpgradeForce bool   `json:"useUpgradeForce"`
}

type overrideProperties struct {
	Chart                  string `json:"chart"`
	HasClusterValuesSecret *bool  `json:"hasClusterValuesSecret,omitempty"`
	Namespace              string `json:"namespace"`
	UseUpgradeForce        *bool  `json:"useUpgradeForce,omitempty"`
}

type overrideConfig map[string]overrideProperties

// New creates a new configured app state getter resource managing
// app CRs.
//
//     https://pkg.go.dev/github.com/giantswarm/resource/v2/appresource#StateGetter
//
func New(config Config) (*Resource, error) {
	if config.ChartName == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.ChartName must not be empty", config)
	}
	if config.G8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.G8sClient must not be empty", config)
	}
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}
	if config.ReleaseVersion == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.ReleaseVersion must not be empty", config)
	}

	if config.RawAppDefaultConfig == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.RawDefaultConfig must not be empty", config)
	}
	if config.RawAppOverrideConfig == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.RawOverrideConfig must not be empty", config)
	}

	defaultConfig := defaultConfig{}
	err := yaml.Unmarshal([]byte(config.RawAppDefaultConfig), &defaultConfig)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	overrideConfig := overrideConfig{}
	err = yaml.Unmarshal([]byte(config.RawAppOverrideConfig), &overrideConfig)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r := &Resource{
		chartName:      config.ChartName,
		g8sClient:      config.G8sClient,
		k8sClient:      config.K8sClient,
		logger:         config.Logger,
		releaseVersion: config.ReleaseVersion,

		defaultConfig:  defaultConfig,
		overrideConfig: overrideConfig,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}
