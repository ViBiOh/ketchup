package githubtest

import (
	"errors"
	"regexp"

	"github.com/ViBiOh/ketchup/pkg/github"
	"github.com/ViBiOh/ketchup/pkg/semver"
)

var _ github.App = app{}

// NewApp creates mock
func NewApp(name *regexp.Regexp, version string) github.App {
	return app{
		name:    name,
		version: version,
	}
}

type app struct {
	name    *regexp.Regexp
	version string
}

func (a app) LatestVersion(repository string) (semver.Version, error) {
	if a.name.MatchString(repository) {
		version, _ := semver.Parse(a.version)
		return version, nil
	}

	return semver.NoneVersion, errors.New("unknown repository")
}
