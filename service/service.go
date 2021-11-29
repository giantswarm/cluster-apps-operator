// Package service implements business logic to create Kubernetes resources
// against the Kubernetes API.
package service

import (
	"context"
	"fmt"
	"sync"

	applicationv1alpha1 "github.com/giantswarm/apiextensions-application/api/v1alpha1"
	"github.com/giantswarm/k8sclient/v6/pkg/k8sclient"
	"github.com/giantswarm/k8sclient/v6/pkg/k8srestconfig"
	"github.com/giantswarm/microendpoint/service/version"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	releasev1alpha1 "github.com/giantswarm/release-operator/v2/api/v1alpha1"
	"github.com/spf13/viper"
	"k8s.io/client-go/rest"
	apiv1alpha3 "sigs.k8s.io/cluster-api/api/v1alpha4"
	bootstrapkubeadmv1alpha3 "sigs.k8s.io/cluster-api/bootstrap/kubeadm/api/v1alpha4"

	capzv1alpha3 "github.com/giantswarm/cluster-apps-operator/api/capz/v1alpha3"
	"github.com/giantswarm/cluster-apps-operator/flag"
	"github.com/giantswarm/cluster-apps-operator/pkg/project"
	"github.com/giantswarm/cluster-apps-operator/service/collector"
	"github.com/giantswarm/cluster-apps-operator/service/controller"
	"github.com/giantswarm/cluster-apps-operator/service/controller/key"
	"github.com/giantswarm/cluster-apps-operator/service/internal/chartname"
	"github.com/giantswarm/cluster-apps-operator/service/internal/podcidr"
	"github.com/giantswarm/cluster-apps-operator/service/internal/releaseversion"
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
	if config.Viper.GetString(config.Flag.Service.Kubernetes.Address) != "" {
		serviceAddress = config.Viper.GetString(config.Flag.Service.Kubernetes.Address)
	} else {
		serviceAddress = ""
	}

	// Dependencies.
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "logger must not be empty")
	}

	var err error

	baseDomain := config.Viper.GetString(config.Flag.Service.Workload.Cluster.BaseDomain)
	calicoSubnet := config.Viper.GetString(config.Flag.Service.Workload.Cluster.Calico.Subnet)
	calicoCIDR := config.Viper.GetString(config.Flag.Service.Workload.Cluster.Calico.CIDR)
	clusterIPRange := config.Viper.GetString(config.Flag.Service.Workload.Cluster.Kubernetes.API.ClusterIPRange)
	registryDomain := config.Viper.GetString(config.Flag.Service.Image.Registry.Domain)

	var dnsIP string
	{
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
				apiv1alpha3.AddToScheme,
				bootstrapkubeadmv1alpha3.AddToScheme,
				releasev1alpha1.AddToScheme,
				capzv1alpha3.AddToScheme,
				applicationv1alpha1.AddToScheme,
			},

			RestConfig: restConfig,
		}

		k8sClient, err = k8sclient.NewClients(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var cn chartname.Interface
	{
		c := chartname.Config{
			G8sClient: k8sClient.CtrlClient(),
			Logger:    config.Logger,
		}

		cn, err = chartname.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var pc podcidr.Interface
	{
		c := podcidr.Config{
			InstallationCIDR: fmt.Sprintf("%s/%s", calicoSubnet, calicoCIDR),
		}

		pc, err = podcidr.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var rv releaseversion.Interface
	{
		c := releaseversion.Config{
			K8sClient: k8sClient,
		}

		rv, err = releaseversion.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var clusterController *controller.Cluster
	{

		c := controller.ClusterConfig{
			ChartName:      cn,
			K8sClient:      k8sClient,
			Logger:         config.Logger,
			PodCIDR:        pc,
			ReleaseVersion: rv,

			BaseDomain:           baseDomain,
			ClusterIPRange:       clusterIPRange,
			DNSIP:                dnsIP,
			RawAppDefaultConfig:  config.Viper.GetString(config.Flag.Service.Release.App.Config.Default),
			RawAppOverrideConfig: config.Viper.GetString(config.Flag.Service.Release.App.Config.Override),
			RegistryDomain:       registryDomain,
		}

		clusterController, err = controller.NewCluster(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var operatorCollector *collector.Set
	{
		c := collector.SetConfig{
			K8sClient: k8sClient.K8sClient(),
			Logger:    config.Logger,
		}

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
