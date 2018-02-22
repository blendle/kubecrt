package chartsconfig

import (
	"errors"
	"html/template"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/blendle/kubecrt/chart"
	yaml "gopkg.in/yaml.v2"
	"k8s.io/helm/pkg/engine"
	hchart "k8s.io/helm/pkg/proto/hapi/chart"
)

// ChartsConfiguration ...
type ChartsConfiguration struct {
	APIVersion string                    `yaml:"apiVersion"`
	Name       string                    `yaml:"name"`
	Namespace  string                    `yaml:"namespace"`
	ChartsMap  []map[string]*chart.Chart `yaml:"charts"`
	ChartsList []*chart.Chart
}

// NewChartsConfiguration initializes a new ChartsConfiguration.
func NewChartsConfiguration(input []byte, tpath string) (*ChartsConfiguration, error) {
	m := &ChartsConfiguration{}

	renderer := engine.New()

	funcs := template.FuncMap{
		"env":       func(s string) string { return os.Getenv(s) },
		"expandenv": func(s string) string { return os.ExpandEnv(s) },
	}

	for k, v := range funcs {
		renderer.FuncMap[k] = v
	}

	t, err := stubChart(input, tpath)
	if err != nil {
		return nil, err
	}

	tpls, err := renderer.Render(t, map[string]interface{}{})
	if err != nil {
		return nil, err
	}

	out := []byte(tpls["kubecrt/charts.yml"])

	if err = yaml.Unmarshal(out, m); err != nil {
		return nil, err
	}

	for _, a := range m.ChartsMap {
		for loc, c := range a {
			c.Location = loc
			m.ChartsList = append(m.ChartsList, c)
		}
	}

	return m, nil
}

// ParseCharts loops through all charts, and returns the parsed resources.
func (cc *ChartsConfiguration) ParseCharts() ([]byte, error) {
	var out []byte

	for _, c := range cc.ChartsList {
		resources, err := c.ParseChart(cc.Name, cc.Namespace)
		if err != nil {
			return nil, err
		}

		out = append(out, resources...)
	}

	return out, nil
}

// Validate makes sure the charts configuration is configured as expected.
func (cc *ChartsConfiguration) Validate() error {
	if cc.APIVersion == "" {
		return errors.New("Missing API version, please add \"apiVersion: v1\"")
	}

	if cc.APIVersion != "v1" {
		return errors.New("Unknown API version, please set apiVersion to \"v1\"")
	}

	if cc.Name == "" {
		return errors.New("Missing name, please add \"name: my-app-name\" or pass \"--name=my-app-name\"")
	}

	if cc.Namespace == "" {
		return errors.New("Missing namespace, please add \"namespace: my-namespace\" or pass \"--namespace=my-namespace\"")
	}

	if len(cc.ChartsList) == 0 {
		return errors.New("Missing charts, you need to define at least one chart")
	}

	for _, c := range cc.ChartsList {
		if c.Location == "" {
			return errors.New("Invalid or missing chart name")
		}

		if c.Version != "" {
			if _, err := semver.NewConstraint(c.Version); err != nil {
				return errors.New(c.Version + ": " + err.Error())
			}
		}
	}

	return nil
}

func stubChart(b []byte, partialPath string) (*hchart.Chart, error) {
	tpls := []*hchart.Template{{Data: []byte(b), Name: "charts.yml"}}

	if partialPath != "" {
		err := filepath.Walk(partialPath, func(path string, f os.FileInfo, err error) error {
			if info, err := os.Stat(path); err == nil && info.IsDir() {
				return nil
			}

			content, err := ioutil.ReadFile(path)
			if err != nil {
				return err
			}

			tpls = append(tpls, &hchart.Template{Data: content, Name: strings.Replace(path, partialPath, "", 1)})
			return nil
		})

		if err != nil {
			return nil, err
		}
	}

	chart := &hchart.Chart{
		Metadata: &hchart.Metadata{
			Name: "kubecrt",
		},
		Templates: tpls,
	}

	return chart, nil
}
