package collector

import (
	"github.com/giantswarm/microerror"
)

var invalidConfigError = &microerror.Error{
	Kind: "invalidConfigError",
}

func IsInvalidConfig(err error) bool {
	return microerror.Cause(err) == invalidConfigError
}
