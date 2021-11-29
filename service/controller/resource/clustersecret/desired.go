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
	"sigs.k8s.io/yaml"

	"github.com/giantswarm/cluster-apps-operator/pkg/project"
	"github.com/giantswarm/cluster-apps-operator/service/controller/key"
)

func (r *Resource) GetDesiredState(ctx context.Context, obj interface{}) ([]*corev1.Secret, error) {
	if !r.enabled {
		return nil, nil
	}

	cr, err := key.ToCluster(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var secrets []*corev1.Secret

	if key.IsDeleted(&cr) {
		r.logger.Debugf(ctx, "deleting cluster secrets for cluster '%s/%s'", cr.GetNamespace(), key.ClusterID(&cr))
		return secrets, nil
	}

	secretSpecs := []secretSpec{
		{
			Name:      key.ClusterValuesResourceName(&cr),
			Namespace: key.ClusterID(&cr),
			Values:    map[string]interface{}{},
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
