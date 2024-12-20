package cluster

import (
	"github.com/giantswarm/cluster-apps-operator/v3/flag/service/proxy"
	"github.com/giantswarm/cluster-apps-operator/v3/flag/service/workload/cluster/calico"
	"github.com/giantswarm/cluster-apps-operator/v3/flag/service/workload/cluster/kubernetes"
)

// Cluster is a data structure to hold cluster specific configuration flags.
type Cluster struct {
	BaseDomain string
	Calico     calico.Calico
	Kubernetes kubernetes.Kubernetes
	Owner      string
	Proxy      proxy.Proxy
}
