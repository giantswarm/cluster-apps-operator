package clustersecret

import (
	"context"
	"strings"

	corev1 "k8s.io/api/core/v1"
	capi "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"

	appv1alpha1 "github.com/giantswarm/apiextensions-application/api/v1alpha1"
	"github.com/giantswarm/microerror"
)

func vsphereProxyEnabled(ctx context.Context, ctrlClient client.Client, cluster capi.Cluster) (bool, error) {
	var userConfig string
	{
		var clusterApp appv1alpha1.App

		err := ctrlClient.Get(ctx, client.ObjectKey{Namespace: cluster.GetNamespace(), Name: cluster.GetName()}, &clusterApp)
		if err != nil {
			return false, microerror.Mask(err)
		}

		if clusterApp.Spec.Name != "cluster-vsphere" {
			return false, nil
		}
		userConfig = clusterApp.Spec.UserConfig.ConfigMap.Name
	}

	if userConfig == "" {
		return false, nil
	}

	var configMap corev1.ConfigMap
	{
		err := ctrlClient.Get(ctx, client.ObjectKey{Namespace: cluster.GetNamespace(), Name: userConfig}, &configMap)
		if err != nil {
			return false, microerror.Mask(err)
		}
	}

	return getProxyEnabledValueFromConfigMap(configMap)
}

func getProxyEnabledValueFromConfigMap(configMap corev1.ConfigMap) (bool, error) {

	var values map[string]interface{}
	{
		if configMap.Data == nil {
			return false, nil
		}
		if configMap.Data["values"] == "" {
			return false, nil
		}
		data := configMap.Data["values"]
		data = strings.TrimPrefix(data, "|")
		data = strings.TrimPrefix(data, "\n")

		err := yaml.Unmarshal([]byte(data), &values)
		if err != nil {
			return false, microerror.Mask(err)
		}
	}

	// check for global.connectivity.proxy.enabled key
	if value, ok := values["global"].(map[string]interface{}); ok {
		if value, ok := value["connectivity"].(map[string]interface{}); ok {
			if value, ok := value["proxy"].(map[string]interface{}); ok {
				if value, ok := value["enabled"].(bool); ok {
					return value, nil
				}
			}
		}
	}

	return false, nil
}
