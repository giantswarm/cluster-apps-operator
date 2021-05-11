package workload

import "github.com/giantswarm/cluster-apps-operator/flag/workload/cluster"

// Workload is a data structure to hold workload cluster specific configuration
// flags.
type Workload struct {
	Cluster cluster.Cluster
}
