package app

import (
	"context"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	apiv1alpha3 "sigs.k8s.io/cluster-api/api/v1alpha4"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"

	"github.com/giantswarm/cluster-apps-operator/service/controller/key"
	"github.com/giantswarm/cluster-apps-operator/service/internal/chartname"
	"github.com/giantswarm/cluster-apps-operator/service/internal/releaseversion"
)

const (
	// Name is the identifier of the resource.
	Name = "app"

	uniqueOperatorVersion = "0.0.0"
)

// Config represents the configuration used to create a new app resource.
type Config struct {
	ChartName      chartname.Interface
	G8sClient      client.Client
	K8sClient      kubernetes.Interface
	Logger         micrologger.Logger
	ReleaseVersion releaseversion.Interface

	RawAppDefaultConfig  string
	RawAppOverrideConfig string
}

// Resource implements the app resource.
type Resource struct {
	chartName      chartname.Interface
	g8sClient      client.Client
	k8sClient      kubernetes.Interface
	logger         micrologger.Logger
	releaseVersion releaseversion.Interface

	defaultConfig  defaultConfig
	overrideConfig overrideConfig
}

type defaultConfig struct {
	Catalog         string `json:"catalog"`
	Namespace       string `json:"namespace"`
	UseUpgradeForce bool   `json:"useUpgradeForce"`
}

type overrideProperties struct {
	Chart                  string `json:"chart"`
	HasClusterValuesSecret *bool  `json:"hasClusterValuesSecret,omitempty"`
	Namespace              string `json:"namespace"`
	UseUpgradeForce        *bool  `json:"useUpgradeForce,omitempty"`
}

type overrideConfig map[string]overrideProperties

// New creates a new configured app state getter resource managing
// app CRs.
//
//     https://pkg.go.dev/github.com/giantswarm/resource/v2/appresource#StateGetter
//
func New(config Config) (*Resource, error) {
	if config.ChartName == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.ChartName must not be empty", config)
	}
	if config.G8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.G8sClient must not be empty", config)
	}
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}
	if config.ReleaseVersion == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.ReleaseVersion must not be empty", config)
	}

	if config.RawAppDefaultConfig == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.RawDefaultConfig must not be empty", config)
	}
	if config.RawAppOverrideConfig == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.RawOverrideConfig must not be empty", config)
	}

	defaultConfig := defaultConfig{}
	err := yaml.Unmarshal([]byte(config.RawAppDefaultConfig), &defaultConfig)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	overrideConfig := overrideConfig{}
	err = yaml.Unmarshal([]byte(config.RawAppOverrideConfig), &overrideConfig)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r := &Resource{
		chartName:      config.ChartName,
		g8sClient:      config.G8sClient,
		k8sClient:      config.K8sClient,
		logger:         config.Logger,
		releaseVersion: config.ReleaseVersion,

		defaultConfig:  defaultConfig,
		overrideConfig: overrideConfig,
	}

	return r, nil
}

func (r *Resource) Name() string {
	return Name
}

func (r *Resource) getConfigMaps(ctx context.Context, cr apiv1alpha3.Cluster) (map[string]corev1.ConfigMap, error) {
	configMaps := map[string]corev1.ConfigMap{}

	r.logger.Debugf(ctx, "finding configMaps in namespace %#q", key.ClusterID(&cr))

	list, err := r.k8sClient.CoreV1().ConfigMaps(key.ClusterID(&cr)).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, microerror.Mask(err)
	}

	for _, cm := range list.Items {
		configMaps[cm.Name] = cm
	}

	r.logger.Debugf(ctx, "found %d configMaps in namespace %#q", len(configMaps), key.ClusterID(&cr))

	return configMaps, nil
}

func (r *Resource) getSecrets(ctx context.Context, cr apiv1alpha3.Cluster) (map[string]corev1.Secret, error) {
	secrets := map[string]corev1.Secret{}

	r.logger.Debugf(ctx, "finding secrets in namespace %#q", key.ClusterID(&cr))

	list, err := r.k8sClient.CoreV1().Secrets(key.ClusterID(&cr)).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, microerror.Mask(err)
	}

	for _, s := range list.Items {
		secrets[s.Name] = s
	}

	r.logger.Debugf(ctx, "found %d secrets in namespace %#q", len(secrets), key.ClusterID(&cr))

	return secrets, nil
}
