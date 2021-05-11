package flag

import (
	"github.com/giantswarm/microkit/flag"

	"github.com/giantswarm/cluster-apps-operator/flag/service"
	"github.com/giantswarm/cluster-apps-operator/flag/workload"
)

// Flag provides data structure for service command line flags.
type Flag struct {
	Service  service.Service
	Workload workload.Workload
}

// New constructs fills new Flag structure with given command line flags.
func New() *Flag {
	f := &Flag{}
	flag.Init(f)

	return f
}
