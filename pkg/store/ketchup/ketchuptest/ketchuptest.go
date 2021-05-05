package ketchuptest

import (
	"context"

	"github.com/ViBiOh/ketchup/pkg/model"
	"github.com/ViBiOh/ketchup/pkg/store/ketchup"
)

var _ ketchup.App = &App{}

// App mocks
type App struct {
	doAtomicErr error

	listKetchups []model.Ketchup
	listTotal    uint64
	listErr      error

	listByRepositoriesIDKetchups []model.Ketchup
	listByRepositoriesIDErr      error

	getByRepositoryKetchup model.Ketchup
	getByRepositoryErr     error

	createID  uint64
	createErr error

	updateErr error

	deleteErr error
}

// New create new instances
func New() *App {
	return &App{}
}

// SetDoAtomic mocks
func (a *App) SetDoAtomic(err error) *App {
	a.doAtomicErr = err

	return a
}

// SetList mocks
func (a *App) SetList(ketchups []model.Ketchup, total uint64, err error) *App {
	a.listKetchups = ketchups
	a.listTotal = total
	a.listErr = err

	return a
}

// SetListByRepositoriesID mocks
func (a *App) SetListByRepositoriesID(ketchups []model.Ketchup, err error) *App {
	a.listByRepositoriesIDKetchups = ketchups
	a.listByRepositoriesIDErr = err

	return a
}

// SetGetByRepositoryID mocks
func (a *App) SetGetByRepositoryID(ketchup model.Ketchup, err error) *App {
	a.getByRepositoryKetchup = ketchup
	a.getByRepositoryErr = err

	return a
}

// SetCreate mocks
func (a *App) SetCreate(id uint64, err error) *App {
	a.createID = id
	a.createErr = err

	return a
}

// SetUpdate mocks
func (a *App) SetUpdate(err error) *App {
	a.updateErr = err

	return a
}

// SetDelete mocks
func (a *App) SetDelete(err error) *App {
	a.deleteErr = err

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
func (a *App) List(_ context.Context, page, pageSize uint) ([]model.Ketchup, uint64, error) {
	return a.listKetchups, a.listTotal, a.listErr
}

// ListByRepositoriesID mocks
func (a *App) ListByRepositoriesID(_ context.Context, _ []uint64, _ model.KetchupFrequency) ([]model.Ketchup, error) {
	return a.listByRepositoriesIDKetchups, a.listByRepositoriesIDErr
}

// GetByRepositoryID mocks
func (a *App) GetByRepositoryID(_ context.Context, _ uint64, _ bool) (model.Ketchup, error) {
	return a.getByRepositoryKetchup, a.getByRepositoryErr
}

// Create mocks
func (a *App) Create(_ context.Context, o model.Ketchup) (uint64, error) {
	return a.createID, a.createErr
}

// Update mocks
func (a *App) Update(_ context.Context, o model.Ketchup) error {
	return a.updateErr
}

// Delete mocks
func (a *App) Delete(_ context.Context, o model.Ketchup) error {
	return a.deleteErr
}
