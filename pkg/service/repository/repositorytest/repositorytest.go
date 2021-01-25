package repositorytest

import (
	"context"
	"errors"
	"regexp"

	httpModel "github.com/ViBiOh/httputils/v3/pkg/model"
	"github.com/ViBiOh/ketchup/pkg/model"
	"github.com/ViBiOh/ketchup/pkg/semver"
	"github.com/ViBiOh/ketchup/pkg/service/repository"
)

var _ repository.App = &App{}

// App mock app
type App struct {
	multiple bool
	name     *regexp.Regexp
	version  string

	list    []model.Repository
	total   uint64
	listErr error

	updateErr error

	latestVersions    map[string]semver.Version
	latestVersionsErr error
}

// New creates raw mock
func New() *App {
	return &App{}
}

// NewApp creates mock
func NewApp(multiple bool, name *regexp.Regexp, version string) repository.App {
	return &App{
		multiple: multiple,
		name:     name,
		version:  version,
	}
}

// SetList mocks
func (a *App) SetList(list []model.Repository, total uint64, err error) *App {
	a.list = list
	a.total = total
	a.listErr = err

	return a
}

// SetUpdate mocks
func (a *App) SetUpdate(err error) *App {
	a.updateErr = err

	return a
}

// SetLatestVersions mocks
func (a *App) SetLatestVersions(latestVersions map[string]semver.Version, err error) *App {
	a.latestVersions = latestVersions
	a.latestVersionsErr = err

	return a
}

// List mocks
func (a *App) List(_ context.Context, _, _ uint) ([]model.Repository, uint64, error) {
	return a.list, a.total, a.listErr
}

// Suggest mocks
func (a *App) Suggest(_ context.Context, _ []uint64, _ uint64) ([]model.Repository, error) {
	return nil, nil
}

// GetOrCreate mocks
func (a *App) GetOrCreate(_ context.Context, name string, repositoryKind model.RepositoryKind) (model.Repository, error) {
	if len(name) == 0 {
		return model.NoneRepository, httpModel.WrapInvalid(errors.New("invalid name"))
	}

	return model.Repository{ID: 1, Name: "vibioh/ketchup", Versions: map[string]string{model.DefaultPattern: "1.2.3"}}, nil
}

// Update mocks
func (a *App) Update(_ context.Context, _ model.Repository) error {
	return a.updateErr
}

// Clean mocks
func (a *App) Clean(_ context.Context) error {
	return nil
}

// LatestVersions mocks
func (a *App) LatestVersions(_ model.Repository) (map[string]semver.Version, error) {
	return a.latestVersions, a.latestVersionsErr
}
