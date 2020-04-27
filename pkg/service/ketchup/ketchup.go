package ketchup

import (
	"context"
	"encoding/json"
	"fmt"

	authModel "github.com/ViBiOh/auth/v2/pkg/model"
	"github.com/ViBiOh/httputils/v3/pkg/crud"
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
	ketchupStore    store.KetchupStore
	repositoryStore store.RepositoryStore
}

// New creates new App from Config
func New(ketchupStore store.KetchupStore, repositoryStore store.RepositoryStore) App {
	return app{
		ketchupStore:    ketchupStore,
		repositoryStore: repositoryStore,
	}
}

func (a app) Unmarshal(data []byte, contentType string) (interface{}, error) {
	var item model.Ketchup
	err := json.Unmarshal(data, &item)
	return item, err
}

func (a app) List(ctx context.Context, page, pageSize uint, sortKey string, sortAsc bool, filters map[string][]string) ([]interface{}, uint, error) {
	list, total, err := a.ketchupStore.List(ctx, page, pageSize, sortKey, sortAsc)
	if err != nil {
		return nil, 0, fmt.Errorf("unable to list: %w", err)
	}

	itemsList := make([]interface{}, len(list))
	for index, item := range list {
		repository, err := a.repositoryStore.Get(ctx, item.Repository.ID)
		if err != nil {
			return nil, 0, fmt.Errorf("unable to get repository for %d: %s", item.Repository.ID, err)
		}

		item.Repository = repository

		itemsList[index] = item
	}

	return itemsList, total, nil
}

func (a app) Get(ctx context.Context, ID uint64) (interface{}, error) {
	item, err := a.ketchupStore.Get(ctx, ID)
	if err != nil {
		return nil, fmt.Errorf("unable to get: %w", err)
	}

	if item == model.NoneKetchup {
		return nil, crud.ErrNotFound
	}

	repository, err := a.repositoryStore.Get(ctx, item.Repository.ID)
	if err != nil {
		return nil, fmt.Errorf("unable to get repository: %w", err)
	}

	item.Repository = repository

	return item, nil
}

func (a app) Create(ctx context.Context, o interface{}) (interface{}, error) {
	item := o.(model.Ketchup)

	if _, err := a.ketchupStore.Create(ctx, item); err != nil {
		return o, fmt.Errorf("unable to create: %w", err)
	}

	return item, nil
}

func (a app) Update(ctx context.Context, o interface{}) (interface{}, error) {
	item := o.(model.Ketchup)

	if err := a.ketchupStore.Update(ctx, item); err != nil {
		return item, fmt.Errorf("unable to update: %w", err)
	}

	return item, nil
}

func (a app) Delete(ctx context.Context, o interface{}) (err error) {
	item := o.(model.Ketchup)

	if err = a.ketchupStore.Delete(ctx, item); err != nil {
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

	return errors
}
