package cluster

import (
	"github.com/giantswarm/cluster-apps-operator/flag/workload/cluster/calico"
	"github.com/giantswarm/cluster-apps-operator/flag/workload/cluster/kubernetes"
)

// Cluster is a data structure to hold cluster specific configuration flags.
type Cluster struct {
	BaseDomain string
	Calico     calico.Calico
	Kubernetes kubernetes.Kubernetes
}
