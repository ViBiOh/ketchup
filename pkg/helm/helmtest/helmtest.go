package helmtest

import (
	"github.com/ViBiOh/ketchup/pkg/helm"
	"github.com/ViBiOh/ketchup/pkg/semver"
)

var _ helm.App = &App{}

// App mock app
type App struct {
	latestVersions    map[string]semver.Version
	latestVersionsErr error
}

// New creates mock
func New() *App {
	return &App{}
}

// SetLatestVersions mock
func (a *App) SetLatestVersions(latestVersions map[string]semver.Version, err error) *App {
	a.latestVersions = latestVersions
	a.latestVersionsErr = err

	return a
}

// LatestVersions mock
func (a App) LatestVersions(_ string, _ []string) (map[string]semver.Version, error) {
	return a.latestVersions, a.latestVersionsErr
}
