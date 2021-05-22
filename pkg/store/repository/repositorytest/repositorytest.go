package repositorytest

import (
	"context"

	"github.com/ViBiOh/ketchup/pkg/model"
	"github.com/ViBiOh/ketchup/pkg/store/repository"
)

var _ repository.App = &App{}

// App mocks
type App struct {
	doAtomicErr error

	listRepositories []model.Repository
	listTotal        uint64
	listErr          error

	listByKindsRepositories []model.Repository
	listByKindsTotal        uint64
	listByKindsErr          error

	suggestRepositories []model.Repository
	suggestErr          error

	getRepository model.Repository
	getErr        error

	getByNameRepository model.Repository
	getByNameErr        error

	createID  uint64
	createErr error

	updateVersionsErr error

	deleteUnusedErr error

	deleteUnusedVersionsErr error
}

// New creates mock instance
func New() *App {
	return &App{}
}

// SetDoAtomic mocks
func (a *App) SetDoAtomic(err error) *App {
	a.doAtomicErr = err

	return a
}

// SetList mocks
func (a *App) SetList(repositories []model.Repository, total uint64, err error) *App {
	a.listRepositories = repositories
	a.listTotal = total
	a.listErr = err

	return a
}

// SetListByKinds mocks
func (a *App) SetListByKinds(repositories []model.Repository, total uint64, err error) *App {
	a.listByKindsRepositories = repositories
	a.listByKindsTotal = total
	a.listByKindsErr = err

	return a
}

// SetSuggest mocks
func (a *App) SetSuggest(repositories []model.Repository, err error) *App {
	a.suggestRepositories = repositories
	a.suggestErr = err

	return a
}

// SetGet mocks
func (a *App) SetGet(repository model.Repository, err error) *App {
	a.getRepository = repository
	a.getErr = err

	return a
}

// SetGetByName mocks
func (a *App) SetGetByName(repository model.Repository, err error) *App {
	a.getByNameRepository = repository
	a.getByNameErr = err

	return a
}

// SetCreate mocks
func (a *App) SetCreate(id uint64, err error) *App {
	a.createID = id
	a.createErr = err

	return a
}

// SetUpdateVersions mocks
func (a *App) SetUpdateVersions(err error) *App {
	a.updateVersionsErr = err

	return a
}

// SetDeleteUnused mocks
func (a *App) SetDeleteUnused(err error) *App {
	a.deleteUnusedErr = err

	return a
}

// SetDeleteUnusedVersions mocks
func (a *App) SetDeleteUnusedVersions(err error) *App {
	a.deleteUnusedVersionsErr = err

	return a
}

// DoAtomic mocks
func (a *App) DoAtomic(ctx context.Context, action func(context.Context) error) error {
	if ctx == context.TODO() {
		return a.doAtomicErr
	}

	return action(ctx)
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
	return a.suggestRepositories, a.suggestErr
}

// Get mocks
func (a *App) Get(_ context.Context, _ uint64, _ bool) (model.Repository, error) {
	return a.getRepository, a.getErr
}

// GetByName mocks
func (a *App) GetByName(_ context.Context, _ model.RepositoryKind, _, _ string) (model.Repository, error) {
	return a.getByNameRepository, a.getByNameErr
}

// Create mocks
func (a *App) Create(_ context.Context, o model.Repository) (uint64, error) {
	return a.createID, a.createErr
}

// UpdateVersions mocks
func (a *App) UpdateVersions(_ context.Context, o model.Repository) error {
	return a.updateVersionsErr
}

// DeleteUnused mocks
func (a *App) DeleteUnused(_ context.Context) error {
	return a.deleteUnusedErr
}

// DeleteUnusedVersions mocks
func (a *App) DeleteUnusedVersions(_ context.Context) error {
	return a.deleteUnusedVersionsErr
}
