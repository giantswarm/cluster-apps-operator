//go:build k8srequired
// +build k8srequired

package flux

import "github.com/giantswarm/microerror"

var executionFailedError = &microerror.Error{
	Kind: "executionFailedError",
}
