package migration

import (
	"github.com/giantswarm/microerror"
)

var invalidConfigError = &microerror.Error{
	Kind: "invalidConfigError",
}

// IsInvalidConfigError asserts invalidConfigError.
func IsInvalidConfigError(err error) bool {
	return microerror.Cause(err) == invalidConfigError
}

var notDeletedError = &microerror.Error{
	Kind: "notDeletedError",
}

// IsNotDeleted asserts notDeletedError.
func IsNotDeleted(err error) bool {
	return microerror.Cause(err) == notDeletedError
}
