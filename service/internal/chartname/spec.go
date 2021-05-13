package chartname

import (
	"context"
)

type Interface interface {
	// ChartName returns the name of the chart as it appears in the index.yaml
	// for the catalog.
	ChartName(ctx context.Context, catalogName, appName, appVersion string) (string, error)
}
