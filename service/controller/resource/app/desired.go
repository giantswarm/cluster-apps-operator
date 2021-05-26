package app

import (
	"context"
	"fmt"
	"strconv"

	applicationv1alpha1 "github.com/giantswarm/apiextensions/v3/pkg/apis/application/v1alpha1"
	"github.com/giantswarm/k8smetadata/pkg/annotation"
	"github.com/giantswarm/k8smetadata/pkg/label"
	"github.com/giantswarm/microerror"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apiv1alpha3 "sigs.k8s.io/cluster-api/api/v1alpha3"

	"github.com/giantswarm/cluster-apps-operator/pkg/project"
	"github.com/giantswarm/cluster-apps-operator/service/controller/key"
	"github.com/giantswarm/cluster-apps-operator/service/internal/releaseversion"
)

func (r *Resource) GetDesiredState(ctx context.Context, obj interface{}) ([]*applicationv1alpha1.App, error) {
	cr, err := key.ToCluster(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var apps []*applicationv1alpha1.App

	if key.IsDeleted(&cr) {
		r.logger.Debugf(ctx, "deleting apps for workload cluster %#q", key.ClusterID(&cr))
		return apps, nil
	}

	appSpecs, err := r.newAppSpecs(ctx, cr)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	componentVersions, err := r.releaseVersion.ComponentVersion(ctx, &cr)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	configMaps, err := r.getConfigMaps(ctx, cr)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	secrets, err := r.getSecrets(ctx, cr)
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
	apps = append(apps, r.newApp(uniqueOperatorVersion, cr, appOperatorAppSpec, applicationv1alpha1.AppSpecUserConfig{}))

	for _, appSpec := range appSpecs {
		userConfig := newUserConfig(cr, appSpec, configMaps, secrets)
		apps = append(apps, r.newApp(appOperatorVersion, cr, appSpec, userConfig))
	}

	return apps, nil
}

func (r *Resource) newApp(appOperatorVersion string, cr apiv1alpha3.Cluster, appSpec key.AppSpec, userConfig applicationv1alpha1.AppSpecUserConfig) *applicationv1alpha1.App {
	configMapName := key.ClusterConfigMapName(&cr)

	// Override config map name when specified.
	if appSpec.ConfigMapName != "" {
		configMapName = appSpec.ConfigMapName
	}

	var appName string

	if appSpec.AppName != "" {
		appName = appSpec.AppName
	} else {
		appName = appSpec.App
	}

	var config applicationv1alpha1.AppSpecConfig

	if appSpec.InCluster {
		config = applicationv1alpha1.AppSpecConfig{}
	} else {
		config = applicationv1alpha1.AppSpecConfig{
			ConfigMap: applicationv1alpha1.AppSpecConfigConfigMap{
				Name:      configMapName,
				Namespace: key.ClusterID(&cr),
			},
		}
	}

	var kubeConfig applicationv1alpha1.AppSpecKubeConfig

	if appSpec.InCluster {
		kubeConfig = applicationv1alpha1.AppSpecKubeConfig{
			InCluster: true,
		}
	} else {
		kubeConfig = applicationv1alpha1.AppSpecKubeConfig{
			Context: applicationv1alpha1.AppSpecKubeConfigContext{
				Name: key.KubeConfigSecretName(&cr),
			},
			Secret: applicationv1alpha1.AppSpecKubeConfigSecret{
				Name:      key.KubeConfigSecretName(&cr),
				Namespace: cr.GetNamespace(),
			},
		}
	}

	return &applicationv1alpha1.App{
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
			Namespace: key.ClusterID(&cr),
		},
		Spec: applicationv1alpha1.AppSpec{
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

func (r *Resource) newAppSpecs(ctx context.Context, cr apiv1alpha3.Cluster) ([]key.AppSpec, error) {
	apps, err := r.releaseVersion.Apps(ctx, &cr)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var specs []key.AppSpec

	for appName, app := range apps {
		var catalog string
		if app.Catalog == "" {
			catalog = r.defaultConfig.Catalog
		} else {
			catalog = app.Catalog
		}

		chart, err := r.chartName.ChartName(ctx, catalog, appName, app.Version)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		spec := key.AppSpec{
			App:             appName,
			Catalog:         catalog,
			Chart:           chart,
			Namespace:       r.defaultConfig.Namespace,
			UseUpgradeForce: r.defaultConfig.UseUpgradeForce,
			Version:         app.Version,
		}
		// For some apps we can't use default settings. We check ConfigExceptions map
		// for these differences.
		// We are looking into ConfigException map to see if this chart is the case.
		if val, ok := r.overrideConfig[appName]; ok {
			if val.Chart != "" {
				spec.Chart = val.Chart
			}
			if val.Namespace != "" {
				spec.Namespace = val.Namespace
			}
			if val.UseUpgradeForce != nil {
				spec.UseUpgradeForce = *val.UseUpgradeForce
			}
		}

		specs = append(specs, spec)
	}

	return specs, nil
}

func newAppOperatorAppSpec(cr apiv1alpha3.Cluster, component releaseversion.ReleaseComponent) key.AppSpec {
	var operatorAppVersion string

	// Setting the reference allows us to deploy from a test catalog.
	if component.Reference != "" {
		operatorAppVersion = component.Reference
	} else {
		operatorAppVersion = component.Version
	}

	return key.AppSpec{
		App: releaseversion.AppOperator,
		// Override app name to include the cluster ID.
		AppName:         fmt.Sprintf("%s-%s", releaseversion.AppOperator, key.ClusterID(&cr)),
		Catalog:         component.Catalog,
		Chart:           releaseversion.AppOperator,
		InCluster:       true,
		Namespace:       key.ClusterID(&cr),
		UseUpgradeForce: false,
		Version:         operatorAppVersion,
	}
}

func newUserConfig(cr apiv1alpha3.Cluster, appSpec key.AppSpec, configMaps map[string]corev1.ConfigMap, secrets map[string]corev1.Secret) applicationv1alpha1.AppSpecUserConfig {
	userConfig := applicationv1alpha1.AppSpecUserConfig{}

	_, ok := configMaps[key.AppUserConfigMapName(appSpec)]
	if ok {
		configMapSpec := applicationv1alpha1.AppSpecUserConfigConfigMap{
			Name:      key.AppUserConfigMapName(appSpec),
			Namespace: key.ClusterID(&cr),
		}

		userConfig.ConfigMap = configMapSpec
	}

	_, ok = secrets[key.AppUserSecretName(appSpec)]
	if ok {
		secretSpec := applicationv1alpha1.AppSpecUserConfigSecret{
			Name:      key.AppUserSecretName(appSpec),
			Namespace: key.ClusterID(&cr),
		}

		userConfig.Secret = secretSpec
	}

	return userConfig
}
