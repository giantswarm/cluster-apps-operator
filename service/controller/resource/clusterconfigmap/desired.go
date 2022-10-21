package clusterconfigmap

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/giantswarm/k8smetadata/pkg/annotation"
	"github.com/giantswarm/k8smetadata/pkg/label"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/v7/pkg/controller/context/resourcecanceledcontext"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"

	capi "sigs.k8s.io/cluster-api/api/v1alpha4"

	capz "github.com/giantswarm/cluster-apps-operator/v2/api/capz/v1alpha4"
	"github.com/giantswarm/cluster-apps-operator/v2/flag/service/proxy"
	"github.com/giantswarm/cluster-apps-operator/v2/pkg/project"
	"github.com/giantswarm/cluster-apps-operator/v2/service/controller/key"
	"github.com/giantswarm/cluster-apps-operator/v2/service/internal/podcidr"
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

	// clusterDNSIP contains the `coredns` k8s `Service` IP.
	// This IP needs to belong to the `Services` CIDR configured for the k8s cluster, which can be set in the
	// "serviceSubnet" field of the KubeadmControlPlane CR. If this field is set we want to take the IP from that CIDR.
	// If it's not, we take the IP from the CIDR passed as parameter, which will probably be the default Service CIDR.
	var clusterDNSIP = r.dnsIP
	{
		if cr.Spec.ControlPlaneRef != nil {
			kubeadmControlPlane := &unstructured.Unstructured{}
			kubeadmControlPlane.SetGroupVersionKind(schema.GroupVersionKind{
				Group:   cr.Spec.ControlPlaneRef.GroupVersionKind().Group,
				Kind:    cr.Spec.ControlPlaneRef.Kind,
				Version: cr.Spec.ControlPlaneRef.GroupVersionKind().Version,
			})
			err = r.k8sClient.CtrlClient().Get(ctx, client.ObjectKey{
				Namespace: cr.Namespace,
				Name:      cr.Spec.ControlPlaneRef.Name,
			}, kubeadmControlPlane)
			if err != nil {
				return nil, microerror.Mask(err)
			}

			serviceCidr, serviceCidrFound, err := unstructured.NestedString(kubeadmControlPlane.Object, []string{"spec", "kubeadmConfigSpec", "clusterConfiguration", "networking", "serviceSubnet"}...)
			if err != nil {
				return nil, microerror.Mask(err)
			}

			if serviceCidrFound {
				clusterDNSIP, err = key.DNSIP(serviceCidr)
				if err != nil {
					return nil, microerror.Mask(err)
				}
			}
		}
	}

	var (
		// clusterCIDR is only used on azure.
		clusterCIDR = ""
		// gcpProject is only used on gcp.
		gcpProject      = ""
		gcpProjectFound bool
	)
	{
		infrastructureRef := cr.Spec.InfrastructureRef
		if infrastructureRef != nil {
			switch r.provider {
			case "azure":
				var azureCluster capz.AzureCluster
				err = r.k8sClient.CtrlClient().Get(ctx, client.ObjectKey{Namespace: infrastructureRef.Namespace, Name: infrastructureRef.Name}, &azureCluster)
				if err != nil {
					return nil, microerror.Mask(err)
				}

				blocks := azureCluster.Spec.NetworkSpec.Vnet.CIDRBlocks
				if len(blocks) > 0 {
					clusterCIDR = blocks[0]
				}
			case "aws":
			case "openstack":
			case "vsphere":
			case "gcp":
				gcpCluster := &unstructured.Unstructured{}
				gcpCluster.SetGroupVersionKind(schema.GroupVersionKind{
					Group:   infrastructureRef.GroupVersionKind().Group,
					Kind:    infrastructureRef.Kind,
					Version: infrastructureRef.GroupVersionKind().Version,
				})
				err = r.k8sClient.CtrlClient().Get(ctx, client.ObjectKey{
					Namespace: cr.Namespace,
					Name:      infrastructureRef.Name,
				}, gcpCluster)
				if err != nil {
					return nil, microerror.Mask(err)
				}

				gcpProject, gcpProjectFound, err = unstructured.NestedString(gcpCluster.Object, []string{"spec", "project"}...)
				if err != nil || !gcpProjectFound {
					return nil, fieldNotFoundOnInfrastructureTypeError
				}
			default:
				r.logger.Debugf(ctx, "unable to extract infrastructure provider-specific clusterValues for cluster. Unsupported infrastructure kind %q", r.provider)
			}
		}
	}

	appOperatorValues := map[string]interface{}{
		"app": map[string]interface{}{
			"watchNamespace":    cr.GetNamespace(),
			"workloadClusterID": key.ClusterID(&cr),
		},
		"provider": map[string]interface{}{
			"kind": r.provider,
		},
		"registry": map[string]interface{}{
			"domain": r.registryDomain,
		},
	}

	appValuesYaml, err := yaml.Marshal(appOperatorValues)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	appValuesConfigMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      key.AppOperatorValuesResourceName(&cr),
			Namespace: cr.GetNamespace(),
			Annotations: map[string]string{
				annotation.Notes: fmt.Sprintf("DO NOT EDIT. Values managed by %s.", project.Name()),
			},
			Labels: map[string]string{
				label.Cluster:   key.ClusterID(&cr),
				label.ManagedBy: project.Name(),
			},
		},
		Data: map[string]string{
			"values": string(appValuesYaml),
		},
	}

	clusterValues := ClusterValuesConfig{
		BaseDomain: key.BaseDomain(&cr, r.baseDomain),
		BootstrapMode: ChartOperatorBootstrapMode{
			Enabled:          true,
			ApiServerPodPort: 6443,
		},
		ChartOperator: ChartOperatorConfig{Cni: map[string]bool{"install": true}},
		Cluster: ClusterConfig{
			Calico: map[string]string{"CIDR": podCIDR},
			Kubernetes: KubernetesConfig{
				API: map[string]string{"clusterIPRange": r.clusterIPRange},
				DNS: map[string]string{"IP": clusterDNSIP},
			},
			Proxy: proxy.Proxy{
				HttpProxy:  r.proxy.HttpProxy,
				HttpsProxy: r.proxy.HttpsProxy,
				NoProxy:    noProxy(cr, r.proxy.NoProxy),
			},
		},
		ClusterCA:    clusterCA,
		ClusterDNSIP: clusterDNSIP,
		ClusterID:    key.ClusterID(&cr),
		ClusterCIDR:  clusterCIDR,
		GcpProject:   gcpProject,
		Provider:     r.provider,
	}

	clusterValuesYaml, err := yaml.Marshal(clusterValues)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	clusterValuesConfigMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      key.ClusterValuesResourceName(&cr),
			Namespace: cr.GetNamespace(),
			Annotations: map[string]string{
				annotation.Notes: fmt.Sprintf("DO NOT EDIT. Values managed by %s.", project.Name()),
			},
			Labels: map[string]string{
				label.Cluster:   key.ClusterID(&cr),
				label.ManagedBy: project.Name(),
			},
		},
		Data: map[string]string{
			"values": string(clusterValuesYaml),
		},
	}
	configMaps = append(configMaps, appValuesConfigMap, clusterValuesConfigMap)

	return configMaps, nil
}

