package clustersecret

type secretSpec struct {
	Name      string
	Namespace string
	Data      map[string][]byte
}
