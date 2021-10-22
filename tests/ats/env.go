//go:build functional || smoke
// +build functional smoke

package ats

import (
	"fmt"
	"os"
)

const (
	EnvVarKubeConfigPath = "ATS_KUBE_CONFIG_PATH"
)

var (
	kubeConfigPath string
)

func init() {
	kubeConfigPath = os.Getenv(EnvVarKubeConfigPath)
	if kubeConfigPath == "" {
		panic(fmt.Sprintf("env var '%s' must not be empty", EnvVarKubeConfigPath))
	}
}

func KubeConfigPath() string {
	return kubeConfigPath
}
