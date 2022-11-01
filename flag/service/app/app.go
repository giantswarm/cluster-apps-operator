package app

import (
	"github.com/giantswarm/cluster-apps-operator/v2/flag/service/app/appoperator"
	"github.com/giantswarm/cluster-apps-operator/v2/flag/service/app/chartoperator"
	"github.com/giantswarm/cluster-apps-operator/v2/flag/service/app/observabilitybundle"
)

type App struct {
	AppOperator         appoperator.AppOperator
	ChartOperator       chartoperator.ChartOperator
	ObservabilityBundle observabilitybundle.ObservabilityBundle
}
