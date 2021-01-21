package repositorytest

import (
	"context"
	"errors"
	"regexp"

	httpModel "github.com/ViBiOh/httputils/v3/pkg/model"
	"github.com/ViBiOh/ketchup/pkg/model"
	"github.com/ViBiOh/ketchup/pkg/semver"
	"github.com/ViBiOh/ketchup/pkg/service/repository"
)

var _ repository.App = app{}

type app struct {
	multiple bool
	name     *regexp.Regexp
	version  string
}

// NewApp creates mock
func NewApp(multiple bool, name *regexp.Regexp, version string) repository.App {
	return app{
		multiple: multiple,
		name:     name,
		version:  version,
	}
}

func (a app) List(ctx context.Context, page, _ uint) ([]model.Repository, uint64, error) {
	if ctx == context.TODO() {
		return nil, 0, errors.New("invalid context")
	}

	if a.multiple {
		if page == 1 {
			return []model.Repository{
				{
					ID:   1,
					Name: "vibioh/viws",
				},
			}, 2, nil
		} else if page == 2 {
			return []model.Repository{
				{
					ID:   2,
					Name: "vibioh/ketchup",
				},
			}, 2, nil
		}
	}

	return []model.Repository{
		{
			ID:   1,
			Name: "vibioh/ketchup",
		},
	}, 1, nil
}

func (a app) Suggest(_ context.Context, _ []uint64, _ uint64) ([]model.Repository, error) {
	return nil, nil
}

func (a app) GetOrCreate(_ context.Context, name string, repositoryKind model.RepositoryKind) (model.Repository, error) {
	if len(name) == 0 {
		return model.NoneRepository, httpModel.WrapInvalid(errors.New("invalid name"))
	}

	return model.Repository{ID: 1, Name: "vibioh/ketchup"}, nil
}

func (a app) Update(_ context.Context, item model.Repository) error {
	if item.Name == "1.0.1" {
		return errors.New("update error")
	}

	return nil
}

func (a app) Clean(_ context.Context) error {
	return nil
}

func (a app) LatestVersion(repo model.Repository) (semver.Version, error) {
	if a.name.MatchString(repo.Name) {
		version, _ := semver.Parse(a.version)
		return version, nil
	}

	return semver.NoneVersion, errors.New("unknown repository")
}
