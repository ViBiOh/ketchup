package service

import (
	"context"
	"database/sql"
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

// Unmarshal User
func (a app) Unmarshal(data []byte, contentType string) (interface{}, error) {
	var user model.User
	err := json.Unmarshal(data, &user)
	return user, err
}

// List Users
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

// Get User
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

// Create User
func (a app) Create(ctx context.Context, o interface{}) (output interface{}, err error) {
	output = model.NoneUser
	user := o.(model.User)

	if inputErrors := a.authService.Check(ctx, nil, user.Login); len(inputErrors) != 0 {
		err = fmt.Errorf("%s: %w", inputErrors, crud.ErrInvalid)
		return
	}

	ctx, err = a.userStore.StartAtomic(ctx)
	if err != nil {
		return
	}

	defer func() {
		a.userStore.EndAtomic(ctx, err)
	}()

	var loginUser interface{}
	loginUser, err = a.authService.Create(ctx, user.Login)
	if err != nil {
		err = fmt.Errorf("unable to create auth: %w", err)
	}

	user.Login = loginUser.(authModel.User)

	var id uint64
	id, err = a.userStore.Create(ctx, user)
	if err != nil {
		err = fmt.Errorf("unable to create: %w", err)
		return
	}

	user.ID = id
	output = user

	return
}

// Update User
func (a app) Update(ctx context.Context, o interface{}) (interface{}, error) {
	user := o.(model.User)

	if err := a.userStore.Update(ctx, user); err != nil {
		return user, fmt.Errorf("unable to update: %w", err)
	}

	return user, nil
}

// Delete User
func (a app) Delete(ctx context.Context, o interface{}) (err error) {
	user := o.(model.User)

	if inputErrors := a.authService.Check(ctx, user.Login, nil); len(inputErrors) != 0 {
		err = fmt.Errorf("%s: %w", inputErrors, crud.ErrInvalid)
		return
	}

	ctx, err = a.userStore.StartAtomic(ctx)
	if err != nil {
		return
	}

	defer func() {
		a.userStore.EndAtomic(ctx, err)
	}()

	err = a.authService.Delete(ctx, user.Login)
	if err != nil {
		err = fmt.Errorf("unable to delete auth: %w", err)
	}

	if err = a.userStore.Delete(ctx, user); err != nil {
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

	newUser := new.(model.User)

	if old != nil && new != nil && !(user.ID == newUser.ID || a.authProvider.IsAuthorized(ctx, user, "admin")) {
		errors = append(errors, crud.NewError("context", "you're not authorized to interact with other user"))
	}

	if len(strings.TrimSpace(newUser.Email)) == 0 {
		errors = append(errors, crud.NewError("email", "email is required"))
	}

	userWithEmail, err := a.userStore.GetByEmail(ctx, newUser.Email)
	if err != nil {
		if err != sql.ErrNoRows {
			errors = append(errors, crud.NewError("email", "unable to check if email already exists"))
		}
	} else if userWithEmail.ID != newUser.ID {
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
