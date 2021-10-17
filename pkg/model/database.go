package model

import (
	"context"

	"github.com/jackc/pgx/v4"
)

// Database interface needed
//go:generate mockgen -destination ../mocks/database.go -mock_names Database=Database -package mocks github.com/ViBiOh/ketchup/pkg/model Database
type Database interface {
	List(context.Context, func(pgx.Rows) error, string, ...interface{}) error
	Get(context.Context, func(pgx.Row) error, string, ...interface{}) error
	Create(context.Context, string, ...interface{}) (uint64, error)
	Exec(context.Context, string, ...interface{}) error
	One(context.Context, string, ...interface{}) error
	DoAtomic(context.Context, func(context.Context) error) error
}
