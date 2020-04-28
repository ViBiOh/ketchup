package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/ViBiOh/httputils/v3/pkg/crud"
	"github.com/ViBiOh/ketchup/pkg/github"
	"github.com/ViBiOh/ketchup/pkg/model"
	"github.com/ViBiOh/ketchup/pkg/store"
)

// App of package
type App interface {
	List(ctx context.Context, page, pageSize uint, sortKey string, sortAsc bool, filters map[string][]string) ([]interface{}, uint, error)
	Get(ctx context.Context, ID uint64) (interface{}, error)
	GetOrCreate(ctx context.Context, name string) (interface{}, error)
	Create(ctx context.Context, o interface{}) (interface{}, error)
	Update(ctx context.Context, o interface{}) (interface{}, error)
	Delete(ctx context.Context, o interface{}) error
}

type app struct {
	repositoryStore store.RepositoryStore
	githubApp       github.App
}

// New creates new App from Config
func New(repositoryStore store.RepositoryStore, githubApp github.App) App {
	return app{
		repositoryStore: repositoryStore,
		githubApp:       githubApp,
	}
}

func (a app) List(ctx context.Context, page, pageSize uint, sortKey string, sortAsc bool, filters map[string][]string) ([]interface{}, uint, error) {
	list, total, err := a.repositoryStore.List(ctx, page, pageSize, sortKey, sortAsc)
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
	item, err := a.repositoryStore.Get(ctx, ID)
	if err != nil {
		return nil, fmt.Errorf("unable to get: %w", err)
	}

	if item == model.NoneRepository {
		return nil, crud.ErrNotFound
	}

	return item, nil
}

func (a app) GetOrCreate(ctx context.Context, name string) (interface{}, error) {
	repository, err := a.repositoryStore.GetByName(ctx, name)
	if err != nil {
		return model.NoneRepository, err
	}

	if repository != model.NoneRepository {
		return repository, nil
	}

	repository = model.Repository{
		Name: name,
	}

	if inputErrors := a.Check(ctx, nil, repository); len(inputErrors) != 0 {
		return model.NoneRepository, fmt.Errorf("%s: %w", inputErrors, crud.ErrInvalid)
	}

	return a.Create(ctx, repository)
}

func (a app) Create(ctx context.Context, o interface{}) (interface{}, error) {
	item := o.(model.Repository)

	release, err := a.githubApp.LastRelease(item.Name)
	if err != nil {
		return model.NoneRepository, fmt.Errorf("unable to prepare creation: %s", err)
	}

	item.Version = release.TagName

	id, err := a.repositoryStore.Create(ctx, item)
	if err != nil {
		return model.NoneRepository, fmt.Errorf("unable to create: %s", err)
	}

	item.ID = id

	return item, nil
}

func (a app) Update(ctx context.Context, o interface{}) (interface{}, error) {
	item := o.(model.Repository)

	if err := a.repositoryStore.Update(ctx, item); err != nil {
		return model.NoneRepository, fmt.Errorf("unable to update: %w", err)
	}

	return item, nil
}

func (a app) Delete(ctx context.Context, o interface{}) error {
	item := o.(model.Repository)

	if err := a.repositoryStore.Update(ctx, item); err != nil {
		return fmt.Errorf("unable to delete: %w", err)
	}

	return nil
}

func (a app) Check(ctx context.Context, old, new interface{}) []crud.Error {
	errors := make([]crud.Error, 0)

	// TODO check if ketchup used that repository

	if new == nil {
		return errors
	}

	newItem := new.(model.Repository)

	if len(strings.TrimSpace(newItem.Name)) == 0 {
		errors = append(errors, crud.NewError("name", "name is required"))
	}

	repositoryWithName, err := a.repositoryStore.GetByName(ctx, newItem.Name)
	if err != nil {
		errors = append(errors, crud.NewError("name", "unable to check if name already exists"))
	} else if repositoryWithName != model.NoneRepository && repositoryWithName.ID != newItem.ID {
		errors = append(errors, crud.NewError("name", "name already exists"))
	}

	return errors
}
