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

type Service struct {
	ketchupStore model.KetchupStore
	repository   model.RepositoryService
}

func New(ketchupStore model.KetchupStore, repositoryService model.RepositoryService) Service {
	return Service{
		ketchupStore: ketchupStore,
		repository:   repositoryService,
	}
}

func (s Service) List(ctx context.Context, pageSize uint, last string) ([]model.Ketchup, uint64, error) {
	list, total, err := s.ketchupStore.List(ctx, pageSize, last)
	if err != nil {
		return nil, 0, httpModel.WrapInternal(fmt.Errorf("list: %w", err))
	}

	enrichedList := enrichKetchupsWithSemver(list)
	sort.Sort(model.KetchupByPriority(enrichedList))

	return enrichedList, total, nil
}

func (s Service) ListForRepositories(ctx context.Context, repositories []model.Repository, frequencies ...model.KetchupFrequency) ([]model.Ketchup, error) {
	ids := make([]model.Identifier, len(repositories))
	for index, repo := range repositories {
		ids[index] = repo.ID
	}

	list, err := s.ketchupStore.ListByRepositoriesIDAndFrequencies(ctx, ids, frequencies...)
	if err != nil {
		return nil, httpModel.WrapInternal(fmt.Errorf("list by ids: %w", err))
	}

	return enrichKetchupsWithSemver(list), nil
}

func (s Service) ListOutdated(ctx context.Context, users ...model.User) ([]model.Ketchup, error) {
	usersIds := make([]model.Identifier, len(users))
	for i, user := range users {
		usersIds[i] = user.ID
	}

	list, err := s.ketchupStore.ListOutdated(ctx, usersIds...)
	if err != nil {
		return nil, httpModel.WrapInternal(fmt.Errorf("list outdated by frequency: %w", err))
	}

	return list, nil
}

func (s Service) Create(ctx context.Context, item model.Ketchup) (model.Ketchup, error) {
	var output model.Ketchup

	err := s.ketchupStore.DoAtomic(ctx, func(ctx context.Context) error {
		repo, err := s.repository.GetOrCreate(ctx, item.Repository.Kind, item.Repository.Name, item.Repository.Part, item.Pattern)
		if err != nil {
			return err
		}

		item.Repository = repo

		if err := s.check(ctx, model.Ketchup{}, item); err != nil {
			return httpModel.WrapInvalid(err)
		}

		if _, err := s.ketchupStore.Create(ctx, item); err != nil {
			return httpModel.WrapInternal(fmt.Errorf("create: %w", err))
		}

		output = item
		return nil
	})

	return output, err
}

func (s Service) UpdateAll(ctx context.Context) error {
	err := s.ketchupStore.DoAtomic(ctx, func(ctx context.Context) error {
		return s.ketchupStore.UpdateAll(ctx)
	})
	if err != nil {
		return fmt.Errorf("update all ketchup: %w", err)
	}

	return nil
}

func (s Service) Update(ctx context.Context, oldPattern string, item model.Ketchup) (model.Ketchup, error) {
	var output model.Ketchup

	err := s.ketchupStore.DoAtomic(ctx, func(ctx context.Context) error {
		old, err := s.ketchupStore.GetByRepository(ctx, item.Repository.ID, oldPattern, true)
		if err != nil {
			return httpModel.WrapInternal(fmt.Errorf("fetch: %w", err))
		}

		if old.Repository.IsZero() {
			return httpModel.WrapNotFound(errors.New("found repository"))
		}

		current := model.Ketchup{
			Pattern:          item.Pattern,
			Version:          item.Version,
			Frequency:        item.Frequency,
			UpdateWhenNotify: item.UpdateWhenNotify,
			Repository:       old.Repository,
			User:             old.User,
		}

		if err := s.check(ctx, old, current); err != nil {
			return httpModel.WrapInvalid(err)
		}

		if old.Pattern != item.Pattern {
			repo, err := s.repository.GetOrCreate(ctx, old.Repository.Kind, old.Repository.Name, old.Repository.Part, item.Pattern)
			if err != nil {
				return httpModel.WrapInternal(fmt.Errorf("get repository version: %w", err))
			}

			current.Repository = repo
		}

		if err := s.ketchupStore.Update(ctx, current, old.Pattern); err != nil {
			return httpModel.WrapInternal(fmt.Errorf("update: %w", err))
		}

		output = current
		return nil
	})

	return output, err
}

func (s Service) UpdateVersion(ctx context.Context, userID, repositoryID model.Identifier, pattern, version string) error {
	if len(pattern) == 0 {
		return errors.New("pattern is required")
	}

	if len(version) == 0 {
		return errors.New("version is required")
	}

	return s.ketchupStore.DoAtomic(ctx, func(ctx context.Context) error {
		return s.ketchupStore.UpdateVersion(ctx, userID, repositoryID, pattern, version)
	})
}

func (s Service) Delete(ctx context.Context, item model.Ketchup) (err error) {
	return s.ketchupStore.DoAtomic(ctx, func(ctx context.Context) error {
		old, err := s.ketchupStore.GetByRepository(ctx, item.Repository.ID, item.Pattern, true)
		if err != nil {
			return httpModel.WrapInternal(fmt.Errorf("fetch current: %w", err))
		}

		if old.Repository.IsZero() {
			return httpModel.WrapNotFound(errors.New("found repository"))
		}

		if err = s.check(ctx, old, model.Ketchup{}); err != nil {
			return httpModel.WrapInvalid(err)
		}

		if err = s.ketchupStore.Delete(ctx, old); err != nil {
			return httpModel.WrapInternal(fmt.Errorf("delete: %w", err))
		}

		return nil
	})
}

func (s Service) check(ctx context.Context, old, new model.Ketchup) error {
	var output []error

	if model.ReadUser(ctx).IsZero() {
		output = append(output, errors.New("you must be logged in for interacting"))
	}

	if new.Repository.IsZero() && new.User.IsZero() {
		return httpModel.ConcatError(output)
	}

	if len(strings.TrimSpace(new.Pattern)) == 0 {
		output = append(output, errors.New("pattern is required"))
	} else if _, err := semver.ParsePattern(new.Pattern); err != nil {
		output = append(output, fmt.Errorf("pattern is invalid: %w", err))
	}

	if len(strings.TrimSpace(new.Version)) == 0 {
		output = append(output, errors.New("version is required"))
	}

	if old.Repository.IsZero() && !new.Repository.IsZero() {
		o, err := s.ketchupStore.GetByRepository(ctx, new.Repository.ID, new.Pattern, false)
		if err != nil {
			output = append(output, errors.New("check if ketchup already exists"))
		} else if !o.Repository.IsZero() {
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
