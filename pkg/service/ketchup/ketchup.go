package ketchup

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"

	httpModel "github.com/ViBiOh/httputils/v3/pkg/model"
	"github.com/ViBiOh/ketchup/pkg/model"
	"github.com/ViBiOh/ketchup/pkg/semver"
	"github.com/ViBiOh/ketchup/pkg/service"
	"github.com/ViBiOh/ketchup/pkg/service/repository"
	"github.com/ViBiOh/ketchup/pkg/store/ketchup"
)

// App of package
type App interface {
	List(ctx context.Context, page, pageSize uint) ([]model.Ketchup, uint64, error)
	ListForRepositories(ctx context.Context, repositories []model.Repository) ([]model.Ketchup, error)
	Create(ctx context.Context, item model.Ketchup) (model.Ketchup, error)
	Update(ctx context.Context, item model.Ketchup) (model.Ketchup, error)
	Delete(ctx context.Context, item model.Ketchup) error
}

type app struct {
	ketchupStore      ketchup.App
	repositoryService repository.App
}

// New creates new App from Config
func New(ketchupStore ketchup.App, repositoryService repository.App) App {
	return app{
		ketchupStore:      ketchupStore,
		repositoryService: repositoryService,
	}
}

func (a app) List(ctx context.Context, page, pageSize uint) ([]model.Ketchup, uint64, error) {
	list, total, err := a.ketchupStore.List(ctx, page, pageSize)
	if err != nil {
		return nil, 0, httpModel.WrapInternal(fmt.Errorf("unable to list: %s", err))
	}

	enrichedList := enrichSemver(list)
	sort.Sort(model.KetchupByPriority(enrichedList))

	return enrichedList, total, nil
}

func (a app) ListForRepositories(ctx context.Context, repositories []model.Repository) ([]model.Ketchup, error) {
	ids := make([]uint64, len(repositories))
	for index, repo := range repositories {
		ids[index] = repo.ID
	}

	list, err := a.ketchupStore.ListByRepositoriesID(ctx, ids)
	if err != nil {
		return nil, httpModel.WrapInternal(fmt.Errorf("unable to list by ids: %s", err))
	}

	return enrichSemver(list), nil
}

func (a app) Create(ctx context.Context, item model.Ketchup) (model.Ketchup, error) {
	var output model.Ketchup

	err := a.ketchupStore.DoAtomic(ctx, func(ctx context.Context) error {
		repo, err := a.repositoryService.GetOrCreate(ctx, item.Repository.Name, item.Repository.Kind)
		if err != nil {
			return err
		}

		item.Repository = repo

		if err := a.check(ctx, model.NoneKetchup, item); err != nil {
			return httpModel.WrapInvalid(err)
		}

		if _, err := a.ketchupStore.Create(ctx, item); err != nil {
			return httpModel.WrapInternal(fmt.Errorf("unable to create: %s", err))
		}

		output = item
		return nil
	})

	return output, err
}

func (a app) Update(ctx context.Context, item model.Ketchup) (model.Ketchup, error) {
	var output model.Ketchup

	err := a.ketchupStore.DoAtomic(ctx, func(ctx context.Context) error {
		old, err := a.ketchupStore.GetByRepositoryID(ctx, item.Repository.ID, true)
		if err != nil {
			return httpModel.WrapInternal(fmt.Errorf("unable to fetch: %s", err))
		}

		current := model.Ketchup{
			Version:    item.Version,
			Repository: old.Repository,
			User:       old.User,
		}

		if err := a.check(ctx, old, current); err != nil {
			return httpModel.WrapInvalid(err)
		}

		if err := a.ketchupStore.Update(ctx, current); err != nil {
			return httpModel.WrapInternal(fmt.Errorf("unable to update: %s", err))
		}

		output = current
		return nil
	})

	return output, err
}

func (a app) Delete(ctx context.Context, item model.Ketchup) (err error) {
	return a.ketchupStore.DoAtomic(ctx, func(ctx context.Context) error {
		old, err := a.ketchupStore.GetByRepositoryID(ctx, item.Repository.ID, true)
		if err != nil {
			return httpModel.WrapInternal(fmt.Errorf("unable to fetch current: %s", err))
		}

		if err = a.check(ctx, old, model.NoneKetchup); err != nil {
			return httpModel.WrapInvalid(err)
		}

		if err = a.ketchupStore.Delete(ctx, old); err != nil {
			return httpModel.WrapInternal(fmt.Errorf("unable to delete: %s", err))
		}

		return nil
	})
}

func (a app) check(ctx context.Context, old, new model.Ketchup) error {
	output := make([]error, 0)

	if model.ReadUser(ctx) == model.NoneUser {
		output = append(output, errors.New("you must be logged in for interacting"))
	}

	if new.Repository.ID == 0 && new.User.ID == 0 {
		return service.ConcatError(output)
	}

	if len(strings.TrimSpace(new.Version)) == 0 {
		output = append(output, errors.New("version is required"))
	}

	if old.Repository.ID == 0 && old.User.ID == 0 {
		o, err := a.ketchupStore.GetByRepositoryID(ctx, new.Repository.ID, false)
		if err != nil {
			output = append(output, errors.New("unable to check if ketchup already exists"))
		} else if o.User.ID != 0 && o.Repository.ID != 0 {
			output = append(output, fmt.Errorf("ketchup for %s already exists", new.Repository.Name))
		}
	}

	return service.ConcatError(output)
}

func enrichSemver(list []model.Ketchup) []model.Ketchup {
	output := make([]model.Ketchup, len(list))

	for index, item := range list {
		repositoryVersion, repositoryErr := semver.Parse(item.Repository.Versions["stable"])
		ketchupVersion, ketchupErr := semver.Parse(item.Version)
		if repositoryErr == nil && ketchupErr == nil {
			item.Semver = repositoryVersion.Compare(ketchupVersion)
		}

		output[index] = item
	}

	return output
}
