package privatecluster

import (
	"github.com/giantswarm/microerror"
)

var fieldNotFoundOnInfrastructureTypeError = &microerror.Error{
	Kind: "fieldNotFoundOnInfrastructureType",
}

// IsFieldNotFoundOnInfrastructureType asserts fieldNotFoundOnInfrastructureTypeError.
func IsFieldNotFoundOnInfrastructureType(err error) bool {
	return microerror.Cause(err) == fieldNotFoundOnInfrastructureTypeError
}
