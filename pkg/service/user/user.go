package user

import (
	"context"
	"errors"
	"fmt"
	"strings"

	authModel "github.com/ViBiOh/auth/v2/pkg/model"
	authService "github.com/ViBiOh/auth/v2/pkg/service"
	"github.com/ViBiOh/httputils/v4/pkg/logger"
	httpModel "github.com/ViBiOh/httputils/v4/pkg/model"
	"github.com/ViBiOh/ketchup/pkg/model"
	"github.com/ViBiOh/ketchup/pkg/store/user"
)

// App of package
type App interface {
	Create(ctx context.Context, o model.User) (model.User, error)
	StoreInContext(ctx context.Context) context.Context
	Count(ctx context.Context) (uint64, error)
}

type app struct {
	userStore user.App

	authService authService.App
}

// New creates new App from Config
func New(userStore user.App, authService authService.App) App {
	return app{
		userStore: userStore,

		authService: authService,
	}
}

func (a app) StoreInContext(ctx context.Context) context.Context {
	id := authModel.ReadUser(ctx).ID
	if id == 0 {
		logger.Warn("no login user in context")
		return ctx
	}

	item, err := a.userStore.GetByLoginID(ctx, id)
	if err != nil || item == model.NoneUser {
		logger.Error("unable to get user with login %d: %s", id, err)
		return ctx
	}

	return model.StoreUser(ctx, item)
}

func (a app) Create(ctx context.Context, item model.User) (model.User, error) {
	if err := a.check(ctx, model.NoneUser, item); err != nil {
		return model.NoneUser, httpModel.WrapInvalid(err)
	}

	if err := a.authService.Check(ctx, authModel.NoneUser, item.Login); err != nil {
		return model.NoneUser, httpModel.WrapInvalid(err)
	}

	var output model.User

	err := a.userStore.DoAtomic(ctx, func(ctx context.Context) error {
		loginUser, err := a.authService.Create(ctx, item.Login)
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

func (a app) check(ctx context.Context, _, new model.User) error {
	if new == model.NoneUser {
		return nil
	}

	output := make([]error, 0)

	if len(strings.TrimSpace(new.Email)) == 0 {
		output = append(output, errors.New("email is required"))
	}

	if userWithEmail, err := a.userStore.GetByEmail(ctx, new.Email); err != nil {
		output = append(output, errors.New("unable to check if email already exists"))
	} else if userWithEmail.ID != 0 {
		output = append(output, errors.New("email already used"))
	}

	return httpModel.ConcatError(output)
}

func (a app) Count(ctx context.Context) (uint64, error) {
	return a.userStore.Count(ctx)
}
