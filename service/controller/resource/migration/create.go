package migration

import (
	"context"
	"time"

	"github.com/giantswarm/apiextensions-application/api/v1alpha1"
	"github.com/giantswarm/backoff"
	"github.com/giantswarm/microerror"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/cluster-apps-operator/v2/service/controller/key"
)

func (r Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	cr, err := key.ToCluster(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	if !key.IsFluxBackendRequested(cr) {
		return nil
	}

	chartAppCR := &v1alpha1.App{
		TypeMeta: metav1.TypeMeta{
			Kind:       "App",
			APIVersion: "application.giantswarm.io",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      key.ChartOperatorAppName(&cr),
			Namespace: cr.GetNamespace(),
		},
	}

	err = r.ctrlClient.Delete(ctx, chartAppCR)
	if err != nil && !apierrors.IsNotFound(err) {
		return microerror.Mask(err)
	}

	o := func() error {
		err := r.ctrlClient.Get(ctx, client.ObjectKey{
			Name:      chartAppCR.Name,
			Namespace: chartAppCR.Namespace,
		}, chartAppCR)
		if apierrors.IsNotFound(err) {
			r.logger.LogCtx(ctx, "level", "debug", "message", "Deleted Chart Operator in favour of using the Flux Helm Controller")
			return nil
		}

		return microerror.Maskf(notDeletedError, "Chart Operator still persists")
	}
	n := func(err error, t time.Duration) {
		r.logger.Errorf(ctx, err, "retrying in %s", t)
	}

	b := backoff.NewConstant(15*time.Second, 5*time.Second)
	err = backoff.RetryNotify(o, b, n)

	return err
}
