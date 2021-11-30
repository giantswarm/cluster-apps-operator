package v1alpha3

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
	NetworkSpec NetworkSpec `json:"network"`
}

type NetworkSpec struct {
	Vnet VnetSpec `json:"vnet"`
}

type VnetSpec struct {
	CIDRBlocks []string `json:"cidrBlocks"`
}

// +kubebuilder:object:root=true
type AzureClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AzureCluster `json:"items"`
}
