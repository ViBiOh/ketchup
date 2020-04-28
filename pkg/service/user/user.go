package user

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/ViBiOh/auth/v2/pkg/auth"
	authModel "github.com/ViBiOh/auth/v2/pkg/model"
	authService "github.com/ViBiOh/auth/v2/pkg/service"
	"github.com/ViBiOh/httputils/v3/pkg/logger"
	"github.com/ViBiOh/ketchup/pkg/model"
	"github.com/ViBiOh/ketchup/pkg/store/user"
)

// App of package
type App interface {
	Unmarshal(data []byte, contentType string) (model.User, error)
	Check(ctx context.Context, old, new model.User) []error
	Create(ctx context.Context, o model.User) (model.User, error)
	StoreInContext(ctx context.Context) context.Context
}

type app struct {
	userStore user.App

	authService  authService.App
	authProvider auth.Provider
}

// New creates new App from Config
func New(userStore user.App, authService authService.App, authProvider auth.Provider) App {
	return app{
		userStore: userStore,

		authService:  authService,
		authProvider: authProvider,
	}
}

func (a app) Unmarshal(data []byte, contentType string) (model.User, error) {
	var item model.User
	err := json.Unmarshal(data, &item)
	return item, err
}

func (a app) StoreInContext(ctx context.Context) context.Context {
	id := authModel.ReadUser(ctx).ID

	if id == 0 {
		logger.Warn("no login user in context")
		return ctx
	}

	item, err := a.userStore.GetByLoginID(ctx, id)
	if err != nil {
		logger.Error("unable to get user with login %d: %s", id, err)
		return ctx
	}

	return model.StoreUser(ctx, item)
}

func (a app) Create(ctx context.Context, item model.User) (output model.User, err error) {
	ctx, err = a.userStore.StartAtomic(ctx)
	if err != nil {
		return
	}

	defer func() {
		if endErr := a.userStore.EndAtomic(ctx, err); endErr != nil {
			err = fmt.Errorf("%s: %w", err.Error(), endErr)
		}
	}()

	if errs := a.Check(ctx, model.NoneUser, item); len(errs) != 0 {
		err = fmt.Errorf("invalid payload: %s", errs)
		return
	}

	if errs := a.authService.Check(ctx, nil, item.Login); len(errs) != 0 {
		err = fmt.Errorf("invalid login payload: %s", errs)
		return
	}

	var loginUser interface{}
	loginUser, err = a.authService.Create(ctx, item.Login)
	if err != nil {
		err = fmt.Errorf("unable to create login: %w", err)
	}

	item.Login = loginUser.(authModel.User)

	var id uint64
	id, err = a.userStore.Create(ctx, item)
	if err != nil {
		err = fmt.Errorf("unable to create: %w", err)
		return
	}

	item.ID = id

	return
}

func (a app) Check(ctx context.Context, old, new model.User) []error {
	output := make([]error, 0)

	if new == model.NoneUser {
		return output
	}

	if len(strings.TrimSpace(new.Email)) == 0 {
		output = append(output, errors.New("email is required"))
	}

	if userWithEmail, err := a.userStore.GetByEmail(ctx, new.Email); err != nil {
		output = append(output, errors.New("unable to check if email already exists"))
	} else if userWithEmail.ID != new.ID {
		output = append(output, errors.New("email already used"))
	}

	return output
}
