package repository

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	httpModel "github.com/ViBiOh/httputils/v3/pkg/model"
	"github.com/ViBiOh/ketchup/pkg/github"
	"github.com/ViBiOh/ketchup/pkg/helm"
	"github.com/ViBiOh/ketchup/pkg/model"
	"github.com/ViBiOh/ketchup/pkg/semver"
	"github.com/ViBiOh/ketchup/pkg/service"
	"github.com/ViBiOh/ketchup/pkg/store/repository"
)

var (
	nameMatcher = regexp.MustCompile(`(?i)(?:github\.com/)?([^/\n]+/[^/\n]+)`)
)

// App of package
type App interface {
	List(ctx context.Context, page, pageSize uint) ([]model.Repository, uint64, error)
	Suggest(ctx context.Context, ignoreIds []uint64, count uint64) ([]model.Repository, error)
	GetOrCreate(ctx context.Context, name string, repositoryKind model.RepositoryKind) (model.Repository, error)
	Clean(ctx context.Context) error
	LatestVersion(repo model.Repository) (semver.Version, error)
}

type app struct {
	repositoryStore repository.App
	githubApp       github.App
	helmApp         helm.App
}

// New creates new App from Config
func New(repositoryStore repository.App, githubApp github.App, helmApp helm.App) App {
	return app{
		repositoryStore: repositoryStore,
		githubApp:       githubApp,
		helmApp:         helmApp,
	}
}

func (a app) List(ctx context.Context, page, pageSize uint) ([]model.Repository, uint64, error) {
	list, total, err := a.repositoryStore.List(ctx, page, pageSize)
	if err != nil {
		return nil, 0, httpModel.WrapInternal(fmt.Errorf("unable to list: %s", err))
	}

	return list, total, nil
}

func (a app) Suggest(ctx context.Context, ignoreIds []uint64, count uint64) ([]model.Repository, error) {
	list, err := a.repositoryStore.Suggest(ctx, ignoreIds, count)
	if err != nil {
		return nil, httpModel.WrapInternal(fmt.Errorf("unable to suggest: %s", err))
	}

	return list, nil
}

func (a app) GetOrCreate(ctx context.Context, name string, repositoryKind model.RepositoryKind) (model.Repository, error) {
	sanitizedName := name
	if repositoryKind == model.Github {
		sanitizedName = sanitizeName(name)
	}

	repo, err := a.repositoryStore.GetByName(ctx, sanitizedName, repositoryKind)
	if err != nil {
		return model.NoneRepository, httpModel.WrapInternal(err)
	}

	if repo != model.NoneRepository {
		return repo, nil
	}

	return a.create(ctx, model.Repository{Name: sanitizedName, Kind: repositoryKind})
}

func (a app) create(ctx context.Context, item model.Repository) (model.Repository, error) {
	if err := a.check(ctx, model.NoneRepository, item); err != nil {
		return model.NoneRepository, httpModel.WrapInvalid(err)
	}

	return item, a.repositoryStore.DoAtomic(ctx, func(ctx context.Context) error {
		id, err := a.repositoryStore.Create(ctx, item)
		if err != nil {
			return httpModel.WrapInternal(fmt.Errorf("unable to create: %s", err))
		}

		item.ID = id
		return nil
	})
}

func (a app) Clean(ctx context.Context) error {
	return a.repositoryStore.DoAtomic(ctx, func(ctx context.Context) error {
		if err := a.repositoryStore.DeleteUnused(ctx); err != nil {
			return httpModel.WrapInternal(fmt.Errorf("unable to delete: %s", err))
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

	repositoryWithName, err := a.repositoryStore.GetByName(ctx, new.Name, new.Kind)
	if err != nil {
		output = append(output, errors.New("unable to check if name already exists"))
	} else if repositoryWithName != model.NoneRepository && repositoryWithName.ID != new.ID {
		output = append(output, errors.New("name already exists"))
	}

	return service.ConcatError(output)
}

func (a app) LatestVersion(repo model.Repository) (semver.Version, error) {
	switch repo.Kind {
	case model.Github:
		return a.githubApp.LatestVersion(repo.Name)
	case model.Helm:
		return a.helmApp.LatestVersion(repo.Name)
	default:
		return semver.NoneVersion, fmt.Errorf("unknown repository kind %d", repo.Kind)
	}
}

func sanitizeName(name string) string {
	matches := nameMatcher.FindStringSubmatch(name)
	if len(matches) > 0 {
		return strings.TrimSpace(matches[len(matches)-1])
	}

	return strings.TrimSpace(name)
}
