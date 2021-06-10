package usertest

import (
	"context"

	"github.com/ViBiOh/ketchup/pkg/model"
	"github.com/ViBiOh/ketchup/pkg/store/user"
)

var _ user.App = &App{}

// App mock app
type App struct {
	doAtomicErr error

	getByEmailUser model.User
	getByEmailErr  error

	getByLoginIDUser model.User
	getByLoginIDErr  error

	createID  uint64
	createErr error

	count    uint64
	countErr error
}

// New creates raw mock
func New() *App {
	return &App{}
}

// SetDoAtomic mocks
func (a *App) SetDoAtomic(err error) *App {
	a.doAtomicErr = err

	return a
}

// SetGetByEmail mocks
func (a *App) SetGetByEmail(user model.User, err error) *App {
	a.getByEmailUser = user
	a.getByEmailErr = err

	return a
}

// SetGetByLoginID mocks
func (a *App) SetGetByLoginID(user model.User, err error) *App {
	a.getByLoginIDUser = user
	a.getByLoginIDErr = err

	return a
}

// SetCreate mocks
func (a *App) SetCreate(id uint64, err error) *App {
	a.createID = id
	a.createErr = err

	return a
}

// SetCount mocks
func (a *App) SetCount(count uint64, err error) *App {
	a.count = count
	a.countErr = err

	return a
}

// DoAtomic mocks
func (a *App) DoAtomic(ctx context.Context, action func(context.Context) error) error {
	if ctx == context.TODO() {
		return a.doAtomicErr
	}
	return action(ctx)
}

// GetByEmail mocks
func (a *App) GetByEmail(ctx context.Context, email string) (model.User, error) {
	return a.getByEmailUser, a.getByEmailErr
}

// GetByLoginID mocks
func (a *App) GetByLoginID(ctx context.Context, loginID uint64) (model.User, error) {
	return a.getByLoginIDUser, a.getByLoginIDErr
}

// Create mocks
func (a *App) Create(ctx context.Context, o model.User) (uint64, error) {
	return a.createID, a.createErr
}

// Count mocks
func (a *App) Count(ctx context.Context) (uint64, error) {
	return a.count, a.countErr
}
