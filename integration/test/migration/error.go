//go:build k8srequired
// +build k8srequired

package migration

import "github.com/giantswarm/microerror"

var executionFailedError = &microerror.Error{
	Kind: "executionFailedError",
}
