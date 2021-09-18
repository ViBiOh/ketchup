package model

import (
	"context"

	"github.com/jackc/pgx/v4"
)

// Database interface needed
type Database interface {
	List(context.Context, func(pgx.Rows) error, string, ...interface{}) error
	Get(context.Context, func(pgx.Row) error, string, ...interface{}) error
	Create(context.Context, string, ...interface{}) (uint64, error)
	Exec(context.Context, string, ...interface{}) error
	DoAtomic(context.Context, func(context.Context) error) error
}
