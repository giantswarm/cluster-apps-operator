package key

import (
	"fmt"
	"net"
	"strings"

	"github.com/giantswarm/apiextensions-application/api/v1alpha1"

	"github.com/giantswarm/k8smetadata/pkg/label"
	"github.com/giantswarm/microerror"
	capi "sigs.k8s.io/cluster-api/api/v1beta1"
)

const (
	// appOperatorFluxBackend is the label added to Cluster CR
	appOperatorFluxBackend = "app-operator.giantswarm.io/flux-backend"

	// defaultDNSLastOctet is the last octect for the DNS service IP, the first
	// 3 octets come from the cluster IP range.
	defaultDNSLastOctet = 10

	fluxLabelKustomizationName      = "kustomize.toolkit.fluxcd.io/name"
	fluxLabelKustomizationNamespace = "kustomize.toolkit.fluxcd.io/namespace"
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
		clusterID = getter.GetLabels()[capi.ClusterNameLabel]
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

func IsBundle(appName string) bool {
	return strings.HasSuffix(appName, "-bundle")
}

// IsEKS check if the cluster is EKS cluster
func IsEKS(cluster capi.Cluster) bool {
	return cluster.Spec.ControlPlaneRef != nil &&
		cluster.Spec.ControlPlaneRef.Kind == "AWSManagedControlPlane" &&
		cluster.Spec.InfrastructureRef != nil &&
		cluster.Spec.InfrastructureRef.Kind == "AWSManagedCluster"
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

// IsFluxBackendRequested checks if App is considered to be managed by Flux.
// Returns true if the flux kustomization labels are set on the App.
func IsFluxBackendRequested(cluster capi.Cluster) bool {
	labels := cluster.GetLabels()

	if _, ok := labels[appOperatorFluxBackend]; ok {
		return true
	}

	return false
}

// IsManagedByFlux checks if App is considered to be managed by Flux.
// Returns true if the flux kustomization labels are set on the App.
func IsManagedByFlux(app v1alpha1.App) bool {
	labels := app.GetLabels()

	if _, ok := labels[fluxLabelKustomizationName]; !ok {
		return false
	}

	if _, ok := labels[fluxLabelKustomizationNamespace]; !ok {
		return false
	}

	return true
}
