// +kubebuilder:object:generate=true
package v1alpha4

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const Group = "infrastructure.cluster.x-k8s.io"
const Version = "v1alpha4"

var (
	SchemeGroupVersion = schema.GroupVersion{Group: Group, Version: Version}
	// SchemeBuilder is the scheme builder with scheme init functions to run for this API package
	SchemeBuilder = runtime.NewSchemeBuilder(addKnownTypes)
	// AddToScheme is a common registration function for mapping packaged scoped group & version keys to a scheme
	AddToScheme = SchemeBuilder.AddToScheme
)

// Adds the list of known types to the given scheme.
func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(SchemeGroupVersion,
		&AzureCluster{},
		&AzureClusterList{},
	)
	metav1.AddToGroupVersion(scheme, SchemeGroupVersion)
	return nil
}
