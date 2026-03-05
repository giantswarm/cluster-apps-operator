package integration

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/giantswarm/cluster-apps-operator/v3/test/harness"
)

func TestMain(m *testing.M) {
	if err := harness.InitManager(context.Background()); err != nil {
		fmt.Fprintf(os.Stderr, "failed to initialize test harness: %v\n", err)
		os.Exit(1)
	}
	code := m.Run()
	harness.ShutdownManager()
	os.Exit(code)
}
