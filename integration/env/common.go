//go:build k8srequired
// +build k8srequired

package env

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	// EnvVarCircleCI is the process environment variable representing the
	// CIRCLECI env var.
	EnvVarCircleCI = "CIRCLECI"
	// EnvVarE2EAppVersion is the process environment variable representing the
	// E2E_APP_VERSION env var. It optionally overrides the version of the app
	// under test. When empty the version is read from the .build_version file
	// the architect orb persists to the workspace.
	EnvVarE2EAppVersion = "E2E_APP_VERSION"
	// EnvVarE2EKubeconfig is the process environment variable representing the
	// E2E_KUBECONFIG env var.
	EnvVarE2EKubeconfig = "E2E_KUBECONFIG"
	// EnvVarKeepResources is the process environment variable representing the
	// KEEP_RESOURCES env var.
	EnvVarKeepResources = "KEEP_RESOURCES"

	// buildVersionFile is written and persisted to the workspace by the
	// architect orb, and holds the version the chart under test is published
	// with.
	buildVersionFile = ".build_version"
)

var (
	appVersion    string
	circleCI      string
	keepResources string
	kubeconfig    string
)

func init() {
	circleCI = os.Getenv(EnvVarCircleCI)
	keepResources = os.Getenv(EnvVarKeepResources)

	appVersion = os.Getenv(EnvVarE2EAppVersion)
	if appVersion == "" {
		appVersion = buildVersion()
	}
	if appVersion == "" {
		panic(fmt.Sprintf("env var '%s' must not be empty when no '%s' file is present", EnvVarE2EAppVersion, buildVersionFile))
	}

	kubeconfig = os.Getenv(EnvVarE2EKubeconfig)
	if kubeconfig == "" {
		panic(fmt.Sprintf("env var '%s' must not be empty", EnvVarE2EKubeconfig))
	}
}

// AppVersion returns the version the app under test is published with in the
// test catalog.
func AppVersion() string {
	return appVersion
}

func CircleCI() bool {
	return circleCI == strings.ToLower("true")
}

func KeepResources() bool {
	return keepResources == strings.ToLower("true")
}

func KubeConfigPath() string {
	return kubeconfig
}

// buildVersion reads the version from the .build_version file. `go test` runs
// the test binary with the package directory as working directory, so we walk
// up to the repository root to find it.
func buildVersion() string {
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}

	for {
		// #nosec G304 -- the path is built from the working directory and a
		// constant file name, not from external input.
		data, err := os.ReadFile(filepath.Join(dir, buildVersionFile))
		if err == nil {
			return strings.TrimSpace(string(data))
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}
