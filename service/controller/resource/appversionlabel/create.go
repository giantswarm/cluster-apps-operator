package appversionlabel

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/giantswarm/apiextensions-application/api/v1alpha1"
	"github.com/giantswarm/k8smetadata/pkg/label"
	"github.com/giantswarm/microerror"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/cluster-apps-operator/pkg/project"
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

// EnsureCreated checks for optional apps and ensures the app-operator version
// label has the correct value.
func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	cr, err := key.ToCluster(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	var apps []*v1alpha1.App
	{
		r.logger.Debugf(ctx, "finding optional apps for cluster '%s/%s'", cr.GetNamespace(), key.ClusterID(&cr))

		labelSelector, err := labels.Parse(fmt.Sprintf("%s!=%s", label.ManagedBy, project.Name()))
		if err != nil {
			return microerror.Mask(err)
		}

		o := client.ListOptions{
			Namespace: key.ClusterID(&cr),
			LabelSelector: labelSelector,
		}
		var appList v1alpha1.AppList
		err = r.g8sClient.List(ctx, &appList, &o)
		if err != nil {
			return microerror.Mask(err)
		}

		for _, item := range appList.Items {
			apps = append(apps, item.DeepCopy())
		}

		r.logger.Debugf(ctx, "found %d optional apps for cluster '%s/%s'", len(apps), cr.GetNamespace(), key.ClusterID(&cr))

		if len(apps) == 0 {
			// Return early as there is nothing to do.
			return nil
		}
	}

	{
		var updatedAppCount int

		componentVersions, err := r.releaseVersion.ComponentVersion(ctx, &cr)
		if err != nil {
			return microerror.Mask(err)
		}

		appOperatorComponent := componentVersions[releaseversion.AppOperator]
		appOperatorVersion := appOperatorComponent.Version
		if appOperatorVersion == "" {
			return microerror.Maskf(notFoundError, "app-operator component version not found")
		}

		r.logger.Debugf(ctx, "updating version label for optional apps in cluster '%s/%s'", cr.GetNamespace(), key.ClusterID(&cr))

		for _, app := range apps {
			currentVersion := app.Labels[label.AppOperatorVersion]

			if currentVersion != appOperatorVersion {
				patches := []patch{}

				if len(app.Labels) == 0 {
					patches = append(patches, patch{
						Op:    "add",
						Path:  "/metadata/labels",
						Value: map[string]string{},
					})
				}

				patches = append(patches, patch{
					Op:    "add",
					Path:  fmt.Sprintf("/metadata/labels/%s", replaceToEscape(label.AppOperatorVersion)),
					Value: appOperatorVersion,
				})

				bytes, err := json.Marshal(patches)
				if err != nil {
					return microerror.Mask(err)
				}

				err = r.g8sClient.Patch(ctx, app, appPatch{
					data: bytes,
				})
				if err != nil {
					return microerror.Mask(err)
				}

				updatedAppCount++
			}
		}

		if updatedAppCount > 0 {
			r.logger.Debugf(ctx, "updating version label for %d optional apps in cluster '%s/%s'", updatedAppCount, cr.GetNamespace(), key.ClusterID(&cr))
		} else {
			r.logger.Debugf(ctx, "no version labels to update for cluster '%s/%s'", cr.GetNamespace(), key.ClusterID(&cr))
		}
	}

	return nil
}
