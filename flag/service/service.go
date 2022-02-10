package service

import (
	"github.com/giantswarm/operatorkit/v7/pkg/flag/service/kubernetes"

	"github.com/giantswarm/cluster-apps-operator/flag/service/app"
	"github.com/giantswarm/cluster-apps-operator/flag/service/image"
	"github.com/giantswarm/cluster-apps-operator/flag/service/management"
	"github.com/giantswarm/cluster-apps-operator/flag/service/provider"
	"github.com/giantswarm/cluster-apps-operator/flag/service/workload"
)

// Service is an intermediate data structure for command line configuration flags.
type Service struct {
	App        app.App
	Image      image.Image
	Kubernetes kubernetes.Kubernetes
	Management management.Management
	Provider   provider.Provider
	Workload   workload.Workload
}
