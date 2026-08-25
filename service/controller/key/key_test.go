package key

import (
	"reflect"
	"testing"

	"github.com/giantswarm/apiextensions-application/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/k8smetadata/pkg/label"
	"github.com/giantswarm/microerror"
	capi "sigs.k8s.io/cluster-api/api/core/v1beta1"
)

// A mock object that implements LabelsGetter interface
type testObject struct {
	labels map[string]string
}

func (to *testObject) GetLabels() map[string]string {
	return to.labels
}

func Test_ClusterID(t *testing.T) {
	testCases := []struct {
		description  string
		customObject LabelsGetter
		expectedID   string
	}{
		{
			description:  "empty value object produces empty ID",
			customObject: &testObject{},
			expectedID:   "",
		},
		{
			description:  "present ID value returned as ClusterID",
			customObject: &testObject{map[string]string{label.Cluster: "cluster-1"}},
			expectedID:   "cluster-1",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			if cid := ClusterID(tc.customObject); cid != tc.expectedID {
				t.Fatalf("ClusterID %s doesn't match. expected: %s", cid, tc.expectedID)
			}
		})
	}
}

func Test_ToCluster(t *testing.T) {
	testCases := []struct {
		description          string
		inputObject          interface{}
		expectedCustomObject capi.Cluster
		expectedError        error
	}{
		{
			description:          "reference to empty value Cluster returns empty Cluster",
			inputObject:          &capi.Cluster{},
			expectedCustomObject: capi.Cluster{},
			expectedError:        nil,
		},
		{
			description:          "non-pointer value of Cluster must return wrongTypeError",
			inputObject:          capi.Cluster{},
			expectedCustomObject: capi.Cluster{},
			expectedError:        wrongTypeError,
		},
		{
			description:          "wrong type must return wrongTypeError",
			inputObject:          &capi.Machine{},
			expectedCustomObject: capi.Cluster{},
			expectedError:        wrongTypeError,
		},
		{
			description:          "nil interface{} must return wrongTypeError",
			inputObject:          nil,
			expectedCustomObject: capi.Cluster{},
			expectedError:        wrongTypeError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			object, err := ToCluster(tc.inputObject)
			if microerror.Cause(err) != tc.expectedError {
				t.Errorf("Received error %#v doesn't match expected %#v",
					err, tc.expectedError)
			}

			if !reflect.DeepEqual(object, tc.expectedCustomObject) {
				t.Fatalf("object %#v doesn't match expected %#v",
					object, tc.expectedCustomObject)
			}
		})
	}
}

func Test_IsManagedByFlux(t *testing.T) {
	testCases := []struct {
		description string
		input       *v1alpha1.App
		expected    bool
	}{
		{
			"case 1: No Flux kustomization labels are set",
			&v1alpha1.App{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "eggs2-net-exporter",
					Namespace: "org-test",
					Labels: map[string]string{
						"foo": "bar",
					},
				},
				Spec: v1alpha1.AppSpec{
					Catalog: "control-plane-catalog",
					Name:    "app-operator",
				},
			},
			false,
		},
		{
			"case 2: Partial Flux kustomization labels are set (name missing)",
			&v1alpha1.App{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "eggs2-net-exporter",
					Namespace: "org-test",
					Labels: map[string]string{
						"foo":                                   "bar",
						"kustomize.toolkit.fluxcd.io/namespace": "default",
					},
				},
				Spec: v1alpha1.AppSpec{
					Catalog: "control-plane-catalog",
					Name:    "app-operator",
				},
			},
			false,
		},
		{
			"case 3: Partial Flux kustomization labels are set (namespace missing)",
			&v1alpha1.App{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "eggs2-net-exporter",
					Namespace: "org-test",
					Labels: map[string]string{
						"kustomize.toolkit.fluxcd.io/name": "test-cluster-eggs2",
						"foo":                              "bar",
					},
				},
				Spec: v1alpha1.AppSpec{
					Catalog: "control-plane-catalog",
					Name:    "app-operator",
				},
			},
			false,
		},
		{
			"case 4: All Flux kustomization labels are set",
			&v1alpha1.App{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "eggs2-net-exporter",
					Namespace: "org-test",
					Labels: map[string]string{
						"foo":                                   "bar",
						"kustomize.toolkit.fluxcd.io/name":      "test-cluster-eggs2",
						"bar":                                   "baz",
						"kustomize.toolkit.fluxcd.io/namespace": "default",
						"not":                                   "used",
					},
				},
				Spec: v1alpha1.AppSpec{
					Catalog: "control-plane-catalog",
					Name:    "app-operator",
				},
			},
			true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			result := IsManagedByFlux(*tc.input)

			if result != tc.expected {
				t.Fatalf("Got the unexpected result for managed by Flux check for: %#v", tc.input)
			}
		})
	}
}

