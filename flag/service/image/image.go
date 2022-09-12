package image

import "github.com/giantswarm/cluster-apps-operator/v2/flag/service/image/registry"

// Image is a data structure to hold container image specific configuration
// flags.
type Image struct {
	Registry registry.Registry
}
