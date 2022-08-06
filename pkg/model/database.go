package model

import (
	"context"

	"github.com/jackc/pgx/v4"
)

// Database interface needed
//
//go:generate mockgen -destination ../mocks/database.go -mock_names Database=Database -package mocks github.com/ViBiOh/ketchup/pkg/model Database
type Database interface {
	List(context.Context, func(pgx.Rows) error, string, ...any) error
	Get(context.Context, func(pgx.Row) error, string, ...any) error
	Create(context.Context, string, ...any) (uint64, error)
	Exec(context.Context, string, ...any) error
	One(context.Context, string, ...any) error
	DoAtomic(context.Context, func(context.Context) error) error
}
