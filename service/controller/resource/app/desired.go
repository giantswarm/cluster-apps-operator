package app

import (
	"context"
	"fmt"
	"strconv"

	appv1alpha1 "github.com/giantswarm/apiextensions-application/api/v1alpha1"
	"github.com/giantswarm/k8smetadata/pkg/annotation"
	"github.com/giantswarm/k8smetadata/pkg/label"
	"github.com/giantswarm/microerror"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	capi "sigs.k8s.io/cluster-api/api/v1alpha4"

	"github.com/giantswarm/cluster-apps-operator/pkg/project"
	"github.com/giantswarm/cluster-apps-operator/service/controller/key"
	"github.com/giantswarm/cluster-apps-operator/service/internal/releaseversion"
)

func (r *Resource) GetDesiredState(ctx context.Context, obj interface{}) ([]*appv1alpha1.App, error) {
	cr, err := key.ToCluster(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var apps []*appv1alpha1.App

	if key.IsDeleted(&cr) {
		r.logger.Debugf(ctx, "deleting apps for cluster '%s/%s'", cr.GetNamespace(), key.ClusterID(&cr))
		return apps, nil
	}

	componentVersions, err := r.releaseVersion.ComponentVersion(ctx, &cr)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	appOperatorComponent := componentVersions[releaseversion.AppOperator]
	appOperatorVersion := appOperatorComponent.Version
	if appOperatorVersion == "" {
		return nil, microerror.Maskf(notFoundError, "%#q component version not found", releaseversion.AppOperator)
	}

	// Define app CR for app-operator in the management cluster namespace.
	appOperatorAppSpec := newAppOperatorAppSpec(cr, appOperatorComponent)
	apps = append(apps, r.newApp(uniqueOperatorVersion, cr, appOperatorAppSpec, appv1alpha1.AppSpecUserConfig{}))

	return apps, nil
}

func (r *Resource) newApp(appOperatorVersion string, cr capi.Cluster, appSpec key.AppSpec, userConfig appv1alpha1.AppSpecUserConfig) *appv1alpha1.App {
	configMapName := key.ClusterValuesResourceName(&cr)
	secretName := key.ClusterValuesResourceName(&cr)

	// Override config map name when specified.
	if appSpec.ConfigMapName != "" {
		configMapName = appSpec.ConfigMapName
	}

	// Override secret name when specified.
	if appSpec.SecretName != "" {
		secretName = appSpec.SecretName
	}

	var appName string

	if appSpec.AppName != "" {
		appName = appSpec.AppName
	} else {
		appName = appSpec.App
	}

	var config appv1alpha1.AppSpecConfig

	if appSpec.InCluster {
		config = appv1alpha1.AppSpecConfig{
			ConfigMap: appv1alpha1.AppSpecConfigConfigMap{
				Name:      appSpec.ConfigMapName,
				Namespace: appSpec.ConfigMapNamespace,
			},
		}
		if appSpec.HasClusterValuesSecret {
			config.Secret = appv1alpha1.AppSpecConfigSecret{
				Name:      appSpec.SecretName,
				Namespace: appSpec.SecretNamespace,
			}
		}
	} else {
		config = appv1alpha1.AppSpecConfig{
			ConfigMap: appv1alpha1.AppSpecConfigConfigMap{
				Name:      configMapName,
				Namespace: key.ClusterID(&cr),
			},
		}
		if appSpec.HasClusterValuesSecret {
			config.Secret = appv1alpha1.AppSpecConfigSecret{
				Name:      secretName,
				Namespace: key.ClusterID(&cr),
			}
		}
	}

	var kubeConfig appv1alpha1.AppSpecKubeConfig

	if appSpec.InCluster {
		kubeConfig = appv1alpha1.AppSpecKubeConfig{
			InCluster: true,
		}
	} else {
		kubeConfig = appv1alpha1.AppSpecKubeConfig{
			Context: appv1alpha1.AppSpecKubeConfigContext{
				Name: key.KubeConfigSecretName(&cr),
			},
			Secret: appv1alpha1.AppSpecKubeConfigSecret{
				Name: key.KubeConfigSecretName(&cr),
				// The kubeconfig secret is created in the same namespace as
				// the cluster CR.
				Namespace: cr.GetNamespace(),
			},
		}
	}

	return &appv1alpha1.App{
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
		Spec: appv1alpha1.AppSpec{
			Catalog:    appSpec.Catalog,
			Name:       appSpec.Chart,
			Namespace:  appSpec.Namespace,
			Version:    appSpec.Version,
			Config:     config,
			KubeConfig: kubeConfig,
			UserConfig: userConfig,
		},
	}
}

func newAppOperatorAppSpec(cr capi.Cluster, component releaseversion.ReleaseComponent) key.AppSpec {
	var operatorAppVersion string

	// Setting the reference allows us to deploy from a test catalog.
	if component.Reference != "" {
		operatorAppVersion = component.Reference
	} else {
		operatorAppVersion = component.Version
	}

	return key.AppSpec{
		App: releaseversion.AppOperator,
		// Override app name to include the cluster ID prefix.
		AppName:            fmt.Sprintf("%s-%s", key.ClusterID(&cr), releaseversion.AppOperator),
		Catalog:            component.Catalog,
		Chart:              releaseversion.AppOperator,
		ConfigMapName:      key.AppOperatorValuesResourceName(&cr),
		ConfigMapNamespace: cr.GetNamespace(),
		InCluster:          true,
		Namespace:          cr.GetNamespace(),
		UseUpgradeForce:    false,
		Version:            operatorAppVersion,
	}
}
