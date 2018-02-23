package config

import "errors"

// DefaultPartialTemplatesPath is the default path used for partials.
const DefaultPartialTemplatesPath = "config/deploy/partials"

// CLIOptions contains all the options set through the CLI arguments
type CLIOptions struct {
	ChartsConfigurationPath    string
	PartialTemplatesPath       string
	ChartsConfigurationOptions *ChartsConfigurationOptions
}

// ChartsConfigurationOptions contains the CLI options relevant for the charts
// configuration.
type ChartsConfigurationOptions struct {
	Name      string
	Namespace string
}

// NewCLIOptions takes CLI arguments, and returns a CLIOptions struct.
func NewCLIOptions(cli map[string]interface{}) (*CLIOptions, error) {
	path, ok := cli["CHARTS_CONFIG"].(string)
	if !ok {
		return nil, errors.New("Invalid argument: CHARTS_CONFIG")
	}

	name, _ := cli["--name"].(string)
	namespace, _ := cli["--namespace"].(string)

	c := &CLIOptions{
		ChartsConfigurationPath: path,
		ChartsConfigurationOptions: &ChartsConfigurationOptions{
			Name:      name,
			Namespace: namespace,
		},
	}

	if cli["--partials-dir"] != nil {
		c.PartialTemplatesPath, _ = cli["--partials-dir"].(string)
	}

	return c, nil
}
