package appfinalizer

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/giantswarm/apiextensions-application/api/v1alpha1"
	"github.com/giantswarm/backoff"
	"github.com/giantswarm/k8smetadata/pkg/label"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/v6/pkg/controller/context/finalizerskeptcontext"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/cluster-apps-operator/service/controller/key"
	"github.com/giantswarm/cluster-apps-operator/service/internal/releaseversion"
)

type appPatch struct {
	data []byte
}

func (a appPatch) Type() types.PatchType {
	return types.JSONPatchType
}

func (a appPatch) Data(client.Object) ([]byte, error) {
	return a.data, nil
}

// EnsureDeleted removes finalizers for workload cluster app CRs. These
// resources are deleted with the cluster by the provider operator.
func (r *Resource) EnsureDeleted(ctx context.Context, obj interface{}) error {
	cr, err := key.ToCluster(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	// Make sure there is no running app-operator in the cluster namespace.
	o := func() error {
		name := fmt.Sprintf("%s-%s", releaseversion.AppOperator, key.ClusterID(&cr))
		_, err = r.k8sClient.AppsV1().Deployments(key.ClusterID(&cr)).Get(ctx, name, metav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			// no-op
			return nil
		} else if err != nil {
			return microerror.Mask(err)
		}

		return microerror.Maskf(executionFailedError, "app operator %#q still persists", name)
	}

	b := backoff.NewExponential(3*time.Minute, 10*time.Second)
	err = backoff.Retry(o, b)

	r.logger.Debugf(ctx, "finding apps to remove finalizers for")

	// We keep the finalizer for the app-operator app CR so the resources in
	// the management cluster are deleted.
	selector, err := labels.Parse(fmt.Sprintf("%s!=%s", label.AppKubernetesName, "app-operator"))
	if err != nil {
		return microerror.Mask(err)
	}

	listOptions := client.ListOptions{
		Namespace:     key.ClusterID(&cr),
		LabelSelector: selector,
	}

	var appList v1alpha1.AppList
	err = r.g8sClient.List(ctx, &appList, &listOptions)
	if err != nil {
		return microerror.Mask(err)
	}

	r.logger.Debugf(ctx, "found %d apps to remove finalizers for", len(appList.Items))

	var skipAppCount int

	for _, app := range appList.Items {
		if app.DeletionTimestamp == nil {
			r.logger.Debugf(ctx, "skipping removal of finalizer for app %#q as it is not deleted", app.Name)
			skipAppCount++

			continue
		}

		r.logger.Debugf(ctx, "removing finalizer for app %#q", app.Name)

		index := getFinalizerIndex(app.Finalizers)
		if index >= 0 {
			patches := []patch{
				{
					Op:   "remove",
					Path: fmt.Sprintf("/metadata/finalizers/%d", index),
				},
			}
			bytes, err := json.Marshal(patches)
			if err != nil {
				return microerror.Mask(err)
			}

			err = r.g8sClient.Patch(ctx, app.DeepCopy(), appPatch{
				data: bytes,
			})
			if err != nil {
				return microerror.Mask(err)
			}

			r.logger.Debugf(ctx, "removed finalizer for app %#q", app.Name)
		} else {
			r.logger.Debugf(ctx, "finalizer already removed for app %#q", app.Name)
		}
	}

	// If we skipped any app CRs we need to keep the cluster CR finalizer.
	// So we retry in the next loop.
	if skipAppCount > 0 {
		r.logger.Debugf(ctx, "%d app CRs have not been deleted yet", skipAppCount)
		r.logger.Debugf(ctx, "keeping finalizers")
		finalizerskeptcontext.SetKept(ctx)
	}

	return nil
}

func getFinalizerIndex(finalizers []string) int {
	for i, f := range finalizers {
		if f == "operatorkit.giantswarm.io/app-operator-app" {
			return i
		}
	}

	return -1
}
