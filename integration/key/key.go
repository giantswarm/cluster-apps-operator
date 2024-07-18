//go:build k8srequired
// +build k8srequired

package key

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	capi "sigs.k8s.io/cluster-api/api/v1beta1"
)

func ClusterID() string {
	return "kind"
}

func ControlPlaneTestCatalogName() string {
	return "control-plane-test-catalog"
}

func KindAppOperatorName() string {
	return "kind-app-operator"
}

func KindAppOperatorValuesName() string {
	return "kind-app-operator-values"
}

func KindChartOperatorName() string {
	return "kind-chart-operator"
}

func Namespace() string {
	return "giantswarm"
}

func OrganizationNamespace() string {
	return "org-test"
}

func TestKindCluster(fluxBackend bool) *capi.Cluster {
	cluster := &capi.Cluster{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Cluster",
			APIVersion: "cluster.x-k8s.io",
		},
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				"cluster-apps-operator.giantswarm.io/watching": "",
				"cluster.x-k8s.io/cluster-name":                "kind",
			},
			Name:      "kind",
			Namespace: "org-test",
		},
		Spec: capi.ClusterSpec{
			InfrastructureRef: &corev1.ObjectReference{
				Kind: "KindCluster",
			},
		},
	}

	if fluxBackend {
		cluster.ObjectMeta.Labels["app-operator.giantswarm.io/flux-backend"] = ""
	}

	return cluster
}
