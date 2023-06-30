package clusterconfigmap

import (
	"context"
	"strconv"
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
	capi "sigs.k8s.io/cluster-api/api/v1beta1"
	clientfake "sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/yaml"

	"github.com/giantswarm/cluster-apps-operator/v2/service/internal/podcidr"
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
			ClusterNetwork: &capi.ClusterNetwork{
				ServiceDomain: "cluster.local",
				Services: &capi.NetworkRanges{
					CIDRBlocks: []string{
						"192.168.10.0/24",
						"192.168.20.0/24",
					},
				},
				Pods: &capi.NetworkRanges{
					CIDRBlocks: []string{
						"192.168.10.0/24",
						"192.168.20.0/24",
					},
				},
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

			if !cmData.BootstrapMode.Enabled {
				t.Fatal("bootstrap mode should be enabled")
			}

			if cmData.BootstrapMode.ApiServerPodPort != 6443 {
				t.Fatal("bootstrap mode should use 6443 on GCP")
			}
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
			InfrastructureRef: &corev1.ObjectReference{
				Kind:       "SomeCluster",
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
			InfrastructureRef: &corev1.ObjectReference{
				Kind:       "SomeCluster",
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

func Test_ClusterValuesGCPProjectOnlyAddedOnGCP(t *testing.T) {
	podCidrConfig := podcidr.Config{InstallationCIDR: "10.0.0.0/16"}
	podCidr, err := podcidr.New(podCidrConfig)
	if err != nil {
		t.Fatal(err)
	}

	openstackCluster := &unstructured.Unstructured{}
	openstackCluster.Object = map[string]interface{}{
		"metadata": map[string]interface{}{
			"name":      "test-cluster",
			"namespace": "default",
		},
		"spec": map[string]interface{}{},
	}
	openstackCluster.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "infrastructure.cluster.x-k8s.io",
		Kind:    "OpenstackCluster",
		Version: "v1beta1",
	})

	cluster := &capi.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-cluster",
			Namespace: "default",
		},
		Spec: capi.ClusterSpec{
			InfrastructureRef: &corev1.ObjectReference{
				Kind:       "OpenstackCluster",
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
				WithRuntimeObjects(openstackCluster, cluster).
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
		Provider:       "vsphere",
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
			assertEquals(t, "", cmData.GcpProject, "GCPProject is only set when using gcp")
			assertEquals(t, "10.96.0.10", cmData.ClusterDNSIP, "Wrong coredns service IP set in cluster-values configmap")
			assertEquals(t, "10.96.0.10", cmData.Cluster.Kubernetes.DNS["IP"], "Wrong coredns service IP set in cluster-values configmap")
		}
	}
}

func Test_ClusterValuesCAPZ(t *testing.T) {
	podCidrConfig := podcidr.Config{InstallationCIDR: "10.200.0.0/24"}
	podCidr, err := podcidr.New(podCidrConfig)
	if err != nil {
		t.Fatal(err)
	}

	capzCluster := &unstructured.Unstructured{}
	capzCluster.Object = map[string]interface{}{
		"metadata": map[string]interface{}{
			"name":      "test-cluster",
			"namespace": "default",
		},
		"spec": map[string]interface{}{
			"resourceGroup":  "group1",
			"subscriptionID": "143d9c06-6015-4a4a-a4f9-74a664207db7",
			"networkSpec": map[string]interface{}{
				"apiServerLB": map[string]interface{}{
					"type": "Public",
				},
			},
		},
	}
	capzCluster.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "infrastructure.cluster.x-k8s.io",
		Kind:    "AzureCluster",
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
				Kind:       "AzureCluster",
				Namespace:  "default",
				Name:       "test-cluster",
				APIVersion: "infrastructure.cluster.x-k8s.io/v1beta1",
			},
			ClusterNetwork: &capi.ClusterNetwork{
				ServiceDomain: "cluster.local",
				Services: &capi.NetworkRanges{
					CIDRBlocks: []string{
						"172.31.0.0/16",
					},
				},
				Pods: &capi.NetworkRanges{
					CIDRBlocks: []string{
						"192.168.0.0/16",
					},
				},
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
				WithRuntimeObjects(capzCluster, cluster).
				Build(),
		})
	}

	config := Config{
		K8sClient:      fakeClient,
		Logger:         microloggertest.New(),
		PodCIDR:        podCidr,
		BaseDomain:     "azuretest.gigantic.io",
		ClusterIPRange: "10.200.0.0/24",
		DNSIP:          "172.31.0.10",
		Provider:       "capz",
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
			assertEquals(t, "test-cluster.azuretest.gigantic.io", cmData.BaseDomain, "Wrong baseDomain set in cluster-values configmap")

			if !cmData.BootstrapMode.Enabled {
				t.Fatal("bootstrap mode should be enabled")
			}

			if cmData.BootstrapMode.ApiServerPodPort != 6443 {
				t.Fatal("bootstrap mode should use 6443 on CAPZ")
			}
		}
	}
}