func noProxy(cluster capi.Cluster, globalNoProxy string) string {

	// generic list of noProxy
	// will be joined with custom defined noProxy targets

	var appendString []string
	if !reflect.ValueOf(cluster.Spec.ClusterNetwork).IsZero() {
		if !reflect.ValueOf(cluster.Spec.ClusterNetwork.ServiceDomain).IsZero() {
			appendString = append(appendString, cluster.Spec.ClusterNetwork.ServiceDomain)
		}

		if !reflect.ValueOf(cluster.Spec.ClusterNetwork.Services).IsZero() && !reflect.ValueOf(cluster.Spec.ClusterNetwork.Services.CIDRBlocks).IsZero() {
			appendString = append(appendString, strings.Join(cluster.Spec.ClusterNetwork.Services.CIDRBlocks, ","))
		}

		if !reflect.ValueOf(cluster.Spec.ClusterNetwork.Pods).IsZero() && !reflect.ValueOf(cluster.Spec.ClusterNetwork.Pods.CIDRBlocks).IsZero() {
			appendString = append(appendString, strings.Join(cluster.Spec.ClusterNetwork.Pods.CIDRBlocks, ","))
		}
	}

	if !reflect.ValueOf(cluster.Spec.ControlPlaneEndpoint.Host).IsZero() {
		appendString = append(appendString, cluster.Spec.ControlPlaneEndpoint.Host)
	}

	if len(globalNoProxy) > 0 {
		appendString = append(appendString, globalNoProxy)
	}

	noProxy := strings.Join([]string{
		strings.Join(appendString, ","),
		"svc",
		"127.0.0.1",
		"localhost",
	}, ",")

	return noProxy
}
