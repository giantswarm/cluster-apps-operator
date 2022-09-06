package api

// API is a data structure to hold guest cluster Kubernetes API specific
// configuration flags.
type API struct {
	// ClusterIPRange is the CIDR for the k8s `Services`.
	ClusterIPRange string
}
