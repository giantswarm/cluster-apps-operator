package clustersecret

import (
	"context"
	"fmt"

	"github.com/giantswarm/k8smetadata/pkg/label"
	"github.com/giantswarm/microerror"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/cluster-apps-operator/pkg/project"
	"github.com/giantswarm/cluster-apps-operator/service/controller/key"
)

func (r *Resource) GetCurrentState(ctx context.Context, obj interface{}) ([]*corev1.Secret, error) {
	cr, err := key.ToCluster(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var secrets []*corev1.Secret
	{
		r.logger.Debugf(ctx, "finding cluster secrets in namespace %#q", cr.GetNamespace())

		lo := metav1.ListOptions{
			LabelSelector: fmt.Sprintf("%s=%s,%s=%s", label.Cluster, key.ClusterID(&cr), label.ManagedBy, project.Name()),
		}

		list, err := r.k8sClient.K8sClient().CoreV1().Secrets(cr.GetNamespace()).List(ctx, lo)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		for _, item := range list.Items {
			secrets = append(secrets, item.DeepCopy())
		}

		r.logger.Debugf(ctx, "found %d secrets in namespace %#q", len(secrets), cr.GetNamespace())
	}

	return secrets, nil
}
