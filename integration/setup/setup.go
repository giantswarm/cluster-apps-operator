//go:build k8srequired
// +build k8srequired

package setup

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/giantswarm/apptest"
	"github.com/giantswarm/backoff"
	"github.com/giantswarm/microerror"

	"github.com/giantswarm/cluster-apps-operator/v3/integration/env"
	"github.com/giantswarm/cluster-apps-operator/v3/integration/key"
	"github.com/giantswarm/cluster-apps-operator/v3/integration/templates"
	"github.com/giantswarm/cluster-apps-operator/v3/pkg/project"
)

func Setup(m *testing.M, config Config) {
	var err error

	ctx := context.Background()

	err = installResources(ctx, config)
	if err != nil {
		config.Logger.LogCtx(ctx, "level", "error", "message", fmt.Sprintf("failed to install %#q", project.Name()), "stack", fmt.Sprintf("%#v", err))
		os.Exit(-1)
	}

	os.Exit(m.Run())
}

func installResources(ctx context.Context, config Config) error {
	var err error

	{
		apps := []apptest.App{
			{
				CatalogName:   key.ControlPlaneTestCatalogName(),
				Name:          project.Name(),
				Namespace:     key.Namespace(),
				SHA:           env.CircleSHA(),
				ValuesYAML:    templates.ClusterAppsOperatorValues,
				WaitForDeploy: true,
			},
		}

		o := func() error {
			err = config.AppTest.InstallApps(ctx, apps)
			if err != nil {
				return microerror.Mask(err)
			}

			return nil
		}

		b := backoff.NewConstant(5*time.Minute, 10*time.Second)
		n := backoff.NewNotifier(config.Logger, ctx)

		err = backoff.RetryNotify(o, b, n)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	return nil
}
