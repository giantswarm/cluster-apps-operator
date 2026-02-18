package clustersecret

import (
	"github.com/giantswarm/k8sclient/v8/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/cluster-apps-operator/v3/flag/service/proxy"
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
	Proxy     proxy.Proxy
}

// Resource implements the clustersecret resource.
type Resource struct {
	k8sClient k8sclient.Interface
	logger    micrologger.Logger
	proxy     proxy.Proxy
}

// New creates a new configured secret state getter resource managing
// cluster secrets.
//
//	https://pkg.go.dev/github.com/giantswarm/operatorkit/v4/pkg/resource/k8s/secretresource#StateGetter
func New(config Config) (*Resource, error) {
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	r := &Resource{
		k8sClient: config.K8sClient,
		logger:    config.Logger,
		proxy:     config.Proxy,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}
