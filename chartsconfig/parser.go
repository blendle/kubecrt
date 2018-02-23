package chartsconfig

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/blendle/kubecrt/chart"
	"github.com/blendle/kubecrt/config"
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
		return nil, wrapError(out, err)
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
	tpls, err := loadTemplates(b, partialPath)
	if err != nil {
		return nil, err
	}

	chart := &hchart.Chart{
		Metadata: &hchart.Metadata{
			Name: "kubecrt",
		},
		Templates: tpls,
	}

	return chart, nil
}

func loadTemplates(b []byte, partialPath string) ([]*hchart.Template, error) {
	tpls := []*hchart.Template{{Data: b, Name: "charts.yml"}}

	if partialPath == config.DefaultPartialTemplatesPath {
		if _, err := os.Stat(partialPath); os.IsNotExist(err) {
			return tpls, nil
		}
	}

	if partialPath == "" {
		return tpls, nil
	}

	err := filepath.Walk(partialPath, func(path string, f os.FileInfo, err error) error {
		if info, e := os.Stat(path); e == nil && info.IsDir() {
			return nil
		}

		content, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}

		tpls = append(tpls, &hchart.Template{Data: content, Name: strings.Replace(path, partialPath, "", 1)})
		return nil
	})

	return tpls, err
}

func wrapError(b []byte, err error) error {
	var lines []string
	shortLines := []string{"\n"}

	el, str := errorLine(err)

	scanner := bufio.NewScanner(bytes.NewReader(b))
	l := 0
	for scanner.Scan() {
		l++
	}
	ll := len(strconv.Itoa(l))

	i := 1
	scanner = bufio.NewScanner(bytes.NewReader(b))
	for scanner.Scan() {
		line := fmt.Sprintf("%*d: %s", ll, i, scanner.Text())

		// if we know the error line, we create an extra summary of the context
		// surrounding the error itself, starting 3 lines before, ending 3 after.
		if el != 0 {
			if i == el {
				line = "\x1b[31;1m" + line + "\x1b[0m"
			}

			if (i >= el-3) && (i <= el+3) {
				shortLines = append(shortLines, line)
			}
		}

		lines = append(lines, line)
		i++
	}

	lines = append(lines, shortLines...)
	lines = append(lines, "\n"+str)

	return errors.New(strings.Join(lines, "\n"))
}

func errorLine(err error) (int, string) {
	var i int
	var p []string
	str := err.Error()

	println(str)

	if strings.HasPrefix(str, "yaml: ") {
		p = strings.SplitN(str, ":", 3)
		i, _ = strconv.Atoi(strings.Replace(p[1], " line ", "", -1))
		str = strings.TrimSpace(p[2])
	}

	if strings.HasPrefix(str, "template: test:") {
		p = strings.SplitN(str, ":", 4)
		i, _ = strconv.Atoi(p[2])
		str = strings.TrimSpace(p[3])
	}

	return i, "Templating error: " + str
}
