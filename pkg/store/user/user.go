package user

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/ViBiOh/httputils/v4/pkg/db"
	"github.com/ViBiOh/ketchup/pkg/model"
)

// App of package
type App interface {
	DoAtomic(ctx context.Context, action func(context.Context) error) error

	GetByEmail(ctx context.Context, email string) (model.User, error)
	GetByLoginID(ctx context.Context, loginID uint64) (model.User, error)
	Create(ctx context.Context, o model.User) (uint64, error)
	Count(ctx context.Context) (uint64, error)
}

type app struct {
	db db.App
}

// New creates new App from Config
func New(db db.App) App {
	return app{
		db: db,
	}
}

func (a app) DoAtomic(ctx context.Context, action func(context.Context) error) error {
	return a.db.DoAtomic(ctx, action)
}

const getByEmailQuery = `
SELECT
  id,
  email,
  login_id
FROM
  ketchup.user
WHERE
  email = $1
`

func (a app) GetByEmail(ctx context.Context, email string) (model.User, error) {
	var item model.User
	scanner := func(row *sql.Row) error {
		err := row.Scan(&item.ID, &item.Email, &item.Login.ID)
		if errors.Is(err, sql.ErrNoRows) {
			item = model.User{}
			return nil
		}

		return err
	}

	return item, a.db.Get(ctx, scanner, getByEmailQuery, email)
}

const getByLoginIDQuery = `
SELECT
  id,
  email,
  login_id
FROM
  ketchup.user
WHERE
  login_id = $1
`

func (a app) GetByLoginID(ctx context.Context, loginID uint64) (model.User, error) {
	var item model.User
	scanner := func(row *sql.Row) error {
		err := row.Scan(&item.ID, &item.Email, &item.Login.ID)
		if errors.Is(err, sql.ErrNoRows) {
			item = model.User{}
			return nil
		}

		return err
	}

	return item, a.db.Get(ctx, scanner, getByLoginIDQuery, loginID)
}

const insertQuery = `
INSERT INTO
  ketchup.user
(
  email,
  login_id
) VALUES (
  $1,
  $2
) RETURNING id
`

func (a app) Create(ctx context.Context, o model.User) (uint64, error) {
	return a.db.Create(ctx, insertQuery, strings.ToLower(o.Email), o.Login.ID)
}

const countQuery = `
SELECT
  COUNT(1)
FROM
  ketchup.user
`

func (a app) Count(ctx context.Context) (uint64, error) {
	var count uint64

	return count, a.db.Get(ctx, func(row *sql.Row) error {
		return row.Scan(&count)
	}, countQuery)
}
