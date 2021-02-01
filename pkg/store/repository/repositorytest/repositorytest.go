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

// DoAtomic mocks
func (a *App) DoAtomic(ctx context.Context, action func(context.Context) error) error {
	if ctx == context.TODO() {
		return a.doAtomicErr
	}

	return action(ctx)
}

// List mocks
func (a *App) List(_ context.Context, _, _ uint) ([]model.Repository, uint64, error) {
	return a.listRepositories, a.listTotal, a.listErr
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
func (a *App) GetByName(_ context.Context, _ string, _ model.RepositoryKind) (model.Repository, error) {
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
