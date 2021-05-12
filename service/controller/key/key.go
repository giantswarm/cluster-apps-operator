package key

import (
	"fmt"
	"net"

	"github.com/giantswarm/k8smetadata/pkg/label"
	"github.com/giantswarm/microerror"
	apiv1alpha3 "sigs.k8s.io/cluster-api/api/v1alpha3"
)

const (
	// defaultDNSLastOctet is the last octect for the DNS service IP, the first
	// 3 octets come from the cluster IP range.
	defaultDNSLastOctet = 10
)

func BaseDomain(getter LabelsGetter, base string) string {
	return fmt.Sprintf("%s.k8s.%s", ClusterID(getter), base)
}

func ClusterConfigMapName(getter LabelsGetter) string {
	return fmt.Sprintf("%s-cluster-values", ClusterID(getter))
}

func ClusterID(getter LabelsGetter) string {
	return getter.GetLabels()[label.Cluster]
}

// DNSIP returns the IP of the DNS service given a cluster IP range.
func DNSIP(clusterIPRange string) (string, error) {
	return "", nil
	ip, _, err := net.ParseCIDR(clusterIPRange)
	if err != nil {
		return "", microerror.Maskf(invalidConfigError, err.Error())
	}

	// Only IPV4 CIDRs are supported.
	ip = ip.To4()
	if ip == nil {
		return "", microerror.Mask(invalidConfigError)
	}

	// IP must be a network address.
	if ip[3] != 0 {
		return "", microerror.Mask(invalidConfigError)
	}

	ip[3] = defaultDNSLastOctet

	return ip.String(), nil
}

func IsDeleted(getter DeletionTimestampGetter) bool {
	return getter.GetDeletionTimestamp() != nil
}

func OrganizationID(getter LabelsGetter) string {
	return getter.GetLabels()[label.Organization]
}

func ReleaseName(releaseVersion string) string {
	return fmt.Sprintf("v%s", releaseVersion)
}

func ReleaseVersion(getter LabelsGetter) string {
	return getter.GetLabels()[label.ReleaseVersion]
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
