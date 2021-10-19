package ketchup

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"

	httpModel "github.com/ViBiOh/httputils/v4/pkg/model"
	"github.com/ViBiOh/ketchup/pkg/model"
	"github.com/ViBiOh/ketchup/pkg/semver"
)

// App of package
type App struct {
	ketchupStore      model.KetchupStore
	repositoryService model.RepositoryService
}

// New creates new App from Config
func New(ketchupStore model.KetchupStore, repositoryService model.RepositoryService) App {
	return App{
		ketchupStore:      ketchupStore,
		repositoryService: repositoryService,
	}
}

// List ketchups
func (a App) List(ctx context.Context, pageSize uint, last string) ([]model.Ketchup, uint64, error) {
	list, total, err := a.ketchupStore.List(ctx, pageSize, last)
	if err != nil {
		return nil, 0, httpModel.WrapInternal(fmt.Errorf("unable to list: %w", err))
	}

	enrichedList := enrichKetchupsWithSemver(list)
	sort.Sort(model.KetchupByPriority(enrichedList))

	return enrichedList, total, nil
}

// ListForRepositories of ketchups
func (a App) ListForRepositories(ctx context.Context, repositories []model.Repository, frequency model.KetchupFrequency) ([]model.Ketchup, error) {
	ids := make([]uint64, len(repositories))
	for index, repo := range repositories {
		ids[index] = repo.ID
	}

	list, err := a.ketchupStore.ListByRepositoriesID(ctx, ids, frequency)
	if err != nil {
		return nil, httpModel.WrapInternal(fmt.Errorf("unable to list by ids: %w", err))
	}

	return enrichKetchupsWithSemver(list), nil
}

// ListOutdatedByFrequency ketchups outdated
func (a App) ListOutdatedByFrequency(ctx context.Context, frequency model.KetchupFrequency) ([]model.Ketchup, error) {
	list, err := a.ketchupStore.ListOutdatedByFrequency(ctx, frequency)
	if err != nil {
		return nil, httpModel.WrapInternal(fmt.Errorf("unable to list outdated by frequency: %w", err))
	}

	return list, nil
}

// Create ketchup
func (a App) Create(ctx context.Context, item model.Ketchup) (model.Ketchup, error) {
	var output model.Ketchup

	err := a.ketchupStore.DoAtomic(ctx, func(ctx context.Context) error {
		repo, err := a.repositoryService.GetOrCreate(ctx, item.Repository.Kind, item.Repository.Name, item.Repository.Part, item.Pattern)
		if err != nil {
			return err
		}

		item.Repository = repo

		if err := a.check(ctx, model.Ketchup{}, item); err != nil {
			return httpModel.WrapInvalid(err)
		}

		if _, err := a.ketchupStore.Create(ctx, item); err != nil {
			return httpModel.WrapInternal(fmt.Errorf("unable to create: %w", err))
		}

		output = item
		return nil
	})

	return output, err
}

// UpdateAll ketchups
func (a App) UpdateAll(ctx context.Context) error {
	err := a.ketchupStore.DoAtomic(ctx, func(ctx context.Context) error {
		return a.ketchupStore.UpdateAll(ctx)
	})
	if err != nil {
		return fmt.Errorf("unable to update all ketchup: %s", err)
	}

	return nil
}

