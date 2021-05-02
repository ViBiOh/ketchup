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

	fetchIndex    map[string]map[string]semver.Version
	fetchIndexErr error
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
func (a App) LatestVersions(_, _ string, _ []string) (map[string]semver.Version, error) {
	return a.latestVersions, a.latestVersionsErr
}

// SetFetchIndex mock
func (a *App) SetFetchIndex(charts map[string]map[string]semver.Version, err error) *App {
	a.fetchIndex = charts
	a.fetchIndexErr = err

	return a
}

// FetchIndex mock
func (a App) FetchIndex(_ string, _ map[string][]string) (map[string]map[string]semver.Version, error) {
	return a.fetchIndex, a.fetchIndexErr
}
