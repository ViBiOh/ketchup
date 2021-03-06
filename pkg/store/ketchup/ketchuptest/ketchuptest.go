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

	listOutdatedByFrequencyKetchups []model.Ketchup
	listOutdatedByFrequencyErr      error

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

// SetListOutdatedByFrequency mocks
func (a *App) SetListOutdatedByFrequency(ketchups []model.Ketchup, err error) *App {
	a.listOutdatedByFrequencyKetchups = ketchups
	a.listOutdatedByFrequencyErr = err

	return a
}

// SetGetByRepository mocks
func (a *App) SetGetByRepository(ketchup model.Ketchup, err error) *App {
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
func (a *App) List(_ context.Context, _ uint, _ string) ([]model.Ketchup, uint64, error) {
	return a.listKetchups, a.listTotal, a.listErr
}

// ListByRepositoriesID mocks
func (a *App) ListByRepositoriesID(_ context.Context, _ []uint64, _ model.KetchupFrequency) ([]model.Ketchup, error) {
	return a.listByRepositoriesIDKetchups, a.listByRepositoriesIDErr
}

// ListOutdatedByFrequency mocks
func (a *App) ListOutdatedByFrequency(_ context.Context, _ model.KetchupFrequency) ([]model.Ketchup, error) {
	return a.listByRepositoriesIDKetchups, a.listByRepositoriesIDErr
}

// GetByRepository mocks
func (a *App) GetByRepository(_ context.Context, _ uint64, _ string, _ bool) (model.Ketchup, error) {
	return a.getByRepositoryKetchup, a.getByRepositoryErr
}

// Create mocks
func (a *App) Create(_ context.Context, _ model.Ketchup) (uint64, error) {
	return a.createID, a.createErr
}

// Update mocks
func (a *App) Update(_ context.Context, _ model.Ketchup, _ string) error {
	return a.updateErr
}

// UpdateAll mocks
func (a *App) UpdateAll(_ context.Context) error {
	return nil
}

// Delete mocks
func (a *App) Delete(_ context.Context, _ model.Ketchup) error {
	return a.deleteErr
}
