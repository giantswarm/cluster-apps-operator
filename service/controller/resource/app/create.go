package app

import (
	"context"
	"fmt"
	"reflect"
	"strconv"

	"github.com/giantswarm/apiextensions-application/api/v1alpha1"
	"github.com/giantswarm/k8smetadata/pkg/annotation"
	"github.com/giantswarm/k8smetadata/pkg/label"
	"github.com/giantswarm/microerror"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	capi "sigs.k8s.io/cluster-api/api/v1alpha4"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/cluster-apps-operator/v2/pkg/project"
	"github.com/giantswarm/cluster-apps-operator/v2/service/controller/key"
)

func (r Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	cr, err := key.ToCluster(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	currentApps, err := r.currentApps(ctx, cr)
	if err != nil {
		return microerror.Mask(err)
	}

	for _, app := range r.desiredApps(ctx, cr) {
		currentApp := findAppByName(currentApps, app.Name, app.Namespace)

		if currentApp == nil {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("creating app '%s/%s'", app.Namespace, app.Name))

			err = r.ctrlClient.Create(ctx, app)
			if apierrors.IsAlreadyExists(err) {
				r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("already created app '%s/%s'", app.Namespace, app.Name))
			} else if err != nil {
				return microerror.Mask(err)
			}

			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("created app '%s/%s'", app.Namespace, app.Name))
		} else if hasAppChanged(currentApp, app) {
			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("updating app '%s/%s'", app.Namespace, app.Name))

			// Get app CR again to ensure the resource version is correct.
			var currentApp v1alpha1.App

			err = r.ctrlClient.Get(
				ctx,
				types.NamespacedName{Name: app.Name, Namespace: app.Namespace},
				&currentApp,
			)
			if err != nil {
				return microerror.Mask(err)
			}

			modifiedApp := currentApp.DeepCopy()
			modifiedApp.Annotations = app.Annotations
			modifiedApp.Labels = app.Labels
			modifiedApp.Spec = app.Spec

			err = r.ctrlClient.Patch(ctx, modifiedApp, client.MergeFrom(&currentApp))
			if err != nil {
				return microerror.Mask(err)
			}

			r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("updated app '%s/%s'", app.Namespace, app.Name))
		}
	}

	return nil
}

func (r *Resource) currentApps(ctx context.Context, cr capi.Cluster) ([]*v1alpha1.App, error) {
	var apps []*v1alpha1.App
	{
		r.logger.Debugf(ctx, "finding apps for cluster '%s/%s'", cr.GetNamespace(), key.ClusterID(&cr))

		selector, err := labels.Parse(fmt.Sprintf("%s=%s,%s=%s", label.Cluster, key.ClusterID(&cr), label.ManagedBy, project.Name()))
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

		r.logger.Debugf(ctx, "found %d apps for cluster '%s/%s'", len(apps), cr.GetNamespace(), key.ClusterID(&cr))
	}

	return apps, nil
}

func (r *Resource) desiredApps(ctx context.Context, cr capi.Cluster) []*v1alpha1.App {
	appSpecs := []AppSpec{
		{
			App: "app-operator",
			// app-operator is deployed by the management cluster
			// instance.
			AppOperatorVersion: uniqueOperatorVersion,
			AppName:            key.AppOperatorAppName(&cr),
			Catalog:            r.appOperatorCatalog,
			ConfigMapName:      key.AppOperatorValuesResourceName(&cr),
			ConfigMapNamespace: cr.GetNamespace(),
			InCluster:          true,
			TargetNamespace:    cr.GetNamespace(),
			UseUpgradeForce:    false,
			Version:            r.appOperatorVersion,
		},
		{
			App: "chart-operator",
			// chart-operator is deployed by the workload cluster
			// instance.
			AppOperatorVersion: r.appOperatorVersion,
			AppName:            key.ChartOperatorAppName(&cr),
			Catalog:            r.chartOperatorCatalog,
			ConfigMapName:      key.ClusterValuesResourceName(&cr),
			ConfigMapNamespace: cr.GetNamespace(),
			InCluster:          false,
			TargetNamespace:    "giantswarm",
			SecretName:         key.ClusterValuesResourceName(&cr),
			SecretNamespace:    cr.GetNamespace(),
			UseUpgradeForce:    false,
			Version:            r.chartOperatorVersion,
		},
	}

	apps := []*v1alpha1.App{}

	for _, spec := range appSpecs {
		apps = append(apps, r.newApp(ctx, cr, spec))
	}

	return apps
}

func (r *Resource) newApp(ctx context.Context, cr capi.Cluster, appSpec AppSpec) *v1alpha1.App {
	var kubeConfig v1alpha1.AppSpecKubeConfig

	if appSpec.InCluster || key.IsBundle(appSpec.App) {
		kubeConfig = v1alpha1.AppSpecKubeConfig{
			InCluster: true,
		}
	} else {
		kubeConfig = v1alpha1.AppSpecKubeConfig{
			Context: v1alpha1.AppSpecKubeConfigContext{
				Name: key.KubeConfigSecretName(&cr),
			},
			Secret: v1alpha1.AppSpecKubeConfigSecret{
				Name:      key.KubeConfigSecretName(&cr),
				Namespace: cr.GetNamespace(),
			},
		}
	}

	appName := appSpec.AppName
	appNamespace := appSpec.TargetNamespace
	// If the app is a bundle, we ensure the MC app operator deploys the apps
	// so the cluster-operator for the wc deploys the apps to the WC.
	appOperatorVersion := appSpec.AppOperatorVersion
	if key.IsBundle(appSpec.App) {
		appName = fmt.Sprintf("%s-%s", key.ClusterID(&cr), appName)
		appOperatorVersion = uniqueOperatorVersion
		appNamespace = cr.GetNamespace()
	}

	return &v1alpha1.App{
		TypeMeta: metav1.TypeMeta{
			Kind:       "App",
			APIVersion: "application.giantswarm.io",
		},
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{
				annotation.ChartOperatorForceHelmUpgrade: strconv.FormatBool(appSpec.UseUpgradeForce),
			},
			Labels: map[string]string{
				label.AppKubernetesName:  appSpec.App,
				label.AppOperatorVersion: appOperatorVersion,
				label.Cluster:            key.ClusterID(&cr),
				label.ManagedBy:          project.Name(),
			},
			Name:      appName,
			Namespace: cr.GetNamespace(),
		},
		Spec: v1alpha1.AppSpec{
			Catalog: appSpec.Catalog,
			Config: v1alpha1.AppSpecConfig{
				ConfigMap: v1alpha1.AppSpecConfigConfigMap{
					Name:      appSpec.ConfigMapName,
					Namespace: appSpec.ConfigMapNamespace,
				},
				Secret: v1alpha1.AppSpecConfigSecret{
					Name:      appSpec.SecretName,
					Namespace: appSpec.SecretNamespace,
				},
			},
			Name:       appSpec.App,
			Namespace:  appNamespace,
			Version:    appSpec.Version,
			KubeConfig: kubeConfig,
		},
	}
}

func hasAppChanged(current, desired *v1alpha1.App) bool {
	if current == nil || desired == nil {
		return false
	}
	if !reflect.DeepEqual(current.Spec, desired.Spec) {
		return true
	}
	if !reflect.DeepEqual(current.Annotations, desired.Annotations) {
		return true
	}
	if !reflect.DeepEqual(current.Labels, desired.Labels) {
		return true
	}

	return false
}
