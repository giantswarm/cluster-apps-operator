package integration

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	capi "sigs.k8s.io/cluster-api/api/core/v1beta1"
	"sigs.k8s.io/yaml"

	configmapresource "github.com/giantswarm/cluster-apps-operator/v3/service/controller/resource/clusterconfigmap"
	"github.com/giantswarm/cluster-apps-operator/v3/test/harness"
)

func randomSuffix() string {
	return fmt.Sprintf("%d", time.Now().UnixNano()%100000)
}

func TestClusterConfigMap_GCP(t *testing.T) {
	env := harness.NewTestEnv(t)
	ctx := context.Background()
	ns := env.CreateNamespace(ctx, "test-cm-gcp-"+randomSuffix())

	env.CreateInfraCluster(ctx, schema.GroupVersionKind{
		Group:   "infrastructure.cluster.x-k8s.io",
		Kind:    "GCPCluster",
		Version: "v1beta1",
	}, ns.Name, "test-cluster", map[string]interface{}{
		"project": "12345",
		"region":  "us-central1",
	})

	cluster := env.CreateCluster(ctx, harness.ClusterOpts{
		Name:      "test-cluster",
		Namespace: ns.Name,
		Labels:    map[string]string{capi.ClusterNameLabel: "test-cluster"},
		InfraRef: &corev1.ObjectReference{
			Kind:       "GCPCluster",
			Namespace:  ns.Name,
			Name:       "test-cluster",
			APIVersion: "infrastructure.cluster.x-k8s.io/v1beta1",
		},
		ClusterNetwork: &capi.ClusterNetwork{
			ServiceDomain: "cluster.local",
			Services: &capi.NetworkRanges{
				CIDRBlocks: []string{"192.168.10.0/24", "192.168.20.0/24"},
			},
			Pods: &capi.NetworkRanges{
				CIDRBlocks: []string{"192.168.10.0/24", "192.168.20.0/24"},
			},
		},
	})

	opts := harness.DefaultConfigMapResourceOpts()
	opts.BaseDomain = "fadi.gigantic.io"
	opts.ClusterIPRange = "10.0.0.0/16"
	opts.DNSIP = "192.168.0.10"
	resource := env.NewConfigMapResource(opts)

	configmaps, err := resource.GetDesiredState(ctx, cluster)
	if err != nil {
		t.Fatal(err)
	}

	for _, configMap := range configmaps {
		if strings.HasSuffix(configMap.Name, "-cluster-values") {
			cmData := &configmapresource.ClusterValuesConfig{}
			if err := yaml.Unmarshal([]byte(configMap.Data["values"]), cmData); err != nil {
				t.Fatal(err)
			}
			assertEquals(t, "test-cluster.fadi.gigantic.io", cmData.BaseDomain, "Wrong baseDomain")
			assertEquals(t, "12345", cmData.GcpProject, "Wrong gcpProject")
			assertEquals(t, "gcp", cmData.Provider, "Wrong provider")
			assertEquals(t, "", cmData.AzureSubscriptionID, "AzureSubscriptionID should be empty for non-CAPZ clusters")
			if !cmData.BootstrapMode.Enabled {
				t.Fatal("bootstrap mode should be enabled")
			}
			if cmData.BootstrapMode.ApiServerPodPort != 6443 {
				t.Fatal("bootstrap mode should use 6443 on GCP")
			}
		} else if strings.HasSuffix(configMap.Name, "-app-operator-values") {
			cmData := &configmapresource.AppOperatorValuesConfig{}
			if err := yaml.Unmarshal([]byte(configMap.Data["values"]), cmData); err != nil {
				t.Fatal(err)
			}
			assertEquals(t, "gcp", cmData.Provider.Kind, "Wrong provider in app-operator-values")
		}
	}
}

