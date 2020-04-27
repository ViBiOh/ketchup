package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	authModel "github.com/ViBiOh/auth/v2/pkg/model"
	"github.com/ViBiOh/httputils/v3/pkg/crud"
	"github.com/ViBiOh/httputils/v3/pkg/logger"
	"github.com/ViBiOh/ketchup/pkg/model"
)

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

	list, total, err := a.storeApp.ListUsers(ctx, page, pageSize, sortKey, sortAsc)
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

	item, err := a.storeApp.GetUser(ctx, ID)
	if err != nil {
		return nil, fmt.Errorf("unable to get: %w", err)
	}

	if item == model.NoneUser {
		return nil, crud.ErrNotFound
	}

	return item, nil
}

// Create User
func (a app) Create(ctx context.Context, o interface{}) (interface{}, error) {
	user := o.(model.User)

	id, err := a.storeApp.CreateUser(ctx, user)
	if err != nil {
		return model.NoneUser, fmt.Errorf("unable to create: %w", err)
	}

	user.ID = id

	return user, nil
}

// Update User
func (a app) Update(ctx context.Context, o interface{}) (interface{}, error) {
	user := o.(model.User)

	if err := a.storeApp.UpdateUser(ctx, user); err != nil {
		return user, fmt.Errorf("unable to update: %w", err)
	}

	return user, nil
}

// Delete User
func (a app) Delete(ctx context.Context, o interface{}) error {
	user := o.(model.User)

	if err := a.storeApp.DeleteUser(ctx, user); err != nil {
		return fmt.Errorf("unable to delete: %w", err)
	}

	return nil
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

	userWithEmail, err := a.storeApp.GetUserByEmail(ctx, newUser.Email)
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
