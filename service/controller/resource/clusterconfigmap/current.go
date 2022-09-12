package clusterconfigmap

import (
	"context"
	"fmt"

	"github.com/giantswarm/k8smetadata/pkg/label"
	"github.com/giantswarm/microerror"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/cluster-apps-operator/v2/pkg/project"
	"github.com/giantswarm/cluster-apps-operator/v2/service/controller/key"
)

func (r *Resource) GetCurrentState(ctx context.Context, obj interface{}) ([]*corev1.ConfigMap, error) {
	cr, err := key.ToCluster(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var configMaps []*corev1.ConfigMap
	{
		r.logger.Debugf(ctx, "finding cluster configmaps in namespace %#q", cr.GetNamespace())

		lo := metav1.ListOptions{
			LabelSelector: fmt.Sprintf("%s=%s,%s=%s", label.Cluster, key.ClusterID(&cr), label.ManagedBy, project.Name()),
		}

		list, err := r.k8sClient.K8sClient().CoreV1().ConfigMaps(cr.GetNamespace()).List(ctx, lo)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		for _, item := range list.Items {
			configMaps = append(configMaps, item.DeepCopy())
		}

		r.logger.Debugf(ctx, "found %d configmaps in namespace %#q", len(configMaps), cr.GetNamespace())
	}

	return configMaps, nil
}
