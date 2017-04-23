package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/blendle/kubecrt/config"
	"github.com/blendle/kubecrt/parser"
)

func main() {
	cli := config.CLI()
	opts, err := config.NewCLIOptions(cli)
	if err != nil {
		fmt.Fprintf(os.Stderr, "kubecrt arguments error: \n\n%s\n", err)
		os.Exit(1)
	}

	cfg, err := readInput(opts.ChartsConfigPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "charts config IO error: \n\n%s\n", err)
		os.Exit(1)
	}

	cc, err := parser.ParseConfig(cfg, opts.ChartsConfigOptions)
	if err != nil {
		fmt.Fprintf(os.Stderr, "charts config parsing error: \n\n%s\n", err)
		os.Exit(1)
	}

	out, err := cc.ParseCharts()
	if err != nil {
		fmt.Fprintf(os.Stderr, "chart parsing error: %s\n", err)
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
