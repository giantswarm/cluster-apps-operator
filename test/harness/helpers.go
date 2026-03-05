package harness

import (
	"context"

	appv1alpha1 "github.com/giantswarm/apiextensions-application/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	capi "sigs.k8s.io/cluster-api/api/core/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// CreateNamespace creates a namespace and registers cleanup to delete it.
func (e *TestEnv) CreateNamespace(ctx context.Context, name string) *corev1.Namespace {
	e.T.Helper()

	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}

	err := e.ctrlClient.Create(ctx, ns)
	if err != nil {
		e.T.Fatalf("harness: create namespace %q: %v", name, err)
	}

	e.T.Cleanup(func() {
		_ = e.ctrlClient.Delete(context.Background(), ns)
	})

	return ns
}

// ClusterOpts configures a CAPI Cluster for creation.
type ClusterOpts struct {
	Name            string
	Namespace       string
	Labels          map[string]string
	InfraRef        *corev1.ObjectReference
	ControlPlaneRef *corev1.ObjectReference
	ClusterNetwork  *capi.ClusterNetwork
}

// CreateCluster creates a CAPI Cluster with the given options.
func (e *TestEnv) CreateCluster(ctx context.Context, opts ClusterOpts) *capi.Cluster {
	e.T.Helper()

	cluster := &capi.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      opts.Name,
			Namespace: opts.Namespace,
			Labels:    opts.Labels,
		},
		Spec: capi.ClusterSpec{
			ClusterNetwork: opts.ClusterNetwork,
		},
	}

	if opts.InfraRef != nil {
		cluster.Spec.InfrastructureRef = opts.InfraRef
	}

	if opts.ControlPlaneRef != nil {
		cluster.Spec.ControlPlaneRef = opts.ControlPlaneRef
	}

	err := e.ctrlClient.Create(ctx, cluster)
	if err != nil {
		e.T.Fatalf("harness: create cluster %s/%s: %v", opts.Namespace, opts.Name, err)
	}

	// Re-apply refs after creation. The CAPI CRD stores objects in v1beta2
	// format where InfrastructureRef uses apiGroup instead of apiVersion.
	// Without the conversion webhook the v1beta2 schema prunes apiVersion
	// during the round-trip. Re-setting the refs ensures the returned
	// Cluster has the full ObjectReference the production code expects.
	if opts.InfraRef != nil {
		cluster.Spec.InfrastructureRef = opts.InfraRef
	}
	if opts.ControlPlaneRef != nil {
		cluster.Spec.ControlPlaneRef = opts.ControlPlaneRef
	}

	return cluster
}

// CreateInfraCluster creates an unstructured infrastructure cluster object.
func (e *TestEnv) CreateInfraCluster(ctx context.Context, gvk schema.GroupVersionKind, namespace, name string, spec map[string]interface{}) *unstructured.Unstructured {
	e.T.Helper()

	obj := &unstructured.Unstructured{}
	obj.SetGroupVersionKind(gvk)
	obj.SetName(name)
	obj.SetNamespace(namespace)

	if spec != nil {
		obj.Object["spec"] = spec
	}

	err := e.ctrlClient.Create(ctx, obj)
	if err != nil {
		e.T.Fatalf("harness: create infra cluster %s/%s (%s): %v", namespace, name, gvk.Kind, err)
	}

	return obj
}

// CreateApp creates an App CR.
func (e *TestEnv) CreateApp(ctx context.Context, app *appv1alpha1.App) {
	e.T.Helper()

	err := e.ctrlClient.Create(ctx, app)
	if err != nil {
		e.T.Fatalf("harness: create app %s/%s: %v", app.Namespace, app.Name, err)
	}
}

// GetConfigMap retrieves a ConfigMap.
func (e *TestEnv) GetConfigMap(ctx context.Context, namespace, name string) *corev1.ConfigMap {
	e.T.Helper()

	cm := &corev1.ConfigMap{}
	err := e.ctrlClient.Get(ctx, types.NamespacedName{Namespace: namespace, Name: name}, cm)
	if err != nil {
		e.T.Fatalf("harness: get configmap %s/%s: %v", namespace, name, err)
	}

	return cm
}

// GetSecret retrieves a Secret.
func (e *TestEnv) GetSecret(ctx context.Context, namespace, name string) *corev1.Secret {
	e.T.Helper()

	secret := &corev1.Secret{}
	err := e.ctrlClient.Get(ctx, types.NamespacedName{Namespace: namespace, Name: name}, secret)
	if err != nil {
		e.T.Fatalf("harness: get secret %s/%s: %v", namespace, name, err)
	}

	return secret
}

// GetApp retrieves an App CR.
func (e *TestEnv) GetApp(ctx context.Context, namespace, name string) *appv1alpha1.App {
	e.T.Helper()

	app := &appv1alpha1.App{}
	err := e.ctrlClient.Get(ctx, types.NamespacedName{Namespace: namespace, Name: name}, app)
	if err != nil {
		e.T.Fatalf("harness: get app %s/%s: %v", namespace, name, err)
	}

	return app
}

// ListApps lists App CRs in the given namespace matching the label selector.
func (e *TestEnv) ListApps(ctx context.Context, namespace string, labelSelector map[string]string) []appv1alpha1.App {
	e.T.Helper()

	list := &appv1alpha1.AppList{}
	opts := []client.ListOption{
		client.InNamespace(namespace),
	}
	if len(labelSelector) > 0 {
		opts = append(opts, client.MatchingLabels(labelSelector))
	}

	err := e.ctrlClient.List(ctx, list, opts...)
	if err != nil {
		e.T.Fatalf("harness: list apps in %s: %v", namespace, err)
	}

	return list.Items
}

// AppExists checks if an App CR exists.
func (e *TestEnv) AppExists(ctx context.Context, namespace, name string) bool {
	e.T.Helper()

	app := &appv1alpha1.App{}
	err := e.ctrlClient.Get(ctx, types.NamespacedName{Namespace: namespace, Name: name}, app)
	return err == nil
}
