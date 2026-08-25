package collector

import (
	"context"
	"fmt"

	"github.com/giantswarm/apiextensions-application/api/v1alpha1"
	"github.com/giantswarm/k8sclient/v8/pkg/k8sclient"
	"github.com/giantswarm/k8smetadata/pkg/label"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/v7/pkg/controller"
	"github.com/prometheus/client_golang/prometheus"
	k8slabels "k8s.io/apimachinery/pkg/labels"
	capi "sigs.k8s.io/cluster-api/api/core/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/cluster-apps-operator/v3/pkg/project"
	"github.com/giantswarm/cluster-apps-operator/v3/service/controller/key"
)

const (
	labelClusterID        = "cluster_id"
	labelClusterNamespace = "cluster_namespace"
)

var (
	danglingApps *prometheus.Desc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, subsystem, "dangling_apps"),
		"Number of apps not yet deleted for a terminating cluster.",
		[]string{
			labelClusterID,
		},
		nil,
	)

	serviceCIDRMissing *prometheus.Desc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, subsystem, "service_cidr_missing"),
		"1 if the Cluster CR has no spec.clusterNetwork.services.cidrBlocks, so clusterDNSIP falls back to the installation default.",
		[]string{
			labelClusterID,
			labelClusterNamespace,
		},
		nil,
	)
)

type ClusterConfig struct {
	K8sClient k8sclient.Interface
	Logger    micrologger.Logger
}

type Cluster struct {
	context   context.Context
	k8sClient k8sclient.Interface
	logger    micrologger.Logger
}

func NewCluster(config ClusterConfig) (*Cluster, error) {
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	np := &Cluster{
		context:   context.Background(),
		k8sClient: config.K8sClient,
		logger:    config.Logger,
	}

	return np, nil
}

func (c *Cluster) Collect(ch chan<- prometheus.Metric) error {
	var clusterList capi.ClusterList
	{
		err := c.k8sClient.CtrlClient().List(
			c.context,
			&clusterList,
		)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	for _, cl := range clusterList.Items {
		// Emitted for every cluster, not only terminating ones. When the
		// services CIDR is absent the clusterconfigmap resource silently falls
		// back to the installation default clusterDNSIP, which chart-operator
		// then uses as its only resolver. See giantswarm/giantswarm#37031.
		serviceCIDRIsMissing := 0.0
		if key.ServiceCIDR(cl) == "" {
			serviceCIDRIsMissing = 1
		}

		ch <- prometheus.MustNewConstMetric(
			serviceCIDRMissing,
			prometheus.GaugeValue,
			serviceCIDRIsMissing,
			cl.GetName(),
			cl.GetNamespace(),
		)

		if cl.DeletionTimestamp.IsZero() || !hasFinalizer(cl.GetFinalizers()) {
			continue
		}

		dangling, err := c.getNumberOfApps(cl.GetName(), cl.GetNamespace())
		if err != nil {
			return microerror.Mask(err)
		}

		ch <- prometheus.MustNewConstMetric(
			danglingApps,
			prometheus.GaugeValue,
			float64(dangling),
			cl.GetName(),
		)

	}

	return nil
}

func (c *Cluster) Describe(ch chan<- *prometheus.Desc) error {
	ch <- danglingApps
	ch <- serviceCIDRMissing

	return nil
}

func (c *Cluster) getNumberOfApps(name, namespace string) (int, error) {
	var appList v1alpha1.AppList
	{
		var err error

		selector, err := k8slabels.Parse(fmt.Sprintf("%s=%s,%s!=%s", label.Cluster, name, label.ManagedBy, project.Name()))
		if err != nil {
			return -1, microerror.Mask(err)
		}

		o := client.ListOptions{
			Namespace:     namespace,
			LabelSelector: selector,
		}

		err = c.k8sClient.CtrlClient().List(c.context, &appList, &o)
		if err != nil {
			return -1, microerror.Mask(err)
		}
	}

	return len(appList.Items), nil
}

func hasFinalizer(finalizers []string) bool {
	for _, f := range finalizers {
		if f == controller.GetFinalizerName(project.Name()+"-cluster-controller") {
			return true
		}
	}

	return false
}
