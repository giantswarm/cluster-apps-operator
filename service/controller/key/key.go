package key

import (
	"fmt"
	"net"

	"github.com/giantswarm/k8smetadata/pkg/label"
	"github.com/giantswarm/microerror"
	capi "sigs.k8s.io/cluster-api/api/v1alpha4"
)

const (
	// defaultDNSLastOctet is the last octect for the DNS service IP, the first
	// 3 octets come from the cluster IP range.
	defaultDNSLastOctet = 10
)

func AppOperatorAppName(getter LabelsGetter) string {
	return fmt.Sprintf("%s-app-operator", ClusterID(getter))
}

func AppOperatorValuesResourceName(getter LabelsGetter) string {
	return fmt.Sprintf("%s-app-operator-values", ClusterID(getter))
}

func BaseDomain(getter LabelsGetter, base string) string {
	return fmt.Sprintf("%s.%s", ClusterID(getter), base)
}

func ChartOperatorAppName(getter LabelsGetter) string {
	return fmt.Sprintf("%s-chart-operator", ClusterID(getter))
}

func ClusterValuesResourceName(getter LabelsGetter) string {
	return fmt.Sprintf("%s-cluster-values", ClusterID(getter))
}

func ClusterCAName(getter LabelsGetter) string {
	return fmt.Sprintf("%s-ca", ClusterID(getter))
}

func ClusterID(getter LabelsGetter) string {
	clusterID := getter.GetLabels()[label.Cluster]
	// If the Giant Swarm cluster name is empty, attempt to retrieve it from the
	// upstream label.
	if clusterID == "" {
		clusterID = getter.GetLabels()[capi.ClusterLabelName]
	}
	return clusterID
}

// DNSIP returns the IP of the DNS service given a cluster IP range.
func DNSIP(clusterIPRange string) (string, error) {
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

func KubeConfigSecretName(getter LabelsGetter) string {
	return fmt.Sprintf("%s-kubeconfig", ClusterID(getter))
}

func PodCIDR(cr capi.Cluster) string {
	if cr.Spec.ClusterNetwork == nil {
		return ""
	}
	if cr.Spec.ClusterNetwork.Pods == nil {
		return ""
	}
	if len(cr.Spec.ClusterNetwork.Pods.CIDRBlocks) == 0 {
		return ""
	}

	return cr.Spec.ClusterNetwork.Pods.CIDRBlocks[0]
}

func ToCluster(v interface{}) (capi.Cluster, error) {
	if v == nil {
		return capi.Cluster{}, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", &capi.Cluster{}, v)
	}

	p, ok := v.(*capi.Cluster)
	if !ok {
		return capi.Cluster{}, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", &capi.Cluster{}, v)
	}

	return *p, nil
}