func TestClusterConfigMap_DNSIP_FromKCP(t *testing.T) {
	env := harness.NewTestEnv(t)
	ctx := context.Background()
	ns := env.CreateNamespace(ctx, "test-cm-dns-kcp-"+randomSuffix())

	env.CreateInfraCluster(ctx, schema.GroupVersionKind{
		Group:   "infrastructure.cluster.x-k8s.io",
		Kind:    "GCPCluster",
		Version: "v1beta1",
	}, ns.Name, "test-cluster", map[string]interface{}{
		"project": "12345",
		"region":  "us-central1",
	})

	// Create KubeadmControlPlane as unstructured with serviceSubnet
	env.CreateInfraCluster(ctx, schema.GroupVersionKind{
		Group:   "controlplane.cluster.x-k8s.io",
		Kind:    "KubeadmControlPlane",
		Version: "v1beta1",
	}, ns.Name, "test-cluster", map[string]interface{}{
		"version": "v1.28.0",
		"machineTemplate": map[string]interface{}{
			"infrastructureRef": map[string]interface{}{
				"apiVersion": "infrastructure.cluster.x-k8s.io/v1beta1",
				"kind":       "GCPMachineTemplate",
				"name":       "test-machine-template",
			},
		},
		"kubeadmConfigSpec": map[string]interface{}{
			"clusterConfiguration": map[string]interface{}{
				"networking": map[string]interface{}{
					"serviceSubnet": "172.16.0.0/16",
				},
			},
		},
	})

	cluster := env.CreateCluster(ctx, harness.ClusterOpts{
		Name:      "test-cluster",
		Namespace: ns.Name,
		InfraRef: &corev1.ObjectReference{
			Kind:       "GCPCluster",
			Namespace:  ns.Name,
			Name:       "test-cluster",
			APIVersion: "infrastructure.cluster.x-k8s.io/v1beta1",
		},
		ControlPlaneRef: &corev1.ObjectReference{
			Kind:       "KubeadmControlPlane",
			Namespace:  ns.Name,
			Name:       "test-cluster",
			APIVersion: "controlplane.cluster.x-k8s.io/v1beta1",
		},
	})

	opts := harness.DefaultConfigMapResourceOpts()
	opts.BaseDomain = "fadi.gigantic.io"
	opts.ClusterIPRange = "10.0.0.0/16"
	opts.DNSIP = "192.168.0.10"
	resource := env.NewConfigMapResource(opts)

	configmaps, err := resource.GetDesiredState(ctx, cluster)
	if err != nil {
		t.Fatal(err)
	}

	for _, configMap := range configmaps {
		if strings.HasSuffix(configMap.Name, "-cluster-values") {
			cmData := &configmapresource.ClusterValuesConfig{}
			if err := yaml.Unmarshal([]byte(configMap.Data["values"]), cmData); err != nil {
				t.Fatal(err)
			}
			assertEquals(t, "172.16.0.10", cmData.ClusterDNSIP, "Wrong coredns service IP")
			assertEquals(t, "172.16.0.10", cmData.Cluster.Kubernetes.DNS["IP"], "Wrong coredns service IP in cluster config")
			assertEquals(t, "gcp", cmData.Provider, "Wrong provider")
		} else if strings.HasSuffix(configMap.Name, "-app-operator-values") {
			cmData := &configmapresource.AppOperatorValuesConfig{}
			if err := yaml.Unmarshal([]byte(configMap.Data["values"]), cmData); err != nil {
				t.Fatal(err)
			}
			assertEquals(t, "gcp", cmData.Provider.Kind, "Wrong provider in app-operator-values")
		}
	}
}

