package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/blendle/kubecrt/chartsconfig"
	"github.com/blendle/kubecrt/config"
	"github.com/blendle/kubecrt/helm"
)

func main() {
	cli := config.CLI()
	opts, err := config.NewCLIOptions(cli)
	if err != nil {
		fmt.Fprintf(os.Stderr, "kubecrt arguments error: \n\n%s\n", err)
		os.Exit(1)
	}

	if err = helm.Init(); err != nil {
		fmt.Fprintf(os.Stderr, "error initialising helm: \n\n%s\n", err)
		os.Exit(1)
	}

	if err = helm.AddRepository("stable", "https://kubernetes-charts.storage.googleapis.com"); err != nil {
		fmt.Fprintf(os.Stderr, "error adding repository: \n\n%s\n", err)
		os.Exit(1)
	}

	if cli["--repo"] != nil {
		for _, r := range strings.Split(cli["--repo"].(string), ",") {
			p := strings.SplitN(r, "=", 2)
			repo := strings.TrimSpace(string(p[0]))
			url := strings.TrimSpace(string(p[1]))

			if err = helm.AddRepository(repo, url); err != nil {
				fmt.Fprintf(os.Stderr, "error adding repository: \n\n%s\n", err)
				os.Exit(1)
			}
		}
	}

	cfg, err := readInput(opts.ChartsConfigurationPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "charts config IO error: \n\n%s\n", err)
		os.Exit(1)
	}

	cc, err := chartsconfig.NewChartsConfiguration(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "charts config parsing error: \n\n%s\n", err)
		os.Exit(1)
	}

	name := opts.ChartsConfigurationOptions.Name
	if name != "" {
		cc.Name = name
	}

	namespace := opts.ChartsConfigurationOptions.Namespace
	if namespace != "" {
		cc.Namespace = namespace
	}

	if err = cc.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "charts validation error: \n\n%s\n", err)
		os.Exit(1)
	}

	out, err := cc.ParseCharts()
	if err != nil {
		fmt.Fprintf(os.Stderr, "chart parsing error: %s\n", err)
		os.Exit(1)
	}

	if cli["--output"] == nil {
		fmt.Print(string(out))
		return
	}

	if err = ioutil.WriteFile(cli["--output"].(string), out, 0644); err != nil {
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
