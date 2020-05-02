package repository

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/ViBiOh/httputils/v3/pkg/logger"
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
	List(ctx context.Context, page, pageSize uint) ([]model.Repository, uint, error)
	Get(ctx context.Context, id uint64) (model.Repository, error)
	GetOrCreate(ctx context.Context, name string) (model.Repository, error)
	Create(ctx context.Context, item model.Repository) (model.Repository, error)
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

func (a app) List(ctx context.Context, page, pageSize uint) ([]model.Repository, uint, error) {
	list, total, err := a.repositoryStore.List(ctx, page, pageSize)
	if err != nil {
		return nil, 0, service.WrapInternal(fmt.Errorf("unable to list: %s", err))
	}

	return list, total, nil
}

func (a app) Get(ctx context.Context, id uint64) (model.Repository, error) {
	repository, err := a.repositoryStore.Get(ctx, id)
	if err != nil {
		return model.NoneRepository, service.WrapInternal(fmt.Errorf("unable to get: %s", err))
	}

	return repository, nil
}

func (a app) GetOrCreate(ctx context.Context, name string) (model.Repository, error) {
	matches := nameMatcher.FindAllStringSubmatch(name, -1)
	if len(matches) > 0 {
		name = matches[0][len(matches[0])-1]
	}

	repository, err := a.repositoryStore.GetByName(ctx, name)
	if err != nil {
		return model.NoneRepository, err
	}

	if repository != model.NoneRepository {
		return repository, nil
	}

	return a.Create(ctx, model.Repository{
		Name: name,
	})
}

func (a app) Create(ctx context.Context, item model.Repository) (model.Repository, error) {
	if err := a.check(ctx, model.NoneRepository, item); err != nil {
		return model.NoneRepository, service.WrapInvalid(err)
	}

	release, err := a.githubApp.LastRelease(item.Name)
	if err != nil {
		logger.Error("%s", err)
		return model.NoneRepository, fmt.Errorf("no release found for %s: %w", item.Name, service.ErrNotFound)
	}

	item.Version = release.TagName

	id, err := a.repositoryStore.Create(ctx, item)
	if err != nil {
		return model.NoneRepository, service.WrapInternal(fmt.Errorf("unable to create: %s", err))
	}

	item.ID = id

	return item, nil
}

func (a app) Update(ctx context.Context, item model.Repository) (err error) {
	ctx, err = a.repositoryStore.StartAtomic(ctx)
	if err != nil {
		return
	}

	defer func() {
		err = a.repositoryStore.EndAtomic(ctx, err)
	}()

	var old model.Repository
	old, err = a.repositoryStore.Get(ctx, item.ID)
	if err != nil {
		err = service.WrapInternal(fmt.Errorf("unable to fetch: %s", err))
	}

	if err = a.check(ctx, old, item); err != nil {
		err = service.WrapInvalid(err)
		return
	}

	if err = a.repositoryStore.Update(ctx, item); err != nil {
		err = service.WrapInternal(fmt.Errorf("unable to update: %s", err))
	}

	return
}

func (a app) Clean(ctx context.Context) error {
	if err := a.repositoryStore.DeleteUnused(ctx); err != nil {
		return service.WrapInternal(fmt.Errorf("unable to delete: %s", err))
	}

	return nil
}

func (a app) check(ctx context.Context, old, new model.Repository) error {
	output := make([]error, 0)

	if new == model.NoneRepository {
		return service.ConcatError(output)
	}

	if len(strings.TrimSpace(new.Name)) == 0 {
		output = append(output, errors.New("name is required"))
	}

	repositoryWithName, err := a.repositoryStore.GetByName(ctx, new.Name)
	if err != nil {
		output = append(output, errors.New("unable to check if name already exists"))
	} else if repositoryWithName != model.NoneRepository && repositoryWithName.ID != new.ID {
		output = append(output, errors.New("name already exists"))
	}

	return service.ConcatError(output)
}
