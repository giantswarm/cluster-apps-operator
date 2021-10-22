//go:build functional || smoke
// +build functional smoke

package ats

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/giantswarm/backoff"
	"github.com/giantswarm/microerror"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/cluster-apps-operator/pkg/project"
)

const (
	namespace = metav1.NamespaceDefault
)

// TestBasic is a smoke test to check the helm chart is installed and the
// operator starts without crashing.
//
// The operator functionality is tested via Tekton in the releases repo.
//
func TestBasic(t *testing.T) {
	var err error

	ctx := context.Background()

	var logger micrologger.Logger
	{
		c := micrologger.Config{}

		logger, err = micrologger.New(c)
		if err != nil {
			t.Fatalf("could not create logger %v", err)
		}
	}

	var k8sClients *k8sclient.Clients
	{
		c := k8sclient.ClientsConfig{
			Logger: logger,

			KubeConfigPath: KubeConfigPath(),
		}

		k8sClients, err = k8sclient.NewClients(c)
		if err != nil {
			t.Fatalf("could not create k8sclients %v", err)
		}
	}

	{
		logger.Debugf(ctx, "waiting for ready %#q deployment", project.Name())

		err = waitForReadyDeployment(ctx, k8sClients)
		if err != nil {
			t.Fatalf("could not get ready %#q deployment %#v", project.Name(), err)
		}

		logger.Debugf(ctx, "waited for ready %#q deployment", project.Name())
	}
}

func waitForReadyDeployment(ctx context.Context, k8sClients *k8sclient.Clients) error {
	var err error

	o := func() error {
		lo := metav1.ListOptions{
			LabelSelector: fmt.Sprintf("app=%s", project.Name()),
		}
		deploys, err := k8sClients.K8sClient().AppsV1().Deployments(namespace).List(ctx, lo)
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
