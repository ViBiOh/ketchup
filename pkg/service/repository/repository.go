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
	GetOrCreate(ctx context.Context, name string, repositoryKind model.RepositoryKind, pattern string) (model.Repository, error)
	Update(ctx context.Context, item model.Repository) error
	Clean(ctx context.Context) error
	LatestVersions(repo model.Repository) (map[string]semver.Version, error)
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
		return nil, 0, httpModel.WrapInternal(fmt.Errorf("unable to list: %w", err))
	}

	return list, total, nil
}

func (a app) Suggest(ctx context.Context, ignoreIds []uint64, count uint64) ([]model.Repository, error) {
	list, err := a.repositoryStore.Suggest(ctx, ignoreIds, count)
	if err != nil {
		return nil, httpModel.WrapInternal(fmt.Errorf("unable to suggest: %w", err))
	}

	return list, nil
}

func (a app) GetOrCreate(ctx context.Context, name string, repositoryKind model.RepositoryKind, pattern string) (model.Repository, error) {
	sanitizedName := name
	if repositoryKind == model.Github {
		sanitizedName = sanitizeName(name)
	}

	repo, err := a.repositoryStore.GetByName(ctx, sanitizedName, repositoryKind)
	if err != nil {
		return model.NoneRepository, httpModel.WrapInternal(err)
	}

	if repo.ID == 0 {
		return a.create(ctx, model.NewRepository(0, repositoryKind, sanitizedName).AddVersion(pattern, ""))
	}

	if repo.Versions[pattern] != "" {
		return repo, nil
	}

	repo.Versions[pattern] = ""
	versions, err := a.LatestVersions(repo)
	if err != nil {
		return model.NoneRepository, httpModel.WrapInternal(fmt.Errorf("unable to get releases for `%s`: %w", repo.Name, err))
	}

	version, ok := versions[pattern]
	if !ok || version == semver.NoneVersion {
		return model.NoneRepository, httpModel.WrapNotFound(fmt.Errorf("no release with pattern `%s` found for repository `%s`", pattern, repo.Name))
	}

	repo.Versions[pattern] = version.Name
	if err := a.repositoryStore.UpdateVersions(ctx, repo); err != nil {
		return model.NoneRepository, httpModel.WrapInternal(fmt.Errorf("unable to update repository versions `%s`: %w", repo.Name, err))
	}

	return repo, nil
}

func (a app) create(ctx context.Context, item model.Repository) (model.Repository, error) {
	if err := a.check(ctx, model.NoneRepository, item); err != nil {
		return model.NoneRepository, httpModel.WrapInvalid(err)
	}

	versions, err := a.LatestVersions(item)
	if err != nil {
		return model.NoneRepository, httpModel.WrapNotFound(fmt.Errorf("unable to get releases for `%s`: %w", item.Name, err))
	}

	for pattern, version := range versions {
		item.Versions[pattern] = version.Name
	}

	err = a.repositoryStore.DoAtomic(ctx, func(ctx context.Context) error {
		id, err := a.repositoryStore.Create(ctx, item)
		if err != nil {
			return httpModel.WrapInternal(fmt.Errorf("unable to create: %w", err))
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
			return httpModel.WrapInternal(fmt.Errorf("unable to fetch: %w", err))
		}

		current := model.Repository{
			ID:       old.ID,
			Kind:     old.Kind,
			Name:     old.Name,
			Versions: item.Versions,
		}

		if err := a.check(ctx, old, current); err != nil {
			return httpModel.WrapInvalid(err)
		}

		if err := a.repositoryStore.UpdateVersions(ctx, current); err != nil {
			return httpModel.WrapInternal(fmt.Errorf("unable to update: %w", err))
		}

		return nil
	})
}

func (a app) Clean(ctx context.Context) error {
	return a.repositoryStore.DoAtomic(ctx, func(ctx context.Context) error {
		if err := a.repositoryStore.DeleteUnused(ctx); err != nil {
			return httpModel.WrapInternal(fmt.Errorf("unable to delete unused repository: %w", err))
		}

		if err := a.repositoryStore.DeleteUnusedVersions(ctx); err != nil {
			return httpModel.WrapInternal(fmt.Errorf("unable to delete unused repository versions: %w", err))
		}

		return nil
	})
}

func (a app) check(ctx context.Context, old, new model.Repository) error {
	if new.ID == 0 && new.Kind == 0 && len(new.Name) == 0 {
		return nil
	}

	output := make([]error, 0)

	if len(strings.TrimSpace(new.Name)) == 0 {
		output = append(output, errors.New("name is required"))
	}

	if old.ID != 0 && new.Kind != old.Kind {
		output = append(output, errors.New("kind cannot be changed"))
	}

	if old.ID != 0 && len(new.Versions) == 0 {
		output = append(output, errors.New("version is required"))
	}

	repositoryWithName, err := a.repositoryStore.GetByName(ctx, new.Name, new.Kind)
	if err != nil {
		output = append(output, errors.New("unable to check if name already exists"))
	} else if repositoryWithName.ID != 0 && repositoryWithName.ID != new.ID {
		output = append(output, errors.New("name already exists"))
	}

	return service.ConcatError(output)
}

func (a app) LatestVersions(repo model.Repository) (map[string]semver.Version, error) {
	if len(repo.Versions) == 0 {
		return nil, errors.New("no pattern for fetching latest versions")
	}

	index := 0
	patterns := make([]string, len(repo.Versions))
	for pattern := range repo.Versions {
		patterns[index] = pattern
		index++
	}

	switch repo.Kind {
	case model.Github:
		return a.githubApp.LatestVersions(repo.Name, patterns)
	case model.Helm:
		return a.helmApp.LatestVersions(repo.Name, patterns)
	default:
		return nil, fmt.Errorf("unknown repository kind %d", repo.Kind)
	}
}

func sanitizeName(name string) string {
	matches := nameMatcher.FindStringSubmatch(name)
	if len(matches) > 0 {
		return strings.TrimSpace(matches[len(matches)-1])
	}

	return strings.TrimSpace(name)
}
