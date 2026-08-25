package collector

import (
	"fmt"
	"testing"

	"github.com/giantswarm/k8smetadata/pkg/label"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	capi "sigs.k8s.io/cluster-api/api/core/v1beta1"
	clientfake "sigs.k8s.io/controller-runtime/pkg/client/fake"

	applicationv1alpha1 "github.com/giantswarm/apiextensions-application/api/v1alpha1"
	"github.com/giantswarm/k8sclient/v8/pkg/k8sclienttest"
	"github.com/giantswarm/micrologger/microloggertest"
)

func TestClusterCollector(t *testing.T) {
	testcases := []struct {
		name      string
		resources []runtime.Object
		expected  []string
		err       error
	}{
		{
			name: "flawless",
			resources: []runtime.Object{
				newCAPIV1alpha4Cluster("1abc2", "org-test"),
				newV1alpha1App("hello-world", "org-test", "1abc2", ""),
			},
			expected: []string{
				prometheus.NewDesc(
					"cluster_apps_operator_cluster_service_cidr_missing",
					"1 if the Cluster CR has no spec.clusterNetwork.services.cidrBlocks, so clusterDNSIP falls back to the installation default.",
					[]string{
						labelClusterID,
						labelClusterNamespace,
					},
					nil).String(),
				prometheus.NewDesc(
					"cluster_apps_operator_cluster_dangling_apps",
					"Number of apps not yet deleted for a terminating cluster.",
					[]string{
						labelClusterID,
					},
					nil).String(),
			},
		},
	}

	for _, test := range testcases {
		t.Run(test.name, func(t *testing.T) {
			var err error

			var fakeClient *k8sclienttest.Clients
			{
				schemeBuilder := runtime.SchemeBuilder{
					applicationv1alpha1.AddToScheme,
					capi.AddToScheme,
				}

				err = schemeBuilder.AddToScheme(scheme.Scheme)
				if err != nil {
					t.Fatal(err)
				}

				fakeClient = k8sclienttest.NewClients(k8sclienttest.ClientsConfig{
					CtrlClient: clientfake.NewClientBuilder().
						WithScheme(scheme.Scheme).
						WithRuntimeObjects(test.resources...).
						Build(),
				})
			}

			var clusterCollector *Cluster
			{
				clusterConfig := ClusterConfig{
					K8sClient: fakeClient,
					Logger:    microloggertest.New(),
				}

				clusterCollector, err = NewCluster(clusterConfig)
				if err != nil {
					t.Fatal(err)
				}
			}

			ch := make(chan prometheus.Metric)
			go func() {
				err = clusterCollector.Collect(ch)
				if err != nil {
					panic(fmt.Sprintf("failed to collect metrics: %v", err))
				}
			}()

			for _, expected := range test.expected {
				got := (<-ch).Desc().String()
				if expected != got {
					t.Fatalf("Expected '%s' but got '%s'", expected, got)
				}
			}

		})
	}
}

func Test_getNumberOfApps(t *testing.T) {
	testcases := []struct {
		name             string
		clusterName      string
		clusterNamespace string
		resources        []runtime.Object
		expected         int
	}{
		{
			name:             "flawless with a single app for cluster",
			clusterName:      "1abc2",
			clusterNamespace: "org-test",
			resources: []runtime.Object{
				newV1alpha1App("hello-world", "org-test", "1abc2", ""),
			},
			expected: 1,
		},
		{
			name:             "flawless for ignoring other cluster apps",
			clusterName:      "1abc2",
			clusterNamespace: "org-test",
			resources: []runtime.Object{
				newV1alpha1App("hello-world", "org-test", "3def4", ""),
			},
			expected: 0,
		},
		{
			name:             "flawless for ignoring managed apps",
			clusterName:      "1abc2",
			clusterNamespace: "org-test",
			resources: []runtime.Object{
				newV1alpha1App("app-operator", "org-test", "1abc2", "cluster-apps-operator"),
				newV1alpha1App("chart-operator", "org-test", "1abc2", "cluster-apps-operator"),
			},
			expected: 0,
		},
		{
			name:             "flawless with mixed resources",
			clusterName:      "1abc2",
			clusterNamespace: "org-test",
			resources: []runtime.Object{
				newV1alpha1App("app-operator", "org-test", "1abc2", "cluster-apps-operator"),
				newV1alpha1App("chart-operator", "org-test", "1abc2", "cluster-apps-operator"),
				newV1alpha1App("hello-world-0", "org-test", "3def4", "flux"),
				newV1alpha1App("hello-world-1", "org-test", "1abc2", ""),
				newV1alpha1App("hello-world-2", "org-test", "1abc2", "flux"),
			},
			expected: 2,
		},
	}

	for _, test := range testcases {
		t.Run(test.name, func(t *testing.T) {
			var err error

			var fakeClient *k8sclienttest.Clients
			{
				schemeBuilder := runtime.SchemeBuilder{
					applicationv1alpha1.AddToScheme,
				}

				err = schemeBuilder.AddToScheme(scheme.Scheme)
				if err != nil {
					t.Fatal(err)
				}

				fakeClient = k8sclienttest.NewClients(k8sclienttest.ClientsConfig{
					CtrlClient: clientfake.NewClientBuilder().
						WithScheme(scheme.Scheme).
						WithRuntimeObjects(test.resources...).
						Build(),
				})
			}

			var clusterCollector *Cluster
			{
				clusterConfig := ClusterConfig{
					K8sClient: fakeClient,
					Logger:    microloggertest.New(),
				}

				clusterCollector, err = NewCluster(clusterConfig)
				if err != nil {
					t.Fatal(err)
				}
			}

			got, err := clusterCollector.getNumberOfApps(test.clusterName, test.clusterNamespace)
			if err != nil {
				t.Fatal(err)
			}

			if test.expected != got {
				t.Fatalf("Expected '%d' but got '%d'", test.expected, got)
			}
		})
	}
}

