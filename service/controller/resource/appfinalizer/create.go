package appfinalizer

import (
	"context"
)

// EnsureCreated is not needed for this resource.
func (r *Resource) EnsureCreated(ctx context.Context, obj interface{}) error {
	return nil
}