func TestClusterConfigMap_DNSIP_Fallback(t *testing.T) {
	env := harness.NewTestEnv(t)
	ctx := context.Background()
	ns := env.CreateNamespace(ctx, "test-cm-dns-fb-"+randomSuffix())

	env.CreateInfraCluster(ctx, schema.GroupVersionKind{
		Group:   "infrastructure.cluster.x-k8s.io",
		Kind:    "GCPCluster",
		Version: "v1beta1",
	}, ns.Name, "test-cluster", map[string]interface{}{
		"project": "12345",
		"region":  "us-central1",
	})

	// KubeadmControlPlane without serviceSubnet
	env.CreateInfraCluster(ctx, schema.GroupVersionKind{
		Group:   "controlplane.cluster.x-k8s.io",
		Kind:    "KubeadmControlPlane",
		Version: "v1beta1",
	}, ns.Name, "test-cluster", nil)

	cluster := env.CreateCluster(ctx, harness.ClusterOpts{
		Name:      "test-cluster",
		Namespace: ns.Name,
		InfraRef: &corev1.ObjectReference{
			Kind:       "GCPCluster",
			Namespace:  ns.Name,
			Name:       "test-cluster",
			APIVersion: "infrastructure.cluster.x-k8s.io/v1beta1",
		},
		ControlPlaneRef: &corev1.ObjectReference{
			Kind:       "KubeadmControlPlane",
			Namespace:  ns.Name,
			Name:       "test-cluster",
			APIVersion: "controlplane.cluster.x-k8s.io/v1beta1",
		},
	})

	opts := harness.DefaultConfigMapResourceOpts()
	opts.BaseDomain = "fadi.gigantic.io"
	opts.ClusterIPRange = "10.96.0.0/12"
	opts.DNSIP = "10.96.0.10"
	resource := env.NewConfigMapResource(opts)

	configmaps, err := resource.GetDesiredState(ctx, cluster)
	if err != nil {
		t.Fatal(err)
	}

	for _, configMap := range configmaps {
		if strings.HasSuffix(configMap.Name, "-cluster-values") {
			cmData := &configmapresource.ClusterValuesConfig{}
			if err := yaml.Unmarshal([]byte(configMap.Data["values"]), cmData); err != nil {
				t.Fatal(err)
			}
			assertEquals(t, "10.96.0.10", cmData.ClusterDNSIP, "Wrong fallback coredns service IP")
			assertEquals(t, "10.96.0.10", cmData.Cluster.Kubernetes.DNS["IP"], "Wrong fallback coredns service IP in cluster config")
			assertEquals(t, "gcp", cmData.Provider, "Wrong provider")
		} else if strings.HasSuffix(configMap.Name, "-app-operator-values") {
			cmData := &configmapresource.AppOperatorValuesConfig{}
			if err := yaml.Unmarshal([]byte(configMap.Data["values"]), cmData); err != nil {
				t.Fatal(err)
			}
			assertEquals(t, "gcp", cmData.Provider.Kind, "Wrong provider in app-operator-values")
		}
	}
}

func TestClusterConfigMap_GCPProjectOnlyOnGCP(t *testing.T) {
	env := harness.NewTestEnv(t)
	ctx := context.Background()
	ns := env.CreateNamespace(ctx, "test-cm-nogcp-"+randomSuffix())

	env.CreateInfraCluster(ctx, schema.GroupVersionKind{
		Group:   "infrastructure.cluster.x-k8s.io",
		Kind:    "AzureCluster",
		Version: "v1beta1",
	}, ns.Name, "test-cluster", map[string]interface{}{
		"location":       "eastus",
		"resourceGroup":  "group1",
		"subscriptionID": "143d9c06-6015-4a4a-a4f9-74a664207db7",
		"networkSpec": map[string]interface{}{
			"apiServerLB": map[string]interface{}{
				"type": "Public",
			},
		},
	})

	cluster := env.CreateCluster(ctx, harness.ClusterOpts{
		Name:      "test-cluster",
		Namespace: ns.Name,
		Labels:    map[string]string{capi.ClusterNameLabel: "test-cluster"},
		InfraRef: &corev1.ObjectReference{
			Kind:       "AzureCluster",
			Namespace:  ns.Name,
			Name:       "test-cluster",
			APIVersion: "infrastructure.cluster.x-k8s.io/v1beta1",
		},
		ClusterNetwork: &capi.ClusterNetwork{
			ServiceDomain: "cluster.local",
			Services:      &capi.NetworkRanges{CIDRBlocks: []string{"172.31.0.0/16"}},
			Pods:          &capi.NetworkRanges{CIDRBlocks: []string{"192.168.0.0/16"}},
		},
	})

	opts := harness.DefaultConfigMapResourceOpts()
	opts.BaseDomain = "fadi.gigantic.io"
	opts.ClusterIPRange = "10.96.0.0/12"
	opts.DNSIP = "10.96.0.10"
	resource := env.NewConfigMapResource(opts)

	configmaps, err := resource.GetDesiredState(ctx, cluster)
	if err != nil {
		t.Fatal(err)
	}

	for _, configMap := range configmaps {
		if strings.HasSuffix(configMap.Name, "-cluster-values") {
			cmData := &configmapresource.ClusterValuesConfig{}
			if err := yaml.Unmarshal([]byte(configMap.Data["values"]), cmData); err != nil {
				t.Fatal(err)
			}
			assertEquals(t, "", cmData.GcpProject, "GCPProject should only be set on GCP")
			assertEquals(t, "10.96.0.10", cmData.ClusterDNSIP, "Wrong coredns service IP")
			assertEquals(t, "10.96.0.10", cmData.Cluster.Kubernetes.DNS["IP"], "Wrong coredns service IP in cluster config")
			assertEquals(t, "capz", cmData.Provider, "Wrong provider")
		} else if strings.HasSuffix(configMap.Name, "-app-operator-values") {
			cmData := &configmapresource.AppOperatorValuesConfig{}
			if err := yaml.Unmarshal([]byte(configMap.Data["values"]), cmData); err != nil {
				t.Fatal(err)
			}
			assertEquals(t, "capz", cmData.Provider.Kind, "Wrong provider in app-operator-values")
		}
	}
}

