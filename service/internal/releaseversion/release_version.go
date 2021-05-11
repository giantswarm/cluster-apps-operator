package releaseversion

import (
	"context"

	"github.com/giantswarm/k8sclient/v5/pkg/k8sclient"
	"github.com/giantswarm/microerror"
)

type Config struct {
	K8sClient k8sclient.Interface
}

type ReleaseVersion struct {
	k8sClient k8sclient.Interface
}

func New(c Config) (*ReleaseVersion, error) {
	if c.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", c)
	}

	rv := &ReleaseVersion{
		k8sClient: c.K8sClient,
	}

	return rv, nil
}

func (rv *ReleaseVersion) Apps(ctx context.Context, obj interface{}) (map[string]ReleaseApp, error) {
	// TODO Add logic.
	return nil, nil
}

func (rv *ReleaseVersion) ComponentVersion(ctx context.Context, obj interface{}) (map[string]ReleaseComponent, error) {
	// TODO Add logic.
	return nil, nil
}
