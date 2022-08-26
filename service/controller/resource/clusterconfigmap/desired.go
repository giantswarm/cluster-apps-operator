package clusterconfigmap

import (
	"context"
	"fmt"

	"github.com/giantswarm/k8smetadata/pkg/annotation"
	"github.com/giantswarm/k8smetadata/pkg/label"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/v7/pkg/controller/context/resourcecanceledcontext"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	capi "sigs.k8s.io/cluster-api/api/v1alpha4"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"

	capz "github.com/giantswarm/cluster-apps-operator/api/capz/v1alpha4"
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

	var clusterCA string
	{
		var secret corev1.Secret
		err = r.k8sClient.CtrlClient().Get(ctx, client.ObjectKey{
			Namespace: cr.Namespace,
			Name:      key.ClusterCAName(&cr),
		}, &secret)
		if apierrors.IsNotFound(err) {
			// During cluster creation there may be a delay until the
			// ca is created.
			r.logger.Debugf(ctx, "secret '%s/%s' not found, cannot get cluster CA", cr.Namespace, key.ClusterCAName(&cr))
		} else if err != nil {
			return nil, microerror.Mask(err)
		}

		clusterCA = string(secret.Data["tls.crt"])
	}

	provider := r.provider
	clusterCIDR := ""
	serviceCidr := r.dnsIP
	region := ""
	gcpProject := ""

	{
		infrastructureRef := cr.Spec.InfrastructureRef
		if infrastructureRef != nil {
			switch infrastructureRef.Kind {
			case "AWSCluster":
				provider = "aws"

			case "AzureCluster":
				provider = "azure"

				var azureCluster capz.AzureCluster
				err = r.k8sClient.CtrlClient().Get(ctx, client.ObjectKey{Namespace: infrastructureRef.Namespace, Name: infrastructureRef.Name}, &azureCluster)
				if err != nil {
					return nil, microerror.Mask(err)
				}

				blocks := azureCluster.Spec.NetworkSpec.Vnet.CIDRBlocks
				if len(blocks) > 0 {
					clusterCIDR = blocks[0]
				}
			case "GCPCluster":
				provider = "gcp"

				gcpCluster := &unstructured.Unstructured{}
				gcpCluster.SetGroupVersionKind(schema.GroupVersionKind{
					Group:   infrastructureRef.GroupVersionKind().Group,
					Kind:    infrastructureRef.Kind,
					Version: infrastructureRef.GroupVersionKind().Version,
				})
				err = r.k8sClient.CtrlClient().Get(ctx, client.ObjectKey{
					Namespace: infrastructureRef.Namespace,
					Name:      infrastructureRef.Name,
				}, gcpCluster)
				if err != nil {
					return nil, microerror.Mask(err)
				}

				gcpProject, _, err = unstructured.NestedString(gcpCluster.Object, []string{"spec", "project"}...)
				if err != nil {
					return nil, microerror.Mask(err)
				}

				region, _, err = unstructured.NestedString(gcpCluster.Object, []string{"spec", "region"}...)
				if err != nil {
					return nil, microerror.Mask(err)
				}

				kubeadmControlPlane := &unstructured.Unstructured{}
				kubeadmControlPlane.SetGroupVersionKind(schema.GroupVersionKind{
					Group:   cr.Spec.ControlPlaneRef.GroupVersionKind().Group,
					Kind:    cr.Spec.ControlPlaneRef.Kind,
					Version: cr.Spec.ControlPlaneRef.GroupVersionKind().Version,
				})
				err = r.k8sClient.CtrlClient().Get(ctx, client.ObjectKey{
					Namespace: cr.Spec.ControlPlaneRef.Namespace,
					Name:      cr.Spec.ControlPlaneRef.Name,
				}, kubeadmControlPlane)
				if err != nil {
					return nil, microerror.Mask(err)
				}

				serviceCidr, _, err = unstructured.NestedString(gcpCluster.Object, []string{"spec", "kubeadmConfigSpec", "clusterConfiguration", "networking", "serviceSubnet"}...)
				if err != nil {
					return nil, microerror.Mask(err)
				}

			case "OpenStackCluster":
				provider = "openstack"

			case "VSphereCluster":
				provider = "vsphere"

			default:
				r.logger.Debugf(ctx, "unable to extract infrastructure provider-specific clusterValues for cluster. Unsupported infrastructure kind %q", infrastructureRef.Kind)
			}
		}
	}

	appOperatorValues := map[string]interface{}{
		"app": map[string]interface{}{
			"watchNamespace":    cr.GetNamespace(),
			"workloadClusterID": key.ClusterID(&cr),
		},
		"provider": map[string]interface{}{
			"kind": provider,
		},
		"registry": map[string]interface{}{
			"domain": r.registryDomain,
		},
	}

	clusterValues := map[string]interface{}{
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
					"IP": serviceCidr,
				},
			},
		},
		"clusterCA":    clusterCA,
		"clusterDNSIP": r.dnsIP,
		"clusterID":    key.ClusterID(&cr),
		"clusterCIDR":  clusterCIDR,
		"provider":     "unknown",
		"region":       region,
		"gcpProject":   gcpProject,
	}

	configMapSpecs := []configMapSpec{
		{
			Name:      key.AppOperatorValuesResourceName(&cr),
			Namespace: cr.GetNamespace(),
			Values:    appOperatorValues,
		},
		{
			Name:      key.ClusterValuesResourceName(&cr),
			Namespace: cr.GetNamespace(),
			Values:    clusterValues,
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

func newConfigMap(cr capi.Cluster, configMapSpec configMapSpec) (*corev1.ConfigMap, error) {
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