// Update ketchup
func (a App) Update(ctx context.Context, oldPattern string, item model.Ketchup) (model.Ketchup, error) {
	var output model.Ketchup

	err := a.ketchupStore.DoAtomic(ctx, func(ctx context.Context) error {
		old, err := a.ketchupStore.GetByRepository(ctx, item.Repository.ID, oldPattern, true)
		if err != nil {
			return httpModel.WrapInternal(fmt.Errorf("unable to fetch: %w", err))
		}

		if old.Repository.ID == 0 {
			return httpModel.WrapNotFound(errors.New("unable to found repository"))
		}

		current := model.Ketchup{
			Pattern:          item.Pattern,
			Version:          item.Version,
			Frequency:        item.Frequency,
			UpdateWhenNotify: item.UpdateWhenNotify,
			Repository:       old.Repository,
			User:             old.User,
		}

		if err := a.check(ctx, old, current); err != nil {
			return httpModel.WrapInvalid(err)
		}

		if old.Pattern != item.Pattern {
			repo, err := a.repositoryService.GetOrCreate(ctx, old.Repository.Kind, old.Repository.Name, old.Repository.Part, item.Pattern)
			if err != nil {
				return httpModel.WrapInternal(fmt.Errorf("unable to get repository version: %w", err))
			}

			current.Repository = repo
		}

		if err := a.ketchupStore.Update(ctx, current, old.Pattern); err != nil {
			return httpModel.WrapInternal(fmt.Errorf("unable to update: %w", err))
		}

		output = current
		return nil
	})

	return output, err
}

// UpdateVersion of a ketchup
func (a App) UpdateVersion(ctx context.Context, userID, repositoryID uint64, pattern, version string) error {
	if len(pattern) == 0 {
		return errors.New("pattern is required")
	}

	if len(version) == 0 {
		return errors.New("version is required")
	}

	return a.ketchupStore.DoAtomic(ctx, func(ctx context.Context) error {
		return a.ketchupStore.UpdateVersion(ctx, userID, repositoryID, pattern, version)
	})
}

// Delete ketchup
func (a App) Delete(ctx context.Context, item model.Ketchup) (err error) {
	return a.ketchupStore.DoAtomic(ctx, func(ctx context.Context) error {
		old, err := a.ketchupStore.GetByRepository(ctx, item.Repository.ID, item.Pattern, true)
		if err != nil {
			return httpModel.WrapInternal(fmt.Errorf("unable to fetch current: %w", err))
		}

		if old.Repository.ID == 0 {
			return httpModel.WrapNotFound(errors.New("unable to found repository"))
		}

		if err = a.check(ctx, old, model.Ketchup{}); err != nil {
			return httpModel.WrapInvalid(err)
		}

		if err = a.ketchupStore.Delete(ctx, old); err != nil {
			return httpModel.WrapInternal(fmt.Errorf("unable to delete: %w", err))
		}

		return nil
	})
}

func (a App) check(ctx context.Context, old, new model.Ketchup) error {
	var output []error

	if model.ReadUser(ctx).IsZero() {
		output = append(output, errors.New("you must be logged in for interacting"))
	}

	if new.Repository.ID == 0 && new.User.ID == 0 {
		return httpModel.ConcatError(output)
	}

	if len(strings.TrimSpace(new.Pattern)) == 0 {
		output = append(output, errors.New("pattern is required"))
	} else if _, err := semver.ParsePattern(new.Pattern); err != nil {
		output = append(output, fmt.Errorf("pattern is invalid: %s", err))
	}

	if len(strings.TrimSpace(new.Version)) == 0 {
		output = append(output, errors.New("version is required"))
	}

	if old.Repository.ID == 0 && new.Repository.ID != 0 {
		o, err := a.ketchupStore.GetByRepository(ctx, new.Repository.ID, new.Pattern, false)
		if err != nil {
			output = append(output, errors.New("unable to check if ketchup already exists"))
		} else if o.Repository.ID != 0 {
			output = append(output, fmt.Errorf("ketchup for `%s` with pattern `%s` already exists", new.Repository.Name, new.Pattern))
		}
	}

	return httpModel.ConcatError(output)
}

func enrichKetchupsWithSemver(list []model.Ketchup) []model.Ketchup {
	output := make([]model.Ketchup, len(list))

	for index, item := range list {
		output[index] = enrichKetchupWithSemver(item)
	}

	return output
}

func enrichKetchupWithSemver(item model.Ketchup) model.Ketchup {
	repositoryVersion, err := semver.Parse(item.Repository.Versions[item.Pattern])
	if err != nil {
		return item
	}

	ketchupVersion, err := semver.Parse(item.Version)
	if err != nil {
		return item
	}

	item.Semver = repositoryVersion.Compare(ketchupVersion)
	return item
}
