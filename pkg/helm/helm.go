package helm

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/ViBiOh/httputils/v3/pkg/request"
	"github.com/ViBiOh/ketchup/pkg/semver"
	"gopkg.in/yaml.v2"
)

const (
	indexName = "index.yaml"
)

type charts struct {
	Entries map[string][]chart `yaml:"entries"`
}

type chart struct {
	Version string `yaml:"version"`
}

// App of package
type App interface {
	LatestVersion(string) (semver.Version, error)
}

type app struct{}

// New creates new App from Config
func New() App {
	return app{}
}

func (a app) LatestVersion(repository string) (semver.Version, error) {
	var version semver.Version

	parts := strings.SplitN(repository, "@", 2)
	if len(parts) != 2 {
		return version, errors.New("invalid name for helm chart")
	}

	resp, err := request.New().Get(fmt.Sprintf("%s/%s", parts[1], indexName)).Send(context.Background(), nil)
	if err != nil {
		return version, fmt.Errorf("unable to request repository: %s", err)
	}

	payload, err := request.ReadBodyResponse(resp)
	if err != nil {
		return version, fmt.Errorf("unable to read body `%s`: %s", payload, err)
	}

	var index charts
	if err := yaml.Unmarshal(payload, &index); err != nil {
		return version, fmt.Errorf("unable to parse index: %s", err)
	}

	charts, ok := index.Entries[parts[0]]
	if !ok {
		return version, fmt.Errorf("no chart `%s` in repository", parts[0])
	}

	for _, chart := range charts {
		chartVersion, err := semver.Parse(chart.Version)
		if err == nil && chartVersion.IsGreater(version) {
			version = chartVersion
		}
	}

	return version, nil
}
