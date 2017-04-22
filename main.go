package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"k8s.io/helm/pkg/chartutil"
	"k8s.io/helm/pkg/engine"
	"k8s.io/helm/pkg/proto/hapi/chart"
	"k8s.io/helm/pkg/timeconv"

	"github.com/blendle/epp/epp"
	"github.com/docopt/docopt-go"

	yaml "gopkg.in/yaml.v2"
)

var usage = `kubecrt - convert Helm charts to Kubernetes resources

Usage:
  kubecrt [options] CHARTS_CONFIG
  kubecrt -h | --help
  kubecrt --version

Where CHARTS_CONFIG is the location of the YAML file
containing the Kubernetes Chart configurations.

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

Arguments:
  CHARTS_CONFIG           Charts configuration file

Options:
  -h, --help              Show this message
  --version               Display the kubecrt version
  --namespace=<ns>        Sets the .Release.Namespace Helm variable, which can
                          be used by Charts during compilation
  -n, --name=<name>       release name. If unspecified, it will autogenerate one
                          for you
  -o, --output=<path>     Write output to a file, instead of stdout
`

var (
	version = "unknown"
	gitrev  = "unknown"

	charts []*Chart
)

// ChartsConfiguration ...
type ChartsConfiguration struct {
	APIVersion string                 `yaml:"apiVersion"`
	Charts     map[string]interface{} `yaml:"charts"`
}

// Chart ...
type Chart struct {
	Name   string
	Config interface{}
}

func main() {
	cli, err := docopt.Parse(usage, nil, true, "kubecrt "+version+" ("+gitrev+")", true)
	if err != nil {
		panic(err)
	}

	fileContents, err := readInput(cli["CHARTS_CONFIG"].(string))
	if err != nil {
		fmt.Fprintf(os.Stderr, "charts config IO error: %s\n", err)
		os.Exit(1)
	}

	out, err := epp.Parse(fileContents)
	if err != nil {
		fmt.Fprintf(os.Stderr, "charts config templating error: %s\n", err)
		os.Exit(1)
	}

	out, err = Parse(out)
	if err != nil {
		fmt.Fprintf(os.Stderr, "charts config parsing error: %s\n", err)
		os.Exit(1)
	}

	if cli["--output"] == nil {
		fmt.Printf(string(out))
		return
	}

	err = ioutil.WriteFile(cli["--output"].(string), out, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "output IO error: %s\n", err)
		os.Exit(1)
	}
}

func readInput(input string) ([]byte, error) {
	if input == "-" {
		return ioutil.ReadAll(os.Stdin)
	}

	return ioutil.ReadFile(input)
}

// Parse ...
func Parse(input []byte) ([]byte, error) {
	c := &ChartsConfiguration{}
	err := yaml.Unmarshal(input, c)
	if err != nil {
		return nil, err
	}

	if c.APIVersion == "" {
		return nil, errors.New("Missing API version, please add \"apiVersion: v1\"")
	}

	if c.APIVersion != "v1" {
		return nil, errors.New("Unknown API version, please set apiVersion to \"v1\"")
	}

	var charts []*Chart
	for name, config := range c.Charts {
		charts = append(charts, &Chart{Name: name, Config: config})
	}

	out, err := parseCharts(charts)
	if err != nil {
		return nil, err
	}

	return out, nil
}

func parseCharts(charts []*Chart) ([]byte, error) {
	var out []byte

	for _, c := range charts {
		d, err := yaml.Marshal(c.Config)
		if err != nil {
			return nil, err
		}

		tmpfile, err := ioutil.TempFile("", "")
		if err != nil {
			return nil, err
		}
		defer os.Remove(tmpfile.Name())

		if _, err = tmpfile.Write(d); err != nil {
			return nil, err
		}

		if err = tmpfile.Close(); err != nil {
			return nil, err
		}

		data, err := parseChart(c.Name, tmpfile.Name())
		if err != nil {
			return nil, err
		}

		out = append(out, data...)
	}

	return out, nil
}

func parseChart(name, values string) ([]byte, error) {
	var output string

	c, err := chartutil.Load(name)
	if err != nil {
		return nil, err
	}

	vv, err := vals(values)
	if err != nil {
		return nil, err
	}

	config := &chart.Config{Raw: string(vv), Values: map[string]*chart.Value{}}

	options := chartutil.ReleaseOptions{
		Name:      "releaseName",
		Time:      timeconv.Now(),
		Namespace: "namespace",
	}

	renderer := engine.New()

	vals, err := chartutil.ToRenderValues(c, config, options)
	if err != nil {
		return nil, err
	}

	out, err := renderer.Render(c, vals)
	if err != nil {
		return nil, err
	}

	for name, data := range out {
		b := filepath.Base(name)
		if b == "NOTES.txt" {
			continue
		}
		if strings.HasPrefix(b, "_") {
			continue
		}

		output = output + data
	}

	return []byte(output), nil
}

func vals(valuesPath string) ([]byte, error) {
	base := map[string]interface{}{}

	bytes, err := ioutil.ReadFile(valuesPath)
	if err != nil {
		return nil, err
	}

	if err := yaml.Unmarshal(bytes, &base); err != nil {
		return nil, fmt.Errorf("failed to parse %s: %s", valuesPath, err)
	}

	return yaml.Marshal(base)
}

// Copied from Helm.
func mergeValues(dest map[string]interface{}, src map[string]interface{}) map[string]interface{} {
	for k, v := range src {
		// If the key doesn't exist already, then just set the key to that value
		if _, exists := dest[k]; !exists {
			dest[k] = v
			continue
		}
		nextMap, ok := v.(map[string]interface{})
		// If it isn't another map, overwrite the value
		if !ok {
			dest[k] = v
			continue
		}
		// If the key doesn't exist already, then just set the key to that value
		if _, exists := dest[k]; !exists {
			dest[k] = nextMap
			continue
		}
		// Edge case: If the key exists in the destination, but isn't a map
		destMap, isMap := dest[k].(map[string]interface{})
		// If the source map has a map for this key, prefer it
		if !isMap {
			dest[k] = v
			continue
		}
		// If we got to this point, it is a map in both, so merge them
		dest[k] = mergeValues(destMap, nextMap)
	}
	return dest
}
