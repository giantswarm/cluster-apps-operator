package key

import (
	"testing"

	"github.com/giantswarm/k8smetadata/pkg/label"
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
