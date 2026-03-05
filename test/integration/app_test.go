package integration

import (
	"context"
	"fmt"
	"testing"

	appv1alpha1 "github.com/giantswarm/apiextensions-application/api/v1alpha1"
	"github.com/giantswarm/k8smetadata/pkg/annotation"
	"github.com/giantswarm/k8smetadata/pkg/label"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	capi "sigs.k8s.io/cluster-api/api/core/v1beta1"

	"github.com/giantswarm/cluster-apps-operator/v3/pkg/project"
	"github.com/giantswarm/cluster-apps-operator/v3/test/harness"
)

func newTestApp(name, namespace, cluster, managedBy string, inCluster bool) *appv1alpha1.App {
	app := &appv1alpha1.App{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    map[string]string{},
		},
		Spec: appv1alpha1.AppSpec{
			Name:      name,
			Namespace: "default",
			Version:   "1.0.0",
			Catalog:   "test",
			KubeConfig: appv1alpha1.AppSpecKubeConfig{
				InCluster: inCluster,
			},
		},
	}

	if cluster != "" {
		app.Labels[label.Cluster] = cluster
	}

	if managedBy != "" {
		app.Labels[label.ManagedBy] = managedBy

		if managedBy == "flux" {
			app.Labels["kustomize.toolkit.fluxcd.io/name"] = cluster
			app.Labels["kustomize.toolkit.fluxcd.io/namespace"] = namespace
		}
	}

	return app
}

// TestApp_EnsureCreated verifies that EnsureCreated creates the expected
// app-operator and chart-operator App CRs for a cluster.
func TestApp_EnsureCreated(t *testing.T) {
	env := harness.NewTestEnv(t)
	ctx := context.Background()

	clusterID := "ec1-" + randomSuffix()
	ns := env.CreateNamespace(ctx, "test-app-cr-"+randomSuffix())

	env.CreateCluster(ctx, harness.ClusterOpts{
		Name:      clusterID,
		Namespace: ns.Name,
		Labels: map[string]string{
			label.Cluster: clusterID,
		},
	})

	resource := env.NewAppResource(harness.DefaultAppResourceOpts())

	cluster := &capi.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      clusterID,
			Namespace: ns.Name,
			Labels: map[string]string{
				label.Cluster: clusterID,
			},
		},
	}

	err := resource.EnsureCreated(ctx, cluster)
	if err != nil {
		t.Fatalf("EnsureCreated returned error: %v", err)
	}

	// Verify app-operator App CR was created.
	appOpName := fmt.Sprintf("%s-app-operator", clusterID)
	if !env.AppExists(ctx, ns.Name, appOpName) {
		t.Fatalf("expected app %s to exist", appOpName)
	}

	appOp := env.GetApp(ctx, ns.Name, appOpName)
	if appOp.Labels[label.Cluster] != clusterID {
		t.Fatalf("expected cluster label %q, got %q", clusterID, appOp.Labels[label.Cluster])
	}
	if appOp.Labels[label.ManagedBy] != project.Name() {
		t.Fatalf("expected managed-by label %q, got %q", project.Name(), appOp.Labels[label.ManagedBy])
	}
	if appOp.Spec.Name != "app-operator" {
		t.Fatalf("expected spec.name %q, got %q", "app-operator", appOp.Spec.Name)
	}
	if appOp.Spec.Catalog != "control-plane-catalog" {
		t.Fatalf("expected spec.catalog %q, got %q", "control-plane-catalog", appOp.Spec.Catalog)
	}
	if appOp.Spec.Version != "1.0.0" {
		t.Fatalf("expected spec.version %q, got %q", "1.0.0", appOp.Spec.Version)
	}
	if !appOp.Spec.KubeConfig.InCluster {
		t.Fatal("expected app-operator to be in-cluster")
	}
	if appOp.Annotations[annotation.ChartOperatorForceHelmUpgrade] != "false" {
		t.Fatalf("expected force-helm-upgrade annotation to be false, got %q", appOp.Annotations[annotation.ChartOperatorForceHelmUpgrade])
	}

	// Verify chart-operator App CR was created.
	chartOpName := fmt.Sprintf("%s-chart-operator", clusterID)
	if !env.AppExists(ctx, ns.Name, chartOpName) {
		t.Fatalf("expected app %s to exist", chartOpName)
	}

	chartOp := env.GetApp(ctx, ns.Name, chartOpName)
	if chartOp.Labels[label.Cluster] != clusterID {
		t.Fatalf("expected cluster label %q, got %q", clusterID, chartOp.Labels[label.Cluster])
	}
	if chartOp.Labels[label.ManagedBy] != project.Name() {
		t.Fatalf("expected managed-by label %q, got %q", project.Name(), chartOp.Labels[label.ManagedBy])
	}
	if chartOp.Spec.Name != "chart-operator" {
		t.Fatalf("expected spec.name %q, got %q", "chart-operator", chartOp.Spec.Name)
	}
	if chartOp.Spec.Catalog != "default" {
		t.Fatalf("expected spec.catalog %q, got %q", "default", chartOp.Spec.Catalog)
	}
	if chartOp.Spec.KubeConfig.InCluster {
		t.Fatal("expected chart-operator to NOT be in-cluster")
	}
}

