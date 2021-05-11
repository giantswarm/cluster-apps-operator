package appfinalizer

import (
	"context"
)

// EnsureDeleted removes finalizers for workload cluster app CRs. These are
// deleted with the cluster by the provider operator.
func (r *Resource) EnsureDeleted(ctx context.Context, obj interface{}) error {
	return nil
}
