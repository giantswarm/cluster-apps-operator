package clustersecret

import (
	"github.com/giantswarm/k8sclient/v6/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"sigs.k8s.io/yaml"
)

const (
	// Name is the identifier of the resource.
	Name = "clustersecret"
)

// Config represents the configuration used to create a new clustersecret
// resource.
type Config struct {
	K8sClient k8sclient.Interface
	Logger    micrologger.Logger

	RawAppOverrideConfig string
}

// Resource implements the clustersecret resource.
type Resource struct {
	k8sClient k8sclient.Interface
	logger    micrologger.Logger

	enabled bool
}

type overrideProperties struct {
	HasClusterValuesSecret *bool `json:"hasClusterValuesSecret,omitempty"`
}

type overrideConfig map[string]overrideProperties

// New creates a new configured secret state getter resource managing
// cluster secrets.
//
//     https://pkg.go.dev/github.com/giantswarm/operatorkit/v4/pkg/resource/k8s/secretresource#StateGetter
//
func New(config Config) (*Resource, error) {
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	if config.RawAppOverrideConfig == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.RawOverrideConfig must not be empty", config)
	}

	overrideConfig := overrideConfig{}
	err := yaml.Unmarshal([]byte(config.RawAppOverrideConfig), &overrideConfig)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var enabled bool
	for _, appOverrides := range overrideConfig {
		if appOverrides.HasClusterValuesSecret != nil && *appOverrides.HasClusterValuesSecret {
			enabled = true
			break
		}
	}

	r := &Resource{
		k8sClient: config.K8sClient,
		logger:    config.Logger,

		enabled: enabled,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}
