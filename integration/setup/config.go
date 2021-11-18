//go:build k8srequired
// +build k8srequired

package setup

import (
	"github.com/giantswarm/apptest"
	"github.com/giantswarm/k8sclient/v6/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/cluster-apps-operator/integration/env"
)

type Config struct {
	AppTest    apptest.Interface
	K8sClients k8sclient.Interface
	Logger     micrologger.Logger
}

func NewConfig() (Config, error) {
	var err error

	var logger micrologger.Logger
	{
		c := micrologger.Config{}

		logger, err = micrologger.New(c)
		if err != nil {
			return Config{}, microerror.Mask(err)
		}
	}

	var k8sClients *k8sclient.Clients
	{
		c := k8sclient.ClientsConfig{
			Logger: logger,

			KubeConfigPath: env.KubeConfigPath(),
		}

		k8sClients, err = k8sclient.NewClients(c)
		if err != nil {
			return Config{}, microerror.Mask(err)
		}
	}

	var appTest apptest.Interface
	{
		c := apptest.Config{
			Logger: logger,

			KubeConfigPath: env.KubeConfigPath(),
		}

		appTest, err = apptest.New(c)
		if err != nil {
			return Config{}, microerror.Mask(err)
		}
	}

	c := Config{
		AppTest:    appTest,
		K8sClients: k8sClients,
		Logger:     logger,
	}

	return c, nil
}
