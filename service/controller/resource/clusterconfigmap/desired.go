package clusterconfigmap

import (
	"context"
	"fmt"

	"github.com/giantswarm/k8smetadata/pkg/annotation"
	"github.com/giantswarm/k8smetadata/pkg/label"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/v5/pkg/controller/context/resourcecanceledcontext"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	capzv1alpha3 "sigs.k8s.io/cluster-api-provider-azure/api/v1alpha3"
	capov1alpha4 "sigs.k8s.io/cluster-api-provider-openstack/api/v1alpha4"
	apiv1alpha3 "sigs.k8s.io/cluster-api/api/v1alpha3"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"

	"github.com/giantswarm/cluster-apps-operator/pkg/project"
	"github.com/giantswarm/cluster-apps-operator/service/controller/key"
	"github.com/giantswarm/cluster-apps-operator/service/internal/podcidr"
)

func (r *Resource) GetDesiredState(ctx context.Context, obj interface{}) ([]*corev1.ConfigMap, error) {
	cr, err := key.ToCluster(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var configMaps []*corev1.ConfigMap

	if key.IsDeleted(&cr) {
		r.logger.Debugf(ctx, "deleting cluster configmaps for cluster '%s/%s'", cr.GetNamespace(), key.ClusterID(&cr))
		return configMaps, nil
	}

	var podCIDR string
	{
		podCIDR, err = r.podCIDR.PodCIDR(ctx, &cr)
		if podcidr.IsNotFound(err) {
			r.logger.Debugf(ctx, "pod cidr not available yet for cluster '%s/%s'", cr.GetNamespace(), key.ClusterID(&cr))
			r.logger.Debugf(ctx, "canceling resource")
			resourcecanceledcontext.SetCanceled(ctx)
			return nil, nil
		} else if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var clusterCIDR string
	var subnetID string
	var networkID string
	{
		infrastructureRef := cr.Spec.InfrastructureRef
		if infrastructureRef != nil {
			switch infrastructureRef.Kind {
			case "AzureCluster":
				var azureCluster capzv1alpha3.AzureCluster
				err = r.k8sClient.CtrlClient().Get(ctx, client.ObjectKey{Namespace: infrastructureRef.Namespace, Name: infrastructureRef.Name}, &azureCluster)
				if err != nil {
					return nil, microerror.Mask(err)
				}

				blocks := azureCluster.Spec.NetworkSpec.Vnet.CIDRBlocks
				if len(blocks) > 0 {
					clusterCIDR = blocks[0]
				}
			case "OpenStackCluster":
				var infraCluster capov1alpha4.OpenStackCluster
				err = r.k8sClient.CtrlClient().Get(ctx, client.ObjectKey{Namespace: infrastructureRef.Namespace, Name: infrastructureRef.Name}, &infraCluster)
				if err != nil {
					return nil, microerror.Mask(err)
				}

				subnetID = infraCluster.Status.Network.Subnet.ID
				networkID = infraCluster.Status.Network.ID
			default:
				r.logger.Debugf(ctx, "unable to extract clusterCIDR for cluster. Unsupported infrastructure kind %q", infrastructureRef.Kind)
			}
		}
	}

	var provider string
	{
		infrastructureRef := cr.Spec.InfrastructureRef

		switch infrastructureRef.Kind {
		case "AWSCluster":
			provider = "aws"

		case "AzureCluster":
			provider = "azure"

		case "OpenStackCluster":
			provider = "openstack"

		default:
			provider = "unknown"
		}
	}

	configMapSpecs := []configMapSpec{
		{
			Name:      key.ClusterConfigMapName(&cr),
			Namespace: key.ClusterID(&cr),
			Values: map[string]interface{}{
				"baseDomain": key.BaseDomain(&cr, r.baseDomain),
				"chartOperator": map[string]interface{}{
					"cni": map[string]interface{}{
						"install": true,
					},
				},
				"cluster": map[string]interface{}{
					"calico": map[string]interface{}{
						"CIDR": podCIDR,
					},
					"kubernetes": map[string]interface{}{
						"API": map[string]interface{}{
							"clusterIPRange": r.clusterIPRange,
						},
						"DNS": map[string]interface{}{
							"IP": r.dnsIP,
						},
					},
				},
				"clusterDNSIP": r.dnsIP,
				"clusterID":    key.ClusterID(&cr),
				"clusterCIDR":  clusterCIDR,
				"provider":     provider,
				"subnetID":     subnetID,
				"networkID":    networkID,
			},
		},
	}

	for _, spec := range configMapSpecs {
		configMap, err := newConfigMap(cr, spec)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		configMaps = append(configMaps, configMap)
	}

	return configMaps, nil
}

func newConfigMap(cr apiv1alpha3.Cluster, configMapSpec configMapSpec) (*corev1.ConfigMap, error) {
	yamlValues, err := yaml.Marshal(configMapSpec.Values)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      configMapSpec.Name,
			Namespace: configMapSpec.Namespace,
			Annotations: map[string]string{
				annotation.Notes: fmt.Sprintf("DO NOT EDIT. Values managed by %s.", project.Name()),
			},
			Labels: map[string]string{
				label.Cluster:   key.ClusterID(&cr),
				label.ManagedBy: project.Name(),
			},
		},
		Data: map[string]string{
			"values": string(yamlValues),
		},
	}

	return cm, nil
}
