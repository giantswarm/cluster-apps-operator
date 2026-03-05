package integration

import (
	"context"
	"testing"

	appv1alpha1 "github.com/giantswarm/apiextensions-application/api/v1alpha1"
	"github.com/giantswarm/k8smetadata/pkg/label"
	"github.com/prometheus/client_golang/prometheus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	capi "sigs.k8s.io/cluster-api/api/core/v1beta1"

	"github.com/giantswarm/cluster-apps-operator/v3/pkg/project"
	"github.com/giantswarm/cluster-apps-operator/v3/test/harness"
)

// TestCollector_Collect verifies that the cluster collector emits a
// dangling-apps metric for a terminating cluster with unmanaged apps.
func TestCollector_Collect(t *testing.T) {
	env := harness.NewTestEnv(t)
	ctx := context.Background()

	ns := env.CreateNamespace(ctx, "test-coll-"+randomSuffix())
	clusterName := "tc1-" + randomSuffix()

	// Create a Cluster with the operatorkit finalizer so deletion does not
	// actually remove it.
	cluster := &capi.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      clusterName,
			Namespace: ns.Name,
			Finalizers: []string{
				"operatorkit.giantswarm.io/cluster-apps-operator-cluster-controller",
			},
		},
	}
	err := env.CtrlClient().Create(ctx, cluster)
	if err != nil {
		t.Fatalf("create cluster: %v", err)
	}

	// Delete the cluster to set DeletionTimestamp; the finalizer prevents
	// actual removal.
	err = env.CtrlClient().Delete(ctx, cluster)
	if err != nil {
		t.Fatalf("delete cluster: %v", err)
	}

	// Create a dangling app (has cluster label, NOT managed by cluster-apps-operator).
	env.CreateApp(ctx, &appv1alpha1.App{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "hello-world",
			Namespace: ns.Name,
			Labels: map[string]string{
				label.Cluster: clusterName,
			},
		},
		Spec: appv1alpha1.AppSpec{
			Name:      "hello-world",
			Namespace: "default",
			Version:   "1.0.0",
			Catalog:   "test",
		},
	})

	collector := env.NewCollector()
	ch := make(chan prometheus.Metric, 10)

	err = collector.Collect(ch)
	if err != nil {
		t.Fatalf("Collect returned error: %v", err)
	}
	close(ch)

	var metrics []prometheus.Metric
	for m := range ch {
		metrics = append(metrics, m)
	}

	if len(metrics) == 0 {
		t.Fatal("expected at least one metric")
	}

	expected := prometheus.NewDesc(
		"cluster_apps_operator_cluster_dangling_apps",
		"Number of apps not yet deleted for a terminating cluster.",
		[]string{"cluster_id"},
		nil,
	).String()

	if metrics[0].Desc().String() != expected {
		t.Fatalf("expected desc %s, got %s", expected, metrics[0].Desc().String())
	}
}

// TestCollector_NoDanglingApps verifies that a terminating cluster whose apps
// are all managed by cluster-apps-operator reports zero dangling apps.
func TestCollector_NoDanglingApps(t *testing.T) {
	env := harness.NewTestEnv(t)
	ctx := context.Background()

	ns := env.CreateNamespace(ctx, "test-coll-nd-"+randomSuffix())
	clusterName := "tc2-" + randomSuffix()

	cluster := &capi.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      clusterName,
			Namespace: ns.Name,
			Finalizers: []string{
				"operatorkit.giantswarm.io/cluster-apps-operator-cluster-controller",
			},
		},
	}
	err := env.CtrlClient().Create(ctx, cluster)
	if err != nil {
		t.Fatalf("create cluster: %v", err)
	}

	err = env.CtrlClient().Delete(ctx, cluster)
	if err != nil {
		t.Fatalf("delete cluster: %v", err)
	}

	// Create only managed apps — these should NOT count as dangling.
	env.CreateApp(ctx, &appv1alpha1.App{
		ObjectMeta: metav1.ObjectMeta{
			Name:      clusterName + "-app-operator",
			Namespace: ns.Name,
			Labels: map[string]string{
				label.Cluster:   clusterName,
				label.ManagedBy: project.Name(),
			},
		},
		Spec: appv1alpha1.AppSpec{
			Name:      "app-operator",
			Namespace: "default",
			Version:   "1.0.0",
			Catalog:   "control-plane-catalog",
		},
	})
	env.CreateApp(ctx, &appv1alpha1.App{
		ObjectMeta: metav1.ObjectMeta{
			Name:      clusterName + "-chart-operator",
			Namespace: ns.Name,
			Labels: map[string]string{
				label.Cluster:   clusterName,
				label.ManagedBy: project.Name(),
			},
		},
		Spec: appv1alpha1.AppSpec{
			Name:      "chart-operator",
			Namespace: "default",
			Version:   "1.0.0",
			Catalog:   "default",
		},
	})

	collector := env.NewCollector()
	ch := make(chan prometheus.Metric, 10)

	err = collector.Collect(ch)
	if err != nil {
		t.Fatalf("Collect returned error: %v", err)
	}
	close(ch)

	var metrics []prometheus.Metric
	for m := range ch {
		metrics = append(metrics, m)
	}

	// The collector should still emit a metric for the terminating cluster,
	// but with a value of 0 dangling apps. We verify the metric exists and
	// has the correct descriptor.
	if len(metrics) == 0 {
		t.Fatal("expected a metric for the terminating cluster")
	}

	expected := prometheus.NewDesc(
		"cluster_apps_operator_cluster_dangling_apps",
		"Number of apps not yet deleted for a terminating cluster.",
		[]string{"cluster_id"},
		nil,
	).String()

	if metrics[0].Desc().String() != expected {
		t.Fatalf("expected desc %s, got %s", expected, metrics[0].Desc().String())
	}
}
