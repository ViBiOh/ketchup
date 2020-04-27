package service

import (
	"context"

	"github.com/ViBiOh/auth/v2/pkg/auth"
	authService "github.com/ViBiOh/auth/v2/pkg/service"
	"github.com/ViBiOh/httputils/v3/pkg/crud"
	"github.com/ViBiOh/ketchup/pkg/store"
)

var (
	_ crud.Service = app{}
)

// App of package
type App interface {
	Unmarshal(data []byte, contentType string) (interface{}, error)
	Check(ctx context.Context, old, new interface{}) []crud.Error
	List(ctx context.Context, page, pageSize uint, sortKey string, sortDesc bool, filters map[string][]string) ([]interface{}, uint, error)
	Get(ctx context.Context, ID uint64) (interface{}, error)
	Create(ctx context.Context, o interface{}) (interface{}, error)
	Update(ctx context.Context, o interface{}) (interface{}, error)
	Delete(ctx context.Context, o interface{}) error
}

type app struct {
	storeApp store.App

	authService  authService.App
	authProvider auth.Provider
}

// New creates new App from Config
func New(storeApp store.App, authService authService.App, authProvider auth.Provider) App {
	return app{
		storeApp: storeApp,

		authService:  authService,
		authProvider: authProvider,
	}
}
