package helm

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/ViBiOh/httputils/v3/pkg/request"
	"github.com/ViBiOh/ketchup/pkg/model"
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
	LatestVersions(string, []string) (map[string]semver.Version, error)
}

type app struct{}

// New creates new App from Config
func New() App {
	return app{}
}

func (a app) LatestVersions(repository string, patterns []string) (map[string]semver.Version, error) {
	parts := strings.SplitN(repository, "@", 2)
	if len(parts) != 2 {
		return nil, errors.New("invalid name for helm chart")
	}

	resp, err := request.New().Get(fmt.Sprintf("%s/%s", parts[1], indexName)).Send(context.Background(), nil)
	if err != nil {
		return nil, fmt.Errorf("unable to request repository: %w", err)
	}

	payload, err := request.ReadBodyResponse(resp)
	if err != nil {
		return nil, fmt.Errorf("unable to read body `%s`: %s", payload, err)
	}

	var index charts
	if err := yaml.Unmarshal(payload, &index); err != nil {
		return nil, fmt.Errorf("unable to parse index: %w", err)
	}

	charts, ok := index.Entries[parts[0]]
	if !ok {
		return nil, fmt.Errorf("no chart `%s` in repository", parts[0])
	}

	versions, compiledPatterns, err := model.PreparePatternMatching(patterns)
	if err != nil {
		return nil, fmt.Errorf("unable to prepare pattern matching: %s", err)
	}

	for _, chart := range charts {
		chartVersion, err := semver.Parse(chart.Version)
		if err != nil {
			continue
		}

		model.CheckPatternsMatching(versions, compiledPatterns, chartVersion)
	}

	return versions, nil
}
