package ketchup

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/ViBiOh/ketchup/pkg/model"
	"github.com/ViBiOh/ketchup/pkg/service/repository"
	"github.com/ViBiOh/ketchup/pkg/service/user"
	"github.com/ViBiOh/ketchup/pkg/store/ketchup"
)

// App of package
type App interface {
	Check(ctx context.Context, old, new model.Ketchup) []error
	List(ctx context.Context, page, pageSize uint) ([]model.Ketchup, uint, error)
	ListForRepositories(ctx context.Context, repositories []model.Repository) ([]model.Ketchup, error)
	Create(ctx context.Context, item model.Ketchup) (model.Ketchup, error)
	Update(ctx context.Context, item model.Ketchup) error
	Delete(ctx context.Context, id uint64) error
}

type app struct {
	ketchupStore      ketchup.App
	repositoryService repository.App
	userService       user.App
}

// New creates new App from Config
func New(ketchupStore ketchup.App, repositoryService repository.App, userService user.App) App {
	return app{
		ketchupStore:      ketchupStore,
		repositoryService: repositoryService,
		userService:       userService,
	}
}

func (a app) List(ctx context.Context, page, pageSize uint) ([]model.Ketchup, uint, error) {
	list, total, err := a.ketchupStore.List(ctx, page, pageSize)
	if err != nil {
		return nil, 0, fmt.Errorf("unable to list: %s", err)
	}

	return list, total, nil
}

func (a app) Create(ctx context.Context, item model.Ketchup) (output model.Ketchup, err error) {
	output = model.NoneKetchup

	ctx, err = a.ketchupStore.StartAtomic(ctx)
	if err != nil {
		return
	}

	defer func() {
		if endErr := a.ketchupStore.EndAtomic(ctx, err); endErr != nil {
			err = fmt.Errorf("%s: %w", err.Error(), endErr)
		}
	}()

	var repository model.Repository
	repository, err = a.repositoryService.GetOrCreate(ctx, item.Repository.Name)
	if err != nil {
		return
	}

	item.Repository = repository

	if _, err = a.ketchupStore.Create(ctx, item); err != nil {
		return model.NoneKetchup, fmt.Errorf("unable to create: %s", err)
	}

	return item, nil
}

func (a app) Update(ctx context.Context, item model.Ketchup) (err error) {
	ctx, err = a.ketchupStore.StartAtomic(ctx)
	if err != nil {
		return
	}

	defer func() {
		if endErr := a.ketchupStore.EndAtomic(ctx, err); endErr != nil {
			err = fmt.Errorf("%s: %s", err.Error(), endErr)
		}
	}()

	var old model.Ketchup
	old, err = a.ketchupStore.GetByRepositoryID(ctx, item.Repository.ID, true)
	if err != nil {
		err = fmt.Errorf("unable to fetch current: %s", err)
	}

	if errs := a.Check(ctx, old, item); len(errs) > 0 {
		err = fmt.Errorf("invalid payload: %s", errs)
		return
	}

	if err = a.ketchupStore.Update(ctx, item); err != nil {
		err = fmt.Errorf("unable to update: %s", err)
	}

	return
}

func (a app) Delete(ctx context.Context, id uint64) (err error) {
	ctx, err = a.ketchupStore.StartAtomic(ctx)
	if err != nil {
		return
	}

	defer func() {
		if endErr := a.ketchupStore.EndAtomic(ctx, err); endErr != nil {
			err = fmt.Errorf("%s: %s", err.Error(), endErr)
		}
	}()

	var old model.Ketchup
	old, err = a.ketchupStore.GetByRepositoryID(ctx, id, true)
	if err != nil {
		err = fmt.Errorf("unable to fetch current: %s", err)
	}

	if errs := a.Check(ctx, old, model.NoneKetchup); len(errs) > 0 {
		err = fmt.Errorf("invalid payload: %s", errs)
		return
	}

	if err := a.ketchupStore.Delete(ctx, old); err != nil {
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

	return list, nil
}

func (a app) Check(ctx context.Context, old, new model.Ketchup) []error {
	output := make([]error, 0)

	if model.ReadUser(ctx) == model.NoneUser {
		output = append(output, errors.New("you must be logged in for interacting"))
	}

	if new == model.NoneKetchup {
		return output
	}

	if len(strings.TrimSpace(new.Version)) == 0 {
		output = append(output, errors.New("version is required"))
	}

	return output
}
