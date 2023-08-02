package collector

import (
	"context"
	"fmt"

	"github.com/giantswarm/apiextensions-application/api/v1alpha1"
	"github.com/giantswarm/k8sclient/v7/pkg/k8sclient"
	"github.com/giantswarm/k8smetadata/pkg/label"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/v8/pkg/controller"
	"github.com/prometheus/client_golang/prometheus"
	k8slabels "k8s.io/apimachinery/pkg/labels"
	capi "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/cluster-apps-operator/v2/pkg/project"
)

const (
	labelClusterID = "cluster_id"
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
