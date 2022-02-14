// Package service implements business logic to create Kubernetes resources
// against the Kubernetes API.
package service

import (
	"context"
	"fmt"
	"sync"

	appv1alpha1 "github.com/giantswarm/apiextensions-application/api/v1alpha1"
	"github.com/giantswarm/k8sclient/v7/pkg/k8sclient"
	"github.com/giantswarm/k8sclient/v7/pkg/k8srestconfig"
	"github.com/giantswarm/microendpoint/service/version"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/spf13/viper"
	"k8s.io/client-go/rest"
	capi "sigs.k8s.io/cluster-api/api/v1alpha4"

	capo "github.com/giantswarm/cluster-apps-operator/api/capo/v1alpha4"
	capz "github.com/giantswarm/cluster-apps-operator/api/capz/v1alpha4"
	"github.com/giantswarm/cluster-apps-operator/flag"
	"github.com/giantswarm/cluster-apps-operator/pkg/project"
	"github.com/giantswarm/cluster-apps-operator/service/collector"
	"github.com/giantswarm/cluster-apps-operator/service/controller"
	"github.com/giantswarm/cluster-apps-operator/service/controller/key"
	"github.com/giantswarm/cluster-apps-operator/service/internal/podcidr"
)

// Config represents the configuration used to create a new service.
type Config struct {
	Logger micrologger.Logger

	Flag  *flag.Flag
	Viper *viper.Viper
}

type Service struct {
	Version *version.Service

	bootOnce          sync.Once
	clusterController *controller.Cluster
	operatorCollector *collector.Set
}

// New creates a new configured service object.
func New(config Config) (*Service, error) {
	var serviceAddress string
	// Settings.
	if config.Flag == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Flag must not be empty")
	}
	if config.Viper == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Viper must not be empty")
	}
	if config.Viper.GetString(config.Flag.Service.Kubernetes.KubeConfig) == "" {
		serviceAddress = config.Viper.GetString(config.Flag.Service.Kubernetes.Address)
	} else {
		serviceAddress = ""
	}

	// Dependencies.
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "logger must not be empty")
	}

	clusterIPRange := config.Viper.GetString(config.Flag.Service.Workload.Cluster.Kubernetes.API.ClusterIPRange)
	var dnsIP string
	{
		var err error
		dnsIP, err = key.DNSIP(clusterIPRange)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var restConfig *rest.Config
	{
		c := k8srestconfig.Config{
			Logger: config.Logger,

			Address:    serviceAddress,
			InCluster:  config.Viper.GetBool(config.Flag.Service.Kubernetes.InCluster),
			KubeConfig: config.Viper.GetString(config.Flag.Service.Kubernetes.KubeConfig),
			TLS: k8srestconfig.ConfigTLS{
				CAFile:  config.Viper.GetString(config.Flag.Service.Kubernetes.TLS.CAFile),
				CrtFile: config.Viper.GetString(config.Flag.Service.Kubernetes.TLS.CrtFile),
				KeyFile: config.Viper.GetString(config.Flag.Service.Kubernetes.TLS.KeyFile),
			},
		}

		var err error
		restConfig, err = k8srestconfig.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var k8sClient k8sclient.Interface
	{
		c := k8sclient.ClientsConfig{
			Logger: config.Logger,
			SchemeBuilder: k8sclient.SchemeBuilder{
				appv1alpha1.AddToScheme,
				capi.AddToScheme,
				capo.AddToScheme,
				capz.AddToScheme,
			},

			RestConfig: restConfig,
		}

		var err error
		k8sClient, err = k8sclient.NewClients(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var pc podcidr.Interface
	{
		calicoSubnet := config.Viper.GetString(config.Flag.Service.Workload.Cluster.Calico.Subnet)
		calicoCIDR := config.Viper.GetString(config.Flag.Service.Workload.Cluster.Calico.CIDR)
		c := podcidr.Config{
			InstallationCIDR: fmt.Sprintf("%s/%s", calicoSubnet, calicoCIDR),
		}

		var err error
		pc, err = podcidr.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var clusterController *controller.Cluster
	{
		c := controller.ClusterConfig{
			K8sClient: k8sClient,
			Logger:    config.Logger,
			PodCIDR:   pc,

			AppOperatorCatalog:   config.Viper.GetString(config.Flag.Service.App.AppOperator.Catalog),
			AppOperatorVersion:   config.Viper.GetString(config.Flag.Service.App.AppOperator.Version),
			ChartOperatorCatalog: config.Viper.GetString(config.Flag.Service.App.ChartOperator.Catalog),
			ChartOperatorVersion: config.Viper.GetString(config.Flag.Service.App.ChartOperator.Version),
			BaseDomain:           config.Viper.GetString(config.Flag.Service.Workload.Cluster.BaseDomain),
			ClusterIPRange:       clusterIPRange,
			DNSIP:                dnsIP,
			Provider:             config.Viper.GetString(config.Flag.Service.Provider.Kind),
			RegistryDomain:       config.Viper.GetString(config.Flag.Service.Image.Registry.Domain),
		}

		var err error
		clusterController, err = controller.NewCluster(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var operatorCollector *collector.Set
	{
		c := collector.SetConfig{
			K8sClient: k8sClient,
			Logger:    config.Logger,
		}

		var err error
		operatorCollector, err = collector.NewSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var versionService *version.Service
	{
		c := version.Config{
			Description: project.Description(),
			GitCommit:   project.GitSHA(),
			Name:        project.Name(),
			Source:      project.Source(),
			Version:     project.Version(),
		}

		var err error
		versionService, err = version.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	s := &Service{
		Version: versionService,

		bootOnce:          sync.Once{},
		clusterController: clusterController,
		operatorCollector: operatorCollector,
	}

	return s, nil
}

func (s *Service) Boot(ctx context.Context) {
	s.bootOnce.Do(func() {
		go s.operatorCollector.Boot(ctx) // nolint:errcheck

		go s.clusterController.Boot(ctx)
	})
}