// TestApp_EnsureCreated_Idempotent verifies that calling EnsureCreated
// twice produces no error and the Apps remain unchanged.
func TestApp_EnsureCreated_Idempotent(t *testing.T) {
	env := harness.NewTestEnv(t)
	ctx := context.Background()

	clusterID := "ec2-" + randomSuffix()
	ns := env.CreateNamespace(ctx, "test-app-idem-"+randomSuffix())

	env.CreateCluster(ctx, harness.ClusterOpts{
		Name:      clusterID,
		Namespace: ns.Name,
		Labels: map[string]string{
			label.Cluster: clusterID,
		},
	})

	resource := env.NewAppResource(harness.DefaultAppResourceOpts())

	cluster := &capi.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      clusterID,
			Namespace: ns.Name,
			Labels: map[string]string{
				label.Cluster: clusterID,
			},
		},
	}

	// First call.
	err := resource.EnsureCreated(ctx, cluster)
	if err != nil {
		t.Fatalf("first EnsureCreated returned error: %v", err)
	}

	appOpName := fmt.Sprintf("%s-app-operator", clusterID)
	first := env.GetApp(ctx, ns.Name, appOpName)

	// Second call (idempotent).
	err = resource.EnsureCreated(ctx, cluster)
	if err != nil {
		t.Fatalf("second EnsureCreated returned error: %v", err)
	}

	second := env.GetApp(ctx, ns.Name, appOpName)

	if first.ResourceVersion != second.ResourceVersion {
		t.Fatalf("expected resource version unchanged, got %q then %q", first.ResourceVersion, second.ResourceVersion)
	}
}

// TestApp_EnsureDeleted_Basic migrates the "flawless" test case: managed apps
// for the target cluster are deleted, while apps for other clusters remain.
func TestApp_EnsureDeleted_Basic(t *testing.T) {
	env := harness.NewTestEnv(t)
	ctx := context.Background()

	clusterID := "demo0"
	ns := env.CreateNamespace(ctx, "test-app-del-"+randomSuffix())

	// Create apps belonging to the target cluster, managed by cluster-apps-operator.
	env.CreateApp(ctx, newTestApp(clusterID+"-app-operator", ns.Name, clusterID, project.Name(), true))
	env.CreateApp(ctx, newTestApp(clusterID+"-chart-operator", ns.Name, clusterID, project.Name(), false))
	env.CreateApp(ctx, newTestApp("other0-app-operator", ns.Name, clusterID, project.Name(), true))
	env.CreateApp(ctx, newTestApp("other0-chart-operator", ns.Name, clusterID, project.Name(), false))
	// An app for the target cluster NOT managed by cluster-apps-operator.
	env.CreateApp(ctx, newTestApp(clusterID+"-hello-world", ns.Name, clusterID, "", false))
	// Apps for a different cluster.
	env.CreateApp(ctx, newTestApp("other0-hello-world", ns.Name, "other0", "", false))
	env.CreateApp(ctx, newTestApp("other0-kyverno-policies", ns.Name, "other0", "", false))

	cluster := &capi.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      clusterID,
			Namespace: ns.Name,
			Labels: map[string]string{
				label.Cluster: clusterID,
			},
		},
	}

	resource := env.NewAppResource(harness.DefaultAppResourceOpts())

	err := resource.EnsureDeleted(ctx, cluster)
	if err != nil {
		t.Fatalf("EnsureDeleted returned error: %v", err)
	}

	// The unmanaged app for demo0 should be deleted.
	if env.AppExists(ctx, ns.Name, clusterID+"-hello-world") {
		t.Fatal("expected demo0-hello-world to be removed")
	}

	// Apps for other clusters must remain.
	for _, name := range []string{"other0-app-operator", "other0-chart-operator", "other0-hello-world", "other0-kyverno-policies"} {
		if !env.AppExists(ctx, ns.Name, name) {
			t.Fatalf("expected %s to still exist", name)
		}
	}
}

