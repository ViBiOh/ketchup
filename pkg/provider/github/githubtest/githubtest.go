package githubtest

import (
	"github.com/ViBiOh/ketchup/pkg/provider/github"
	"github.com/ViBiOh/ketchup/pkg/semver"
	"github.com/prometheus/client_golang/prometheus"
)

var _ github.App = &App{}

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

// Start mock
func (a App) Start(_ prometheus.Registerer, _ <-chan struct{}) {
}
