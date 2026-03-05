package integration

import (
	"context"
	"strings"
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	capi "sigs.k8s.io/cluster-api/api/core/v1beta1"
	"sigs.k8s.io/yaml"

	configmapresource "github.com/giantswarm/cluster-apps-operator/v3/service/controller/resource/clusterconfigmap"
	"github.com/giantswarm/cluster-apps-operator/v3/test/harness"
)

// TestFullReconcile_GCP exercises the full configmap + secret + app resource
// pipeline for a GCP cluster and verifies all expected objects are created.
func TestFullReconcile_GCP(t *testing.T) {
	env := harness.NewTestEnv(t)
	ctx := context.Background()
	ns := env.CreateNamespace(ctx, "test-rec-gcp-"+randomSuffix())

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
			Services: &capi.NetworkRanges{CIDRBlocks: []string{"10.96.0.0/12"}},
			Pods:     &capi.NetworkRanges{CIDRBlocks: []string{"10.244.0.0/16"}},
		},
	})

	// Step 1: ConfigMap resource — compute desired state and persist.
	cmOpts := harness.DefaultConfigMapResourceOpts()
	cmResource := env.NewConfigMapResource(cmOpts)

	configmaps, err := cmResource.GetDesiredState(ctx, cluster)
	if err != nil {
		t.Fatalf("configmap GetDesiredState: %v", err)
	}
	if len(configmaps) != 2 {
		t.Fatalf("expected 2 configmaps, got %d", len(configmaps))
	}

	for _, cm := range configmaps {
		if err := env.CtrlClient().Create(ctx, cm); err != nil {
			t.Fatalf("create configmap %s: %v", cm.Name, err)
		}
	}

	// Step 2: Secret resource — compute desired state and persist.
	secretResource := env.NewSecretResource(harness.SecretResourceOpts{})

	secrets, err := secretResource.GetDesiredState(ctx, cluster)
	if err != nil {
		t.Fatalf("secret GetDesiredState: %v", err)
	}
	if len(secrets) != 1 {
		t.Fatalf("expected 1 secret, got %d", len(secrets))
	}

	for _, s := range secrets {
		if err := env.CtrlClient().Create(ctx, s); err != nil {
			t.Fatalf("create secret %s: %v", s.Name, err)
		}
	}

	// Step 3: App resource — EnsureCreated creates App CRs directly.
	appResource := env.NewAppResource(harness.DefaultAppResourceOpts())
	if err := appResource.EnsureCreated(ctx, cluster); err != nil {
		t.Fatalf("app EnsureCreated: %v", err)
	}

	// Step 4: Verify all objects exist with correct content.
	clusterValuesCM := env.GetConfigMap(ctx, ns.Name, "test-cluster-cluster-values")
	cmData := &configmapresource.ClusterValuesConfig{}
	if err := yaml.Unmarshal([]byte(clusterValuesCM.Data["values"]), cmData); err != nil {
		t.Fatalf("unmarshal cluster-values: %v", err)
	}
	assertEquals(t, "gcp", cmData.Provider, "Wrong provider")
	assertEquals(t, "my-gcp-project", cmData.GcpProject, "Wrong gcpProject")
	assertEquals(t, "test-cluster", cmData.ClusterID, "Wrong clusterID")

	appOpCM := env.GetConfigMap(ctx, ns.Name, "test-cluster-app-operator-values")
	if appOpCM.Data["values"] == "" {
		t.Fatal("app-operator-values configmap should have values")
	}

	secret := env.GetSecret(ctx, ns.Name, "test-cluster-cluster-values")
	if secret.Data["values"] == nil {
		t.Fatal("cluster-values secret should have values key")
	}

	appOp := env.GetApp(ctx, ns.Name, "test-cluster-app-operator")
	assertEquals(t, "app-operator", appOp.Spec.Name, "Wrong app-operator spec name")
	assertEquals(t, "control-plane-catalog", appOp.Spec.Catalog, "Wrong app-operator catalog")
	assertEquals(t, "1.0.0", appOp.Spec.Version, "Wrong app-operator version")
	if !appOp.Spec.KubeConfig.InCluster {
		t.Fatal("app-operator should use in-cluster kubeconfig")
	}

	chartOp := env.GetApp(ctx, ns.Name, "test-cluster-chart-operator")
	assertEquals(t, "chart-operator", chartOp.Spec.Name, "Wrong chart-operator spec name")
	assertEquals(t, "default", chartOp.Spec.Catalog, "Wrong chart-operator catalog")
	assertEquals(t, "1.0.0", chartOp.Spec.Version, "Wrong chart-operator version")
	assertEquals(t, "giantswarm", chartOp.Spec.Namespace, "Wrong chart-operator target namespace")
	if chartOp.Spec.KubeConfig.InCluster {
		t.Fatal("chart-operator should NOT use in-cluster kubeconfig")
	}
	assertEquals(t, "test-cluster-kubeconfig", chartOp.Spec.KubeConfig.Secret.Name, "Wrong chart-operator kubeconfig secret")
}

