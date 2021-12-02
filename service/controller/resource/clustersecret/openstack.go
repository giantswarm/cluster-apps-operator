package clustersecret

import (
	"context"

	"github.com/giantswarm/microerror"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"

	capo "github.com/giantswarm/cluster-apps-operator/api/capo/v1alpha4"
)

type openStackClouds struct {
	Clouds map[string]openStackCloudConfig `json:"clouds"`
}

type openStackCloudConfigAuth struct {
	AuthURL        string `json:"auth_url"`
	Username       string `json:"username"`
	Password       string `json:"password"`
	UserDomainName string `json:"user_domain_name"`
	ProjectID      string `json:"project_id"`
}

type openStackCloudConfig struct {
	Auth               openStackCloudConfigAuth `json:"auth"`
	Verify             bool                     `json:"verify"`
	RegionName         string                   `json:"region_name"`
	Interface          string                   `json:"interface"`
	IdentityAPIVersion int                      `json:"identity_api_version"`
}

func (r *Resource) generateOpenStackCloudConfig(ctx context.Context, cluster capo.OpenStackCluster) (map[string]interface{}, error) {
	if cluster.Spec.IdentityRef == nil || cluster.Spec.IdentityRef.Name == "" || cluster.Spec.IdentityRef.Kind != "Secret" {
		return nil, microerror.Mask(invalidConfigError)
	}

	var cloudConfigSecret corev1.Secret
	err := r.k8sClient.CtrlClient().Get(ctx, client.ObjectKey{Namespace: cluster.Namespace, Name: cluster.Spec.IdentityRef.Name}, &cloudConfigSecret)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	cloudsYAML, ok := cloudConfigSecret.Data["clouds.yaml"]
	if !ok {
		return nil, microerror.Mask(invalidConfigError)
	}

	var clouds openStackClouds
	err = yaml.Unmarshal(cloudsYAML, &clouds)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	cloudConfig, ok := clouds.Clouds["openstack"]
	if !ok {
		return nil, microerror.Mask(invalidConfigError)
	}

	networkID := cluster.Status.Network.ID
	subnetID := cluster.Status.Network.Subnet.ID
	floatingNetworkID := cluster.Status.ExternalNetwork.ID
	publicNetworkName := cluster.Status.ExternalNetwork.Name

	authURL := cloudConfig.Auth.AuthURL
	username := cloudConfig.Auth.Username
	password := cloudConfig.Auth.Password
	tenantID := cloudConfig.Auth.ProjectID
	domainName := cloudConfig.Auth.UserDomainName
	region := cloudConfig.RegionName

	return map[string]interface{}{
		"global": map[string]interface{}{
			"auth-url":    authURL,
			"username":    username,
			"password":    password,
			"tenant-id":   tenantID,
			"domain-name": domainName,
			"region":      region,
		},
		"networking": map[string]interface{}{
			"ipv6-support-disabled": true,
			"public-network-name":   publicNetworkName,
		},
		"loadBalancer": map[string]interface{}{
			"internal-lb":         false,
			"floating-network-id": floatingNetworkID,
			"network-id":          networkID,
			"subnet-id":           subnetID,
		},
	}, nil
}
