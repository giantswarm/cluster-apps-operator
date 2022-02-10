package clusterconfigmap

import (
	"strings"

	"github.com/giantswarm/k8sclient/v7/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/cluster-apps-operator/service/internal/podcidr"
)

const (
	// Name is the identifier of the resource.
	Name = "clusterconfigmap"
)

// Config represents the configuration used to create a new clusterConfigMap
// resource.
type Config struct {
	K8sClient k8sclient.Interface
	Logger    micrologger.Logger
	PodCIDR   podcidr.Interface

	BaseDomain        string
	ClusterIPRange    string
	DNSIP             string
	ManagementCluster string
	Provider          string
	RegistryDomain    string
}

// Resource implements the clusterConfigMap resource.
type Resource struct {
	k8sClient k8sclient.Interface
	logger    micrologger.Logger
	podCIDR   podcidr.Interface

	baseDomain        string
	clusterIPRange    string
	dnsIP             string
	managementCluster string
	provider          string
	registryDomain    string
}

// New creates a new configured config map state getter resource managing
// cluster config maps.
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
	if config.PodCIDR == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.PodCIDR must not be empty", config)
	}

	if config.BaseDomain == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.BaseDomain must not be empty", config)
	}
	if config.ClusterIPRange == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.ClusterIPRange must not be empty", config)
	}
	if config.DNSIP == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.DNSIP must not be empty", config)
	}
	if config.ManagementCluster == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.ManagementCluster must not be empty", config)
	}
	if config.Provider == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.Provider must not be empty", config)
	}
	if config.RegistryDomain == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.RegistryDomain must not be empty", config)
	}

	r := &Resource{
		k8sClient: config.K8sClient,
		logger:    config.Logger,
		podCIDR:   config.PodCIDR,

		baseDomain:        strings.TrimPrefix(config.BaseDomain, "k8s."),
		clusterIPRange:    config.ClusterIPRange,
		dnsIP:             config.DNSIP,
		managementCluster: config.ManagementCluster,
		provider:          config.Provider,
		registryDomain:    config.RegistryDomain,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}
