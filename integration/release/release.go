//go:build k8srequired
// +build k8srequired

package release

import (
	"context"
	"fmt"
	"time"

	"github.com/giantswarm/apiextensions-application/api/v1alpha1"
	"github.com/giantswarm/backoff"
	"github.com/giantswarm/k8sclient/v7/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/giantswarm/cluster-apps-operator/v2/pkg/project"
)

type Config struct {
	K8sClients k8sclient.Interface
	Logger     micrologger.Logger
}

type Release struct {
	k8sClients k8sclient.Interface
	logger     micrologger.Logger
}

func New(config Config) (*Release, error) {
	if config.K8sClients == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	r := &Release{
		k8sClients: config.K8sClients,
		logger:     config.Logger,
	}

	return r, nil
}

func (r *Release) WaitForAppCreate(ctx context.Context, namespace, name string) error {
	var app v1alpha1.App
	var err error

	o := func() error {
		err = r.k8sClients.CtrlClient().Get(
			ctx,
			types.NamespacedName{Namespace: namespace, Name: name},
			&app,
		)
		if err != nil {
			return microerror.Mask(err)
		}

		return nil
	}

	n := func(err error, t time.Duration) {
		r.logger.Log("level", "debug", "message", fmt.Sprintf("failed to get app created '%s': retrying in %s", name, t), "stack", fmt.Sprintf("%v", err))
	}

	b := backoff.NewExponential(10*time.Minute, 60*time.Second)
	err = backoff.RetryNotify(o, b, n)
	if err != nil {
		return microerror.Mask(err)
	}
	return nil
}

func (r *Release) WaitForAppDelete(ctx context.Context, namespace, name string) error {
	var app v1alpha1.App
	var err error

	o := func() error {
		err = r.k8sClients.CtrlClient().Get(
			ctx,
			types.NamespacedName{Namespace: namespace, Name: name},
			&app,
		)
		if apierrors.IsNotFound(err) {
			// Fall through.
			return nil
		} else if err != nil {
			return microerror.Mask(err)
		}

		return waitError
	}

	n := func(err error, t time.Duration) {
		r.logger.Log("level", "debug", "message", fmt.Sprintf("failed to get app deleted '%s': retrying in %s", name, t), "stack", fmt.Sprintf("%v", err))
	}

	b := backoff.NewExponential(10*time.Minute, 60*time.Second)
	err = backoff.RetryNotify(o, b, n)
	if err != nil {
		return microerror.Mask(err)
	}
	return nil
}

func (r *Release) WaitForReadyDeployment(ctx context.Context, namespace string) error {
	var err error

	o := func() error {
		lo := metav1.ListOptions{
			LabelSelector: fmt.Sprintf("app=%s", project.Name()),
		}
		deploys, err := r.k8sClients.K8sClient().AppsV1().Deployments(namespace).List(ctx, lo)
		if err != nil {
			return microerror.Mask(err)
		}
		if len(deploys.Items) != 1 {
			return microerror.Maskf(executionFailedError, "expected 1 deployment got %d", len(deploys.Items))
		}

		deploy := deploys.Items[0]
		if *deploy.Spec.Replicas != deploy.Status.ReadyReplicas {
			return microerror.Maskf(executionFailedError, "expected %d ready pods got %d", *deploy.Spec.Replicas, deploy.Status.ReadyReplicas)
		}

		return nil
	}

	n := func(err error, t time.Duration) {
		r.logger.Log("level", "debug", "message", fmt.Sprintf("faild to get deployment for %s", t), "stack", fmt.Sprintf("%v", err))
	}

	err = backoff.RetryNotify(o, backoff.NewConstant(5*time.Minute, 15*time.Second), n)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}
