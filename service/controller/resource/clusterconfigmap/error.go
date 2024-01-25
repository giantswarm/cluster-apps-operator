package clusterconfigmap

import "github.com/giantswarm/microerror"

var infrastructureRefNotFoundError = &microerror.Error{
	Kind: "infrastructureRefNotFoundError",
}

// IsInvalidConfig asserts invalidConfigError.
func IsInfrastructureRefNotFoundError(err error) bool {
	return microerror.Cause(err) == infrastructureRefNotFoundError
}

var invalidConfigError = &microerror.Error{
	Kind: "invalidConfigError",
}

// IsInvalidConfig asserts invalidConfigError.
func IsInvalidConfig(err error) bool {
	return microerror.Cause(err) == invalidConfigError
}

var wrongTypeError = &microerror.Error{
	Kind: "wrongTypeError",
}

// IsWrongType asserts wrongTypeError.
func IsWrongType(err error) bool {
	return microerror.Cause(err) == wrongTypeError
}

var fieldNotFoundOnInfrastructureTypeError = &microerror.Error{
	Kind: "fieldNotFoundOnInfrastructureType",
}

// IsFieldNotFoundOnInfrastructureType asserts fieldNotFoundOnInfrastructureTypeError.
func IsFieldNotFoundOnInfrastructureType(err error) bool {
	return microerror.Cause(err) == fieldNotFoundOnInfrastructureTypeError
}
