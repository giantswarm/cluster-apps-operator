package key

import (
	"reflect"
	"testing"

	"github.com/giantswarm/apiextensions-application/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/k8smetadata/pkg/label"
	"github.com/giantswarm/microerror"
	capi "sigs.k8s.io/cluster-api/api/v1alpha4"
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
