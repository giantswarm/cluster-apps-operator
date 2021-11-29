package clustersecret

type secretSpec struct {
	Name      string
	Namespace string
	Values    map[string]interface{}
}