func TestClusterConfigMap_CAPZ(t *testing.T) {
	env := harness.NewTestEnv(t)
	ctx := context.Background()
	ns := env.CreateNamespace(ctx, "test-cm-capz-"+randomSuffix())

	env.CreateInfraCluster(ctx, schema.GroupVersionKind{
		Group:   "infrastructure.cluster.x-k8s.io",
		Kind:    "AzureCluster",
		Version: "v1beta1",
	}, ns.Name, "test-cluster", map[string]interface{}{
		"location":       "eastus",
		"resourceGroup":  "group1",
		"subscriptionID": "143d9c06-6015-4a4a-a4f9-74a664207db7",
		"networkSpec": map[string]interface{}{
			"apiServerLB": map[string]interface{}{
				"type": "Public",
			},
		},
	})

	cluster := env.CreateCluster(ctx, harness.ClusterOpts{
		Name:      "test-cluster",
		Namespace: ns.Name,
		Labels:    map[string]string{capi.ClusterNameLabel: "test-cluster"},
		InfraRef: &corev1.ObjectReference{
			Kind:       "AzureCluster",
			Namespace:  ns.Name,
			Name:       "test-cluster",
			APIVersion: "infrastructure.cluster.x-k8s.io/v1beta1",
		},
		ClusterNetwork: &capi.ClusterNetwork{
			ServiceDomain: "cluster.local",
			Services:      &capi.NetworkRanges{CIDRBlocks: []string{"172.31.0.0/16"}},
			Pods:          &capi.NetworkRanges{CIDRBlocks: []string{"192.168.0.0/16"}},
		},
	})

	opts := harness.DefaultConfigMapResourceOpts()
	opts.BaseDomain = "azuretest.gigantic.io"
	opts.ClusterIPRange = "10.200.0.0/24"
	opts.DNSIP = "172.31.0.10"
	resource := env.NewConfigMapResource(opts)

	configmaps, err := resource.GetDesiredState(ctx, cluster)
	if err != nil {
		t.Fatal(err)
	}

	for _, configMap := range configmaps {
		if strings.HasSuffix(configMap.Name, "-cluster-values") {
			cmData := &configmapresource.ClusterValuesConfig{}
			if err := yaml.Unmarshal([]byte(configMap.Data["values"]), cmData); err != nil {
				t.Fatal(err)
			}
			assertEquals(t, "test-cluster.azuretest.gigantic.io", cmData.BaseDomain, "Wrong baseDomain")
			assertEquals(t, "capz", cmData.Provider, "Wrong provider")
			assertEquals(t, "143d9c06-6015-4a4a-a4f9-74a664207db7", cmData.AzureSubscriptionID, "Wrong AzureSubscriptionID")
			if !cmData.BootstrapMode.Enabled {
				t.Fatal("bootstrap mode should be enabled")
			}
			if cmData.BootstrapMode.ApiServerPodPort != 6443 {
				t.Fatal("bootstrap mode should use 6443 on CAPZ")
			}
		} else if strings.HasSuffix(configMap.Name, "-app-operator-values") {
			cmData := &configmapresource.AppOperatorValuesConfig{}
			if err := yaml.Unmarshal([]byte(configMap.Data["values"]), cmData); err != nil {
				t.Fatal(err)
			}
			assertEquals(t, "capz", cmData.Provider.Kind, "Wrong provider in app-operator-values")
		}
	}
}

