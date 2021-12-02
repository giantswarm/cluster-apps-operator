package clustersecret

import (
	"context"
	"fmt"

	"github.com/giantswarm/k8smetadata/pkg/annotation"
	"github.com/giantswarm/k8smetadata/pkg/label"
	"github.com/giantswarm/microerror"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apiv1alpha3 "sigs.k8s.io/cluster-api/api/v1alpha4"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"

	capo "github.com/giantswarm/cluster-apps-operator/api/capo/v1alpha4"
	"github.com/giantswarm/cluster-apps-operator/pkg/project"
	"github.com/giantswarm/cluster-apps-operator/service/controller/key"
)

func (r *Resource) GetDesiredState(ctx context.Context, obj interface{}) ([]*corev1.Secret, error) {
	cr, err := key.ToCluster(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var secrets []*corev1.Secret

	if key.IsDeleted(&cr) {
		r.logger.Debugf(ctx, "deleting cluster secrets for cluster '%s/%s'", cr.GetNamespace(), key.ClusterID(&cr))
		return secrets, nil
	}

	values := map[string]interface{}{}

	{
		infrastructureRef := cr.Spec.InfrastructureRef
		if infrastructureRef != nil {
			switch infrastructureRef.Kind {
			case "OpenStackCluster":
				var infraCluster capo.OpenStackCluster
				err = r.k8sClient.CtrlClient().Get(ctx, client.ObjectKey{Namespace: infrastructureRef.Namespace, Name: infrastructureRef.Name}, &infraCluster)
				if err != nil {
					return nil, microerror.Mask(err)
				}

				values["cloudConfig"], err = r.generateOpenStackCloudConfig(ctx, infraCluster)
				if err != nil {
					return nil, microerror.Mask(err)
				}
			}
		}
	}

	secretSpecs := []secretSpec{
		{
			Name:      key.ClusterValuesResourceName(&cr),
			Namespace: key.ClusterID(&cr),
			Values:    values,
		},
	}

	for _, spec := range secretSpecs {
		secret, err := newSecret(cr, spec)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		secrets = append(secrets, secret)
	}

	return secrets, nil
}

func newSecret(cr apiv1alpha3.Cluster, secretSpec secretSpec) (*corev1.Secret, error) {
	yamlValues, err := yaml.Marshal(secretSpec.Values)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	cm := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretSpec.Name,
			Namespace: secretSpec.Namespace,
			Annotations: map[string]string{
				annotation.Notes: fmt.Sprintf("DO NOT EDIT. Values managed by %s.", project.Name()),
			},
			Labels: map[string]string{
				label.Cluster:   key.ClusterID(&cr),
				label.ManagedBy: project.Name(),
			},
		},
		Data: map[string][]byte{
			"values": yamlValues,
		},
	}

	return cm, nil
}
