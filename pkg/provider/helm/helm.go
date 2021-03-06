package helm

import (
	"context"
	"fmt"
	"io"

	"github.com/ViBiOh/httputils/v4/pkg/logger"
	"github.com/ViBiOh/httputils/v4/pkg/request"
	"github.com/ViBiOh/ketchup/pkg/model"
	"github.com/ViBiOh/ketchup/pkg/semver"
	"gopkg.in/yaml.v3"
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
	LatestVersions(string, string, []string) (map[string]semver.Version, error)
	FetchIndex(string, map[string][]string) (map[string]map[string]semver.Version, error)
}

type app struct{}

// New creates new App from Config
func New() App {
	return app{}
}

func (a app) FetchIndex(url string, chartsPatterns map[string][]string) (map[string]map[string]semver.Version, error) {
	resp, err := request.New().Get(fmt.Sprintf("%s/%s", url, indexName)).Send(context.Background(), nil)
	if err != nil {
		return nil, fmt.Errorf("unable to request repository: %w", err)
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			logger.Error("unable to close response body: %s", err)
		}
	}()

	var index charts
	if err := yaml.NewDecoder(resp.Body).Decode(&index); err != nil && err != io.EOF {
		return nil, fmt.Errorf("unable to parse yaml index: %s", err)
	}

	output := make(map[string]map[string]semver.Version, len(index.Entries))
	for key, charts := range index.Entries {
		patterns, ok := chartsPatterns[key]
		if !ok {
			continue
		}

		versions, compiledPatterns, err := model.PreparePatternMatching(patterns)
		if err != nil {
			return nil, fmt.Errorf("unable to prepare pattern matching for `%s`: %s", key, err)
		}

		for _, chart := range charts {
			chartVersion, err := semver.Parse(chart.Version)
			if err != nil {
				continue
			}

			model.CheckPatternsMatching(versions, compiledPatterns, chartVersion)
		}

		output[key] = versions
	}

	return output, nil
}

func (a app) LatestVersions(name, part string, patterns []string) (map[string]semver.Version, error) {
	index, err := a.FetchIndex(name, map[string][]string{part: patterns})
	if err != nil {
		return nil, err
	}

	charts, ok := index[part]
	if !ok {
		return nil, fmt.Errorf("no chart `%s` in repository", part)
	}

	return charts, nil
}
