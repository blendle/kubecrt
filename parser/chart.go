package parser

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	yaml "gopkg.in/yaml.v2"
	"k8s.io/helm/pkg/chartutil"
	"k8s.io/helm/pkg/engine"
	"k8s.io/helm/pkg/proto/hapi/chart"
	"k8s.io/helm/pkg/timeconv"
)

// ParseCharts ...
func (cc *ChartsConfiguration) ParseCharts() ([]byte, error) {
	var out []byte

	for _, c := range cc.Charts {
		resources, err := c.ToResources(cc.Name, cc.Namespace)
		if err != nil {
			return nil, err
		}

		out = append(out, resources...)
	}

	return out, nil
}

// ToResources ...
func (c *Chart) ToResources(name, namespace string) ([]byte, error) {
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

	resources, err := chartToResources(c.Location, name, namespace, tmpfile.Name())
	if err != nil {
		return nil, err
	}

	return resources, nil
}

func chartToResources(location, releaseName, namespace, values string) ([]byte, error) {
	var output string

	c, err := chartutil.Load(location)
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
