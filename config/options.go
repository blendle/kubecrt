package config

import "errors"

// CLIOptions ...
type CLIOptions struct {
	ChartsConfigPath    string
	ChartsConfigOptions *ChartsConfigOptions
}

// ChartsConfigOptions ...
type ChartsConfigOptions struct {
	Name      string
	Namespace string
}

// NewCLIOptions ...
func NewCLIOptions(cli map[string]interface{}) (*CLIOptions, error) {
	path, ok := cli["CHARTS_CONFIG"].(string)
	if !ok {
		return nil, errors.New("Invalid argument: CHARTS_CONFIG")
	}

	name, _ := cli["--name"].(string)
	namespace, _ := cli["--namespace"].(string)

	c := &CLIOptions{
		ChartsConfigPath: path,
		ChartsConfigOptions: &ChartsConfigOptions{
			Name:      name,
			Namespace: namespace,
		},
	}

	return c, nil
}
