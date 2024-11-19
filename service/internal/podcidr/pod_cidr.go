package podcidr

import (
	"context"

	"github.com/giantswarm/microerror"

	"github.com/giantswarm/cluster-apps-operator/v3/service/controller/key"
)

type Config struct {
	InstallationCIDR string
}

type PodCIDR struct {
	installationCIDR string
}

func New(c Config) (*PodCIDR, error) {
	if c.InstallationCIDR == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.InstallationCIDR must not be empty", c)
	}

	p := &PodCIDR{
		installationCIDR: c.InstallationCIDR,
	}

	return p, nil
}

func (p *PodCIDR) PodCIDR(ctx context.Context, obj interface{}) (string, error) {
	cr, err := key.ToCluster(obj)
	if err != nil {
		return "", microerror.Mask(err)
	}

	var podCIDR string
	podCIDR = key.PodCIDR(cr)
	if podCIDR == "" {
		podCIDR = p.installationCIDR
	}

	return podCIDR, nil
}
