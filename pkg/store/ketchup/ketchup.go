package ketchup

import (
	"context"
	"database/sql"
	"errors"

	"github.com/ViBiOh/httputils/v3/pkg/db"
	"github.com/ViBiOh/ketchup/pkg/model"
	"github.com/lib/pq"
)

// App of package
type App interface {
	DoAtomic(ctx context.Context, action func(context.Context) error) error

	List(ctx context.Context, page, pageSize uint) ([]model.Ketchup, uint64, error)
	ListByRepositoriesID(ctx context.Context, ids []uint64) ([]model.Ketchup, error)
	GetByRepositoryID(ctx context.Context, id uint64, forUpdate bool) (model.Ketchup, error)
	Create(ctx context.Context, o model.Ketchup) (uint64, error)
	Update(ctx context.Context, o model.Ketchup) error
	Delete(ctx context.Context, o model.Ketchup) error
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

const listQuery = `
SELECT
  k.version,
  k.repository_id,
  r.name,
  r.kind,
  count(1) OVER() AS full_count
FROM
  ketchup.ketchup k,
  ketchup.repository r
WHERE
  user_id = $3
  AND repository_id = id
LIMIT $1
OFFSET $2
`

func (a app) List(ctx context.Context, page, pageSize uint) ([]model.Ketchup, uint64, error) {
	user := model.ReadUser(ctx)

	var totalCount uint64
	list := make([]model.Ketchup, 0)

	scanner := func(rows *sql.Rows) error {
		item := model.Ketchup{
			User: user,
		}
		var rawRepositoryKind string

		if err := rows.Scan(&item.Version, &item.Repository.ID, &item.Repository.Name, &rawRepositoryKind, &totalCount); err != nil {
			return err
		}

		repositoryKind, err := model.ParseRepositoryKind(rawRepositoryKind)
		if err != nil {
			return err
		}
		item.Repository.Kind = repositoryKind

		list = append(list, item)
		return nil
	}

	return list, totalCount, db.List(ctx, a.db, scanner, listQuery, pageSize, (page-1)*pageSize, user.ID)
}

const listByRepositoriesIDQuery = `
SELECT
  k.version,
  k.repository_id,
  k.user_id,
  u.email
FROM
  ketchup.ketchup k,
  ketchup.user u
WHERE
  repository_id = ANY ($1)
  AND k.user_id = u.id
`

func (a app) ListByRepositoriesID(ctx context.Context, ids []uint64) ([]model.Ketchup, error) {
	list := make([]model.Ketchup, 0)

	scanner := func(rows *sql.Rows) error {
		var item model.Ketchup
		if err := rows.Scan(&item.Version, &item.Repository.ID, &item.User.ID, &item.User.Email); err != nil {
			return err
		}

		list = append(list, item)
		return nil
	}

	return list, db.List(ctx, a.db, scanner, listByRepositoriesIDQuery, pq.Array(ids))
}

const getQuery = `
SELECT
  version,
  repository_id,
  user_id
FROM
  ketchup.ketchup
WHERE
  repository_id = $1
  AND user_id = $2
`

func (a app) GetByRepositoryID(ctx context.Context, id uint64, forUpdate bool) (model.Ketchup, error) {
	query := getQuery
	if forUpdate {
		query += " FOR UPDATE"
	}

	user := model.ReadUser(ctx)
	item := model.Ketchup{
		User: user,
	}

	scanner := func(row *sql.Row) error {
		err := row.Scan(&item.Version, &item.Repository.ID, &item.User.ID)
		if errors.Is(err, sql.ErrNoRows) {
			item = model.NoneKetchup
			return nil
		}

		return err
	}

	return item, db.Get(ctx, a.db, scanner, query, id, user.ID)
}

const insertQuery = `
INSERT INTO
  ketchup.ketchup
(
  version,
  repository_id,
  user_id
) VALUES (
  $1,
  $2,
  $3
) RETURNING 1
`

func (a app) Create(ctx context.Context, o model.Ketchup) (uint64, error) {
	return db.Create(ctx, insertQuery, o.Version, o.Repository.ID, model.ReadUser(ctx).ID)
}

const updateQuery = `
UPDATE
  ketchup.ketchup
SET
  version = $3
WHERE
  repository_id = $1
  AND user_id = $2
`

func (a app) Update(ctx context.Context, o model.Ketchup) error {
	return db.Exec(ctx, updateQuery, o.Repository.ID, model.ReadUser(ctx).ID, o.Version)
}

const deleteQuery = `
DELETE FROM
  ketchup.ketchup
WHERE
  repository_id = $1
  AND user_id = $2
`

func (a app) Delete(ctx context.Context, o model.Ketchup) error {
	return db.Exec(ctx, deleteQuery, o.Repository.ID, model.ReadUser(ctx).ID)
}
