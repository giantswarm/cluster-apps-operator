package clusterconfigmap

import (
	"context"

	corev1 "k8s.io/api/core/v1"
)

func (r *Resource) GetDesiredState(ctx context.Context, obj interface{}) ([]*corev1.ConfigMap, error) {
	return nil, nil
}
