package collector

import (
	"context"

	"github.com/giantswarm/apiextensions-application/api/v1alpha1"
	"github.com/giantswarm/k8sclient/v6/pkg/k8sclient"
	"github.com/giantswarm/k8smetadata/pkg/label"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/prometheus/client_golang/prometheus"
	k8slabels "k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	apiv1alpha3 "sigs.k8s.io/cluster-api/api/v1alpha3"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/cluster-apps-operator/pkg/project"
)

const (
	labelInstallation = "installation"
	labelClusterID    = "cluster_id"
)

var (
	danglingApps *prometheus.Desc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, subsystem, "dangling_apps"),
		"Number of dangling apps for a terminating cluster.",
		[]string{
			labelClusterID,
			labelInstallation,
		},
		nil,
	)
)

type ClusterConfig struct {
	K8sClient k8sclient.Interface
	Logger    micrologger.Logger
}

type Cluster struct {
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
		k8sClient: config.K8sClient,
		logger:    config.Logger,
	}

	return np, nil
}

func (c *Cluster) Collect(ch chan<- prometheus.Metric) error {
	ctx := context.Background()

	var clusterList apiv1alpha3.ClusterList
	{
		err := c.k8sClient.CtrlClient().List(
			ctx,
			&clusterList,
		)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	for _, cl := range clusterList.Items {
		if cl.DeletionTimestamp.IsZero() || hasDesiredFinalizer(cl.GetFinalizers()) {
			continue
		}

		var appList v1alpha1.AppList
		{
			var err error

			selector := k8slabels.NewSelector()
			clusterLabel, err := k8slabels.NewRequirement(label.Cluster, selection.Equals, []string{cl.GetName()})
			if err != nil {
				return microerror.Mask(err)
			}

			managedByLabel, err := k8slabels.NewRequirement(label.ManagedBy, selection.DoesNotExist, []string{})
			if err != nil {
				return microerror.Mask(err)
			}

			selector = selector.Add(*clusterLabel)
			selector = selector.Add(*managedByLabel)

			err = c.k8sClient.CtrlClient().List(
				ctx,
				&appList,
				client.InNamespace(cl.GetNamespace()),
				client.MatchingLabelsSelector{Selector: selector},
			)
			if err != nil {
				return microerror.Mask(err)
			}
		}

		ch <- prometheus.MustNewConstMetric(
			danglingApps,
			prometheus.GaugeValue,
			float64(len(appList.Items)),
			cl.GetName(),
			"test",
		)

	}

	return nil
}

func (c *Cluster) Describe(ch chan<- *prometheus.Desc) error {
	ch <- danglingApps

	return nil
}

func hasDesiredFinalizer(finalizers []string) bool {
	for _, f := range finalizers {
		if f == project.Name()+"-cluster-controller" {
			return true
		}
	}

	return false
}
