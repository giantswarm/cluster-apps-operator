package clustersecret

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	capvcd "github.com/giantswarm/cluster-apps-operator/v2/api/capvcd/v1beta1"

	"github.com/giantswarm/microerror"
)

func (r *Resource) generateCloudDirectorConfig(ctx context.Context, vcdCluster capvcd.VCDCluster) (map[string]interface{}, error) {
	if vcdCluster.Spec.UserCredentialsContext.SecretRef == nil ||
		vcdCluster.Spec.UserCredentialsContext.SecretRef.Name == "" ||
		vcdCluster.Spec.UserCredentialsContext.SecretRef.Namespace == "" {
		return nil, microerror.Mask(invalidConfigError)
	}

	var userContextSecret corev1.Secret
	err := r.k8sClient.CtrlClient().Get(ctx, client.ObjectKey{Namespace: vcdCluster.Spec.UserCredentialsContext.SecretRef.Namespace, Name: vcdCluster.Spec.UserCredentialsContext.SecretRef.Name}, &userContextSecret)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	refreshToken, ok := userContextSecret.Data["refreshToken"]
	if !ok {
		return nil, microerror.Mask(invalidConfigError)
	}

	return map[string]interface{}{
		"basicAuthSecret": map[string]string{
			"refreshToken": string(refreshToken),
		},
		"vcdConfig": map[string]string{
			"site":        vcdCluster.Spec.Site,
			"org":         vcdCluster.Spec.Org,
			"ovdc":        vcdCluster.Spec.Ovdc,
			"ovdcNetwork": vcdCluster.Spec.OvdcNetwork,
			"vipSubnet":   vcdCluster.Spec.LoadBalancer.VipSubnet,
			"clusterid":   vcdCluster.Spec.RDEId,
			"vAppName":    vcdCluster.Name,
		},
	}, nil
}
