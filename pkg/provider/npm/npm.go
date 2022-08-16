package npm

import (
	"context"
	"fmt"

	"github.com/ViBiOh/httputils/v4/pkg/httpjson"
	"github.com/ViBiOh/httputils/v4/pkg/request"
	"github.com/ViBiOh/ketchup/pkg/model"
	"github.com/ViBiOh/ketchup/pkg/semver"
)

const (
	registryURL = "https://registry.npmjs.org/"
)

type packageResp struct {
	Versions map[string]versionResp `json:"versions"`
}

type versionResp struct {
	Version string `json:"version"`
}

// App of package
type App struct{}

// New creates new App from Config
func New() App {
	return App{}
}

// LatestVersions retrieves latest version for package name on given patterns
func (a App) LatestVersions(ctx context.Context, name string, patterns []string) (map[string]semver.Version, error) {
	versions, compiledPatterns, err := model.PreparePatternMatching(patterns)
	if err != nil {
		return nil, fmt.Errorf("prepare pattern matching: %w", err)
	}

	resp, err := request.Get(fmt.Sprintf("%s/%s", registryURL, name)).Header("Accept", "application/vnd.npm.install-v1+json").Send(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("fetch registry: %w", err)
	}

	var content packageResp
	if err := httpjson.Read(resp, &content); err != nil {
		return nil, fmt.Errorf("read versions: %w", err)
	}

	for _, version := range content.Versions {
		tagVersion, err := semver.Parse(version.Version)
		if err != nil {
			continue
		}

		model.CheckPatternsMatching(versions, compiledPatterns, tagVersion)
	}

	return versions, nil
}
