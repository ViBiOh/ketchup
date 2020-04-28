package user

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ViBiOh/auth/v2/pkg/auth"
	authModel "github.com/ViBiOh/auth/v2/pkg/model"
	authService "github.com/ViBiOh/auth/v2/pkg/service"
	"github.com/ViBiOh/httputils/v3/pkg/crud"
	"github.com/ViBiOh/httputils/v3/pkg/logger"
	"github.com/ViBiOh/ketchup/pkg/model"
	"github.com/ViBiOh/ketchup/pkg/store"
)

var (
	_ crud.Service = app{}
)

// App of package
type App interface {
	Unmarshal(data []byte, contentType string) (interface{}, error)
	Check(ctx context.Context, old, new interface{}) []crud.Error
	List(ctx context.Context, page, pageSize uint, sortKey string, sortDesc bool, filters map[string][]string) ([]interface{}, uint, error)
	Get(ctx context.Context, ID uint64) (interface{}, error)
	Create(ctx context.Context, o interface{}) (interface{}, error)
	Update(ctx context.Context, o interface{}) (interface{}, error)
	Delete(ctx context.Context, o interface{}) error

	GetFromContext(ctx context.Context) (interface{}, error)
}

type app struct {
	userStore store.UserStore

	authService  authService.App
	authProvider auth.Provider
}

// New creates new App from Config
func New(userStore store.UserStore, authService authService.App, authProvider auth.Provider) App {
	return app{
		userStore: userStore,

		authService:  authService,
		authProvider: authProvider,
	}
}

func (a app) Unmarshal(data []byte, contentType string) (interface{}, error) {
	var item model.User
	err := json.Unmarshal(data, &item)
	return item, err
}

func (a app) List(ctx context.Context, page, pageSize uint, sortKey string, sortAsc bool, filters map[string][]string) ([]interface{}, uint, error) {
	if err := a.CheckRights(ctx, 0); err != nil {
		return nil, 0, err
	}

	list, total, err := a.userStore.List(ctx, page, pageSize, sortKey, sortAsc)
	if err != nil {
		return nil, 0, fmt.Errorf("unable to list: %w", err)
	}

	itemsList := make([]interface{}, len(list))
	for index, item := range list {
		itemsList[index] = item
	}

	return itemsList, total, nil
}

func (a app) Get(ctx context.Context, ID uint64) (interface{}, error) {
	if err := a.CheckRights(ctx, ID); err != nil {
		return nil, err
	}

	item, err := a.userStore.Get(ctx, ID)
	if err != nil {
		return nil, fmt.Errorf("unable to get: %w", err)
	}

	if item == model.NoneUser {
		return nil, crud.ErrNotFound
	}

	login, err := a.authService.Get(ctx, item.Login.ID)
	if err != nil {
		return nil, fmt.Errorf("unable to get auth: %w", err)
	}

	item.Login = login.(authModel.User)

	return item, nil
}

func (a app) GetFromContext(ctx context.Context) (interface{}, error) {
	item, err := a.userStore.GetByLoginID(ctx, authModel.ReadUser(ctx).ID)
	if err != nil {
		return nil, fmt.Errorf("unable to get: %w", err)
	}

	if item == model.NoneUser {
		return nil, crud.ErrNotFound
	}

	return item, nil
}

func (a app) Create(ctx context.Context, o interface{}) (output interface{}, err error) {
	output = model.NoneUser
	item := o.(model.User)

	if inputErrors := a.authService.Check(ctx, nil, item.Login); len(inputErrors) != 0 {
		err = fmt.Errorf("%s: %w", inputErrors, crud.ErrInvalid)
		return
	}

	ctx, err = a.userStore.StartAtomic(ctx)
	if err != nil {
		return
	}

	defer func() {
		if endErr := a.userStore.EndAtomic(ctx, err); endErr != nil {
			err = fmt.Errorf("%s: %w", err.Error(), endErr)
		}
	}()

	var loginUser interface{}
	loginUser, err = a.authService.Create(ctx, item.Login)
	if err != nil {
		err = fmt.Errorf("unable to create auth: %w", err)
	}

	item.Login = loginUser.(authModel.User)

	var id uint64
	id, err = a.userStore.Create(ctx, item)
	if err != nil {
		err = fmt.Errorf("unable to create: %w", err)
		return
	}

	item.ID = id
	output = item

	return
}

func (a app) Update(ctx context.Context, o interface{}) (interface{}, error) {
	item := o.(model.User)

	if err := a.userStore.Update(ctx, item); err != nil {
		return item, fmt.Errorf("unable to update: %w", err)
	}

	return item, nil
}

func (a app) Delete(ctx context.Context, o interface{}) (err error) {
	item := o.(model.User)

	if inputErrors := a.authService.Check(ctx, item.Login, nil); len(inputErrors) != 0 {
		err = fmt.Errorf("%s: %w", inputErrors, crud.ErrInvalid)
		return
	}

	ctx, err = a.userStore.StartAtomic(ctx)
	if err != nil {
		return
	}

	defer func() {
		if endErr := a.userStore.EndAtomic(ctx, err); endErr != nil {
			err = fmt.Errorf("%s: %w", err.Error(), endErr)
		}
	}()

	err = a.authService.Delete(ctx, item.Login)
	if err != nil {
		err = fmt.Errorf("unable to delete auth: %w", err)
	}

	if err = a.userStore.Delete(ctx, item); err != nil {
		err = fmt.Errorf("unable to delete: %w", err)
		return
	}

	return
}

func (a app) Check(ctx context.Context, old, new interface{}) []crud.Error {
	errors := make([]crud.Error, 0)

	user := authModel.ReadUser(ctx)
	if old != nil && user == authModel.NoneUser {
		errors = append(errors, crud.NewError("context", "you must be logged in for interacting"))
	}

	if new == nil && !a.authProvider.IsAuthorized(ctx, user, "admin") {
		errors = append(errors, crud.NewError("context", "you must be an admin for deleting"))
	}

	if new == nil {
		return errors
	}

	newItem := new.(model.User)

	if old != nil && new != nil && !(user.ID == newItem.ID || a.authProvider.IsAuthorized(ctx, user, "admin")) {
		errors = append(errors, crud.NewError("context", "you're not authorized to interact with other user"))
	}

	if len(strings.TrimSpace(newItem.Email)) == 0 {
		errors = append(errors, crud.NewError("email", "email is required"))
	}

	userWithEmail, err := a.userStore.GetByEmail(ctx, newItem.Email)
	if err != nil {
		errors = append(errors, crud.NewError("email", "unable to check if email already exists"))
	} else if userWithEmail.ID != newItem.ID {
		errors = append(errors, crud.NewError("email", "email already used by another user"))
	}

	return errors
}

func (a app) CheckRights(ctx context.Context, id uint64) error {
	user := authModel.ReadUser(ctx)
	if user == authModel.NoneUser {
		return crud.ErrUnauthorized
	}

	if id != 0 && user.ID == id || a.authProvider.IsAuthorized(ctx, user, "admin") {
		return nil
	}

	logger.Info("unauthorized access for login=%s", user.Login)

	return crud.ErrForbidden
}
