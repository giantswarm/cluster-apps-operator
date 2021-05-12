package key

import (
	"fmt"

	"github.com/giantswarm/k8smetadata/pkg/label"
)

func ClusterID(getter LabelsGetter) string {
	return getter.GetLabels()[label.Cluster]
}

func ReleaseName(releaseVersion string) string {
	return fmt.Sprintf("v%s", releaseVersion)
}

func ReleaseVersion(getter LabelsGetter) string {
	return getter.GetLabels()[label.ReleaseVersion]
}
