package integration

import (
	"context"
	"strconv"
	"strings"
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	capi "sigs.k8s.io/cluster-api/api/core/v1beta1"
	"sigs.k8s.io/yaml"

	configmapresource "github.com/giantswarm/cluster-apps-operator/v3/service/controller/resource/clusterconfigmap"
	"github.com/giantswarm/cluster-apps-operator/v3/test/harness"
)

// TestPrivateCluster_AWSPrivate verifies that an AWSCluster with the
// vpc-mode=private annotation is detected as private through GetDesiredState.
func TestPrivateCluster_AWSPrivate(t *testing.T) {
	env := harness.NewTestEnv(t)
	ctx := context.Background()
	ns := env.CreateNamespace(ctx, "test-pc-awspriv-"+randomSuffix())

	// Create AWSCluster with private annotation.
	obj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "infrastructure.cluster.x-k8s.io/v1beta2",
			"kind":       "AWSCluster",
			"metadata": map[string]interface{}{
				"name":      "test-cluster",
				"namespace": ns.Name,
				"annotations": map[string]interface{}{
					"aws.giantswarm.io/vpc-mode": "private",
				},
			},
			"spec": map[string]interface{}{
				"region": "us-east-1",
			},
		},
	}
	if err := env.CtrlClient().Create(ctx, obj); err != nil {
		t.Fatalf("create AWSCluster: %v", err)
	}

	cluster := env.CreateCluster(ctx, harness.ClusterOpts{
		Name:      "test-cluster",
		Namespace: ns.Name,
		Labels:    map[string]string{capi.ClusterNameLabel: "test-cluster"},
		InfraRef: &corev1.ObjectReference{
			Kind:       "AWSCluster",
			Namespace:  ns.Name,
			Name:       "test-cluster",
			APIVersion: "infrastructure.cluster.x-k8s.io/v1beta2",
		},
		ClusterNetwork: &capi.ClusterNetwork{
			Services: &capi.NetworkRanges{CIDRBlocks: []string{"172.31.0.0/16"}},
			Pods:     &capi.NetworkRanges{CIDRBlocks: []string{"192.168.0.0/16"}},
		},
	})

	opts := harness.DefaultConfigMapResourceOpts()
	resource := env.NewConfigMapResource(opts)

	configmaps, err := resource.GetDesiredState(ctx, cluster)
	if err != nil {
		t.Fatal(err)
	}

	assertPrivateClusterValues(t, configmaps, true)
}

// TestPrivateCluster_AWSPublic verifies that an AWSCluster with the
// vpc-mode=public annotation is detected as non-private.
func TestPrivateCluster_AWSPublic(t *testing.T) {
	env := harness.NewTestEnv(t)
	ctx := context.Background()
	ns := env.CreateNamespace(ctx, "test-pc-awspub-"+randomSuffix())

	obj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "infrastructure.cluster.x-k8s.io/v1beta2",
			"kind":       "AWSCluster",
			"metadata": map[string]interface{}{
				"name":      "test-cluster",
				"namespace": ns.Name,
				"annotations": map[string]interface{}{
					"aws.giantswarm.io/vpc-mode": "public",
				},
			},
			"spec": map[string]interface{}{
				"region": "us-east-1",
			},
		},
	}
	if err := env.CtrlClient().Create(ctx, obj); err != nil {
		t.Fatalf("create AWSCluster: %v", err)
	}

	cluster := env.CreateCluster(ctx, harness.ClusterOpts{
		Name:      "test-cluster",
		Namespace: ns.Name,
		Labels:    map[string]string{capi.ClusterNameLabel: "test-cluster"},
		InfraRef: &corev1.ObjectReference{
			Kind:       "AWSCluster",
			Namespace:  ns.Name,
			Name:       "test-cluster",
			APIVersion: "infrastructure.cluster.x-k8s.io/v1beta2",
		},
		ClusterNetwork: &capi.ClusterNetwork{
			Services: &capi.NetworkRanges{CIDRBlocks: []string{"172.31.0.0/16"}},
			Pods:     &capi.NetworkRanges{CIDRBlocks: []string{"192.168.0.0/16"}},
		},
	})

	opts := harness.DefaultConfigMapResourceOpts()
	resource := env.NewConfigMapResource(opts)

	configmaps, err := resource.GetDesiredState(ctx, cluster)
	if err != nil {
		t.Fatal(err)
	}

	assertPrivateClusterValues(t, configmaps, false)
}

// TestPrivateCluster_AzurePrivate verifies that an AzureCluster with an
// Internal API server LB is detected as private.
func TestPrivateCluster_AzurePrivate(t *testing.T) {
	env := harness.NewTestEnv(t)
	ctx := context.Background()
	ns := env.CreateNamespace(ctx, "test-pc-azpriv-"+randomSuffix())

	env.CreateInfraCluster(ctx, schema.GroupVersionKind{
		Group:   "infrastructure.cluster.x-k8s.io",
		Kind:    "AzureCluster",
		Version: "v1beta1",
	}, ns.Name, "test-cluster", map[string]interface{}{
		"location":       "eastus",
		"subscriptionID": "sub-123",
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
			Services: &capi.NetworkRanges{CIDRBlocks: []string{"172.31.0.0/16"}},
			Pods:     &capi.NetworkRanges{CIDRBlocks: []string{"192.168.0.0/16"}},
		},
	})

	opts := harness.DefaultConfigMapResourceOpts()
	resource := env.NewConfigMapResource(opts)

	configmaps, err := resource.GetDesiredState(ctx, cluster)
	if err != nil {
		t.Fatal(err)
	}

	assertPrivateClusterValues(t, configmaps, true)
}

