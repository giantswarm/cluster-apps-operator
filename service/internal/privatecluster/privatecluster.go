package privatecluster

import (
	"context"

	"github.com/giantswarm/k8smetadata/pkg/annotation"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	capi "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	infra "github.com/giantswarm/cluster-apps-operator/v2/service/internal/infrastructure"
)

func IsPrivateCluster(ctx context.Context, logger micrologger.Logger, ctrlclient client.Client, cr capi.Cluster) (bool, error) {
	var privateCluster bool
	var err error

	infrastructureRef := cr.Spec.InfrastructureRef
	if infrastructureRef != nil {
		switch infrastructureRef.Kind {
		case infra.AzureClusterKind, infra.AzureManagedClusterKind:
			capzCluster := &unstructured.Unstructured{}
			capzCluster.SetGroupVersionKind(schema.GroupVersionKind{
				Group:   infrastructureRef.GroupVersionKind().Group,
				Kind:    infrastructureRef.Kind,
				Version: infrastructureRef.GroupVersionKind().Version,
			})
			err = ctrlclient.Get(ctx, client.ObjectKey{
				Namespace: cr.Namespace,
				Name:      infrastructureRef.Name,
			}, capzCluster)
			if err != nil {
				return false, microerror.Mask(err)
			}

			apiServerLbType, apiServerLbFound, err := unstructured.NestedString(capzCluster.Object, []string{"spec", "networkSpec", "apiServerLB", "type"}...)
			if err != nil || !apiServerLbFound {
				return false, microerror.Mask(fieldNotFoundOnInfrastructureTypeError)
			}

			privateCluster = apiServerLbType == "Internal"
		case infra.AWSClusterKind, infra.AWSManagedClusterKind:
			awsCluster := &unstructured.Unstructured{}
			awsCluster.SetGroupVersionKind(schema.GroupVersionKind{
				Group:   infrastructureRef.GroupVersionKind().Group,
				Kind:    infrastructureRef.Kind,
				Version: infrastructureRef.GroupVersionKind().Version,
			})
			err = ctrlclient.Get(ctx, client.ObjectKey{
				Namespace: cr.Namespace,
				Name:      infrastructureRef.Name,
			}, awsCluster)
			if err != nil {
				return false, microerror.Mask(err)
			}

			annotationValue, annotationFound, err := unstructured.NestedString(awsCluster.Object, []string{"metadata", "annotations", annotation.AWSVPCMode}...)
			if err != nil || !annotationFound {
				return false, microerror.Mask(fieldNotFoundOnInfrastructureTypeError)
			}

			privateCluster = annotationValue == annotation.AWSVPCModePrivate
		default:
			logger.Debugf(ctx, "privatecluster.IsPrivateCluster in not implemented for infrastructure kind %q", infrastructureRef.Kind)
		}
	}

	return privateCluster, nil
}
