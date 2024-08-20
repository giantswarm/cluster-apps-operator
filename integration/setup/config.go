//go:build k8srequired
// +build k8srequired

package setup

import (
	"github.com/giantswarm/apptest"
	"github.com/giantswarm/k8sclient/v7/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	capi "sigs.k8s.io/cluster-api/api/v1beta1"

	"github.com/giantswarm/cluster-apps-operator/v2/integration/env"
	"github.com/giantswarm/cluster-apps-operator/v2/integration/release"
)

type Config struct {
	AppTest    apptest.Interface
	K8s        *k8sclient.Setup
	K8sClients k8sclient.Interface
	Logger     micrologger.Logger
	Release    *release.Release
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
			SchemeBuilder: k8sclient.SchemeBuilder{
				capi.AddToScheme,
			},

			KubeConfigPath: env.KubeConfigPath(),
		}

		k8sClients, err = k8sclient.NewClients(c)
		if err != nil {
			return Config{}, microerror.Mask(err)
		}
	}

	var k8sSetup *k8sclient.Setup
	{
		c := k8sclient.SetupConfig{
			Clients: k8sClients,
			Logger:  logger,
		}

		k8sSetup, err = k8sclient.NewSetup(c)
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

	var releaseMgmt *release.Release
	{
		c := release.Config{
			K8sClients: k8sClients,
			Logger:     logger,
		}

		releaseMgmt, err = release.New(c)
		if err != nil {
			return Config{}, microerror.Mask(err)
		}
	}

	c := Config{
		AppTest:    appTest,
		K8sClients: k8sClients,
		K8s:        k8sSetup,
		Logger:     logger,
		Release:    releaseMgmt,
	}

	return c, nil
}
