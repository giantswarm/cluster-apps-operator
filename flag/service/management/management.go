package management

import "github.com/giantswarm/cluster-apps-operator/flag/service/management/cluster"

// Management is a data structure to hold management cluster specific configuration
// flags.
type Management struct {
	Cluster cluster.Cluster
}