// TestApp_EnsureDeleted_InCluster migrates the "flawless with in-cluster" test
// case: in-cluster apps for the target cluster are also deleted.
func TestApp_EnsureDeleted_InCluster(t *testing.T) {
	env := harness.NewTestEnv(t)
	ctx := context.Background()

	clusterID := "demo0"
	ns := env.CreateNamespace(ctx, "test-app-ic-"+randomSuffix())

	env.CreateApp(ctx, newTestApp(clusterID+"-app-operator", ns.Name, clusterID, project.Name(), true))
	env.CreateApp(ctx, newTestApp(clusterID+"-chart-operator", ns.Name, clusterID, project.Name(), false))
	env.CreateApp(ctx, newTestApp("other0-app-operator", ns.Name, clusterID, project.Name(), true))
	env.CreateApp(ctx, newTestApp("other0-chart-operator", ns.Name, clusterID, project.Name(), false))
	// In-cluster and remote apps for the target cluster, not managed by us.
	env.CreateApp(ctx, newTestApp(clusterID+"-security-pack", ns.Name, clusterID, "", true))
	env.CreateApp(ctx, newTestApp(clusterID+"-trivy", ns.Name, clusterID, "", false))
	env.CreateApp(ctx, newTestApp(clusterID+"-falco", ns.Name, clusterID, "", false))
	// Different cluster.
	env.CreateApp(ctx, newTestApp("other0-hello-world", ns.Name, "other0", "", false))

	cluster := &capi.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      clusterID,
			Namespace: ns.Name,
			Labels: map[string]string{
				label.Cluster: clusterID,
			},
		},
	}

	resource := env.NewAppResource(harness.DefaultAppResourceOpts())

	err := resource.EnsureDeleted(ctx, cluster)
	if err != nil {
		t.Fatalf("EnsureDeleted returned error: %v", err)
	}

	// Unmanaged cluster apps should be removed.
	for _, name := range []string{clusterID + "-security-pack", clusterID + "-trivy", clusterID + "-falco"} {
		if env.AppExists(ctx, ns.Name, name) {
			t.Fatalf("expected %s to be removed", name)
		}
	}

	// Other cluster apps must remain.
	for _, name := range []string{"other0-app-operator", "other0-chart-operator", "other0-hello-world"} {
		if !env.AppExists(ctx, ns.Name, name) {
			t.Fatalf("expected %s to still exist", name)
		}
	}
}

