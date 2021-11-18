package v1alpha3

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type AzureCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`

	Spec AzureClusterSpec `json:"spec"`
}

func (a AzureCluster) DeepCopyObject() runtime.Object {
	panic("implement me")
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
