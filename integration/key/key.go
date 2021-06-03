// +build k8srequired

package key

func ControlPlaneTestCatalogName() string {
	return "control-plane-test-catalog"
}

func Namespace() string {
	return "giantswarm"
}
