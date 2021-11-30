package app

import (
	"context"
	"fmt"

	"github.com/giantswarm/apiextensions-application/api/v1alpha1"
	"github.com/giantswarm/k8smetadata/pkg/label"
	"github.com/giantswarm/microerror"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/cluster-apps-operator/pkg/project"
	"github.com/giantswarm/cluster-apps-operator/service/controller/key"
)

func (r *Resource) GetCurrentState(ctx context.Context, obj interface{}) ([]*v1alpha1.App, error) {
	cr, err := key.ToCluster(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var apps []*v1alpha1.App
	{
		r.logger.Debugf(ctx, "finding apps for cluster '%s/%s'", cr.GetNamespace(), key.ClusterID(&cr))

		selector, err := labels.Parse(fmt.Sprintf("%s=%s", label.ManagedBy, project.Name()))
		if err != nil {
			return nil, microerror.Mask(err)
		}

		o := client.ListOptions{
			Namespace:     key.ClusterID(&cr),
			LabelSelector: selector,
		}

		var appList v1alpha1.AppList
		err = r.g8sClient.List(ctx, &appList, &o)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		for _, item := range appList.Items {
			apps = append(apps, item.DeepCopy())
		}

		r.logger.Debugf(ctx, "found %d apps for cluster '%s/%s'", len(apps), cr.GetNamespace(), key.ClusterID(&cr))
	}

	return apps, nil
}
