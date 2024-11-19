package app

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/giantswarm/apiextensions-application/api/v1alpha1"
	"github.com/giantswarm/backoff"
	"github.com/giantswarm/k8smetadata/pkg/label"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/v7/pkg/controller/context/finalizerskeptcontext"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	capi "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/cluster-apps-operator/v3/pkg/project"
	"github.com/giantswarm/cluster-apps-operator/v3/service/controller/key"
)

func (r Resource) EnsureDeleted(ctx context.Context, obj interface{}) error {
	cr, err := key.ToCluster(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	// Get apps for the given cluster, not managed by the cluster-apps-operator.
	// Note: this list may be incomplete depending on the label missing or present
	// by mistake on apps that shouldn't have it.
	apps, err := r.getClusterApps(ctx, cr)
	if err != nil {
		return microerror.Mask(err)
	}

	// For the apps returned in the previous step, let's try to remove them,
	// skipping apps managed by Flux and the ones whose deletion has already
	// been requested.
	err = r.deleteClusterApps(ctx, apps)
	if err != nil {
		r.logger.Errorf(ctx, err, "encountered problem removing apps")
		return r.cancel(ctx)
	}

	// We don't want to initiate deletion of the `app-operator` and `chart-operator`
	// apps in case the apps we requested deletion for are still not gone from the
	// cluster, or we have Flux managed apps.
	apps, err = r.getClusterApps(ctx, cr)
	if err != nil {
		return microerror.Mask(err)
	} else if len(apps) > 0 {
		var appNames []string
		for _, app := range apps {
			appNames = append(appNames, app.Name)
		}
		r.logger.Debugf(ctx, "waiting for %d apps to be deleted for cluster '%s/%s': %s", len(apps), cr.GetNamespace(), key.ClusterID(&cr), strings.Join(appNames, ", "))
		return r.cancel(ctx)
	}

	desiredApps := r.desiredApps(ctx, cr)

	// There are no more app CRs to manage so we can delete chart-operator.
	err = r.waitForAppDeletion(ctx, cr, key.ChartOperatorAppName(&cr), desiredApps)
	if IsNotDeleted(err) {
		r.logger.Debugf(ctx, "%s not deleted yet", key.ChartOperatorAppName(&cr))

		finalizerskeptcontext.SetKept(ctx)
		r.logger.Debugf(ctx, "keeping finalizers")

		r.logger.Debugf(ctx, "canceling resource")
		return nil
	} else if err != nil {
		return microerror.Mask(err)
	}

	// Once chart-operator is deleted we can delete app-operator and remove
	// the finalizer.
	err = r.waitForAppDeletion(ctx, cr, fmt.Sprintf("%s-app-operator", key.ClusterID(&cr)), desiredApps)
	if IsNotDeleted(err) {
		r.logger.Debugf(ctx, "%s not deleted yet", key.AppOperatorAppName(&cr))

		finalizerskeptcontext.SetKept(ctx)
		r.logger.Debugf(ctx, "keeping finalizers")

		r.logger.Debugf(ctx, "canceling resource")
		return nil
	} else if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (r Resource) cancel(ctx context.Context) error {
	finalizerskeptcontext.SetKept(ctx)

	r.logger.Debugf(ctx, "keeping finalizers")
	r.logger.Debugf(ctx, "canceling resource")

	return nil
}

// deleteClusterApps tries to delete given apps, skipping apps with
// Flux managed-by label.
func (r Resource) deleteClusterApps(ctx context.Context, apps []*v1alpha1.App) error {
	for _, app := range apps {
		// No need to delete app whose deletion has already been requested,
		// or when managed by Flux as the app may be recreated in such case.
		// There is one valid case when deleting Flux-managed app makes sense,
		// namely when `prune: false` is used and app is gone in the repository,
		// but there is no way to recognize this case here.
		if key.IsManagedByFlux(*app) {
			r.logger.Debugf(ctx, "skipping Flux-managed '%s/%s' app in deletion", app.Namespace, app.Name)
			continue
		}

		if !app.DeletionTimestamp.IsZero() {
			r.logger.Debugf(ctx, "deletion already requested for '%s/%s' app", app.Namespace, app.Name)
			continue
		}

		r.logger.Debugf(ctx, "requesting deletion of '%s/%s' app", app.Namespace, app.Name)

		err := r.ctrlClient.Delete(ctx, app)
		if apierrors.IsNotFound(err) {
			r.logger.Debugf(ctx, "app '%s/%s' already deleted", app.Namespace, app.Name)
			continue
		} else if err != nil {
			return microerror.Mask(err)
		}

		r.logger.Debugf(ctx, "successfully requested deletion of '%s/%s' app", app.Namespace, app.Name)
	}

	return nil
}

// getClusterApps gets all the App CRs with matching selector, not managed
// by the `cluster-apps-operator`.
func (r Resource) getClusterApps(ctx context.Context, cr capi.Cluster) ([]*v1alpha1.App, error) {
	var apps []*v1alpha1.App

	r.logger.Debugf(ctx, "finding apps for cluster '%s/%s'", cr.GetNamespace(), key.ClusterID(&cr))

	// Get all apps for cluster not being managed by cluster-apps-operator.
	selector, err := labels.Parse(fmt.Sprintf("%s=%s,%s!=%s", label.Cluster, key.ClusterID(&cr), label.ManagedBy, project.Name()))
	if err != nil {
		return nil, microerror.Mask(err)
	}

	o := client.ListOptions{
		Namespace:     cr.GetNamespace(),
		LabelSelector: selector,
	}

	var appList v1alpha1.AppList

	err = r.ctrlClient.List(ctx, &appList, &o)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	for _, item := range appList.Items {
		apps = append(apps, item.DeepCopy())
	}

	r.logger.Debugf(ctx, "found %d app(s) for cluster '%s/%s'", len(apps), cr.GetNamespace(), key.ClusterID(&cr))

	return apps, nil
}

// waitForAppDeletion deletes the app CR and waits for a short period. If the
// deletion takes longer we check again in the next resync period to allow
// processing other clusters.
func (r Resource) waitForAppDeletion(ctx context.Context, cr capi.Cluster, appName string, desiredApps []*v1alpha1.App) error {
	var err error

	app := findAppByName(desiredApps, appName, cr.GetNamespace())
	if app != nil {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("deleting app '%s/%s'", app.Namespace, app.Name))

		err = r.ctrlClient.Delete(ctx, app)
		if apierrors.IsNotFound(err) {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("app '%s/%s' already deleted", app.Namespace, app.Name))
			return nil
		}
	}

	o := func() error {
		err := r.ctrlClient.Get(ctx, client.ObjectKey{
			Namespace: cr.GetNamespace(),
			Name:      appName,
		}, app)
		if apierrors.IsNotFound(err) {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("deleted app '%s/%s'", app.Namespace, app.Name))
			return nil
		}

		return microerror.Maskf(notDeletedError, "'%s/%s' app still persists", app.Namespace, app.Name)
	}
	n := func(err error, t time.Duration) {
		r.logger.Errorf(ctx, err, "retrying in %s", t)
	}

	b := backoff.NewConstant(15*time.Second, 5*time.Second)
	err = backoff.RetryNotify(o, b, n)

	return err
}