func TestClusterConfigMap_PrivateCAPZ(t *testing.T) {
	env := harness.NewTestEnv(t)
	ctx := context.Background()
	ns := env.CreateNamespace(ctx, "test-cm-privcapz-"+randomSuffix())

	env.CreateInfraCluster(ctx, schema.GroupVersionKind{
		Group:   "infrastructure.cluster.x-k8s.io",
		Kind:    "AzureCluster",
		Version: "v1beta1",
	}, ns.Name, "test-cluster", map[string]interface{}{
		"location":       "eastus",
		"resourceGroup":  "group1",
		"subscriptionID": "143d9c06-6015-4a4a-a4f9-74a664207db7",
		"networkSpec": map[string]interface{}{
			"apiServerLB": map[string]interface{}{
				"type": "Internal",
			},
		},
	})

	cluster := env.CreateCluster(ctx, harness.ClusterOpts{
		Name:      "test-cluster",
		Namespace: ns.Name,
		Labels:    map[string]string{capi.ClusterNameLabel: "test-cluster"},
		InfraRef: &corev1.ObjectReference{
			Kind:       "AzureCluster",
			Namespace:  ns.Name,
			Name:       "test-cluster",
			APIVersion: "infrastructure.cluster.x-k8s.io/v1beta1",
		},
		ClusterNetwork: &capi.ClusterNetwork{
			ServiceDomain: "cluster.local",
			Services:      &capi.NetworkRanges{CIDRBlocks: []string{"172.31.0.0/16"}},
			Pods:          &capi.NetworkRanges{CIDRBlocks: []string{"192.168.0.0/16"}},
		},
	})

	opts := harness.DefaultConfigMapResourceOpts()
	opts.BaseDomain = "azuretest.gigantic.io"
	opts.ClusterIPRange = "10.200.0.0/24"
	opts.DNSIP = "172.31.0.10"
	resource := env.NewConfigMapResource(opts)

	configmaps, err := resource.GetDesiredState(ctx, cluster)
	if err != nil {
		t.Fatal(err)
	}

	for _, configMap := range configmaps {
		if strings.HasSuffix(configMap.Name, "-cluster-values") {
			cmData := &configmapresource.ClusterValuesConfig{}
			if err := yaml.Unmarshal([]byte(configMap.Data["values"]), cmData); err != nil {
				t.Fatal(err)
			}
			assertEquals(t, "test-cluster.azuretest.gigantic.io", cmData.BaseDomain, "Wrong baseDomain")
			assertEquals(t, "", *cmData.ExternalDNSIP, "externalDNSIP should be empty for private cluster")
			assertEquals(t, "true", strconv.FormatBool(cmData.Cluster.Private), "cluster.private should be true for private cluster")
			assertEquals(t, "capz", cmData.Provider, "Wrong provider")
			if !cmData.BootstrapMode.Enabled {
				t.Fatal("bootstrap mode should be enabled")
			}
			if cmData.BootstrapMode.ApiServerPodPort != 6443 {
				t.Fatal("bootstrap mode should use 6443 on CAPZ")
			}
		} else if strings.HasSuffix(configMap.Name, "-app-operator-values") {
			cmData := &configmapresource.AppOperatorValuesConfig{}
			if err := yaml.Unmarshal([]byte(configMap.Data["values"]), cmData); err != nil {
				t.Fatal(err)
			}
			assertEquals(t, "capz", cmData.Provider.Kind, "Wrong provider in app-operator-values")
		}
	}
}

func assertEquals(t *testing.T, expected, actual, message string) {
	t.Helper()
	if expected != actual {
		t.Fatalf("%s, expected %q, actual %q", message, expected, actual)
	}
}
