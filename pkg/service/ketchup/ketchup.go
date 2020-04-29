package ketchup

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/ViBiOh/ketchup/pkg/model"
	"github.com/ViBiOh/ketchup/pkg/service"
	"github.com/ViBiOh/ketchup/pkg/service/repository"
	"github.com/ViBiOh/ketchup/pkg/service/user"
	"github.com/ViBiOh/ketchup/pkg/store/ketchup"
)

// App of package
type App interface {
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
		return nil, 0, fmt.Errorf("unable to list: %s: %w", err, service.ErrInternalError)
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
		err = a.ketchupStore.EndAtomic(ctx, err)
	}()

	var repository model.Repository
	repository, err = a.repositoryService.GetOrCreate(ctx, item.Repository.Name)
	if err != nil {
		return
	}

	item.Repository = repository

	if err = a.check(ctx, model.NoneKetchup, item); err != nil {
		err = fmt.Errorf("%s: %w", err, service.ErrInvalid)
		return
	}

	if _, err = a.ketchupStore.Create(ctx, item); err != nil {
		return model.NoneKetchup, fmt.Errorf("unable to create: %s: %w", err, service.ErrInternalError)
	}

	return item, nil
}

func (a app) Update(ctx context.Context, item model.Ketchup) (err error) {
	ctx, err = a.ketchupStore.StartAtomic(ctx)
	if err != nil {
		return
	}

	defer func() {
		err = a.ketchupStore.EndAtomic(ctx, err)
	}()

	var old model.Ketchup
	old, err = a.ketchupStore.GetByRepositoryID(ctx, item.Repository.ID, true)
	if err != nil {
		err = fmt.Errorf("unable to fetch current: %s: %w", err, service.ErrInternalError)
	}

	if err = a.check(ctx, old, item); err != nil {
		err = fmt.Errorf("%s: %w", err, service.ErrInvalid)
		return
	}

	if err = a.ketchupStore.Update(ctx, item); err != nil {
		err = fmt.Errorf("unable to update: %s: %w", err, service.ErrInternalError)
	}

	return
}

func (a app) Delete(ctx context.Context, id uint64) (err error) {
	ctx, err = a.ketchupStore.StartAtomic(ctx)
	if err != nil {
		return
	}

	defer func() {
		err = a.ketchupStore.EndAtomic(ctx, err)
	}()

	var old model.Ketchup
	old, err = a.ketchupStore.GetByRepositoryID(ctx, id, true)
	if err != nil {
		err = fmt.Errorf("unable to fetch current: %s: %w", err, service.ErrInternalError)
	}

	if err = a.check(ctx, old, model.NoneKetchup); err != nil {
		err = fmt.Errorf("%s: %w", err, service.ErrInvalid)
		return
	}

	if err := a.ketchupStore.Delete(ctx, old); err != nil {
		return fmt.Errorf("unable to delete: %s: %w", err, service.ErrInternalError)
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
		return nil, fmt.Errorf("unable to list by ids: %s: %w", err, service.ErrInternalError)
	}

	return list, nil
}

func (a app) check(ctx context.Context, old, new model.Ketchup) error {
	output := make([]error, 0)

	if model.ReadUser(ctx) == model.NoneUser {
		output = append(output, errors.New("you must be logged in for interacting"))
	}

	if new == model.NoneKetchup {
		return service.ConcatError(output)
	}

	if len(strings.TrimSpace(new.Version)) == 0 {
		output = append(output, errors.New("version is required"))
	}

	if old == model.NoneKetchup {
		ketchup, err := a.ketchupStore.GetByRepositoryID(ctx, new.Repository.ID, false)
		if err != nil {
			output = append(output, errors.New("unable to check if ketchup already exists"))
		} else if ketchup != model.NoneKetchup {
			output = append(output, fmt.Errorf("ketchup for %s already exists", new.Repository.Name))
		}
	}

	return service.ConcatError(output)
}
