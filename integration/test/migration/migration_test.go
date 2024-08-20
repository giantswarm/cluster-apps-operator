//go:build k8srequired
// +build k8srequired

package migration

import (
	"context"
	"fmt"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
	capi "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/yaml"

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

		err = config.Release.WaitForReadyDeployment(ctx, key.Namespace())
		if err != nil {
			t.Fatalf("could not get ready %#q deployment %#v", project.Name(), err)
		}
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
	}

	// Create Cluster CR and wait for App CRs to be created
	{
		config.Logger.Debugf(ctx, "creating %#q Cluster CR in %#q namespace", key.ClusterID(), key.OrganizationNamespace())

		err = config.K8sClients.CtrlClient().Create(ctx, key.TestKindCluster(false))
		if err != nil {
			t.Fatalf("expected nil got %#v", err)
		}

		config.Logger.Debugf(
			ctx,
			"waiting for %#q and %#q App CRs in %#q namespace",
			key.KindAppOperatorName(),
			key.KindChartOperatorName(),
			key.OrganizationNamespace(),
		)

		err = config.Release.WaitForAppCreate(ctx, key.OrganizationNamespace(), key.KindAppOperatorName())
		if err != nil {
			t.Fatalf("expected nil got %#v", err)
		}

		err = config.Release.WaitForAppCreate(ctx, key.OrganizationNamespace(), key.KindChartOperatorName())
		if err != nil {
			t.Fatalf("expected nil got %#v", err)
		}
	}

	// Validate App Operator is configured correct, that is Flux backend is disabled
	{
		config.Logger.Debugf(
			ctx,
			"validating %#q has Flux backend disabled",
			fmt.Sprintf("%s/%s", key.OrganizationNamespace(), key.KindAppOperatorName()),
		)

		appOpCm, err := config.K8sClients.K8sClient().CoreV1().ConfigMaps(key.OrganizationNamespace()).Get(
			ctx,
			key.KindAppOperatorValuesName(),
			metav1.GetOptions{},
		)
		if err != nil {
			t.Fatalf("expected nil got %#v", err)
		}

		type appOperatorValues struct {
			App struct {
				HelmControllerBackend string
			}
		}

		var val appOperatorValues
		err = yaml.Unmarshal([]byte(appOpCm.Data["values"]), &val)
		if err != nil {
			t.Fatalf("expected nil got %#v", err)
		}

		if val.App.HelmControllerBackend != "" {
			t.Fatalf("expected \"\" got %#v", val.App.HelmControllerBackend)
		}
	}

	// Update Cluster CR to enable Flux backend
	{
		var cluster capi.Cluster
		config.Logger.Debugf(ctx, "getting %#q Cluster CR in %#q namespace", key.ClusterID(), key.OrganizationNamespace())
		err = config.K8sClients.CtrlClient().Get(
			ctx,
			types.NamespacedName{Namespace: key.OrganizationNamespace(), Name: key.ClusterID()},
			&cluster,
		)
		if err != nil {
			t.Fatalf("expected nil got %#v", err)
		}

		cluster.ObjectMeta.Labels["app-operator.giantswarm.io/flux-backend"] = ""

		config.Logger.Debugf(ctx, "updating %#q Cluster CR in %#q namespace", key.ClusterID(), key.OrganizationNamespace())

		err = config.K8sClients.CtrlClient().Update(ctx, &cluster)
		if err != nil {
			t.Fatalf("expected nil got %#v", err)
		}
	}

	// Validating App Operator App CR is there and that Chart Operator App CR is gone
	{
		config.Logger.Debugf(
			ctx,
			"validating %#q App CR in %#q namespace is still present",
			key.KindAppOperatorName(),
			key.OrganizationNamespace(),
		)

		err = config.Release.WaitForAppCreate(ctx, key.OrganizationNamespace(), key.KindAppOperatorName())
		if err != nil {
			t.Fatalf("expected nil got %#v", err)
		}

		config.Logger.Debugf(
			ctx,
			"validating %#q App CR in %#q namespace is gone",
			key.KindChartOperatorName(),
			key.OrganizationNamespace(),
		)

		err = config.Release.WaitForAppDelete(ctx, key.OrganizationNamespace(), key.KindChartOperatorName())
		if err != nil {
			t.Fatalf("expected nil got %#v", err)
		}

	}

	// Validate App Operator is configured correct, that is Flux backend is disabled
	{
		config.Logger.Debugf(
			ctx,
			"validating %#q has Flux backend enabled",
			fmt.Sprintf("%s/%s", key.OrganizationNamespace(), key.KindAppOperatorName()),
		)

		appOpCm, err := config.K8sClients.K8sClient().CoreV1().ConfigMaps(key.OrganizationNamespace()).Get(
			ctx,
			key.KindAppOperatorValuesName(),
			metav1.GetOptions{},
		)
		if err != nil {
			t.Fatalf("expected nil got %#v", err)
		}

		type appOperatorValues struct {
			App struct {
				HelmControllerBackend string
			}
		}

		var val appOperatorValues
		err = yaml.Unmarshal([]byte(appOpCm.Data["values"]), &val)
		if err != nil {
			t.Fatalf("expected nil got %#v", err)
		}

		if val.App.HelmControllerBackend != "true" {
			t.Fatalf("expected \"true\" got %#v", val.App.HelmControllerBackend)
		}
	}

	// Delete Cluster CR and wait for App CR to be deleted
	{
		config.Logger.Debugf(ctx, "deleting %#q Cluster CR in %#q namespace", key.ClusterID(), key.OrganizationNamespace())

		err = config.K8sClients.CtrlClient().Delete(ctx, key.TestKindCluster(false))
		if err != nil {
			t.Fatalf("expected nil got %#v", err)
		}

		config.Logger.Debugf(
			ctx,
			"waiting for %#q App CR deletion in %#q namespace",
			key.KindAppOperatorName(),
			key.OrganizationNamespace(),
		)

		err = config.Release.WaitForAppDelete(ctx, key.OrganizationNamespace(), key.KindAppOperatorName())
		if err != nil {
			t.Fatalf("expected nil got %#v", err)
		}
	}
}