// Test_ServiceCIDR pins the contract that clusterconfigmap's fallback and the
// cluster_service_cidr_missing metric both depend on: ServiceCIDR returns the
// empty string for every "not set" shape, and those shapes are
// indistinguishable to callers. See giantswarm/giantswarm#37031.
func Test_ServiceCIDR(t *testing.T) {
	testCases := []struct {
		description string
		cluster     capi.Cluster
		expected    string
	}{
		{
			description: "case 0: services CIDR is set",
			cluster: capi.Cluster{
				Spec: capi.ClusterSpec{
					ClusterNetwork: &capi.ClusterNetwork{
						Services: &capi.NetworkRanges{
							CIDRBlocks: []string{"172.31.0.0/16"},
						},
					},
				},
			},
			expected: "172.31.0.0/16",
		},
		{
			description: "case 1: only the first CIDR block is returned",
			cluster: capi.Cluster{
				Spec: capi.ClusterSpec{
					ClusterNetwork: &capi.ClusterNetwork{
						Services: &capi.NetworkRanges{
							CIDRBlocks: []string{"172.31.0.0/16", "172.32.0.0/16"},
						},
					},
				},
			},
			expected: "172.31.0.0/16",
		},
		{
			description: "case 2: clusterNetwork is nil",
			cluster:     capi.Cluster{},
			expected:    "",
		},
		{
			description: "case 3: services is nil",
			cluster: capi.Cluster{
				Spec: capi.ClusterSpec{
					ClusterNetwork: &capi.ClusterNetwork{},
				},
			},
			expected: "",
		},
		{
			description: "case 4: cidrBlocks is empty",
			cluster: capi.Cluster{
				Spec: capi.ClusterSpec{
					ClusterNetwork: &capi.ClusterNetwork{
						Services: &capi.NetworkRanges{
							CIDRBlocks: []string{},
						},
					},
				},
			},
			expected: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			result := ServiceCIDR(tc.cluster)
			if result != tc.expected {
				t.Fatalf("expected %#q but got %#q", tc.expected, result)
			}
		})
	}
}

// Test_DNSIP covers the ".10 of the services CIDR" convention that
// cluster-apps-operator implements independently of the identical Helm-side
// implementation in giantswarm/cluster.
func Test_DNSIP(t *testing.T) {
	testCases := []struct {
		description    string
		clusterIPRange string
		expected       string
		expectError    bool
	}{
		{
			description:    "case 0: default installation CIDR",
			clusterIPRange: "10.96.0.0/12",
			expected:       "10.96.0.10",
		},
		{
			description:    "case 1: cluster chart default",
			clusterIPRange: "172.31.0.0/16",
			expected:       "172.31.0.10",
		},
		{
			description:    "case 2: not a network address",
			clusterIPRange: "172.31.0.5/16",
			expectError:    true,
		},
		{
			description:    "case 3: not a CIDR",
			clusterIPRange: "172.31.0.0",
			expectError:    true,
		},
		{
			description:    "case 4: IPv6 is unsupported",
			clusterIPRange: "fd00::/108",
			expectError:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			result, err := DNSIP(tc.clusterIPRange)
			if tc.expectError {
				if err == nil {
					t.Fatalf("expected an error but got %#q", result)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if result != tc.expected {
				t.Fatalf("expected %#q but got %#q", tc.expected, result)
			}
		})
	}
}
