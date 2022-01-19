package app

// AppSpec is used to define app custom resources.
type AppSpec struct {
	App                    string
	AppOperatorVersion     string
	AppName                string
	Catalog                string
	ConfigMapName          string
	ConfigMapNamespace     string
	InCluster              bool
	HasClusterValuesSecret bool
	TargetNamespace        string
	UseUpgradeForce        bool
	Version                string
}
