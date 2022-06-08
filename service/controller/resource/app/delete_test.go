package app

import (
	"context"
	"fmt"
	"testing"

	"github.com/giantswarm/apiextensions-application/api/v1alpha1"
	"github.com/giantswarm/k8smetadata/pkg/label"
	"github.com/giantswarm/micrologger/microloggertest"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	capi "sigs.k8s.io/cluster-api/api/v1alpha4"
	"sigs.k8s.io/controller-runtime/pkg/client/fake" //nolint:staticcheck

	"github.com/giantswarm/cluster-apps-operator/pkg/project"
)

func Test_EnsureDeleted(t *testing.T) {
	testCases := []struct {
		name                string
		apps                []*v1alpha1.App
		cluster             *capi.Cluster
		config              Config
		expectedAppsLeft    []types.NamespacedName
		expectedAppsRemoved []types.NamespacedName
	}{
		{
			name: "flawless",
			apps: []*v1alpha1.App{
				newAppCR("demo0-app-operator", "org-acme", "demo0", project.Name(), true),
				newAppCR("demo0-chart-operator", "org-acme", "demo0", project.Name(), false),
				newAppCR("other0-app-operator", "org-acme", "demo0", project.Name(), true),
				newAppCR("other0-chart-operator", "org-acme", "demo0", project.Name(), false),
				newAppCR("demo0-hello-world", "org-acme", "demo0", "", false),
				newAppCR("other0-hello-world", "org-acme", "other0", "", false),
				newAppCR("other0-kyverno-policies", "org-acme", "other0", "", false),
			},
			cluster: &capi.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "demo0",
					Namespace: "org-acme",
					Labels: map[string]string{
						label.Cluster: "demo0",
					},
				},
			},
			config: Config{
				AppOperatorCatalog:   "control-plane-catalog",
				AppOperatorVersion:   "1.0.0",
				ChartOperatorCatalog: "default",
				ChartOperatorVersion: "1.0.0",
			},
			expectedAppsLeft: []types.NamespacedName{
				types.NamespacedName{
					Name:      "other0-app-operator",
					Namespace: "org-acme",
				},
				types.NamespacedName{
					Name:      "other0-chart-operator",
					Namespace: "org-acme",
				},
				types.NamespacedName{
					Name:      "other0-hello-world",
					Namespace: "org-acme",
				},
				types.NamespacedName{
					Name:      "other0-kyverno-policies",
					Namespace: "org-acme",
				},
			},
			expectedAppsRemoved: []types.NamespacedName{
				types.NamespacedName{
					Name:      "demo0-app-operator",
					Namespace: "org-acme",
				},
				types.NamespacedName{
					Name:      "demo0-chart-operator",
					Namespace: "org-acme",
				},
				types.NamespacedName{
					Name:      "demo0-hello-world",
					Namespace: "org-acme",
				},
			},
		},
		{
			name: "flawless with in-cluster",
			apps: []*v1alpha1.App{
				newAppCR("demo0-app-operator", "org-acme", "demo0", project.Name(), true),
				newAppCR("demo0-chart-operator", "org-acme", "demo0", project.Name(), false),
				newAppCR("other0-app-operator", "org-acme", "demo0", project.Name(), true),
				newAppCR("other0-chart-operator", "org-acme", "demo0", project.Name(), false),
				newAppCR("demo0-security-pack", "org-acme", "demo0", "", true),
				newAppCR("demo0-trivy", "org-acme", "demo0", "", false),
				newAppCR("demo0-falco", "org-acme", "demo0", "", false),
				newAppCR("other0-hello-world", "org-acme", "other0", "", false),
			},
			cluster: &capi.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "demo0",
					Namespace: "org-acme",
					Labels: map[string]string{
						label.Cluster: "demo0",
					},
				},
			},
			config: Config{
				AppOperatorCatalog:   "control-plane-catalog",
				AppOperatorVersion:   "1.0.0",
				ChartOperatorCatalog: "default",
				ChartOperatorVersion: "1.0.0",
			},
			expectedAppsLeft: []types.NamespacedName{
				types.NamespacedName{
					Name:      "other0-app-operator",
					Namespace: "org-acme",
				},
				types.NamespacedName{
					Name:      "other0-chart-operator",
					Namespace: "org-acme",
				},
				types.NamespacedName{
					Name:      "other0-hello-world",
					Namespace: "org-acme",
				},
			},
			expectedAppsRemoved: []types.NamespacedName{
				types.NamespacedName{
					Name:      "demo0-app-operator",
					Namespace: "org-acme",
				},
				types.NamespacedName{
					Name:      "demo0-chart-operator",
					Namespace: "org-acme",
				},
				types.NamespacedName{
					Name:      "demo0-falco",
					Namespace: "org-acme",
				},
				types.NamespacedName{
					Name:      "demo0-trivy",
					Namespace: "org-acme",
				},
				types.NamespacedName{
					Name:      "demo0-security-pack",
					Namespace: "org-acme",
				},
			},
		},
		{
			name: "flawless with Flux managed apps",
			apps: []*v1alpha1.App{
				newAppCR("demo0-app-operator", "org-acme", "demo0", project.Name(), true),
				newAppCR("demo0-chart-operator", "org-acme", "demo0", project.Name(), false),
				newAppCR("other0-app-operator", "org-acme", "demo0", project.Name(), true),
				newAppCR("other0-chart-operator", "org-acme", "demo0", project.Name(), false),
				newAppCR("demo0-security-pack", "org-acme", "demo0", "flux", true),
				newAppCR("demo0-trivy", "org-acme", "demo0", "", false),
				newAppCR("demo0-falco", "org-acme", "demo0", "", false),
				newAppCR("demo0-hello-world", "org-acme", "demo0", "flux", false),
				newAppCR("other0-hello-world", "org-acme", "other0", "", false),
			},
			cluster: &capi.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "demo0",
					Namespace: "org-acme",
					Labels: map[string]string{
						label.Cluster: "demo0",
					},
				},
			},
			config: Config{
				AppOperatorCatalog:   "control-plane-catalog",
				AppOperatorVersion:   "1.0.0",
				ChartOperatorCatalog: "default",
				ChartOperatorVersion: "1.0.0",
			},
			expectedAppsLeft: []types.NamespacedName{
				types.NamespacedName{
					Name:      "demo0-app-operator",
					Namespace: "org-acme",
				},
				types.NamespacedName{
					Name:      "demo0-chart-operator",
					Namespace: "org-acme",
				},
				types.NamespacedName{
					Name:      "other0-app-operator",
					Namespace: "org-acme",
				},
				types.NamespacedName{
					Name:      "other0-chart-operator",
					Namespace: "org-acme",
				},
				types.NamespacedName{
					Name:      "other0-hello-world",
					Namespace: "org-acme",
				},
				types.NamespacedName{
					Name:      "demo0-security-pack",
					Namespace: "org-acme",
				},
			},
			expectedAppsRemoved: []types.NamespacedName{
				types.NamespacedName{
					Name:      "demo0-falco",
					Namespace: "org-acme",
				},
				types.NamespacedName{
					Name:      "demo0-trivy",
					Namespace: "org-acme",
				},
			},
		},
		{
			name: "flawless with in-cluster without label",
			apps: []*v1alpha1.App{
				newAppCR("demo0-app-operator", "org-acme", "demo0", project.Name(), true),
				newAppCR("demo0-chart-operator", "org-acme", "demo0", project.Name(), false),
				newAppCR("other0-app-operator", "org-acme", "demo0", project.Name(), true),
				newAppCR("other0-chart-operator", "org-acme", "demo0", project.Name(), false),
				newAppCR("demo0-security-pack", "org-acme", "", "", true),
				newAppCR("demo0-trivy", "org-acme", "demo0", "", false),
				newAppCR("demo0-falco", "org-acme", "demo0", "", false),
				newAppCR("other0-hello-world", "org-acme", "other0", "", false),
			},
			cluster: &capi.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "demo0",
					Namespace: "org-acme",
					Labels: map[string]string{
						label.Cluster: "demo0",
					},
				},
			},
			config: Config{
				AppOperatorCatalog:   "control-plane-catalog",
				AppOperatorVersion:   "1.0.0",
				ChartOperatorCatalog: "default",
				ChartOperatorVersion: "1.0.0",
			},
			expectedAppsLeft: []types.NamespacedName{
				types.NamespacedName{
					Name:      "other0-app-operator",
					Namespace: "org-acme",
				},
				types.NamespacedName{
					Name:      "other0-chart-operator",
					Namespace: "org-acme",
				},
				types.NamespacedName{
					Name:      "other0-hello-world",
					Namespace: "org-acme",
				},
				types.NamespacedName{
					Name:      "demo0-security-pack",
					Namespace: "org-acme",
				},
			},
			expectedAppsRemoved: []types.NamespacedName{
				types.NamespacedName{
					Name:      "demo0-app-operator",
					Namespace: "org-acme",
				},
				types.NamespacedName{
					Name:      "demo0-chart-operator",
					Namespace: "org-acme",
				},
				types.NamespacedName{
					Name:      "demo0-falco",
					Namespace: "org-acme",
				},
				types.NamespacedName{
					Name:      "demo0-trivy",
					Namespace: "org-acme",
				},
			},
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("case %d: %s", i, tc.name), func(t *testing.T) {
			var err error

			ctx := context.TODO()

			g8sObjs := make([]runtime.Object, 0)
			for _, app := range tc.apps {
				g8sObjs = append(g8sObjs, app)
			}

			scheme := runtime.NewScheme()
			_ = v1alpha1.AddToScheme(scheme)

			var resource *Resource
			{
				tc.config.CtrlClient = fake.NewClientBuilder().
					WithScheme(scheme).
					WithRuntimeObjects(g8sObjs...).
					Build()
				tc.config.Logger = microloggertest.New()

				resource, err = New(tc.config)
				if err != nil {
					t.Fatalf("error == %#v, want nil", err)
				}
			}

			err = resource.EnsureDeleted(ctx, tc.cluster)
			if err != nil {
				t.Fatalf("error == %#v, want nil", err)
			}

			var app v1alpha1.App
			for _, a := range tc.expectedAppsLeft {
				err = resource.ctrlClient.Get(ctx, a, &app)

				if apierrors.IsNotFound(err) {
					t.Fatalf("expected %s to be left", a.Name)
				} else if err != nil {
					t.Fatalf("error == %#v, want nil", err)
				}
			}

			for _, a := range tc.expectedAppsRemoved {
				err = resource.ctrlClient.Get(ctx, a, &app)

				if err == nil {
					t.Fatalf("expected %s to be gone", a.Name)
				}

				if err != nil && !apierrors.IsNotFound(err) {
					t.Fatalf("unexpected error == %#v, want 'NotFound'", err)
				}
			}
		})
	}
}

func newAppCR(name, namespace, cluster, managedBy string, inCluster bool) *v1alpha1.App {
	app := v1alpha1.App{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    map[string]string{},
		},
		Spec: v1alpha1.AppSpec{
			KubeConfig: v1alpha1.AppSpecKubeConfig{
				InCluster: false,
			},
		},
	}

	if cluster != "" {
		app.Labels[label.Cluster] = cluster
	}

	if managedBy != "" {
		app.Labels[label.ManagedBy] = managedBy
	}

	if inCluster {
		app.Spec.KubeConfig.InCluster = true
	}

	return &app
}
