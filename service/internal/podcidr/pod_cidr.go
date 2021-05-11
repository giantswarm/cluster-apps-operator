package podcidr

import (
	"context"

	"github.com/giantswarm/k8sclient/v5/pkg/k8sclient"
	"github.com/giantswarm/microerror"
)

type Config struct {
	K8sClient k8sclient.Interface

	InstallationCIDR string
}

type PodCIDR struct {
	k8sClient k8sclient.Interface

	installationCIDR string
}

func New(c Config) (*PodCIDR, error) {
	if c.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", c)
	}

	if c.InstallationCIDR == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.InstallationCIDR must not be empty", c)
	}

	p := &PodCIDR{
		k8sClient: c.K8sClient,

		installationCIDR: c.InstallationCIDR,
	}

	return p, nil
}

func (p *PodCIDR) PodCIDR(ctx context.Context, obj interface{}) (string, error) {
	// TODO Add logic.
	return "", nil
}