// TestFullReconcile_Azure exercises the full pipeline for an Azure cluster
// and verifies CAPZ-specific values like subscriptionID and provider=capz.
func TestFullReconcile_Azure(t *testing.T) {
	env := harness.NewTestEnv(t)
	ctx := context.Background()
	ns := env.CreateNamespace(ctx, "test-rec-az-"+randomSuffix())

	env.CreateInfraCluster(ctx, schema.GroupVersionKind{
		Group:   "infrastructure.cluster.x-k8s.io",
		Kind:    "AzureCluster",
		Version: "v1beta1",
	}, ns.Name, "test-cluster", map[string]interface{}{
		"location":       "eastus",
		"subscriptionID": "sub-azure-123",
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

	// Step 1: ConfigMaps.
	cmOpts := harness.DefaultConfigMapResourceOpts()
	cmResource := env.NewConfigMapResource(cmOpts)

	configmaps, err := cmResource.GetDesiredState(ctx, cluster)
	if err != nil {
		t.Fatalf("configmap GetDesiredState: %v", err)
	}

	for _, cm := range configmaps {
		if err := env.CtrlClient().Create(ctx, cm); err != nil {
			t.Fatalf("create configmap %s: %v", cm.Name, err)
		}
	}

	// Step 2: Secrets.
	secretResource := env.NewSecretResource(harness.SecretResourceOpts{})

	secrets, err := secretResource.GetDesiredState(ctx, cluster)
	if err != nil {
		t.Fatalf("secret GetDesiredState: %v", err)
	}

	for _, s := range secrets {
		if err := env.CtrlClient().Create(ctx, s); err != nil {
			t.Fatalf("create secret %s: %v", s.Name, err)
		}
	}

	// Step 3: Apps.
	appResource := env.NewAppResource(harness.DefaultAppResourceOpts())
	if err := appResource.EnsureCreated(ctx, cluster); err != nil {
		t.Fatalf("app EnsureCreated: %v", err)
	}

	// Step 4: Verify Azure-specific values.
	clusterValuesCM := env.GetConfigMap(ctx, ns.Name, "test-cluster-cluster-values")
	cmData := &configmapresource.ClusterValuesConfig{}
	if err := yaml.Unmarshal([]byte(clusterValuesCM.Data["values"]), cmData); err != nil {
		t.Fatalf("unmarshal cluster-values: %v", err)
	}
	assertEquals(t, "capz", cmData.Provider, "Wrong provider for Azure")
	assertEquals(t, "sub-azure-123", cmData.AzureSubscriptionID, "Wrong subscriptionID")
	assertEquals(t, "", cmData.GcpProject, "gcpProject should be empty for Azure")

	// Verify app-operator-values has CAPZ provider.
	appOpCM := env.GetConfigMap(ctx, ns.Name, "test-cluster-app-operator-values")
	appOpData := &configmapresource.AppOperatorValuesConfig{}
	if err := yaml.Unmarshal([]byte(appOpCM.Data["values"]), appOpData); err != nil {
		t.Fatalf("unmarshal app-operator-values: %v", err)
	}
	assertEquals(t, "capz", appOpData.Provider.Kind, "Wrong provider in app-operator-values")

	// Verify cluster-values secret exists.
	secret := env.GetSecret(ctx, ns.Name, "test-cluster-cluster-values")
	if secret.Data["values"] == nil {
		t.Fatal("cluster-values secret should have values key")
	}

	// Verify both App CRs exist.
	appOp := env.GetApp(ctx, ns.Name, "test-cluster-app-operator")
	assertEquals(t, "app-operator", appOp.Spec.Name, "Wrong app-operator spec name")

	chartOp := env.GetApp(ctx, ns.Name, "test-cluster-chart-operator")
	assertEquals(t, "chart-operator", chartOp.Spec.Name, "Wrong chart-operator spec name")

	// Verify configmap references in chart-operator App.
	assertEquals(t, "test-cluster-cluster-values", chartOp.Spec.Config.ConfigMap.Name, "Wrong chart-operator config configmap")
	assertEquals(t, "test-cluster-cluster-values", chartOp.Spec.Config.Secret.Name, "Wrong chart-operator config secret")

	// Cluster should NOT be private (Public LB).
	if cmData.Cluster.Private {
		t.Fatal("cluster should not be private with Public LB")
	}
	if cmData.ExternalDNSIP != nil {
		t.Fatalf("externalDNSIP should be nil for non-private cluster, got %q", *cmData.ExternalDNSIP)
	}
}

// appOperatorValues is a minimal struct for unmarshaling the app-operator
// values configmap (avoids importing the full type which uses map[string]interface{}).
type appOperatorValues struct {
	App struct {
		WatchNamespace    string `json:"watchNamespace"`
		WorkloadClusterID string `json:"workloadClusterID"`
	} `json:"app"`
	Provider struct {
		Kind string `json:"kind"`
	} `json:"provider"`
}

// TestFullReconcile_VerifyAppOperatorValues ensures the app-operator-values
// configmap contains the correct watch namespace and workload cluster ID.
func TestFullReconcile_VerifyAppOperatorValues(t *testing.T) {
	env := harness.NewTestEnv(t)
	ctx := context.Background()
	ns := env.CreateNamespace(ctx, "test-rec-aov-"+randomSuffix())

	env.CreateInfraCluster(ctx, schema.GroupVersionKind{
		Group:   "infrastructure.cluster.x-k8s.io",
		Kind:    "GCPCluster",
		Version: "v1beta1",
	}, ns.Name, "test-cluster", map[string]interface{}{
		"project": "proj-1",
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
			Services: &capi.NetworkRanges{CIDRBlocks: []string{"10.96.0.0/12"}},
			Pods:     &capi.NetworkRanges{CIDRBlocks: []string{"10.244.0.0/16"}},
		},
	})

	cmOpts := harness.DefaultConfigMapResourceOpts()
	cmResource := env.NewConfigMapResource(cmOpts)

	configmaps, err := cmResource.GetDesiredState(ctx, cluster)
	if err != nil {
		t.Fatalf("configmap GetDesiredState: %v", err)
	}

	for _, cm := range configmaps {
		if strings.HasSuffix(cm.Name, "-app-operator-values") {
			var vals appOperatorValues
			if err := yaml.Unmarshal([]byte(cm.Data["values"]), &vals); err != nil {
				t.Fatalf("unmarshal app-operator-values: %v", err)
			}
			assertEquals(t, ns.Name, vals.App.WatchNamespace, "Wrong watchNamespace")
			assertEquals(t, "test-cluster", vals.App.WorkloadClusterID, "Wrong workloadClusterID")
			assertEquals(t, "gcp", vals.Provider.Kind, "Wrong provider kind")
			return
		}
	}

	t.Fatal("app-operator-values configmap not found")
}
