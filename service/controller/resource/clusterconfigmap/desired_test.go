package clusterconfigmap

import (
	"context"
	"strings"
	"testing"

	"github.com/giantswarm/k8sclient/v7/pkg/k8sclienttest"
	"github.com/giantswarm/micrologger/microloggertest"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/scheme"
	capi "sigs.k8s.io/cluster-api/api/v1alpha4"
	clientfake "sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/yaml"

	"github.com/giantswarm/cluster-apps-operator/service/internal/podcidr"
)

func Test_ClusterValuesGCP(t *testing.T) {
	podCidrConfig := podcidr.Config{InstallationCIDR: "10.0.0.0/16"}
	podCidr, err := podcidr.New(podCidrConfig)
	if err != nil {
		t.Fatal(err)
	}

	gcpCluster := &unstructured.Unstructured{}
	gcpCluster.Object = map[string]interface{}{
		"metadata": map[string]interface{}{
			"name":      "test-cluster",
			"namespace": "default",
		},
		"spec": map[string]interface{}{
			"project": "12345",
		},
	}
	gcpCluster.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "infrastructure.cluster.x-k8s.io",
		Kind:    "GCPCluster",
		Version: "v1beta1",
	})

	cluster := &capi.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-cluster",
			Namespace: "default",
			Labels: map[string]string{
				capi.ClusterLabelName: "test-cluster",
			},
		},
		Spec: capi.ClusterSpec{
			InfrastructureRef: &corev1.ObjectReference{
				Kind:       "GCPCluster",
				Namespace:  "default",
				Name:       "test-cluster",
				APIVersion: "infrastructure.cluster.x-k8s.io/v1beta1",
			},
		},
	}

	var fakeClient *k8sclienttest.Clients
	{
		schemeBuilder := runtime.SchemeBuilder{
			capi.AddToScheme,
		}

		err = schemeBuilder.AddToScheme(scheme.Scheme)
		if err != nil {
			t.Fatal(err)
		}

		fakeClient = k8sclienttest.NewClients(k8sclienttest.ClientsConfig{
			CtrlClient: clientfake.NewClientBuilder().
				WithRuntimeObjects(gcpCluster, cluster).
				Build(),
		})
	}

	config := Config{
		K8sClient:      fakeClient,
		Logger:         microloggertest.New(),
		PodCIDR:        podCidr,
		BaseDomain:     "fadi.gigantic.io",
		ClusterIPRange: "10.0.0.0/16",
		DNSIP:          "192.168.0.10",
		Provider:       "gcp",
		RegistryDomain: "quay.io/giantswarm",
	}
	resource, err := New(config)
	if err != nil {
		t.Fatal(err)
	}

	configmaps, err := resource.GetDesiredState(context.Background(), cluster)
	if err != nil {
		t.Fatal(err)
	}

	for _, configMap := range configmaps {
		if strings.HasSuffix(configMap.Name, "-cluster-values") {
			cmData := &ClusterValuesConfig{}
			err := yaml.Unmarshal([]byte(configMap.Data["values"]), cmData)
			if err != nil {
				t.Fatal(err)
			}
			assertEquals(t, "test-cluster.fadi.gigantic.io", cmData.BaseDomain, "Wrong baseDomain set in cluster-values configmap")
			assertEquals(t, "12345", cmData.GcpProject, "Wrong gcpProject set in cluster-values configmap")
		}
	}
}