func Test_ClusterValuesPrivateCAPZ(t *testing.T) {
	podCidrConfig := podcidr.Config{InstallationCIDR: "10.200.0.0/24"}
	podCidr, err := podcidr.New(podCidrConfig)
	if err != nil {
		t.Fatal(err)
	}

	capzCluster := &unstructured.Unstructured{}
	capzCluster.Object = map[string]interface{}{
		"metadata": map[string]interface{}{
			"name":      "test-cluster",
			"namespace": "default",
		},
		"spec": map[string]interface{}{
			"resourceGroup":  "group1",
			"subscriptionID": "143d9c06-6015-4a4a-a4f9-74a664207db7",
			"networkSpec": map[string]interface{}{
				"apiServerLB": map[string]interface{}{
					"type": "Internal",
				},
			},
		},
	}
	capzCluster.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "infrastructure.cluster.x-k8s.io",
		Kind:    "AzureCluster",
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
				Kind:       "AzureCluster",
				Namespace:  "default",
				Name:       "test-cluster",
				APIVersion: "infrastructure.cluster.x-k8s.io/v1beta1",
			},
			ClusterNetwork: &capi.ClusterNetwork{
				ServiceDomain: "cluster.local",
				Services: &capi.NetworkRanges{
					CIDRBlocks: []string{
						"172.31.0.0/16",
					},
				},
				Pods: &capi.NetworkRanges{
					CIDRBlocks: []string{
						"192.168.0.0/16",
					},
				},
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
				WithRuntimeObjects(capzCluster, cluster).
				Build(),
		})
	}

	config := Config{
		K8sClient:      fakeClient,
		Logger:         microloggertest.New(),
		PodCIDR:        podCidr,
		BaseDomain:     "azuretest.gigantic.io",
		ClusterIPRange: "10.200.0.0/24",
		DNSIP:          "172.31.0.10",
		Provider:       "capz",
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
			assertEquals(t, "test-cluster.azuretest.gigantic.io", cmData.BaseDomain, "Wrong baseDomain set in cluster-values configmap")
			assertEquals(t, "", *cmData.ExternalDNSIP, "Wrong externalDNSIP set in cluster-values configmap for a private cluster")
			assertEquals(t, "true", strconv.FormatBool(cmData.Cluster.Private), "Wrong cluster.private set in cluster-values configmap for a private cluster")

			if !cmData.BootstrapMode.Enabled {
				t.Fatal("bootstrap mode should be enabled")
			}

			if cmData.BootstrapMode.ApiServerPodPort != 6443 {
				t.Fatal("bootstrap mode should use 6443 on CAPZ")
			}
		}
	}
}

func assertEquals(t *testing.T, expected, actual, message string) {
	if expected != actual {
		t.Fatalf("%s, expected %q, actual %q", message, expected, actual)
	}
}
