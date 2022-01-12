package app

import (
	"context"
	"fmt"
	"strconv"

	appv1alpha1 "github.com/giantswarm/apiextensions-application/api/v1alpha1"
	"github.com/giantswarm/k8smetadata/pkg/annotation"
	"github.com/giantswarm/k8smetadata/pkg/label"
	"github.com/giantswarm/microerror"
	corev1 "k8s.io/api/core/v1"
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

	// Define app CR for app-operator in the workload cluster namespace.
	appOperatorAppSpec := newAppOperatorAppSpec(cr, appOperatorComponent)
	apps = append(apps, r.newApp(uniqueOperatorVersion, cr, appOperatorAppSpec, appv1alpha1.AppSpecUserConfig{}))

	for _, appSpec := range appSpecs {
		// These apps are pre-installed when the control plane is
		// managed by AWS.
		if key.InfrastructureRefKind(cr) == "AWSManagedControlPlane" {
			if appSpec.App == "aws-cns" || appSpec.App == "coredns" {
				r.logger.Debugf(ctx, "not creating app %#q for infra ref kind %#q", appSpec.App, key.InfrastructureRefKind(cr))
				continue
			}
		}

		userConfig := newUserConfig(cr, appSpec, configMaps, secrets)
		apps = append(apps, r.newApp(appOperatorVersion, cr, appSpec, userConfig))
	}

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

	// Add cluster ID prefix for app CRs in organization namespace.
	if !appSpec.InCluster {
		appName = fmt.Sprintf("%s=%s", key.ClusterID(&cr), appName)
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

	var appNamespace string

	if appSpec.InCluster {
		appNamespace = key.ClusterID(&cr)
	} else {
		appNamespace = cr.GetNamespace()
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
			Namespace: appNamespace,
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

func (r *Resource) newAppSpecs(ctx context.Context, cr capi.Cluster) ([]key.AppSpec, error) {
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
			if val.HasClusterValuesSecret != nil {
				spec.HasClusterValuesSecret = *val.HasClusterValuesSecret
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
		// Override app name to include the cluster ID.
		AppName: fmt.Sprintf("%s-%s", releaseversion.AppOperator, key.ClusterID(&cr)),
		Catalog: component.Catalog,
		Chart:   releaseversion.AppOperator,
		// Use config map with management cluster config.
		ConfigMapName:      "app-operator-konfigure",
		ConfigMapNamespace: "giantswarm",
		InCluster:          true,
		Namespace:          key.ClusterID(&cr),
		UseUpgradeForce:    false,
		Version:            operatorAppVersion,
	}
}

func newUserConfig(cr capi.Cluster, appSpec key.AppSpec, configMaps map[string]corev1.ConfigMap, secrets map[string]corev1.Secret) appv1alpha1.AppSpecUserConfig {
	userConfig := appv1alpha1.AppSpecUserConfig{}

	_, ok := configMaps[key.AppUserConfigMapName(appSpec)]
	if ok {
		configMapSpec := appv1alpha1.AppSpecUserConfigConfigMap{
			Name:      key.AppUserConfigMapName(appSpec),
			Namespace: key.ClusterID(&cr),
		}

		userConfig.ConfigMap = configMapSpec
	}

	_, ok = secrets[key.AppUserSecretName(appSpec)]
	if ok {
		secretSpec := appv1alpha1.AppSpecUserConfigSecret{
			Name:      key.AppUserSecretName(appSpec),
			Namespace: key.ClusterID(&cr),
		}

		userConfig.Secret = secretSpec
	}

	return userConfig
}