// TestPrivateCluster_AzurePublic verifies that an AzureCluster with a
// Public API server LB is detected as non-private.
func TestPrivateCluster_AzurePublic(t *testing.T) {
	env := harness.NewTestEnv(t)
	ctx := context.Background()
	ns := env.CreateNamespace(ctx, "test-pc-azpub-"+randomSuffix())

	env.CreateInfraCluster(ctx, schema.GroupVersionKind{
		Group:   "infrastructure.cluster.x-k8s.io",
		Kind:    "AzureCluster",
		Version: "v1beta1",
	}, ns.Name, "test-cluster", map[string]interface{}{
		"location":       "eastus",
		"subscriptionID": "sub-456",
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
			Services: &capi.NetworkRanges{CIDRBlocks: []string{"172.31.0.0/16"}},
			Pods:     &capi.NetworkRanges{CIDRBlocks: []string{"192.168.0.0/16"}},
		},
	})

	opts := harness.DefaultConfigMapResourceOpts()
	resource := env.NewConfigMapResource(opts)

	configmaps, err := resource.GetDesiredState(ctx, cluster)
	if err != nil {
		t.Fatal(err)
	}

	assertPrivateClusterValues(t, configmaps, false)
}

// TestPrivateCluster_GCP verifies that a GCPCluster is always non-private.
func TestPrivateCluster_GCP(t *testing.T) {
	env := harness.NewTestEnv(t)
	ctx := context.Background()
	ns := env.CreateNamespace(ctx, "test-pc-gcp-"+randomSuffix())

	env.CreateInfraCluster(ctx, schema.GroupVersionKind{
		Group:   "infrastructure.cluster.x-k8s.io",
		Kind:    "GCPCluster",
		Version: "v1beta1",
	}, ns.Name, "test-cluster", map[string]interface{}{
		"project": "my-gcp-project",
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
			Services: &capi.NetworkRanges{CIDRBlocks: []string{"172.31.0.0/16"}},
			Pods:     &capi.NetworkRanges{CIDRBlocks: []string{"192.168.0.0/16"}},
		},
	})

	opts := harness.DefaultConfigMapResourceOpts()
	resource := env.NewConfigMapResource(opts)

	configmaps, err := resource.GetDesiredState(ctx, cluster)
	if err != nil {
		t.Fatal(err)
	}

	assertPrivateClusterValues(t, configmaps, false)
}

// TestPrivateCluster_GCPManaged verifies that a GCPManagedCluster is always
// non-private.
func TestPrivateCluster_GCPManaged(t *testing.T) {
	env := harness.NewTestEnv(t)
	ctx := context.Background()
	ns := env.CreateNamespace(ctx, "test-pc-gcpmgd-"+randomSuffix())

	env.CreateInfraCluster(ctx, schema.GroupVersionKind{
		Group:   "infrastructure.cluster.x-k8s.io",
		Kind:    "GCPManagedCluster",
		Version: "v1beta1",
	}, ns.Name, "test-cluster", map[string]interface{}{
		"project": "my-gcp-project",
		"region":  "us-central1",
	})

	cluster := env.CreateCluster(ctx, harness.ClusterOpts{
		Name:      "test-cluster",
		Namespace: ns.Name,
		Labels:    map[string]string{capi.ClusterNameLabel: "test-cluster"},
		InfraRef: &corev1.ObjectReference{
			Kind:       "GCPManagedCluster",
			Namespace:  ns.Name,
			Name:       "test-cluster",
			APIVersion: "infrastructure.cluster.x-k8s.io/v1beta1",
		},
		ClusterNetwork: &capi.ClusterNetwork{
			Services: &capi.NetworkRanges{CIDRBlocks: []string{"172.31.0.0/16"}},
			Pods:     &capi.NetworkRanges{CIDRBlocks: []string{"192.168.0.0/16"}},
		},
	})

	opts := harness.DefaultConfigMapResourceOpts()
	resource := env.NewConfigMapResource(opts)

	configmaps, err := resource.GetDesiredState(ctx, cluster)
	if err != nil {
		t.Fatal(err)
	}

	assertPrivateClusterValues(t, configmaps, false)
}

// assertPrivateClusterValues checks the cluster-values configmap for the
// expected private cluster flag and externalDNSIP value.
func assertPrivateClusterValues(t *testing.T, configmaps []*corev1.ConfigMap, expectPrivate bool) {
	t.Helper()

	for _, cm := range configmaps {
		if !strings.HasSuffix(cm.Name, "-cluster-values") {
			continue
		}

		cmData := &configmapresource.ClusterValuesConfig{}
		if err := yaml.Unmarshal([]byte(cm.Data["values"]), cmData); err != nil {
			t.Fatalf("unmarshal cluster-values: %v", err)
		}

		assertEquals(t,
			strconv.FormatBool(expectPrivate),
			strconv.FormatBool(cmData.Cluster.Private),
			"cluster.private mismatch",
		)

		if expectPrivate {
			if cmData.ExternalDNSIP == nil {
				t.Fatal("externalDNSIP should be set (to empty string) for private cluster")
			}
			assertEquals(t, "", *cmData.ExternalDNSIP, "externalDNSIP should be empty for private cluster")
		} else {
			if cmData.ExternalDNSIP != nil {
				t.Fatalf("externalDNSIP should be nil for non-private cluster, got %q", *cmData.ExternalDNSIP)
			}
		}

		return
	}

	t.Fatal("cluster-values configmap not found in GetDesiredState output")
}
