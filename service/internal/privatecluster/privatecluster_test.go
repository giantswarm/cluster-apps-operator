package privatecluster

import (
	"context"
	"testing"

	"github.com/giantswarm/micrologger/microloggertest"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/scheme"
	capi "sigs.k8s.io/cluster-api/api/core/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	clientfake "sigs.k8s.io/controller-runtime/pkg/client/fake"

	infra "github.com/giantswarm/cluster-apps-operator/v3/service/internal/infrastructure"
)

func TestIsPrivateCluster(t *testing.T) {
	privateAwsCluster := &unstructured.Unstructured{}
	privateAwsCluster.Object = map[string]interface{}{
		"metadata": map[string]interface{}{
			"name":      "private-aws-cluster",
			"namespace": "default",
			"annotations": map[string]interface{}{
				"aws.giantswarm.io/vpc-mode": "private",
			},
		},
	}
	privateAwsCluster.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "infrastructure.cluster.x-k8s.io",
		Kind:    infra.AWSClusterKind,
		Version: "v1beta2",
	})

	publicAwsCluster := &unstructured.Unstructured{}
	publicAwsCluster.Object = map[string]interface{}{
		"metadata": map[string]interface{}{
			"name":      "public-aws-cluster",
			"namespace": "default",
			"annotations": map[string]interface{}{
				"aws.giantswarm.io/vpc-mode": "public",
			},
		},
	}
	publicAwsCluster.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "infrastructure.cluster.x-k8s.io",
		Version: "v1beta2",
		Kind:    infra.AWSClusterKind,
	})

	privateAwsManagedCluster := &unstructured.Unstructured{}
	privateAwsManagedCluster.Object = map[string]interface{}{
		"metadata": map[string]interface{}{
			"name":      "private-aws-managed-cluster",
			"namespace": "default",
			"annotations": map[string]interface{}{
				"aws.giantswarm.io/vpc-mode": "private",
			},
		},
	}
	privateAwsManagedCluster.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "infrastructure.cluster.x-k8s.io",
		Version: "v1beta2",
		Kind:    infra.AWSManagedClusterKind,
	})

	publicAwsManagedCluster := &unstructured.Unstructured{}
	publicAwsManagedCluster.Object = map[string]interface{}{
		"metadata": map[string]interface{}{
			"name":      "private-aws-managed-cluster",
			"namespace": "default",
			"annotations": map[string]interface{}{
				"aws.giantswarm.io/vpc-mode": "public",
			},
		},
	}
	publicAwsManagedCluster.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "infrastructure.cluster.x-k8s.io",
		Version: "v1beta2",
		Kind:    infra.AWSManagedClusterKind,
	})

	privateAzureCluster := &unstructured.Unstructured{}
	privateAzureCluster.Object = map[string]interface{}{
		"metadata": map[string]interface{}{
			"name":      "private-aws-cluster",
			"namespace": "default",
		},
		"spec": map[string]interface{}{
			"networkSpec": map[string]interface{}{
				"apiServerLB": map[string]interface{}{
					"type": "Internal",
				},
			},
		},
	}
	privateAzureCluster.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "infrastructure.cluster.x-k8s.io",
		Version: "v1beta1",
		Kind:    infra.AzureClusterKind,
	})

	publicAzureCluster := &unstructured.Unstructured{}
	publicAzureCluster.Object = map[string]interface{}{
		"metadata": map[string]interface{}{
			"name":      "public-aws-cluster",
			"namespace": "default",
		},
		"spec": map[string]interface{}{
			"networkSpec": map[string]interface{}{
				"apiServerLB": map[string]interface{}{
					"type": "Public",
				},
			},
		},
	}
	publicAzureCluster.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "infrastructure.cluster.x-k8s.io",
		Version: "v1beta1",
		Kind:    infra.AzureClusterKind,
	})

	gcpCluster := &unstructured.Unstructured{}
	gcpCluster.Object = map[string]interface{}{
		"metadata": map[string]interface{}{
			"name":      "gcp-cluster",
			"namespace": "default",
		},
	}
	gcpCluster.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "infrastructure.cluster.x-k8s.io",
		Version: "v1beta1",
		Kind:    infra.GCPClusterKind,
	})

	gcpManagedCluster := &unstructured.Unstructured{}
	gcpManagedCluster.Object = map[string]interface{}{
		"metadata": map[string]interface{}{
			"name":      "gcp-cluster",
			"namespace": "default",
		},
	}
	gcpManagedCluster.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "infrastructure.cluster.x-k8s.io",
		Version: "v1beta1",
		Kind:    infra.GCPManagedClusterKind,
	})

	tests := []struct {
		name     string
		infraRef *unstructured.Unstructured
		want     bool
		wantErr  bool
	}{
		{
			name:     "AWS Private cluster",
			infraRef: privateAwsCluster,
			want:     true,
			wantErr:  false,
		},
		{
			name:     "AWS NON Private cluster",
			infraRef: publicAwsCluster,
			want:     false,
			wantErr:  false,
		},
		{
			name:     "AWS Private managed cluster",
			infraRef: privateAwsManagedCluster,
			want:     true,
			wantErr:  false,
		},
		{
			name:     "AWS NON Private managed cluster",
			infraRef: publicAwsManagedCluster,
			want:     false,
			wantErr:  false,
		},
		{
			name:     "Azure Private cluster",
			infraRef: privateAzureCluster,
			want:     true,
			wantErr:  false,
		},
		{
			name:     "Azure NON Private cluster",
			infraRef: publicAzureCluster,
			want:     false,
			wantErr:  false,
		},
		{
			name:     "GCP cluster",
			infraRef: gcpCluster,
			want:     false,
			wantErr:  false,
		},
		{
			name:     "GCP managed cluster",
			infraRef: gcpManagedCluster,
			want:     false,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ctrlclient client.Client
			{
				schemeBuilder := runtime.SchemeBuilder{
					capi.AddToScheme,
				}

				err := schemeBuilder.AddToScheme(scheme.Scheme)
				if err != nil {
					t.Fatal(err)
				}

				ctrlclient = clientfake.NewClientBuilder().
					WithRuntimeObjects(tt.infraRef).
					Build()
			}

			cluster := clusterForInfrastructureRef(tt.infraRef)
			got, err := IsPrivateCluster(context.Background(), microloggertest.New(), ctrlclient, cluster)
			if (err != nil) != tt.wantErr {
				t.Errorf("IsPrivateCluster() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("IsPrivateCluster() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func clusterForInfrastructureRef(ref *unstructured.Unstructured) capi.Cluster {
	return capi.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
		},
		Spec: capi.ClusterSpec{
			InfrastructureRef: &v1.ObjectReference{
				Kind:       ref.GetKind(),
				Namespace:  ref.GetNamespace(),
				Name:       ref.GetName(),
				APIVersion: ref.GetAPIVersion(),
			},
		},
	}
}
