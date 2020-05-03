package user

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/ViBiOh/httputils/v3/pkg/db"
	"github.com/ViBiOh/ketchup/pkg/model"
	"github.com/ViBiOh/ketchup/pkg/store"
)

// App of package
type App interface {
	StartAtomic(ctx context.Context) (context.Context, error)
	EndAtomic(ctx context.Context, err error) error

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

func (a app) StartAtomic(ctx context.Context) (context.Context, error) {
	return store.StartAtomic(ctx, a.db)
}

func (a app) EndAtomic(ctx context.Context, err error) error {
	return store.EndAtomic(ctx, err)
}

const getByEmailQuery = `
SELECT
  id,
  email,
  login_id
FROM
  "user"
WHERE
  email = $1
`

func (a app) GetByEmail(ctx context.Context, email string) (model.User, error) {
	var item model.User
	scanner := func(row db.RowScanner) error {
		err := row.Scan(&item.ID, &item.Email, &item.Login.ID)
		if errors.Is(err, sql.ErrNoRows) {
			item = model.NoneUser
			return nil
		}

		return err
	}

	err := db.GetRow(ctx, a.db, scanner, getByEmailQuery, email)
	return item, err
}

const getByLoginIDQuery = `
SELECT
  id,
  email,
  login_id
FROM
  "user"
WHERE
  login_id = $1
`

func (a app) GetByLoginID(ctx context.Context, loginID uint64) (model.User, error) {
	var item model.User
	scanner := func(row db.RowScanner) error {
		err := row.Scan(&item.ID, &item.Email, &item.Login.ID)
		if errors.Is(err, sql.ErrNoRows) {
			item = model.NoneUser
			return nil
		}

		return err
	}

	err := db.GetRow(ctx, a.db, scanner, getByLoginIDQuery, loginID)
	return item, err
}

const insertQuery = `
INSERT INTO
  "user"
(
  email,
  login_id
) VALUES (
  $1,
  $2
) RETURNING id
`

func (a app) Create(ctx context.Context, o model.User) (uint64, error) {
	return db.Create(ctx, a.db, insertQuery, strings.ToLower(o.Email), o.Login.ID)
}
