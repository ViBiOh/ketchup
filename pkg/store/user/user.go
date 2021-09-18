package user

import (
	"context"

	"github.com/ViBiOh/ketchup/pkg/model"
	"github.com/jackc/pgx/v4"
)

// App of the package
type App struct {
	db model.Database
}

// New creates new App from Config
func New(db model.Database) App {
	return App{
		db: db,
	}
}

// DoAtomic does an atomic operation
func (a App) DoAtomic(ctx context.Context, action func(context.Context) error) error {
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

// GetByEmail retrieve user by email
func (a App) GetByEmail(ctx context.Context, email string) (model.User, error) {
	var item model.User
	scanner := func(row pgx.Row) error {
		err := row.Scan(&item.ID, &item.Email, &item.Login.ID)
		if err == pgx.ErrNoRows {
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

// GetByLoginID retrieves user by id
func (a App) GetByLoginID(ctx context.Context, loginID uint64) (model.User, error) {
	var item model.User
	scanner := func(row pgx.Row) error {
		err := row.Scan(&item.ID, &item.Email, &item.Login.ID)
		if err == pgx.ErrNoRows {
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

// Create user
func (a App) Create(ctx context.Context, o model.User) (uint64, error) {
	return a.db.Create(ctx, insertQuery, o.Email, o.Login.ID)
}

const countQuery = `
SELECT
  COUNT(1)
FROM
  ketchup.user
`

// Count users
func (a App) Count(ctx context.Context) (uint64, error) {
	var count uint64

	return count, a.db.Get(ctx, func(row pgx.Row) error {
		return row.Scan(&count)
	}, countQuery)
}
