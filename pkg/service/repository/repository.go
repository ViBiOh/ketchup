package repository

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/ViBiOh/ketchup/pkg/github"
	"github.com/ViBiOh/ketchup/pkg/model"
	"github.com/ViBiOh/ketchup/pkg/service"
	"github.com/ViBiOh/ketchup/pkg/store/repository"
)

var (
	nameMatcher = regexp.MustCompile(`(?i)(?:github.com/)?([^/\n]+/[^/\n]+)`)
)

// App of package
type App interface {
	List(ctx context.Context, page, pageSize uint) ([]model.Repository, uint64, error)
	GetOrCreate(ctx context.Context, name string) (model.Repository, error)
	Update(ctx context.Context, item model.Repository) error
	Clean(ctx context.Context) error
}

type app struct {
	repositoryStore repository.App
	githubApp       github.App
}

// New creates new App from Config
func New(repositoryStore repository.App, githubApp github.App) App {
	return app{
		repositoryStore: repositoryStore,
		githubApp:       githubApp,
	}
}

func (a app) List(ctx context.Context, page, pageSize uint) ([]model.Repository, uint64, error) {
	list, total, err := a.repositoryStore.List(ctx, page, pageSize)
	if err != nil {
		return nil, 0, service.WrapInternal(fmt.Errorf("unable to list: %s", err))
	}

	return list, total, nil
}

func (a app) GetOrCreate(ctx context.Context, name string) (model.Repository, error) {
	sanitizedName := sanitizeName(name)

	repository, err := a.repositoryStore.GetByName(ctx, sanitizedName)
	if err != nil {
		return model.NoneRepository, service.WrapInternal(err)
	}

	if repository != model.NoneRepository {
		return repository, nil
	}

	return a.create(ctx, model.Repository{Name: sanitizedName})
}

func (a app) create(ctx context.Context, item model.Repository) (model.Repository, error) {
	if err := a.check(ctx, model.NoneRepository, item); err != nil {
		return model.NoneRepository, service.WrapInvalid(err)
	}

	release, err := a.githubApp.LastRelease(item.Name)
	if err != nil {
		return model.NoneRepository, fmt.Errorf("no release found for %s: %w", item.Name, service.ErrNotFound)
	}

	item.Version = release.TagName

	err = a.repositoryStore.DoAtomic(ctx, func(ctx context.Context) error {
		id, err := a.repositoryStore.Create(ctx, item)
		if err != nil {
			return service.WrapInternal(fmt.Errorf("unable to create: %s", err))
		}

		item.ID = id
		return nil
	})

	return item, err
}

func (a app) Update(ctx context.Context, item model.Repository) error {
	return a.repositoryStore.DoAtomic(ctx, func(ctx context.Context) error {
		old, err := a.repositoryStore.Get(ctx, item.ID, true)
		if err != nil {
			return service.WrapInternal(fmt.Errorf("unable to fetch: %s", err))
		}

		new := model.Repository{
			ID:      old.ID,
			Name:    old.Name,
			Version: item.Version,
		}

		if err := a.check(ctx, old, new); err != nil {
			return service.WrapInvalid(err)
		}

		if err := a.repositoryStore.Update(ctx, new); err != nil {
			return service.WrapInternal(fmt.Errorf("unable to update: %s", err))
		}

		return nil
	})
}

func (a app) Clean(ctx context.Context) error {
	return a.repositoryStore.DoAtomic(ctx, func(ctx context.Context) error {
		if err := a.repositoryStore.DeleteUnused(ctx); err != nil {
			return service.WrapInternal(fmt.Errorf("unable to delete: %s", err))
		}

		return nil
	})
}

func (a app) check(ctx context.Context, old, new model.Repository) error {
	if new == model.NoneRepository {
		return nil
	}

	output := make([]error, 0)

	if len(strings.TrimSpace(new.Name)) == 0 {
		output = append(output, errors.New("name is required"))
	}

	if old != model.NoneRepository && len(strings.TrimSpace(new.Version)) == 0 {
		output = append(output, errors.New("version is required"))
	}

	repositoryWithName, err := a.repositoryStore.GetByName(ctx, new.Name)
	if err != nil {
		output = append(output, errors.New("unable to check if name already exists"))
	} else if repositoryWithName != model.NoneRepository && repositoryWithName.ID != new.ID {
		output = append(output, errors.New("name already exists"))
	}

	return service.ConcatError(output)
}

func sanitizeName(name string) string {
	matches := nameMatcher.FindStringSubmatch(name)
	if len(matches) > 0 {
		return matches[len(matches)-1]
	}

	return name
}
