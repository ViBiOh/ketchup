package repository

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	httpModel "github.com/ViBiOh/httputils/v4/pkg/model"
	"github.com/ViBiOh/ketchup/pkg/model"
	"github.com/ViBiOh/ketchup/pkg/semver"
)

var nameMatcher = regexp.MustCompile(`(?i)(?:github\.com/)?([^/\n]+/[^/\n]+)`)

type Service struct {
	repository model.RepositoryStore
	github     model.GenericProvider
	helm       model.HelmProvider
	docker     model.GenericProvider
	npm        model.GenericProvider
	pypi       model.GenericProvider
}

func New(repositoryStore model.RepositoryStore, githubService model.GenericProvider, helmService model.HelmProvider, dockerService model.GenericProvider, npmService model.GenericProvider, pypiService model.GenericProvider) Service {
	return Service{
		repository: repositoryStore,
		github:     githubService,
		helm:       helmService,
		docker:     dockerService,
		npm:        npmService,
		pypi:       pypiService,
	}
}

func (s Service) List(ctx context.Context, pageSize uint, last string) ([]model.Repository, uint64, error) {
	list, total, err := s.repository.List(ctx, pageSize, last)
	if err != nil {
		return nil, 0, httpModel.WrapInternal(fmt.Errorf("list: %w", err))
	}

	return list, total, nil
}

func (s Service) ListByKinds(ctx context.Context, pageSize uint, last string, kinds ...model.RepositoryKind) ([]model.Repository, uint64, error) {
	list, total, err := s.repository.ListByKinds(ctx, pageSize, last, kinds...)
	if err != nil {
		return nil, 0, httpModel.WrapInternal(fmt.Errorf("list by kind: %w", err))
	}

	return list, total, nil
}

func (s Service) Suggest(ctx context.Context, ignoreIds []model.Identifier, count uint64) ([]model.Repository, error) {
	list, err := s.repository.Suggest(ctx, ignoreIds, count)
	if err != nil {
		return nil, httpModel.WrapInternal(fmt.Errorf("suggest: %w", err))
	}

	return list, nil
}

func (s Service) GetOrCreate(ctx context.Context, kind model.RepositoryKind, name, part, pattern string) (model.Repository, error) {
	sanitizedName := name
	if kind == model.Github {
		sanitizedName = sanitizeName(name)
	}

	repo, err := s.repository.GetByName(ctx, kind, sanitizedName, part)
	if err != nil {
		return model.NewEmptyRepository(), httpModel.WrapInternal(err)
	}

	if repo.IsZero() {
		return s.create(ctx, model.NewRepository(0, kind, sanitizedName, part).AddVersion(pattern, ""))
	}

	if repo.Versions[pattern] != "" {
		return repo, nil
	}

	repo.Versions[pattern] = ""
	versions, err := s.LatestVersions(ctx, repo)
	if err != nil {
		return model.NewEmptyRepository(), fmt.Errorf("get releases for `%s`: %w", repo.Name, err)
	}

	version, ok := versions[pattern]
	if !ok || version.IsZero() {
		return model.NewEmptyRepository(), httpModel.WrapNotFound(fmt.Errorf("no release with pattern `%s` found for repository `%s`", pattern, repo.Name))
	}

	repo.Versions[pattern] = version.Name
	if err := s.repository.UpdateVersions(ctx, repo); err != nil {
		return model.NewEmptyRepository(), fmt.Errorf("update repository versions `%s`: %w", repo.Name, err)
	}

	return repo, nil
}

func (s Service) create(ctx context.Context, item model.Repository) (model.Repository, error) {
	if err := s.check(ctx, model.NewEmptyRepository(), item); err != nil {
		return model.NewEmptyRepository(), httpModel.WrapInvalid(err)
	}

	versions, err := s.LatestVersions(ctx, item)
	if err != nil {
		return model.NewEmptyRepository(), httpModel.WrapNotFound(fmt.Errorf("get releases for `%s`: %w", item.Name, err))
	}

	for pattern, version := range versions {
		item.Versions[pattern] = version.Name
	}

	err = s.repository.DoAtomic(ctx, func(ctx context.Context) error {
		id, err := s.repository.Create(ctx, item)
		if err != nil {
			return httpModel.WrapInternal(fmt.Errorf("create: %w", err))
		}

		item.ID = id
		return nil
	})

	return item, err
}

func (s Service) Update(ctx context.Context, item model.Repository) error {
	return s.repository.DoAtomic(ctx, func(ctx context.Context) error {
		old, err := s.repository.Get(ctx, item.ID, true)
		if err != nil {
			return httpModel.WrapInternal(fmt.Errorf("fetch: %w", err))
		}

		current := model.Repository{
			ID:       old.ID,
			Kind:     old.Kind,
			Name:     old.Name,
			Versions: item.Versions,
		}

		if err := s.check(ctx, old, current); err != nil {
			return httpModel.WrapInvalid(err)
		}

		if err := s.repository.UpdateVersions(ctx, current); err != nil {
			return httpModel.WrapInternal(fmt.Errorf("update: %w", err))
		}

		return nil
	})
}

func (s Service) Clean(ctx context.Context) error {
	return s.repository.DoAtomic(ctx, func(ctx context.Context) error {
		if err := s.repository.DeleteUnused(ctx); err != nil {
			return httpModel.WrapInternal(fmt.Errorf("delete unused repository: %w", err))
		}

		if err := s.repository.DeleteUnusedVersions(ctx); err != nil {
			return httpModel.WrapInternal(fmt.Errorf("delete unused repository versions: %w", err))
		}

		return nil
	})
}

func (s Service) check(ctx context.Context, old, new model.Repository) error {
	if new.IsZero() && new.Kind == 0 && len(new.Name) == 0 {
		return nil
	}

	var output []error

	if len(strings.TrimSpace(new.Name)) == 0 {
		output = append(output, errors.New("name is required"))
	}

	if !old.ID.IsZero() && new.Kind != old.Kind {
		output = append(output, errors.New("kind cannot be changed"))
	}

	if !old.ID.IsZero() && len(new.Versions) == 0 {
		output = append(output, errors.New("version is required"))
	}

	repositoryWithName, err := s.repository.GetByName(ctx, new.Kind, new.Name, new.Part)
	if err != nil {
		output = append(output, errors.New("check if name already exists"))
	} else if !repositoryWithName.IsZero() && repositoryWithName.ID != new.ID {
		output = append(output, errors.New("name already exists"))
	}

	return httpModel.ConcatError(output)
}

func (s Service) LatestVersions(ctx context.Context, repo model.Repository) (map[string]semver.Version, error) {
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
		return s.github.LatestVersions(ctx, repo.Name, patterns)
	case model.Helm:
		return s.helm.LatestVersions(ctx, repo.Name, repo.Part, patterns)
	case model.Docker:
		return s.docker.LatestVersions(ctx, repo.Name, patterns)
	case model.NPM:
		return s.npm.LatestVersions(ctx, repo.Name, patterns)
	case model.Pypi:
		return s.pypi.LatestVersions(ctx, repo.Name, patterns)
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
