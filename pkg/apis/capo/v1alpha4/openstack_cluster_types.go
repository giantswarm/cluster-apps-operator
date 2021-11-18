package v1alpha4

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type OpenStackCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Status OpenStackClusterStatus `json:"status"`
}

func (o OpenStackCluster) DeepCopyObject() runtime.Object {
	panic("implement me")
}

type OpenStackClusterStatus struct {
	Network *Network `json:"network"`
}

type Network struct {
	ID string `json:"id"`
	Subnet *Subnet `json:"subnet"`
}

type Subnet struct {
	ID string `json:"id"`
}
