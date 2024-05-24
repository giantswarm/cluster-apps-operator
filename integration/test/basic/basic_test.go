//go:build k8srequired
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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
	capi "sigs.k8s.io/cluster-api/api/v1beta1"

	"github.com/giantswarm/cluster-apps-operator/v2/integration/env"
	"github.com/giantswarm/cluster-apps-operator/v2/integration/key"
	"github.com/giantswarm/cluster-apps-operator/v2/pkg/project"
)

const (
	clusterName    = "kind-kind"
	kubeConfigName = "kube-config"
)

// TestBasic is a smoke test to check the helm chart is installed and the
// operator starts without crashing.
//
// The operator functionality is tested via Tekton in the releases repo.
func TestBasic(t *testing.T) {
	var err error

	ctx := context.Background()

	{
		config.Logger.Debugf(ctx, "waiting for ready %#q deployment", project.Name())

		err = waitForReadyDeployment(ctx)
		if err != nil {
			t.Fatalf("could not get ready %#q deployment %#v", project.Name(), err)
		}

		config.Logger.Debugf(ctx, "waited for ready %#q deployment", project.Name())
	}

	{
		err = config.K8s.EnsureNamespaceCreated(ctx, key.OrganizationNamespace())
		if err != nil {
			t.Fatalf("expected nil got %#v", err)
		}
	}

	// Transform kubeconfig file to restconfig and flatten.
	{
		c := clientcmd.GetConfigFromFileOrDie(env.KubeConfigPath())

		// Extract KIND kubeconfig settings. This is for local testing as
		// api.FlattenConfig does not work with file paths in kubeconfigs.
		clusterKubeConfig := &api.Config{
			AuthInfos: map[string]*api.AuthInfo{
				clusterName: c.AuthInfos[clusterName],
			},
			Clusters: map[string]*api.Cluster{
				clusterName: c.Clusters[clusterName],
			},
			Contexts: map[string]*api.Context{
				clusterName: c.Contexts[clusterName],
			},
		}

		err = api.FlattenConfig(clusterKubeConfig)
		if err != nil {
			t.Fatalf("expected nil got %#v", err)
		}

		// Normally KIND assigns 127.0.0.1 as the server address. For this test
		// that should change to the Kubernetes service.
		clusterKubeConfig.Clusters[clusterName].Server = "https://kubernetes.default.svc.cluster.local"

		bytes, err := clientcmd.Write(*c)
		if err != nil {
			t.Fatalf("expected nil got %#v", err)
		}

		// Create kubeconfig secret for the chart CR watcher in app-operator.
		secret := &corev1.Secret{
			Data: map[string][]byte{
				"value": bytes,
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("%s-kubeconfig", key.ClusterID()),
				Namespace: key.OrganizationNamespace(),
			},
		}
		_, err = config.K8sClients.K8sClient().CoreV1().Secrets(key.OrganizationNamespace()).Create(ctx, secret, metav1.CreateOptions{})
		if err != nil {
			t.Fatalf("expected nil got %#v", err)
		}

		//kubeConfig = string(bytes)
	}

	{
		testCapiCluster := capi.Cluster{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Cluster",
				APIVersion: "cluster.x-k8s.io",
			},
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{
					"cluster-apps-operator.giantswarm.io/watching": "",
					"cluster.x-k8s.io/cluster-name":                "kind",
				},
				Name:      "kind",
				Namespace: "org-test",
			},
			Spec: capi.ClusterSpec{
				InfrastructureRef: &corev1.ObjectReference{
					Kind: "KindCluster",
				},
			},
		}

		err = config.K8sClients.CtrlClient().Create(ctx, &testCapiCluster)
		if err != nil {
			t.Fatalf("expected nil got %#v", err)
		}
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
