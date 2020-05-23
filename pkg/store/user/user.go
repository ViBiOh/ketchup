package user

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/ViBiOh/httputils/v3/pkg/db"
	"github.com/ViBiOh/ketchup/pkg/model"
)

// App of package
type App interface {
	DoAtomic(ctx context.Context, action func(context.Context) error) error

	GetByEmail(ctx context.Context, email string) (model.User, error)
	GetByLoginID(ctx context.Context, loginID uint64) (model.User, error)
	Create(ctx context.Context, o model.User) (uint64, error)
}

type app struct {
	db *sql.DB
}

// New creates new App from Config
func New(db *sql.DB) App {
	return app{
		db: db,
	}
}

func (a app) DoAtomic(ctx context.Context, action func(context.Context) error) error {
	return db.DoAtomic(ctx, a.db, action)
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
			item = model.NoneUser
			return nil
		}

		return err
	}

	err := db.Get(ctx, a.db, scanner, getByEmailQuery, email)
	return item, err
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
			item = model.NoneUser
			return nil
		}

		return err
	}

	err := db.Get(ctx, a.db, scanner, getByLoginIDQuery, loginID)
	return item, err
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
	return db.Create(ctx, insertQuery, strings.ToLower(o.Email), o.Login.ID)
}
