package key

import (
	"reflect"
	"testing"

	apiCoreV1 "k8s.io/api/core/v1"

	"github.com/giantswarm/apiextensions-application/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/k8smetadata/pkg/label"
	"github.com/giantswarm/microerror"
	capi "sigs.k8s.io/cluster-api/api/v1beta1"
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

func Test_IsClusterInTransition(t *testing.T) {
	testCases := []struct {
		description string
		input       *capi.Cluster
		expected    bool
	}{
		{
			"Case 1: Immediately after the cluster resource is created",
			&capi.Cluster{
				Status: capi.ClusterStatus{
					ObservedGeneration: 1,
					Phase:              "Provisioning",
				},
			},
			true,
		},
		{
			"Case 2: First status fields added, control plane is initializing",
			&capi.Cluster{
				Status: capi.ClusterStatus{
					ObservedGeneration: 1,
					Phase:              "Provisioning",
					Conditions: capi.Conditions{
						capi.Condition{
							Type:     "InfrastructureReady",
							Status:   apiCoreV1.ConditionFalse,
							Severity: capi.ConditionSeverityInfo,
							Reason:   "NatGatewaysCreationStarted",
							Message:  "3 of 8 completed",
						},
						capi.Condition{
							Type:     "ControlPlaneInitialized",
							Status:   apiCoreV1.ConditionFalse,
							Severity: capi.ConditionSeverityInfo,
							Reason:   "WaitingForControlPlaneProviderInitialized",
							Message:  "Waiting for control plane provider to indicate the control plane has been initialized",
						},
						capi.Condition{
							Type:     "ControlPlaneReady",
							Status:   apiCoreV1.ConditionFalse,
							Severity: capi.ConditionSeverityWarning,
							Reason:   "ScalingUp",
							Message:  "Scaling up control plane to 3 replicas (actual 0)",
						},
						capi.Condition{
							Type:     "Ready",
							Status:   apiCoreV1.ConditionFalse,
							Severity: capi.ConditionSeverityWarning,
							Reason:   "ScalingUp",
							Message:  "Scaling up control plane to 3 replicas (actual 0)",
						},
					},
				},
			},
			true,
		},
		{
			"Case 3: Provisioned and infrastructure ready is reported, but still in progress",
			&capi.Cluster{
				Status: capi.ClusterStatus{
					ObservedGeneration:  2,
					Phase:               "Provisioned",
					InfrastructureReady: true,
					Conditions: capi.Conditions{
						capi.Condition{
							Type:   "Provisioned",
							Status: apiCoreV1.ConditionTrue,
						},
						capi.Condition{
							Type:     "ControlPlaneInitialized",
							Status:   apiCoreV1.ConditionFalse,
							Severity: capi.ConditionSeverityInfo,
							Reason:   "WaitingForControlPlaneProviderInitialized",
							Message:  "Waiting for control plane provider to indicate the control plane has been initialized",
						},
						capi.Condition{
							Type:     "Ready",
							Status:   apiCoreV1.ConditionFalse,
							Severity: capi.ConditionSeverityWarning,
							Reason:   "ScalingUp",
							Message:  "Scaling up control plane to 3 replicas (actual 1)",
						},
						capi.Condition{
							Type:     "ControlPlaneReady",
							Status:   apiCoreV1.ConditionFalse,
							Severity: capi.ConditionSeverityWarning,
							Reason:   "ScalingUp",
							Message:  "Scaling up control plane to 3 replicas (actual 1)",
						},
					},
				},
			},
			true,
		},
		{
			"Case 4: Control plane is ready as well, but cluster is not fully ready",
			&capi.Cluster{
				Status: capi.ClusterStatus{
					ObservedGeneration:  2,
					Phase:               "Provisioned",
					InfrastructureReady: true,
					ControlPlaneReady:   true,
					Conditions: capi.Conditions{
						capi.Condition{
							Type:   "Provisioned",
							Status: apiCoreV1.ConditionTrue,
						},
						capi.Condition{
							Type:   "ControlPlaneInitialized",
							Status: apiCoreV1.ConditionTrue,
						},
						capi.Condition{
							Type:   "ControlPlaneReady",
							Status: apiCoreV1.ConditionTrue,
						},
						capi.Condition{
							Type:     "Ready",
							Status:   apiCoreV1.ConditionFalse,
							Severity: capi.ConditionSeverityWarning,
							Reason:   "...",
							Message:  "...",
						},
					},
				},
			},
			true,
		},
		{
			"Case 5: Cluster ready condition is set to true",
			&capi.Cluster{
				Status: capi.ClusterStatus{
					ObservedGeneration:  2,
					Phase:               "Provisioned",
					InfrastructureReady: true,
					ControlPlaneReady:   true,
					Conditions: capi.Conditions{
						capi.Condition{
							Type:   "Provisioned",
							Status: apiCoreV1.ConditionTrue,
						},
						capi.Condition{
							Type:   "ControlPlaneInitialized",
							Status: apiCoreV1.ConditionTrue,
						},
						capi.Condition{
							Type:   "ControlPlaneReady",
							Status: apiCoreV1.ConditionTrue,
						},
						capi.Condition{
							Type:   "Ready",
							Status: apiCoreV1.ConditionTrue,
						},
					},
				},
			},
			false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			result := IsClusterInTransition(*tc.input)

			if result != tc.expected {
				t.Fatalf("Got the unexpected result for is cluster in transition check for: %#v", tc.input)
			}
		})
	}
}
