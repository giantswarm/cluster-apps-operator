package key

import (
	"github.com/giantswarm/microerror"
	apiv1alpha3 "sigs.k8s.io/cluster-api/api/v1alpha3"
)

// DNSIP returns the IP of the DNS service given a cluster IP range.
func DNSIP(clusterIPRange string) (string, error) {
	// TODO Add logic.
	return "", nil
}

func ToCluster(v interface{}) (apiv1alpha3.Cluster, error) {
	if v == nil {
		return apiv1alpha3.Cluster{}, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", &apiv1alpha3.Cluster{}, v)
	}

	p, ok := v.(*apiv1alpha3.Cluster)
	if !ok {
		return apiv1alpha3.Cluster{}, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", &apiv1alpha3.Cluster{}, v)
	}

	return *p, nil
}
