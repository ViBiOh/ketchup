package ketchuptest

import (
	"context"
	"errors"

	"github.com/ViBiOh/ketchup/pkg/model"
	"github.com/ViBiOh/ketchup/pkg/service/ketchup"
)

var _ ketchup.App = app{}

// NewApp creates mock
func NewApp() ketchup.App {
	return app{}
}

type app struct{}

func (tks app) List(ctx context.Context, _, _ uint) ([]model.Ketchup, uint64, error) {
	if model.ReadUser(ctx) == model.NoneUser {
		return nil, 0, errors.New("user not found")
	}

	return []model.Ketchup{
		{Kind: "release", Upstream: "1.0.0", Current: "1.0.0", Repository: model.Repository{ID: 1, Name: "vibioh/ketchup"}, User: model.User{ID: 1, Email: "nobody@localhost"}},
	}, 0, nil
}

func (tks app) ListForRepositories(ctx context.Context, repositories []model.Repository) ([]model.Ketchup, error) {
	if ctx == context.TODO() {
		return nil, errors.New("invalid context")
	}

	if len(repositories) == 0 {
		return nil, nil
	}

	if repositories[0].Name == "vibioh/ketchup" {
		return []model.Ketchup{
			{Repository: repositories[0], User: model.User{ID: 1, Email: "nobody@localhost"}},
			{Repository: repositories[0], User: model.User{ID: 2, Email: "guest@nowhere"}},
		}, nil
	}

	if repositories[0].Name == "vibioh/viws" {
		return []model.Ketchup{
			{Repository: repositories[0], User: model.User{ID: 1, Email: "nobody@localhost"}},
			{Repository: repositories[0], User: model.User{ID: 2, Email: "guest@nowhere"}},
			{Repository: repositories[1], User: model.User{ID: 2, Email: "guest@nowhere"}},
		}, nil
	}

	return nil, nil
}

func (tks app) ListKindsByRepositoryID(ctx context.Context, repository model.Repository) ([]string, error) {
	return nil, nil
}

func (tks app) Create(_ context.Context, _ model.Ketchup) (model.Ketchup, error) {
	return model.NoneKetchup, nil
}

func (tks app) Update(_ context.Context, _ model.Ketchup) (model.Ketchup, error) {
	return model.NoneKetchup, nil
}

func (tks app) Delete(_ context.Context, _ model.Ketchup) error {
	return nil
}