func Test_ClusterValuesDNSIP(t *testing.T) {
	podCidrConfig := podcidr.Config{InstallationCIDR: "10.0.0.0/16"}
	podCidr, err := podcidr.New(podCidrConfig)
	if err != nil {
		t.Fatal(err)
	}

	kubeadmControlPlane := &unstructured.Unstructured{}
	kubeadmControlPlane.Object = map[string]interface{}{
		"metadata": map[string]interface{}{
			"name":      "test-cluster",
			"namespace": "default",
		},
		"spec": map[string]interface{}{
			"kubeadmConfigSpec": map[string]interface{}{
				"clusterConfiguration": map[string]interface{}{
					"networking": map[string]interface{}{
						// The coredns service ip must belong to this CIDR
						"serviceSubnet": "172.16.0.0/16",
					},
				},
			},
		},
	}
	kubeadmControlPlane.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "controlplane.cluster.x-k8s.io",
		Kind:    "KubeadmControlPlane",
		Version: "v1beta1",
	})

	cluster := &capi.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-cluster",
			Namespace: "default",
		},
		Spec: capi.ClusterSpec{
			ControlPlaneRef: &corev1.ObjectReference{
				Kind:       "KubeadmControlPlane",
				Namespace:  "default",
				Name:       "test-cluster",
				APIVersion: "controlplane.cluster.x-k8s.io/v1beta1",
			},
		},
	}

	var fakeClient *k8sclienttest.Clients
	{
		schemeBuilder := runtime.SchemeBuilder{
			capi.AddToScheme,
		}

		err = schemeBuilder.AddToScheme(scheme.Scheme)
		if err != nil {
			t.Fatal(err)
		}

		fakeClient = k8sclienttest.NewClients(k8sclienttest.ClientsConfig{
			CtrlClient: clientfake.NewClientBuilder().
				WithRuntimeObjects(kubeadmControlPlane, cluster).
				Build(),
		})
	}

	config := Config{
		K8sClient:      fakeClient,
		Logger:         microloggertest.New(),
		PodCIDR:        podCidr,
		BaseDomain:     "fadi.gigantic.io",
		ClusterIPRange: "10.0.0.0/16",
		DNSIP:          "192.168.0.10",
		Provider:       "gcp",
		RegistryDomain: "quay.io/giantswarm",
	}
	resource, err := New(config)
	if err != nil {
		t.Fatal(err)
	}

	configmaps, err := resource.GetDesiredState(context.Background(), cluster)
	if err != nil {
		t.Fatal(err)
	}

	for _, configMap := range configmaps {
		if strings.HasSuffix(configMap.Name, "-cluster-values") {
			cmData := &ClusterValuesConfig{}
			err := yaml.Unmarshal([]byte(configMap.Data["values"]), cmData)
			if err != nil {
				t.Fatal(err)
			}
			assertEquals(t, "172.16.0.10", cmData.ClusterDNSIP, "Wrong coredns service IP set in cluster-values configmap")
			assertEquals(t, "172.16.0.10", cmData.Cluster.Kubernetes.DNS["IP"], "Wrong coredns service IP set in cluster-values configmap")
		}
	}
}

func Test_ClusterValuesDNSIPWhenServiceCidrIsNotSet(t *testing.T) {
	podCidrConfig := podcidr.Config{InstallationCIDR: "10.0.0.0/16"}
	podCidr, err := podcidr.New(podCidrConfig)
	if err != nil {
		t.Fatal(err)
	}

	kubeadmControlPlane := &unstructured.Unstructured{}
	kubeadmControlPlane.Object = map[string]interface{}{
		"metadata": map[string]interface{}{
			"name":      "test-cluster",
			"namespace": "default",
		},
		"spec": map[string]interface{}{},
	}
	kubeadmControlPlane.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "controlplane.cluster.x-k8s.io",
		Kind:    "KubeadmControlPlane",
		Version: "v1beta1",
	})

	cluster := &capi.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-cluster",
			Namespace: "default",
		},
		Spec: capi.ClusterSpec{
			ControlPlaneRef: &corev1.ObjectReference{
				Kind:       "KubeadmControlPlane",
				Namespace:  "default",
				Name:       "test-cluster",
				APIVersion: "controlplane.cluster.x-k8s.io/v1beta1",
			},
		},
	}

	var fakeClient *k8sclienttest.Clients
	{
		schemeBuilder := runtime.SchemeBuilder{
			capi.AddToScheme,
		}

		err = schemeBuilder.AddToScheme(scheme.Scheme)
		if err != nil {
			t.Fatal(err)
		}

		fakeClient = k8sclienttest.NewClients(k8sclienttest.ClientsConfig{
			CtrlClient: clientfake.NewClientBuilder().
				WithRuntimeObjects(kubeadmControlPlane, cluster).
				Build(),
		})
	}

	config := Config{
		K8sClient:      fakeClient,
		Logger:         microloggertest.New(),
		PodCIDR:        podCidr,
		BaseDomain:     "fadi.gigantic.io",
		ClusterIPRange: "10.96.0.0/12",
		DNSIP:          "10.96.0.10",
		Provider:       "gcp",
		RegistryDomain: "quay.io/giantswarm",
	}
	resource, err := New(config)
	if err != nil {
		t.Fatal(err)
	}

	configmaps, err := resource.GetDesiredState(context.Background(), cluster)
	if err != nil {
		t.Fatal(err)
	}

	for _, configMap := range configmaps {
		if strings.HasSuffix(configMap.Name, "-cluster-values") {
			cmData := &ClusterValuesConfig{}
			err := yaml.Unmarshal([]byte(configMap.Data["values"]), cmData)
			if err != nil {
				t.Fatal(err)
			}
			assertEquals(t, "10.96.0.10", cmData.ClusterDNSIP, "Wrong coredns service IP set in cluster-values configmap")
			assertEquals(t, "10.96.0.10", cmData.Cluster.Kubernetes.DNS["IP"], "Wrong coredns service IP set in cluster-values configmap")
		}
	}
}

func assertEquals(t *testing.T, expected, actual, message string) {
	if expected != actual {
		t.Fatalf("%s, expected %q, actual %q", message, expected, actual)
	}
}