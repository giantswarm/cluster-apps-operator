package service

import (
	"github.com/giantswarm/operatorkit/v6/pkg/flag/service/kubernetes"

	"github.com/giantswarm/cluster-apps-operator/flag/service/image"
	"github.com/giantswarm/cluster-apps-operator/flag/service/release"
	"github.com/giantswarm/cluster-apps-operator/flag/service/workload"
)

// Service is an intermediate data structure for command line configuration flags.
type Service struct {
	Image      image.Image
	Kubernetes kubernetes.Kubernetes
	Release    release.Release
	Workload   workload.Workload
}