func newV1alpha1App(name, namespace, cluster, managedBy string) *applicationv1alpha1.App {
	metaLabels := map[string]string{}

	if cluster != "" {
		metaLabels[label.Cluster] = cluster
	}

	if managedBy != "" {
		metaLabels[label.ManagedBy] = managedBy
	}

	c := &applicationv1alpha1.App{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "application.giantswarm.io/v1alpha1",
			Kind:       "App",
		},
		ObjectMeta: metav1.ObjectMeta{
			Labels:    metaLabels,
			Name:      name,
			Namespace: namespace,
		},
	}

	return c
}

func newCAPIV1alpha4Cluster(id, namespace string) *capi.Cluster {
	timestamp := metav1.Now()

	c := &capi.Cluster{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "cluster.x-k8s.io/v1alpha3",
			Kind:       "Cluster",
		},
		ObjectMeta: metav1.ObjectMeta{
			Finalizers: []string{
				"operatorkit.giantswarm.io/cluster-apps-operator-cluster-controller",
			},
			Name:              id,
			Namespace:         namespace,
			CreationTimestamp: timestamp,
			DeletionTimestamp: &timestamp,
		},
	}

	return c
}

// Test_ClusterCollectorServiceCIDRMissing asserts the emitted value, not just
// the descriptor: 1 exactly when the Cluster CR carries no services CIDR, in
// which case clusterconfigmap falls back to the installation default
// clusterDNSIP.
func Test_ClusterCollectorServiceCIDRMissing(t *testing.T) {
	testcases := []struct {
		name      string
		resources []runtime.Object
		expected  map[string]float64
	}{
		{
			name: "services CIDR set",
			resources: []runtime.Object{
				newCAPIClusterWithServiceCIDR("1abc2", "org-test", "172.31.0.0/16"),
			},
			expected: map[string]float64{"1abc2": 0},
		},
		{
			name: "services CIDR missing",
			resources: []runtime.Object{
				newCAPIClusterWithServiceCIDR("3def4", "org-test", ""),
			},
			expected: map[string]float64{"3def4": 1},
		},
		{
			name: "reported per cluster",
			resources: []runtime.Object{
				newCAPIClusterWithServiceCIDR("1abc2", "org-test", "172.31.0.0/16"),
				newCAPIClusterWithServiceCIDR("3def4", "org-test", ""),
			},
			expected: map[string]float64{"1abc2": 0, "3def4": 1},
		},
	}

	for _, test := range testcases {
		t.Run(test.name, func(t *testing.T) {
			var err error

			var fakeClient *k8sclienttest.Clients
			{
				schemeBuilder := runtime.SchemeBuilder{
					applicationv1alpha1.AddToScheme,
					capi.AddToScheme,
				}

				err = schemeBuilder.AddToScheme(scheme.Scheme)
				if err != nil {
					t.Fatal(err)
				}

				fakeClient = k8sclienttest.NewClients(k8sclienttest.ClientsConfig{
					CtrlClient: clientfake.NewClientBuilder().
						WithScheme(scheme.Scheme).
						WithRuntimeObjects(test.resources...).
						Build(),
				})
			}

			var clusterCollector *Cluster
			{
				clusterCollector, err = NewCluster(ClusterConfig{
					K8sClient: fakeClient,
					Logger:    microloggertest.New(),
				})
				if err != nil {
					t.Fatal(err)
				}
			}

			ch := make(chan prometheus.Metric)
			go func() {
				err := clusterCollector.Collect(ch)
				if err != nil {
					panic(fmt.Sprintf("failed to collect metrics: %v", err))
				}
			}()

			// The fixtures are not terminating, so dangling_apps is skipped and
			// service_cidr_missing is the only metric per cluster.
			got := map[string]float64{}
			for range test.expected {
				metric := <-ch

				var pb dto.Metric
				err = metric.Write(&pb)
				if err != nil {
					t.Fatal(err)
				}

				var clusterID string
				for _, l := range pb.GetLabel() {
					if l.GetName() == labelClusterID {
						clusterID = l.GetValue()
					}
				}

				got[clusterID] = pb.GetGauge().GetValue()
			}

			for clusterID, want := range test.expected {
				if got[clusterID] != want {
					t.Fatalf("cluster %#q: expected %v but got %v", clusterID, want, got[clusterID])
				}
			}
		})
	}
}

// newCAPIClusterWithServiceCIDR returns a live (non-terminating) Cluster CR. An
// empty serviceCIDR leaves spec.clusterNetwork unset entirely.
func newCAPIClusterWithServiceCIDR(id, namespace, serviceCIDR string) *capi.Cluster {
	c := &capi.Cluster{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "cluster.x-k8s.io/v1beta1",
			Kind:       "Cluster",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      id,
			Namespace: namespace,
		},
	}

	if serviceCIDR != "" {
		c.Spec.ClusterNetwork = &capi.ClusterNetwork{
			Services: &capi.NetworkRanges{
				CIDRBlocks: []string{serviceCIDR},
			},
		}
	}

	return c
}
