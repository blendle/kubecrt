package config

import (
	"fmt"
	"os"

	docopt "github.com/docopt/docopt-go"
)

var version = "unknown"
var gitrev = "unknown"

const usage = `kubecrt - convert Helm charts to Kubernetes resources

Given a charts.yml file, compile all the charts with
the provided template variables and return the
resulting Kubernetes resource files to STDOUT, or
write them to a provided file.

Doing this, you can use Kubernetes charts, without
having to use Helm locally, or Tiller on the server.

Usage:
  kubecrt [options] CHARTS_CONFIG
  kubecrt -h | --help
  kubecrt --version
  kubecrt --example-config

Where CHARTS_CONFIG is the location of the YAML file
containing the Kubernetes Charts configuration.

Arguments:
  CHARTS_CONFIG                    Charts configuration file

Options:
  -h, --help                       Show this screen
  --version                        Show version
  -n NS, --namespace=NS            Set the .Release.Namespace chart variable,
                                   used by charts during compilation
  -a NAME, --name=NAME             Set the .Release.Name chart variable, used by
                                   charts during compilation
  -o PATH, --output=PATH           Write output to a file, instead of STDOUT
  -r NAME=URL, --repo=NAME=URL,... List of NAME=URL pairs of repositories to add
                                   to the index before compiling charts config
  --example-config                 Print an example charts.yaml, including
                                   extended documentation on the tunables
`

// CLI returns the parsed command-line arguments
func CLI() map[string]interface{} {
	arguments, err := docopt.Parse(usage, nil, true, "kubecrt "+version+" ("+gitrev+")", true)
	if err != nil {
		panic(err)
	}

	if arguments["--example-config"].(bool) {
		generateExampleConfig()
		os.Exit(0)
	}

	return arguments
}

func generateExampleConfig() {
	fmt.Println(docs)
}

const docs = `
# apiVersion defines the version of the charts.yaml structure. Currently,
# only "v1" is supported.
apiVersion: v1

# name is the .Release.Name template value that charts can use in their
# templates, which can be overridden by the "--name" CLI flag. If omitted,
# "--name" is required.
name: my-bundled-apps

# namespace is the .Release.Namespace template value that charts can use in
# their templates. Note that since kubecrt does not communicate with
# Kubernetes in any way, it is up to you to also use this namespace when
# doing kubectl apply [...]. Can be overridden using "--namespace".  If omitted,
# "--namespace" is required.
namespace: apps

# charts is an array of charts you want to compile into Kubernetes resource
# files.
#
# A single chart might be used to deploy something simple, like a memcached pod,
# or something complex, like a full web app stack with HTTP servers, databases,
# caches, and so on.
charts:

# A Chart can either be in the format REPO/NAME, or a PATH to a local chart.
#
# If using REPO/NAME, kubecrt knows by-default where to locate the "stable"
# repository, all other repositories require the "repo" configuration (see
# below).
- stable/factorio:
    # values is a map of key/value pairs used when compiling the chart. This
    # uses the same format as in regular chart "values.yaml" files.
    #
    # see: https://git.io/v9Tyr
    values:
      resources:
        requests:
          memory: 1024Mi
          cpu: 750m
      factorioServer:
        # charts.yaml supports the same templating as chart templates do,
        # using the "sprig" library.
        #
        # see: https://masterminds.github.io/sprig/
        name: {{ env "MY_SERVER_NAME" | default "hello world!" }}

- stable/minecraft:
    # version is a semantic version constraint.
    #
    # see: https://github.com/Masterminds/semver#basic-comparisons
    version: ~> 0.1.0
    values:
      minecraftServer:
        difficulty: hard

- opsgoodness/prometheus-operator:
    # repo is the location of a repositry, if other than "stable". This is
    # the URL you would normally add using "helm repo add NAME URL".
    repo: http://charts.opsgoodness.com
    values:
      sendAnalytics: false

# For the above charts, see here for the default configurations:
#
#   * stable/factorio: https://git.io/v9Tyr
#   * stable/minecraft: https://git.io/v9Tya
#   * opsgoodness/prometheus-operator: https://git.io/v9SAY
`
