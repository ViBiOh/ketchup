package user

import (
	"context"
	"errors"
	"fmt"
	"strings"

	authModel "github.com/ViBiOh/auth/v2/pkg/model"
	"github.com/ViBiOh/httputils/v4/pkg/logger"
	httpModel "github.com/ViBiOh/httputils/v4/pkg/model"
	"github.com/ViBiOh/ketchup/pkg/model"
)

// App of package
type App struct {
	userStore model.UserStore
	authApp   model.AuthService
}

// New creates new App from Config
func New(userStore model.UserStore, authApp model.AuthService) App {
	return App{
		userStore: userStore,
		authApp:   authApp,
	}
}

// StoreInContext read login user from context and store app user in context
func (a App) StoreInContext(ctx context.Context) context.Context {
	id := authModel.ReadUser(ctx).ID
	if id == 0 {
		logger.Warn("no login user in context")
		return ctx
	}

	item, err := a.userStore.GetByLoginID(ctx, id)
	if err != nil || item.IsZero() {
		logger.Error("unable to get user with login %d: %s", id, err)
		return ctx
	}

	return model.StoreUser(ctx, item)
}

// ListReminderUsers list users that need a reminder
func (a App) ListReminderUsers(ctx context.Context) ([]model.User, error) {
	return a.userStore.ListReminderUsers(ctx)
}

// Create user
func (a App) Create(ctx context.Context, item model.User) (model.User, error) {
	if err := a.check(ctx, model.User{}, item); err != nil {
		return model.User{}, httpModel.WrapInvalid(err)
	}

	if err := a.authApp.Check(ctx, authModel.User{}, item.Login); err != nil {
		return model.User{}, httpModel.WrapInvalid(err)
	}

	var output model.User

	err := a.userStore.DoAtomic(ctx, func(ctx context.Context) error {
		loginUser, err := a.authApp.Create(ctx, item.Login)
		if err != nil {
			return httpModel.WrapInternal(fmt.Errorf("unable to create login: %w", err))
		}

		item.Login = loginUser

		id, err := a.userStore.Create(ctx, item)
		if err != nil {
			return httpModel.WrapInternal(fmt.Errorf("unable to create: %w", err))
		}

		item.ID = id
		output = item

		return nil
	})

	return output, err
}

func (a App) check(ctx context.Context, _, new model.User) error {
	if new.IsZero() {
		return nil
	}

	var output []error

	if len(strings.TrimSpace(new.Email)) == 0 {
		output = append(output, errors.New("email is required"))
	}

	if userWithEmail, err := a.userStore.GetByEmail(ctx, new.Email); err != nil {
		output = append(output, errors.New("unable to check if email already exists"))
	} else if !userWithEmail.ID.IsZero() {
		output = append(output, errors.New("email already used"))
	}

	return httpModel.ConcatError(output)
}

// Count users
func (a App) Count(ctx context.Context) (uint64, error) {
	return a.userStore.Count(ctx)
}