// TestApp_EnsureDeleted_FluxManaged migrates the "flawless with Flux managed
// apps" test case: Flux-managed apps are preserved during deletion.
func TestApp_EnsureDeleted_FluxManaged(t *testing.T) {
	env := harness.NewTestEnv(t)
	ctx := context.Background()

	clusterID := "demo0"
	ns := env.CreateNamespace(ctx, "test-app-fx-"+randomSuffix())

	env.CreateApp(ctx, newTestApp(clusterID+"-app-operator", ns.Name, clusterID, project.Name(), true))
	env.CreateApp(ctx, newTestApp(clusterID+"-chart-operator", ns.Name, clusterID, project.Name(), false))
	env.CreateApp(ctx, newTestApp("other0-app-operator", ns.Name, clusterID, project.Name(), true))
	env.CreateApp(ctx, newTestApp("other0-chart-operator", ns.Name, clusterID, project.Name(), false))
	// Flux-managed apps.
	env.CreateApp(ctx, newTestApp(clusterID+"-security-pack", ns.Name, clusterID, "flux", true))
	env.CreateApp(ctx, newTestApp(clusterID+"-hello-world", ns.Name, clusterID, "flux", false))
	// Non-flux, non-managed apps for target cluster.
	env.CreateApp(ctx, newTestApp(clusterID+"-trivy", ns.Name, clusterID, "", false))
	env.CreateApp(ctx, newTestApp(clusterID+"-falco", ns.Name, clusterID, "", false))
	// Different cluster.
	env.CreateApp(ctx, newTestApp("other0-hello-world", ns.Name, "other0", "", false))

	cluster := &capi.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      clusterID,
			Namespace: ns.Name,
			Labels: map[string]string{
				label.Cluster: clusterID,
			},
		},
	}

	resource := env.NewAppResource(harness.DefaultAppResourceOpts())

	err := resource.EnsureDeleted(ctx, cluster)
	if err != nil {
		t.Fatalf("EnsureDeleted returned error: %v", err)
	}

	// Non-flux unmanaged apps should be removed.
	for _, name := range []string{clusterID + "-trivy", clusterID + "-falco"} {
		if env.AppExists(ctx, ns.Name, name) {
			t.Fatalf("expected %s to be removed", name)
		}
	}

	// Flux-managed apps must remain.
	if !env.AppExists(ctx, ns.Name, clusterID+"-security-pack") {
		t.Fatal("expected flux-managed demo0-security-pack to remain")
	}

	// Operator apps (managed by cluster-apps-operator) should remain because
	// EnsureDeleted only proceeds to delete them after all unmanaged apps are gone,
	// and the flux-managed apps are still present so it cancels.
	for _, name := range []string{clusterID + "-app-operator", clusterID + "-chart-operator"} {
		if !env.AppExists(ctx, ns.Name, name) {
			t.Fatalf("expected %s to remain (flux apps still present)", name)
		}
	}

	// Other cluster apps must remain.
	if !env.AppExists(ctx, ns.Name, "other0-hello-world") {
		t.Fatal("expected other0-hello-world to still exist")
	}
}

// TestApp_EnsureDeleted_InClusterWithoutLabel migrates the "flawless with
// in-cluster without label" test case: an in-cluster app without a cluster
// label is not targeted for deletion.
func TestApp_EnsureDeleted_InClusterWithoutLabel(t *testing.T) {
	env := harness.NewTestEnv(t)
	ctx := context.Background()

	clusterID := "demo0"
	ns := env.CreateNamespace(ctx, "test-app-icnl-"+randomSuffix())

	env.CreateApp(ctx, newTestApp(clusterID+"-app-operator", ns.Name, clusterID, project.Name(), true))
	env.CreateApp(ctx, newTestApp(clusterID+"-chart-operator", ns.Name, clusterID, project.Name(), false))
	env.CreateApp(ctx, newTestApp("other0-app-operator", ns.Name, clusterID, project.Name(), true))
	env.CreateApp(ctx, newTestApp("other0-chart-operator", ns.Name, clusterID, project.Name(), false))
	// In-cluster app WITHOUT cluster label (empty string).
	env.CreateApp(ctx, newTestApp(clusterID+"-security-pack", ns.Name, "", "", true))
	// Non-managed apps for target cluster.
	env.CreateApp(ctx, newTestApp(clusterID+"-trivy", ns.Name, clusterID, "", false))
	env.CreateApp(ctx, newTestApp(clusterID+"-falco", ns.Name, clusterID, "", false))
	// Different cluster.
	env.CreateApp(ctx, newTestApp("other0-hello-world", ns.Name, "other0", "", false))

	cluster := &capi.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      clusterID,
			Namespace: ns.Name,
			Labels: map[string]string{
				label.Cluster: clusterID,
			},
		},
	}

	resource := env.NewAppResource(harness.DefaultAppResourceOpts())

	err := resource.EnsureDeleted(ctx, cluster)
	if err != nil {
		t.Fatalf("EnsureDeleted returned error: %v", err)
	}

	// Non-managed cluster apps should be removed.
	for _, name := range []string{clusterID + "-trivy", clusterID + "-falco"} {
		if env.AppExists(ctx, ns.Name, name) {
			t.Fatalf("expected %s to be removed", name)
		}
	}

	// In-cluster app without cluster label should remain (not matched by selector).
	if !env.AppExists(ctx, ns.Name, clusterID+"-security-pack") {
		t.Fatal("expected demo0-security-pack (no cluster label) to remain")
	}

	// Other cluster apps must remain.
	for _, name := range []string{"other0-app-operator", "other0-chart-operator", "other0-hello-world"} {
		if !env.AppExists(ctx, ns.Name, name) {
			t.Fatalf("expected %s to still exist", name)
		}
	}
}
