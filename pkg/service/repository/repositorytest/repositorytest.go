package repositorytest

import (
	"context"

	"github.com/ViBiOh/ketchup/pkg/model"
	"github.com/ViBiOh/ketchup/pkg/semver"
	"github.com/ViBiOh/ketchup/pkg/service/repository"
)

var _ repository.App = &App{}

// App mock app
type App struct {
	listRepositories []model.Repository
	listTotal        uint64
	listErr          error

	listByKindsRepositories []model.Repository
	listByKindsTotal        uint64
	listByKindsErr          error

	getOrCreateRepo model.Repository
	getOrCreateErr  error

	updateErr error

	latestVersions    map[string]semver.Version
	latestVersionsErr error
}

// New creates raw mock
func New() *App {
	return &App{}
}

// SetList mocks
func (a *App) SetList(list []model.Repository, total uint64, err error) *App {
	a.listRepositories = list
	a.listTotal = total
	a.listErr = err

	return a
}

// SetListByKinds mocks
func (a *App) SetListByKinds(list []model.Repository, total uint64, err error) *App {
	a.listByKindsRepositories = list
	a.listByKindsTotal = total
	a.listByKindsErr = err

	return a
}

// SetGetOrCreate mocks
func (a *App) SetGetOrCreate(repo model.Repository, err error) *App {
	a.getOrCreateRepo = repo
	a.getOrCreateErr = err

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
func (a *App) List(_ context.Context, _ uint, _ string) ([]model.Repository, uint64, error) {
	return a.listRepositories, a.listTotal, a.listErr
}

// ListByKinds mocks
func (a *App) ListByKinds(_ context.Context, _ uint, _ string, _ ...model.RepositoryKind) ([]model.Repository, uint64, error) {
	return a.listByKindsRepositories, a.listByKindsTotal, a.listByKindsErr
}

// Suggest mocks
func (a *App) Suggest(_ context.Context, _ []uint64, _ uint64) ([]model.Repository, error) {
	return nil, nil
}

// GetOrCreate mocks
func (a *App) GetOrCreate(_ context.Context, _ model.RepositoryKind, _ string, _ string, _ string) (model.Repository, error) {
	return a.getOrCreateRepo, a.getOrCreateErr
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
