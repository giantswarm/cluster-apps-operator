package v1alpha4

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:object:root=true

// OpenStackCluster is the Schema for the openstackclusters API.
type OpenStackCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   OpenStackClusterSpec   `json:"spec"`
	Status OpenStackClusterStatus `json:"status"`
}

type OpenStackClusterSpec struct {
	IdentityRef *v1.ObjectReference `json:"identityRef,omitempty"`
}

type OpenStackClusterStatus struct {
	Network         *Network `json:"network,omitempty"`
	ExternalNetwork *Network `json:"externalNetwork,omitempty"`
}

type Network struct {
	Name   string  `json:"name"`
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
