package parser

import (
	"errors"

	"github.com/blendle/epp/epp"
	"github.com/blendle/kubecrt/config"
	yaml "gopkg.in/yaml.v2"
)

// ChartsConfiguration ...
type ChartsConfiguration struct {
	APIVersion string `yaml:"apiVersion"`
	Charts     []*Chart
	Name       string `yaml:"name"`
	Namespace  string `yaml:"namespace"`
}

// Chart ...
type Chart struct {
	Location string
	Config   interface{}
}

// ParseConfig ...
func ParseConfig(input []byte, opts *config.ChartsConfigOptions) (*ChartsConfiguration, error) {
	out, err := parseEpp(input)
	if err != nil {
		return nil, err
	}

	m := make(map[string]interface{})

	err = yaml.Unmarshal(out, m)
	if err != nil {
		return nil, err
	}

	c := NewChartsConfiguration(m, opts)

	err = validateConfig(c)
	if err != nil {
		return nil, err
	}

	return c, nil
}

// NewChartsConfiguration ...
func NewChartsConfiguration(m map[string]interface{}, opts *config.ChartsConfigOptions) *ChartsConfiguration {
	var apiVersion string
	var charts map[interface{}]interface{}

	apiVersion, _ = m["apiVersion"].(string)

	namespace := opts.Namespace
	if namespace == "" {
		namespace, _ = m["namespace"].(string)
	}

	name := opts.Name
	if name == "" {
		name, _ = m["name"].(string)
	}

	cc := &ChartsConfiguration{
		APIVersion: apiVersion,
		Name:       name,
		Namespace:  namespace,
	}

	charts, _ = m["charts"].(map[interface{}]interface{})
	for key, config := range charts {
		location, _ := key.(string)

		c := &Chart{
			Location: location,
			Config:   config,
		}

		cc.Charts = append(cc.Charts, c)
	}

	return cc
}

func parseEpp(input []byte) ([]byte, error) {
	out, err := epp.Parse(input)
	if err != nil {
		return nil, err
	}

	return out, nil
}

func validateConfig(cc *ChartsConfiguration) error {
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

	if len(cc.Charts) == 0 {
		return errors.New("Missing charts, you need to define at least one chart")
	}

	for _, c := range cc.Charts {
		if c.Location == "" {
			return errors.New("Invalid or missing chart name")
		}
	}

	return nil
}
