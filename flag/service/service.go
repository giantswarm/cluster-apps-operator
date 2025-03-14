package service

import (
	"github.com/giantswarm/operatorkit/v7/pkg/flag/service/kubernetes"

	"github.com/giantswarm/cluster-apps-operator/v3/flag/service/app"
	"github.com/giantswarm/cluster-apps-operator/v3/flag/service/controller"
	"github.com/giantswarm/cluster-apps-operator/v3/flag/service/image"
	"github.com/giantswarm/cluster-apps-operator/v3/flag/service/proxy"
	"github.com/giantswarm/cluster-apps-operator/v3/flag/service/workload"
)

// Service is an intermediate data structure for command line configuration flags.
type Service struct {
	App        app.App
	Image      image.Image
	Kubernetes kubernetes.Kubernetes
	Workload   workload.Workload
	Proxy      proxy.Proxy

	Controller controller.Controller
}
