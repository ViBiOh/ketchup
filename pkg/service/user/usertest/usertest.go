package usertest

import (
	"context"

	authModel "github.com/ViBiOh/auth/v2/pkg/model"
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
	return model.StoreUser(ctx, model.NewUser(0, "nobody@localhost", authModel.NewUser(0, "")))
}
