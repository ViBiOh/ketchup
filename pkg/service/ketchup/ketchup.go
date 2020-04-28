package ketchup

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	authModel "github.com/ViBiOh/auth/v2/pkg/model"
	"github.com/ViBiOh/httputils/v3/pkg/crud"
	"github.com/ViBiOh/ketchup/pkg/model"
	repositoryService "github.com/ViBiOh/ketchup/pkg/service/repository"
	userService "github.com/ViBiOh/ketchup/pkg/service/user"
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

	ListForRepositories(ctx context.Context, repositories []model.Repository) ([]model.Ketchup, error)
}

type app struct {
	ketchupStore      store.KetchupStore
	repositoryService repositoryService.App
	userService       userService.App
}

// New creates new App from Config
func New(ketchupStore store.KetchupStore, repositoryService repositoryService.App, userService userService.App) App {
	return app{
		ketchupStore:      ketchupStore,
		repositoryService: repositoryService,
		userService:       userService,
	}
}

func (a app) Unmarshal(data []byte, contentType string) (interface{}, error) {
	var item model.Ketchup
	err := json.Unmarshal(data, &item)
	return item, err
}

func (a app) List(ctx context.Context, page, pageSize uint, sortKey string, sortAsc bool, filters map[string][]string) ([]interface{}, uint, error) {
	ctx, err := a.convertLoginToUser(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("unable to convert user: %s", err)
	}

	list, total, err := a.ketchupStore.List(ctx, page, pageSize, sortKey, sortAsc)
	if err != nil {
		return nil, 0, fmt.Errorf("unable to list: %s", err)
	}

	itemsList := make([]interface{}, len(list))
	for index, item := range list {
		repository, err := a.repositoryService.Get(ctx, item.Repository.ID)
		if err != nil {
			return nil, 0, fmt.Errorf("unable to get repository for %d: %s", item.Repository.ID, err)
		}

		item.Repository = repository.(model.Repository)

		itemsList[index] = item
	}

	return itemsList, total, nil
}

func (a app) Get(ctx context.Context, ID uint64) (interface{}, error) {
	ctx, err := a.convertLoginToUser(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to convert user: %s", err)
	}

	item, err := a.ketchupStore.GetByRepositoryID(ctx, ID)
	if err != nil {
		return nil, fmt.Errorf("unable to get: %s", err)
	}

	if item == model.NoneKetchup {
		return nil, crud.ErrNotFound
	}

	repository, err := a.repositoryService.Get(ctx, item.Repository.ID)
	if err != nil {
		return nil, fmt.Errorf("unable to get repository: %s", err)
	}

	item.Repository = repository.(model.Repository)

	return item, nil
}

func (a app) Create(ctx context.Context, o interface{}) (output interface{}, err error) {
	ctx, err = a.convertLoginToUser(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to convert user: %s", err)
	}

	output = model.NoneKetchup
	item := o.(model.Ketchup)

	ctx, err = a.ketchupStore.StartAtomic(ctx)
	if err != nil {
		return
	}

	defer func() {
		if endErr := a.ketchupStore.EndAtomic(ctx, err); endErr != nil {
			err = fmt.Errorf("%s: %w", err.Error(), endErr)
		}
	}()

	var repository interface{}
	repository, err = a.repositoryService.GetOrCreate(ctx, item.Repository.Name)
	if err != nil {
		return
	}

	item.Repository = repository.(model.Repository)

	if _, err = a.ketchupStore.Create(ctx, item); err != nil {
		return o, fmt.Errorf("unable to create: %s", err)
	}

	return item, nil
}

func (a app) Update(ctx context.Context, o interface{}) (interface{}, error) {
	ctx, err := a.convertLoginToUser(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to convert user: %s", err)
	}

	item := o.(model.Ketchup)

	if err := a.ketchupStore.Update(ctx, item); err != nil {
		return item, fmt.Errorf("unable to update: %s", err)
	}

	return item, nil
}

func (a app) Delete(ctx context.Context, o interface{}) error {
	ctx, err := a.convertLoginToUser(ctx)
	if err != nil {
		return fmt.Errorf("unable to convert user: %s", err)
	}

	item := o.(model.Ketchup)

	if err := a.ketchupStore.Delete(ctx, item); err != nil {
		return fmt.Errorf("unable to delete: %s", err)
	}

	return nil
}

func (a app) ListForRepositories(ctx context.Context, repositories []model.Repository) ([]model.Ketchup, error) {
	ids := make([]uint64, len(repositories))
	for index, repo := range repositories {
		ids[index] = repo.ID
	}

	list, err := a.ketchupStore.ListByRepositoriesID(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("unable to list by ids: %s", err)
	}

	enrichList := make([]model.Ketchup, len(list))
	for index, ketchup := range list {
		user, err := a.userService.Get(ctx, ketchup.User.ID)
		if err != nil {
			return nil, fmt.Errorf("unable to get user for repository %d and user %d: %s", ketchup.Repository.ID, ketchup.User.ID, err)
		}

		ketchup.User = user.(model.User)
		enrichList[index] = ketchup
	}

	return enrichList, nil
}

func (a app) Check(ctx context.Context, old, new interface{}) []crud.Error {
	errors := make([]crud.Error, 0)

	user := authModel.ReadUser(ctx)
	if old != nil && user == authModel.NoneUser {
		errors = append(errors, crud.NewError("context", "you must be logged in for interacting"))
	}

	if new == nil {
		return errors
	}

	newItem := new.(model.Ketchup)

	if len(strings.TrimSpace(newItem.Version)) == 0 {
		errors = append(errors, crud.NewError("version", "version is required"))
	}

	return errors
}

func (a app) convertLoginToUser(ctx context.Context) (context.Context, error) {
	user, err := a.userService.GetFromContext(ctx)
	if err != nil {
		return ctx, err
	}

	return model.StoreUser(ctx, user.(model.User)), nil
}
