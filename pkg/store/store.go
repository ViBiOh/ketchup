package store

import (
	"context"
	"database/sql"

	"github.com/ViBiOh/httputils/v3/pkg/db"
)

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
