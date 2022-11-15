package clustersecret

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"reflect"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apiv1alpha3 "sigs.k8s.io/cluster-api/api/v1alpha4"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"

	"github.com/giantswarm/k8smetadata/pkg/annotation"
	"github.com/giantswarm/k8smetadata/pkg/label"
	"github.com/giantswarm/microerror"

	capi "sigs.k8s.io/cluster-api/api/v1alpha4"

	capvcd "github.com/giantswarm/cluster-apps-operator/v2/api/capvcd/v1beta1"

	capo "github.com/giantswarm/cluster-apps-operator/v2/api/capo/v1alpha4"
	"github.com/giantswarm/cluster-apps-operator/v2/pkg/project"
	"github.com/giantswarm/cluster-apps-operator/v2/service/controller/key"
)

const (
	mainSecretSection      = "values"
	containerdProxySection = "containerdProxy"

	containerdProxyTemplate = `[Service]
Environment="HTTP_PROXY={{ .HttpProxy }}"
Environment="http_proxy={{ .HttpProxy }}"
Environment="HTTPS_PROXY={{ .HttpsProxy }}"
Environment="https_proxy={{ .HttpsProxy }}"
Environment="NO_PROXY={{ .NoProxy}}"
Environment="no_proxy={{ .NoProxy }}"
`
)

func (r *Resource) GetDesiredState(ctx context.Context, obj interface{}) ([]*corev1.Secret, error) {
	cr, err := key.ToCluster(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var secrets []*corev1.Secret

	if key.IsDeleted(&cr) {
		r.logger.Debugf(ctx, "deleting cluster secrets for cluster '%s/%s'", cr.GetNamespace(), key.ClusterID(&cr))
		return secrets, nil
	}

	values := map[string]interface{}{}

	{
		infrastructureRef := cr.Spec.InfrastructureRef
		if infrastructureRef != nil {
			switch infrastructureRef.Kind {
			case "OpenStackCluster":
				var infraCluster capo.OpenStackCluster
				err = r.k8sClient.CtrlClient().Get(ctx, client.ObjectKey{Namespace: infrastructureRef.Namespace, Name: infrastructureRef.Name}, &infraCluster)
				if err != nil {
					return nil, microerror.Mask(err)
				}

				values["cloudConfig"], err = r.generateOpenStackCloudConfig(ctx, infraCluster)
				if err != nil {
					return nil, microerror.Mask(err)
				}
			case "VCDCluster":
				var infraCluster capvcd.VCDCluster
				err = r.k8sClient.CtrlClient().Get(ctx, client.ObjectKey{Namespace: infrastructureRef.Namespace, Name: infrastructureRef.Name}, &infraCluster)
				if err != nil {
					return nil, microerror.Mask(err)
				}

				values["global"], err = r.generateCloudDirectorConfig(ctx, infraCluster)
				if err != nil {
					return nil, microerror.Mask(err)
				}
			}
		}
	}

	secretSpecs := []secretSpec{
		{
			Name:      key.ClusterValuesResourceName(&cr),
			Namespace: cr.GetNamespace(),
			Data:      map[string][]byte{},
		},
	}

	if !reflect.ValueOf(r.proxy).IsZero() {
		r.logger.Debugf(ctx, "proxy secrets for cluster '%s/%s' : %v", cr.GetNamespace(), key.ClusterID(&cr), r.proxy)

		values["cluster"] = map[string]interface{}{
			"proxy": map[string]string{
				"noProxy": noProxy(cr, r.proxy.NoProxy),
				"http":    r.proxy.HttpProxy,
				"https":   r.proxy.HttpsProxy,
			},
		}

		// template containerd proxy configuration
		t := template.Must(template.New("systemd-proxy-template").Parse(containerdProxyTemplate))
		var tpl bytes.Buffer
		if err := t.Execute(&tpl, r.proxy); err != nil {
			return nil, err
		}

		containerdProxy := tpl.String()

		secretSpecs = append(secretSpecs, secretSpec{
			Name:      fmt.Sprintf("%s-systemd-proxy", key.ClusterID(&cr)),
			Namespace: cr.GetNamespace(),
			Data: map[string][]byte{
				containerdProxySection: []byte(containerdProxy),
			},
		})
	}

	yamlValues, err := yaml.Marshal(values)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	secretSpecs[0].Data[mainSecretSection] = []byte(yamlValues)

	for _, spec := range secretSpecs {
		secret := newSecret(cr, spec)
		secrets = append(secrets, secret)
	}

	return secrets, nil
}

func newSecret(cr apiv1alpha3.Cluster, secretSpec secretSpec) *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretSpec.Name,
			Namespace: secretSpec.Namespace,
			Annotations: map[string]string{
				annotation.Notes: fmt.Sprintf("DO NOT EDIT. Values managed by %s.", project.Name()),
			},
			Labels: map[string]string{
				label.Cluster:   key.ClusterID(&cr),
				label.ManagedBy: project.Name(),
			},
		},
		Data: secretSpec.Data,
	}
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
