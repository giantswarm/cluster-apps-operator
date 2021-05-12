package key

import (
	"reflect"
	"testing"

	"github.com/giantswarm/k8smetadata/pkg/label"
	"github.com/giantswarm/microerror"
	apiv1alpha3 "sigs.k8s.io/cluster-api/api/v1alpha3"
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
		expectedCustomObject apiv1alpha3.Cluster
		expectedError        error
	}{
		{
			description:          "reference to empty value Cluster returns empty Cluster",
			inputObject:          &apiv1alpha3.Cluster{},
			expectedCustomObject: apiv1alpha3.Cluster{},
			expectedError:        nil,
		},
		{
			description:          "non-pointer value of Cluster must return wrongTypeError",
			inputObject:          apiv1alpha3.Cluster{},
			expectedCustomObject: apiv1alpha3.Cluster{},
			expectedError:        wrongTypeError,
		},
		{
			description:          "wrong type must return wrongTypeError",
			inputObject:          &apiv1alpha3.Machine{},
			expectedCustomObject: apiv1alpha3.Cluster{},
			expectedError:        wrongTypeError,
		},
		{
			description:          "nil interface{} must return wrongTypeError",
			inputObject:          nil,
			expectedCustomObject: apiv1alpha3.Cluster{},
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
