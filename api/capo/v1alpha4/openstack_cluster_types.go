package v1alpha4

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:object:root=true

// OpenStackCluster is the Schema for the openstackclusters API.
type OpenStackCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Status OpenStackClusterStatus `json:"status"`
}

type OpenStackClusterStatus struct {
	Network *Network `json:"network"`
}

type Network struct {
	ID     string  `json:"id"`
	Subnet *Subnet `json:"subnet"`
}

type Subnet struct {
	ID string `json:"id"`
}

// +kubebuilder:object:root=true

// OpenStackClusterList contains a list of OpenStackCluster.
type OpenStackClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []OpenStackCluster `json:"items"`
}
