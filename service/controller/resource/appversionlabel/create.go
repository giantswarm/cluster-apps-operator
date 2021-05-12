package appversionlabel

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/giantswarm/apiextensions/v3/pkg/apis/application/v1alpha1"
	"github.com/giantswarm/apiextensions/v3/pkg/label"
	"github.com/giantswarm/microerror"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/giantswarm/cluster-apps-operator/pkg/project"
	"github.com/giantswarm/cluster-apps-operator/service/controller/key"
	"github.com/giantswarm/cluster-apps-operator/service/internal/releaseversion"
)

// EnsureCreated checks for optional apps and ensures the app-operator version
// label has the correct value.
func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	cr, err := key.ToCluster(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	var apps []*v1alpha1.App
	{
		r.logger.Debugf(ctx, "finding optional apps for workload cluster %#q", key.ClusterID(&cr))

		o := metav1.ListOptions{
			LabelSelector: fmt.Sprintf("%s!=%s", label.ManagedBy, project.Name()),
		}
		list, err := r.g8sClient.ApplicationV1alpha1().Apps(key.ClusterID(&cr)).List(ctx, o)
		if err != nil {
			return microerror.Mask(err)
		}

		for _, item := range list.Items {
			apps = append(apps, item.DeepCopy())
		}

		r.logger.Debugf(ctx, "found %d optional apps for workload cluster %#q", len(apps), key.ClusterID(&cr))

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

		r.logger.Debugf(ctx, "updating version label for optional apps in workload cluster %#q", key.ClusterID(&cr))

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

				_, err = r.g8sClient.ApplicationV1alpha1().Apps(app.Namespace).Patch(ctx, app.Name, types.JSONPatchType, bytes, metav1.PatchOptions{})
				if err != nil {
					return microerror.Mask(err)
				}

				updatedAppCount++
			}
		}

		if updatedAppCount > 0 {
			r.logger.Debugf(ctx, "updating version label for %d optional apps in workload cluster %#q", updatedAppCount, key.ClusterID(&cr))
		} else {
			r.logger.Debugf(ctx, "no version labels to update for workload cluster %#q", key.ClusterID(&cr))
		}
	}

	return nil
}
