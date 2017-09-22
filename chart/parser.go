package chart

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/blendle/kubecrt/helm"

	yaml "gopkg.in/yaml.v2"
	"k8s.io/helm/pkg/chartutil"
	"k8s.io/helm/pkg/downloader"
	"k8s.io/helm/pkg/engine"
	"k8s.io/helm/pkg/getter"
	"k8s.io/helm/pkg/helm/environment"
	"k8s.io/helm/pkg/helm/helmpath"
	"k8s.io/helm/pkg/proto/hapi/chart"
	"k8s.io/helm/pkg/timeconv"
)

// Chart ...
type Chart struct {
	Version  string      `yaml:"version"`
	Repo     string      `yaml:"repo"`
	Values   interface{} `yaml:"values"`
	Location string
}

// ParseChart ...
func (c *Chart) ParseChart(name, namespace string) ([]byte, error) {
	s := strings.Split(c.Location, "/")

	if len(s) == 2 && c.Repo != "" {
		err := helm.AddRepository(s[0], c.Repo)
		if err != nil {
			return nil, err
		}
	}

	d, err := yaml.Marshal(c.Values)
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

	resources, err := c.compile(name, namespace, tmpfile.Name())
	if err != nil {
		return nil, err
	}

	return resources, nil
}

func (c *Chart) compile(releaseName, namespace, values string) ([]byte, error) {
	var output string

	location, err := locateChartPath(c.Location, c.Version)
	if err != nil {
		return nil, err
	}

	cr, err := chartutil.Load(location)
	if err != nil {
		return nil, err
	}

	vv, err := vals(values)
	if err != nil {
		return nil, err
	}

	config := &chart.Config{Raw: string(vv), Values: map[string]*chart.Value{}}

	options := chartutil.ReleaseOptions{
		Name:      releaseName,
		Time:      timeconv.Now(),
		Namespace: namespace,
	}

	renderer := engine.New()

	vals, err := chartutil.ToRenderValues(cr, config, options)
	if err != nil {
		return nil, err
	}

	out, err := renderer.Render(cr, vals)
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
		if strings.TrimSpace(data) == "" {
			continue
		}

		data = strings.TrimSpace(data)

		if !strings.HasPrefix(data, "---\n") {
			data = "---\n" + data
		}

		if output != "" {
			data = "\n\n" + data
		}

		output = output + data
	}

	return []byte(strings.Trim(output, "\n") + "\n"), nil
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

func locateChartPath(name, version string) (string, error) {
	homepath := filepath.Join(os.Getenv("HOME"), ".helm")

	name = strings.TrimSpace(name)
	version = strings.TrimSpace(version)
	if _, err := os.Stat(name); err == nil {
		abs, err := filepath.Abs(name)
		if err != nil {
			return abs, err
		}

		return abs, nil
	}

	if filepath.IsAbs(name) || strings.HasPrefix(name, ".") {
		return name, fmt.Errorf("path %q not found", name)
	}

	crepo := filepath.Join(helmpath.Home(homepath).Repository(), name)
	if _, err := os.Stat(crepo); err == nil {
		return filepath.Abs(crepo)
	}

	settings := environment.EnvSettings{
		Home: helmpath.Home(environment.DefaultHelmHome),
	}

	dl := downloader.ChartDownloader{
		HelmHome: helmpath.Home(homepath),
		Out:      os.Stdout,
		Getters:  getter.All(settings),
	}

	err := os.MkdirAll(filepath.Dir(crepo), 0755)
	if err != nil {
		return "", fmt.Errorf("Failed to untar (mkdir): %s", err)
	}

	version, err = helm.GetAcceptableVersion(name, version)
	if err != nil {
		return "", err
	}

	filename, _, err := dl.DownloadTo(name, version, filepath.Dir(crepo))
	if err != nil {
		return "", err
	}

	return filename, nil
}
