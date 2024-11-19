package clusterconfigmap

import (
	"strings"

	"github.com/giantswarm/k8sclient/v7/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/cluster-apps-operator/v3/flag/service/proxy"
	"github.com/giantswarm/cluster-apps-operator/v3/service/internal/podcidr"
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

	BaseDomain          string
	ClusterIPRange      string
	DNSIP               string
	ManagementClusterID string
	RegistryDomain      string
	Proxy               proxy.Proxy
}

// Resource implements the clusterConfigMap resource.
type Resource struct {
	k8sClient k8sclient.Interface
	logger    micrologger.Logger
	podCIDR   podcidr.Interface

	baseDomain string
	// clusterIPRange is the CIDR for the k8s `Services`.
	clusterIPRange string
	// dnsIP is the 10th IP within the `clusterIPRange` CIDR, that will be used for the coredns `Service`.
	dnsIP               string
	managementClusterID string
	registryDomain      string
	proxy               proxy.Proxy
}

// New creates a new configured config map state getter resource managing
// cluster config maps.
//
//	https://pkg.go.dev/github.com/giantswarm/operatorkit/v4/pkg/resource/k8s/secretresource#StateGetter
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
	if config.RegistryDomain == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.RegistryDomain must not be empty", config)
	}

	r := &Resource{
		k8sClient: config.K8sClient,
		logger:    config.Logger,
		podCIDR:   config.PodCIDR,

		baseDomain:          strings.TrimPrefix(config.BaseDomain, "k8s."),
		clusterIPRange:      config.ClusterIPRange,
		dnsIP:               config.DNSIP,
		managementClusterID: config.ManagementClusterID,
		registryDomain:      config.RegistryDomain,
		proxy:               config.Proxy,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}
