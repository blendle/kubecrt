package chartsconfig

import (
	"errors"

	"github.com/Masterminds/semver"
	"github.com/blendle/epp/epp"
	"github.com/blendle/kubecrt/chart"
	yaml "gopkg.in/yaml.v2"
)

// ChartsConfiguration ...
type ChartsConfiguration struct {
	APIVersion string                    `yaml:"apiVersion"`
	Name       string                    `yaml:"name"`
	Namespace  string                    `yaml:"namespace"`
	ChartsMap  []map[string]*chart.Chart `yaml:"charts"`
	ChartsList []*chart.Chart
}

// NewChartsConfiguration initialises a new ChartsConfiguration.
func NewChartsConfiguration(input []byte) (*ChartsConfiguration, error) {
	m := &ChartsConfiguration{}

	out, err := parseEpp(input)
	if err != nil {
		return nil, err
	}

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

func parseEpp(input []byte) ([]byte, error) {
	out, err := epp.Parse(input)
	if err != nil {
		return nil, err
	}

	return out, nil
}
