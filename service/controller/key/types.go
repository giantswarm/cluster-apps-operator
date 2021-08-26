package key

// AppSpec is used to define app custom resources.
type AppSpec struct {
	App     string
	AppName string
	Catalog string
	Chart   string
	// ConfigMapName overrides the name, otherwise the cluster values configmap
	// is used.
	ConfigMapName string
	// ConfigMapNamespace overrides the namespace, otherwise the cluster
	// namespace is used.
	ConfigMapNamespace string
	// InCluster determines if the app CR should use in cluster. Otherwise the
	// cluster kubeconfig is specified.
	InCluster       bool
	Namespace       string
	UseUpgradeForce bool
	Version         string
}
