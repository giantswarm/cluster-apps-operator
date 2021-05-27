package clusternamespace

import (
	"context"

	"github.com/giantswarm/k8smetadata/pkg/label"
	"github.com/giantswarm/microerror"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/cluster-apps-operator/pkg/project"
	"github.com/giantswarm/cluster-apps-operator/service/controller/key"
)

// EnsureDeleted deletes the namespace that stores app related resources if it
// is managed by cluster-apps-operator.
func (r *Resource) EnsureDeleted(ctx context.Context, obj interface{}) error {
	cr, err := key.ToCluster(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	r.logger.Debugf(ctx, "deleting namespace %#q for cluster '%s/%s'", key.ClusterID(&cr), cr.GetNamespace(), key.ClusterID(&cr))

	ns, err := r.k8sClient.CoreV1().Namespaces().Get(ctx, key.ClusterID(&cr), metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		r.logger.Debugf(ctx, "already deleted namespace %#q for cluster '%s/%s'", key.ClusterID(&cr), cr.GetNamespace(), key.ClusterID(&cr))
		return nil
	} else if err != nil {
		return microerror.Mask(err)
	}

	val, ok := ns.Labels[label.ManagedBy]
	if !ok || val != project.Name() {
		r.logger.Debugf(ctx, "namespace %#q not managed by %#q", ns.Name, project.Name())
		return nil
	}

	err = r.k8sClient.CoreV1().Namespaces().Delete(ctx, key.ClusterID(&cr), metav1.DeleteOptions{})
	if apierrors.IsNotFound(err) {
		// Fall through.
	} else if err != nil {
		return microerror.Mask(err)
	}

	r.logger.Debugf(ctx, "deleted namespace %#q for cluster '%s/%s'", ns.Name, cr.GetNamespace(), key.ClusterID(&cr))

	return nil
}
