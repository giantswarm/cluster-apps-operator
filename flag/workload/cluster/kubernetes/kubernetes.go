package kubernetes

import (
	"github.com/giantswarm/cluster-apps-operator/flag/workload/cluster/kubernetes/api"
)

// Kubernetes is a data structure to hold guest cluster Kubernetes specific
// configuration flags.
type Kubernetes struct {
	API           api.API
	ClusterDomain string
}
