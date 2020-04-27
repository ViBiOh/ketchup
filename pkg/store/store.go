package store

import (
	"context"
	"database/sql"
	"regexp"

	authStore "github.com/ViBiOh/auth/v2/pkg/store/db"
	"github.com/ViBiOh/ketchup/pkg/model"
)

var (
	sortKeyMatcher = regexp.MustCompile(`[A-Za-z0-9]+`)
)

// App of package
type App interface {
	ListUsers(ctx context.Context, page, pageSize uint, sortKey string, sortAsc bool) ([]model.User, uint, error)
	GetUser(ctx context.Context, id uint64) (model.User, error)
	GetUserByEmail(ctx context.Context, email string) (model.User, error)
	CreateUser(ctx context.Context, o model.User) (uint64, error)
	UpdateUser(ctx context.Context, o model.User) error
	DeleteUser(ctx context.Context, o model.User) error

	ListRepositories(ctx context.Context, page, pageSize uint, sortKey string, sortAsc bool) ([]model.Repository, uint, error)
	GetRepository(ctx context.Context, id uint64) (model.Repository, error)
	CreateRepository(ctx context.Context, o model.Repository) (uint64, error)
	UpdateRepository(ctx context.Context, o model.Repository) error
	DeleteRepository(ctx context.Context, o model.Repository) error
}

type app struct {
	db        *sql.DB
	authStore authStore.App
}

// New creates new App from Config
func New(db *sql.DB) App {
	return app{
		db:        db,
		authStore: authStore.New(db),
	}
}
