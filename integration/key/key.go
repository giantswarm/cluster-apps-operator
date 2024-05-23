//go:build k8srequired
// +build k8srequired

package key

func ClusterID() string {
	return "kind"
}

func ControlPlaneTestCatalogName() string {
	return "control-plane-test-catalog"
}

func Namespace() string {
	return "giantswarm"
}

func OrganizationNamespace() string {
	return "org-test"
}
