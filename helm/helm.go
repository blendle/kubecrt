package helm

import (
	"k8s.io/helm/pkg/helm/environment"
	"k8s.io/helm/pkg/helm/helmpath"
)

var settings environment.EnvSettings

func init() {
	settings.Home = helmpath.Home(environment.DefaultHelmHome)
}
