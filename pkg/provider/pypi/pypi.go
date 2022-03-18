package pypi

import (
	"context"
	"fmt"

	"github.com/ViBiOh/httputils/v4/pkg/httpjson"
	"github.com/ViBiOh/httputils/v4/pkg/request"
	"github.com/ViBiOh/ketchup/pkg/model"
	"github.com/ViBiOh/ketchup/pkg/semver"
)

const (
	registryURL = "https://pypi.org/pypi"
)

type packageResp struct {
	Versions map[string]any `json:"releases"`
}

// App of package
type App struct{}

// New creates new App
func New() App {
	return App{}
}

// LatestVersions retrieves latest version for package name on given patterns
func (a App) LatestVersions(ctx context.Context, name string, patterns []string) (map[string]semver.Version, error) {
	versions, compiledPatterns, err := model.PreparePatternMatching(patterns)
	if err != nil {
		return nil, fmt.Errorf("unable to prepare pattern matching: %s", err)
	}

	resp, err := request.Get(fmt.Sprintf("%s/%s/json", registryURL, name)).Send(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("unable to fetch registry: %s", err)
	}

	var content packageResp
	if err := httpjson.Read(resp, &content); err != nil {
		return nil, fmt.Errorf("unable to read versions: %s", err)
	}

	for version := range content.Versions {
		tagVersion, err := semver.Parse(version)
		if err != nil {
			continue
		}

		model.CheckPatternsMatching(versions, compiledPatterns, tagVersion)
	}

	return versions, nil
}
