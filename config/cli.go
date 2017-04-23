package config

import (
	"fmt"
	"os"

	docopt "github.com/docopt/docopt-go"
)

var version = "unknown"
var gitrev = "unknown"

const usage = `kubecrt - convert Helm charts to Kubernetes resources

Usage:
  kubecrt [options] CHARTS_CONFIG
  kubecrt -h | --help
  kubecrt --version
  kubecrt --config-docs

Where CHARTS_CONFIG is the location of the YAML file
containing the Kubernetes Charts configuration.

Arguments:
  CHARTS_CONFIG           Charts configuration file

Options:
  -h, --help              Show this message
  --version               Display the kubecrt version
  --namespace=<ns>        Sets the .Release.Namespace chart variable, used by
	                        Charts during compilation
  -n, --name=<name>       Sets the .Release.Name chart variable, used by charts
	                        during compilation
  -o, --output=<path>     Write output to a file, instead of stdout
  --config-docs           Show extended documentation on the Charts
                          configuration file

`

// CLI returns the parsed commandline arguments
func CLI() map[string]interface{} {
	arguments, err := docopt.Parse(usage, nil, true, "kubecrt "+version+" ("+gitrev+")", true)
	if err != nil {
		panic(err)
	}

	if arguments["--config-docs"].(bool) {
		generateConfigDocs()
		os.Exit(0)
	}

	return arguments
}

func generateConfigDocs() {
	fmt.Println(docs)
}

const docs = `
The Chart Configuration file has the following structure:

    apiVersion: v1
    charts:
      stable/factorio:
        resources:
          requests:
            memory: 1024Mi
            cpu: 750m
        factorioServer:
      	  name: {{ MY_SERVER_NAME | default:"hello world!" }}

      stable/minecraft:
        minecraftServer:
          difficulty: hard

Each Chart configuration starts with the chart location (a local path, or a
"Chart Repository" location), followed by the configuration for that chart,
which overrides the default configuration.

For the above example, see here for the default configurations:

  * stable/factorio: https://git.io/v9Tyr
  * stable/minecraft: https://git.io/v9Tya

The Chart Configuration file can also contain templated language which is
processed by epp (https://github.com/blendle/epp).

In  the above example, the "MY_SERVER_NAME" value is expanded using your
exported environment variables. If none is found, "hello world!" will be the
default name.

epp uses Pongo2 (https://github.com/flosch/pongo2) for its templating
functionality.
`
