package store

import (
	"context"
	"database/sql"

	"github.com/ViBiOh/httputils/v3/pkg/db"
	"github.com/ViBiOh/ketchup/pkg/model"
)

// UserStore store User
type UserStore interface {
	StartAtomic(ctx context.Context) (context.Context, error)
	EndAtomic(ctx context.Context, err error) error

	List(ctx context.Context, page, pageSize uint, sortKey string, sortAsc bool) ([]model.User, uint, error)
	Get(ctx context.Context, id uint64) (model.User, error)
	GetByEmail(ctx context.Context, email string) (model.User, error)
	Create(ctx context.Context, o model.User) (uint64, error)
	Update(ctx context.Context, o model.User) error
	Delete(ctx context.Context, o model.User) error
}

// RepositoryStore store Repository
type RepositoryStore interface {
	StartAtomic(ctx context.Context) (context.Context, error)
	EndAtomic(ctx context.Context, err error) error

	List(ctx context.Context, page, pageSize uint, sortKey string, sortAsc bool) ([]model.Repository, uint, error)
	Get(ctx context.Context, id uint64) (model.Repository, error)
	Create(ctx context.Context, o model.Repository) (uint64, error)
	Update(ctx context.Context, o model.Repository) error
	Delete(ctx context.Context, o model.Repository) error
}

// KetchupStore store Ketchup
type KetchupStore interface {
	StartAtomic(ctx context.Context) (context.Context, error)
	EndAtomic(ctx context.Context, err error) error

	List(ctx context.Context, page, pageSize uint, sortKey string, sortAsc bool) ([]model.Ketchup, uint, error)
	Get(ctx context.Context, id uint64) (model.Ketchup, error)
	Create(ctx context.Context, o model.Ketchup) (uint64, error)
	Update(ctx context.Context, o model.Ketchup) error
	Delete(ctx context.Context, o model.Ketchup) error
}

// StartAtomic starts atomic work
func StartAtomic(ctx context.Context, usedDB *sql.DB) (context.Context, error) {
	if db.ReadTx(ctx) != nil {
		return ctx, nil
	}

	tx, err := usedDB.Begin()
	if err != nil {
		return ctx, err
	}

	return db.StoreTx(ctx, tx), nil
}

// EndAtomic ends atomic work
func EndAtomic(ctx context.Context, err error) error {
	tx := db.ReadTx(ctx)
	if tx == nil {
		return nil
	}

	return db.EndTx(tx, err)
}
