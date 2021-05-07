package ketchuptest

import (
	"context"

	"github.com/ViBiOh/ketchup/pkg/model"
	"github.com/ViBiOh/ketchup/pkg/service/ketchup"
)

var _ ketchup.App = &App{}

// App mock app
type App struct {
	listKetchups []model.Ketchup
	listTotal    uint64
	listErr      error

	listForRepositoriesKetchups []model.Ketchup
	listForRepositoriesErr      error

	listOutdatedByFrequencyKetchups []model.Ketchup
	listOutdatedByFrequencyErr      error
}

// New creates raw mock
func New() *App {
	return &App{}
}

// SetList mocks
func (a *App) SetList(list []model.Ketchup, total uint64, err error) *App {
	a.listKetchups = list
	a.listTotal = total
	a.listErr = err

	return a
}

// SetListForRepositories mocks
func (a *App) SetListForRepositories(list []model.Ketchup, err error) *App {
	a.listForRepositoriesKetchups = list
	a.listForRepositoriesErr = err

	return a
}

// SetListOutdatedByFrequency mocks
func (a *App) SetListOutdatedByFrequency(list []model.Ketchup, err error) *App {
	a.listOutdatedByFrequencyKetchups = list
	a.listOutdatedByFrequencyErr = err

	return a
}

// List mocks
func (a App) List(_ context.Context, _, _ uint) ([]model.Ketchup, uint64, error) {
	return a.listKetchups, a.listTotal, a.listErr
}

// ListForRepositories mocks
func (a App) ListForRepositories(_ context.Context, _ []model.Repository, _ model.KetchupFrequency) ([]model.Ketchup, error) {
	return a.listForRepositoriesKetchups, a.listForRepositoriesErr
}

// ListOutdatedByFrequency mocks
func (a App) ListOutdatedByFrequency(_ context.Context, _ model.KetchupFrequency) ([]model.Ketchup, error) {
	return a.listOutdatedByFrequencyKetchups, a.listOutdatedByFrequencyErr
}

// Create mocks
func (a App) Create(_ context.Context, _ model.Ketchup) (model.Ketchup, error) {
	return model.NoneKetchup, nil
}

// Update mocks
func (a App) Update(_ context.Context, _ model.Ketchup) (model.Ketchup, error) {
	return model.NoneKetchup, nil
}

// Delete mocks
func (a App) Delete(_ context.Context, _ model.Ketchup) error {
	return nil
}
