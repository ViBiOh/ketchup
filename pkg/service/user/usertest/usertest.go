package usertest

import (
	"context"

	"github.com/ViBiOh/ketchup/pkg/model"
	"github.com/ViBiOh/ketchup/pkg/service/user"
)

var _ user.App = app{}

// NewApp creates mock
func NewApp() user.App {
	return app{}
}

type app struct{}

func (a app) Create(_ context.Context, _ model.User) (model.User, error) {
	return model.NoneUser, nil
}

func (a app) StoreInContext(ctx context.Context) context.Context {
	return model.StoreUser(ctx, model.User{Email: "nobody@localhost"})
}
