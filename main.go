package main

import (
	"context"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/microkit/command"
	microserver "github.com/giantswarm/microkit/server"
	"github.com/giantswarm/micrologger"
	"github.com/spf13/viper"

	"github.com/giantswarm/cluster-apps-operator/v2/flag"
	"github.com/giantswarm/cluster-apps-operator/v2/pkg/project"
	"github.com/giantswarm/cluster-apps-operator/v2/server"
	"github.com/giantswarm/cluster-apps-operator/v2/service"
)

var (
	f *flag.Flag = flag.New()
)

func main() {
	err := mainE(context.Background())
	if err != nil {
		panic(microerror.JSON(err))
	}
}

func mainE(ctx context.Context) error {
	var err error

	var logger micrologger.Logger
	{
		c := micrologger.Config{}

		logger, err = micrologger.New(c)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	// We define a server factory to create the custom server once all command
	// line flags are parsed and all microservice configuration is storted out.
	serverFactory := func(v *viper.Viper) microserver.Server {
		// Create a new custom service which implements business logic.
		var newService *service.Service
		{
			c := service.Config{
				Logger: logger,

				Flag:  f,
				Viper: v,
			}

			newService, err = service.New(c)
			if err != nil {
				panic(microerror.JSON(err))
			}

			go newService.Boot(ctx)
		}

		// Create a new custom server which bundles our endpoints.
		var newServer microserver.Server
		{
			c := server.Config{
				Logger:  logger,
				Service: newService,

				Viper: v,
			}

			newServer, err = server.New(c)
			if err != nil {
				panic(microerror.JSON(err))
			}
		}

		return newServer
	}

	// Create a new microkit command which manages our custom microservice.
	var newCommand command.Command
	{
		c := command.Config{
			Logger:        logger,
			ServerFactory: serverFactory,

			Description: project.Description(),
			GitCommit:   project.GitSHA(),
			Name:        project.Name(),
			Source:      project.Source(),
			Version:     project.Version(),
		}

		newCommand, err = command.New(c)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	daemonCommand := newCommand.DaemonCommand().CobraCommand()

	daemonCommand.PersistentFlags().String(f.Service.App.AppOperator.Catalog, "", "Catalog for app-operator app CR.")
	daemonCommand.PersistentFlags().String(f.Service.App.AppOperator.Version, "", "Version for app-operator app CR.")
	daemonCommand.PersistentFlags().String(f.Service.App.ChartOperator.Catalog, "", "Catalog for chart-operator app CR.")
	daemonCommand.PersistentFlags().String(f.Service.App.ChartOperator.Version, "", "Version for chart-operator app CR.")

	daemonCommand.PersistentFlags().String(f.Service.Image.Registry.Domain, "quay.io", "Image registry.")

	daemonCommand.PersistentFlags().String(f.Service.Kubernetes.Address, "http://127.0.0.1:6443", "Address used to connect to Kubernetes. When empty in-cluster config is created.")
	daemonCommand.PersistentFlags().Bool(f.Service.Kubernetes.InCluster, false, "Whether to use the in-cluster config to authenticate with Kubernetes.")
	daemonCommand.PersistentFlags().String(f.Service.Kubernetes.KubeConfig, "", "KubeConfig used to connect to Kubernetes. When empty other settings are used.")
	daemonCommand.PersistentFlags().String(f.Service.Kubernetes.TLS.CAFile, "", "Certificate authority file path to use to authenticate with Kubernetes.")
	daemonCommand.PersistentFlags().String(f.Service.Kubernetes.TLS.CrtFile, "", "Certificate file path to use to authenticate with Kubernetes.")
	daemonCommand.PersistentFlags().String(f.Service.Kubernetes.TLS.KeyFile, "", "Key file path to use to authenticate with Kubernetes.")

	daemonCommand.PersistentFlags().String(f.Service.Provider.Kind, "", "Provider of management cluster this operator is running in. Used to determine provider-specific cluster values.")

	daemonCommand.PersistentFlags().String(f.Service.Workload.Cluster.BaseDomain, "", "Cluster owner base domain.")
	daemonCommand.PersistentFlags().String(f.Service.Workload.Cluster.Calico.CIDR, "", "Prefix length for the CIDR block used by Calico.")
	daemonCommand.PersistentFlags().String(f.Service.Workload.Cluster.Calico.Subnet, "", "Network address for the CIDR block used by Calico.")
	daemonCommand.PersistentFlags().String(f.Service.Workload.Cluster.Kubernetes.API.ClusterIPRange, "", "CIDR Range for Pods in cluster.")
	daemonCommand.PersistentFlags().String(f.Service.Workload.Cluster.Kubernetes.ClusterDomain, "cluster.local", "Internal Kubernetes domain.")
	daemonCommand.PersistentFlags().String(f.Service.Workload.Cluster.Owner, "", "Management cluster codename.")

	daemonCommand.PersistentFlags().String(f.Service.Proxy.NoProxy, "", "Installation specific no_proxy values.")
	daemonCommand.PersistentFlags().String(f.Service.Proxy.HttpProxy, "", "Installation specific http_proxy value.")
	daemonCommand.PersistentFlags().String(f.Service.Proxy.HttpsProxy, "", "Installation specific https_proxy value.")
	/*
		TODO:
			* set http and https from external
			* inject into cluster values
	*/

	err = newCommand.CobraCommand().Execute()
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}
