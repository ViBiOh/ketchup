package user

import (
	"context"

	"github.com/ViBiOh/ketchup/pkg/model"
	"github.com/jackc/pgx/v5"
)

type Service struct {
	db model.Database
}

func New(db model.Database) Service {
	return Service{
		db: db,
	}
}

func (s Service) DoAtomic(ctx context.Context, action func(context.Context) error) error {
	return s.db.DoAtomic(ctx, action)
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

func (s Service) GetByEmail(ctx context.Context, email string) (model.User, error) {
	var item model.User
	scanner := func(row pgx.Row) (err error) {
		switch err = row.Scan(&item.ID, &item.Email, &item.Login.ID); err {
		case pgx.ErrNoRows:
			err = nil
		}

		return err
	}

	return item, s.db.Get(ctx, scanner, getByEmailQuery, email)
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

func (s Service) GetByLoginID(ctx context.Context, loginID uint64) (model.User, error) {
	var item model.User
	scanner := func(row pgx.Row) (err error) {
		switch err = row.Scan(&item.ID, &item.Email, &item.Login.ID); err {
		case pgx.ErrNoRows:
			return nil
		}

		return err
	}

	return item, s.db.Get(ctx, scanner, getByLoginIDQuery, loginID)
}

const insertQuery = `
INSERT INTO
  ketchup.user
(
  email,
  login_id,
) VALUES (
  $1,
  $2,
) RETURNING id
`

func (s Service) Create(ctx context.Context, o model.User) (model.Identifier, error) {
	id, err := s.db.Create(ctx, insertQuery, o.Email, o.Login.ID)

	return model.Identifier(id), err
}

const countQuery = `
SELECT
  COUNT(1)
FROM
  ketchup.user
`

func (s Service) Count(ctx context.Context) (uint64, error) {
	var count uint64

	return count, s.db.Get(ctx, func(row pgx.Row) error {
		return row.Scan(&count)
	}, countQuery)
}
