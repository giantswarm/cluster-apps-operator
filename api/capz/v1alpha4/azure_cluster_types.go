package v1alpha4

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:object:root=true
type AzureCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`

	Spec AzureClusterSpec `json:"spec"`
}

type AzureClusterSpec struct {
	NetworkSpec NetworkSpec `json:"networkSpec,omitempty"`
}

type NetworkSpec struct {
	Vnet VnetSpec `json:"vnet,omitempty"`
}

type VnetSpec struct {
	// +optional
	CIDRBlocks []string `json:"cidrBlocks,omitempty"`
}

// +kubebuilder:object:root=true
type AzureClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AzureCluster `json:"items"`
}
