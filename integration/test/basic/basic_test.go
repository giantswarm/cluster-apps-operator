// +build k8srequired

package basic

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/giantswarm/backoff"
	"github.com/giantswarm/microerror"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/cluster-apps-operator/integration/key"
	"github.com/giantswarm/cluster-apps-operator/pkg/project"
)

// TestBasic is a smoke test to check the helm chart is installed and the
// operator starts without crashing.
//
// The operator functionality is tested via Tekton in the releases repo.
//
func TestBasic(t *testing.T) {
	var err error

	ctx := context.Background()

	{
		config.Logger.Debugf(ctx, "waiting for %#q pod", project.Name())

		err = waitForReadyDeployment(ctx)
		if err != nil {
			t.Fatalf("could not get %#q pod %#v", project.Name(), err)
		}

		config.Logger.Debugf(ctx, "waited for %#q pod", project.Name())
	}
}

func waitForReadyDeployment(ctx context.Context) error {
	var err error

	o := func() error {
		lo := metav1.ListOptions{
			LabelSelector: fmt.Sprintf("app=%s", project.Name()),
		}
		deploys, err := config.K8sClients.K8sClient().AppsV1().Deployments(key.Namespace()).List(ctx, lo)
		if err != nil {
			return microerror.Mask(err)
		}
		if len(deploys.Items) != 1 {
			return microerror.Maskf(executionFailedError, "expected 1 deployment got %d", len(deploys.Items))
		}

		deploy := deploys.Items[0]
		if *deploy.Spec.Replicas != deploy.Status.ReadyReplicas {
			return microerror.Maskf(executionFailedError, "expected %d ready pods got %d", *deploy.Spec.Replicas, deploy.Status.ReadyReplicas)
		}

		return nil
	}

	n := func(err error, t time.Duration) {
		log.Printf("waiting for ready deployment for %s: %#v", t, err)
	}

	err = backoff.RetryNotify(o, backoff.NewConstant(5*time.Minute, 15*time.Second), n)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}
