package controller

import (
	"github.com/giantswarm/k8sclient/v6/pkg/k8sclient"
	"github.com/giantswarm/k8smetadata/pkg/label"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/v6/pkg/controller"
	"github.com/giantswarm/operatorkit/v6/pkg/resource"
	"github.com/giantswarm/operatorkit/v6/pkg/resource/crud"
	"github.com/giantswarm/operatorkit/v6/pkg/resource/k8s/configmapresource"
	"github.com/giantswarm/operatorkit/v6/pkg/resource/k8s/secretresource"
	"github.com/giantswarm/operatorkit/v6/pkg/resource/wrapper/metricsresource"
	"github.com/giantswarm/operatorkit/v6/pkg/resource/wrapper/retryresource"
	"k8s.io/apimachinery/pkg/labels"
	capi "sigs.k8s.io/cluster-api/api/v1alpha4"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/cluster-apps-operator/pkg/project"
	"github.com/giantswarm/cluster-apps-operator/service/controller/resource/app"
	"github.com/giantswarm/cluster-apps-operator/service/controller/resource/clusterconfigmap"
	"github.com/giantswarm/cluster-apps-operator/service/controller/resource/clustersecret"
	"github.com/giantswarm/cluster-apps-operator/service/internal/chartname"
	"github.com/giantswarm/cluster-apps-operator/service/internal/podcidr"
)

type ClusterConfig struct {
	ChartName      chartname.Interface
	K8sClient      k8sclient.Interface
	Logger         micrologger.Logger
	ReleaseVersion releaseversion.Interface
	PodCIDR        podcidr.Interface

	BaseDomain           string
	ClusterIPRange       string
	DNSIP                string
	Provider             string
	RegistryDomain       string
}

type Cluster struct {
	*controller.Controller
}

func NewCluster(config ClusterConfig) (*Cluster, error) {
	var err error

	resources, err := newClusterResources(config)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var operatorkitController *controller.Controller
	{
		selector, err := labels.Parse(label.ClusterAppsOperatorWatching)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		c := controller.Config{
			K8sClient: config.K8sClient,
			Logger:    config.Logger,
			NewRuntimeObjectFunc: func() client.Object {
				return new(capi.Cluster)
			},
			Resources: resources,

			// Name is used to compute finalizer names. This here results in something
			// like operatorkit.giantswarm.io/cluster-apps-operator-cluster-controller.
			Name:     project.Name() + "-cluster-controller",
			Selector: selector,
		}

		operatorkitController, err = controller.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	c := &Cluster{
		Controller: operatorkitController,
	}

	return c, nil
}

func newClusterResources(config ClusterConfig) ([]resource.Interface, error) {
	var err error

	var appGetter appresource.StateGetter
	{
		c := app.Config{
			ChartName:      config.ChartName,
			G8sClient:      config.K8sClient.CtrlClient(),
			K8sClient:      config.K8sClient.K8sClient(),
			Logger:         config.Logger,
			ReleaseVersion: config.ReleaseVersion,

			RawAppDefaultConfig:  config.RawAppDefaultConfig,
			RawAppOverrideConfig: config.RawAppOverrideConfig,
		}

		appGetter, err = app.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var appResource resource.Interface
	{
		c := appresource.Config{
			G8sClient: config.K8sClient.CtrlClient(),
			Logger:    config.Logger,

			Name:        app.Name,
			StateGetter: appGetter,
		}

		c.AllowedAnnotations = []string{
			"app-operator.giantswarm.io/giantswarm.io/latest-configmap-version",
			"app-operator.giantswarm.io/latest-secret-version",
		}

		ops, err := appresource.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		appResource, err = toCRUDResource(config.Logger, ops)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

		}

		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var clusterConfigMapGetter configmapresource.StateGetter
	{
		c := clusterconfigmap.Config{
			BaseDomain: config.BaseDomain,
			K8sClient:  config.K8sClient,
			Logger:     config.Logger,
			PodCIDR:    config.PodCIDR,

			ClusterIPRange: config.ClusterIPRange,
			DNSIP:          config.DNSIP,
			Provider:       config.Provider,
			RegistryDomain: config.RegistryDomain,
		}

		clusterConfigMapGetter, err = clusterconfigmap.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var clusterConfigMapResource resource.Interface
	{
		c := configmapresource.Config{
			K8sClient: config.K8sClient.K8sClient(),
			Logger:    config.Logger,

			AllowedLabels: []string{
				label.AppOperatorWatching,
			},
			Name:        clusterconfigmap.Name,
			StateGetter: clusterConfigMapGetter,
		}

		ops, err := configmapresource.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		clusterConfigMapResource, err = toCRUDResource(config.Logger, ops)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var clusterSecretGetter secretresource.StateGetter
	{
		c := clustersecret.Config{
			K8sClient: config.K8sClient,
			Logger:    config.Logger,
		}

		clusterSecretGetter, err = clustersecret.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var clusterSecretResource resource.Interface
	{
		c := secretresource.Config{
			K8sClient: config.K8sClient.K8sClient(),
			Logger:    config.Logger,

			Name:        clustersecret.Name,
			StateGetter: clusterSecretGetter,
		}

		ops, err := secretresource.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		clusterSecretResource, err = toCRUDResource(config.Logger, ops)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	resources := []resource.Interface{
		// clusterConfigMapResource is executed before the app resource so the
		// app CRs are accepted by the validation webhook.
		clusterConfigMapResource,
	}

	if config.Provider == "openstack" {
		// clusterSecretResource is executed before the app resource so the
		// app CRs are accepted by the validation webhook.
		resources = append(resources, clusterSecretResource)
	}

	resources = append(resources,
		// appResource manages the per cluster app-operator instance and the
		// workload cluster apps.
		appResource,
	)

	{
		c := retryresource.WrapConfig{
			Logger: config.Logger,
		}

		resources, err = retryresource.Wrap(resources, c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	{
		c := metricsresource.WrapConfig{}

		resources, err = metricsresource.Wrap(resources, c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	return resources, nil
}

func toCRUDResource(logger micrologger.Logger, v crud.Interface) (*crud.Resource, error) {
	c := crud.ResourceConfig{
		CRUD:   v,
		Logger: logger,
	}

	r, err := crud.NewResource(c)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return r, nil
}
