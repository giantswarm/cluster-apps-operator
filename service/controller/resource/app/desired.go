package app

import (
	"context"

	applicationv1alpha1 "github.com/giantswarm/apiextensions/v3/pkg/apis/application/v1alpha1"
)

func (r *Resource) GetDesiredState(ctx context.Context, obj interface{}) ([]*applicationv1alpha1.App, error) {
	return nil, nil
}
