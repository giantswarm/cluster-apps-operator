package clusternamespace

import (
	"context"

	"github.com/giantswarm/k8smetadata/pkg/label"
	"github.com/giantswarm/microerror"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/cluster-apps-operator/pkg/project"
	"github.com/giantswarm/cluster-apps-operator/service/controller/key"
)

// EnsureCreated creates the namespace for storing app related resources if it
// does not exist.
func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	cr, err := key.ToCluster(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: key.ClusterID(&cr),
			Labels: map[string]string{
				label.Cluster:   key.ClusterID(&cr),
				label.ManagedBy: project.Name(),
			},
		},
	}

	r.logger.Debugf(ctx, "creating namespace %#q for cluster '%s/%s'", ns.Name, cr.GetNamespace(), key.ClusterID(&cr))

	_, err = r.k8sClient.CoreV1().Namespaces().Create(ctx, ns, metav1.CreateOptions{})
	if apierrors.IsAlreadyExists(err) {
		r.logger.Debugf(ctx, "namespace %#q already exists for cluster '%s/%s'", ns.Name, cr.GetNamespace(), key.ClusterID(&cr))
		return nil
	} else if err != nil {
		return microerror.Mask(err)
	}

	r.logger.Debugf(ctx, "created namespace %#q for cluster '%s/%s'", ns.Name, cr.GetNamespace(), key.ClusterID(&cr))

	return nil
}
