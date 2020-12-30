package repository

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/ViBiOh/httputils/v3/pkg/db"
	"github.com/ViBiOh/ketchup/pkg/model"
	"github.com/lib/pq"
)

// App of package
type App interface {
	DoAtomic(ctx context.Context, action func(context.Context) error) error

	List(ctx context.Context, page, pageSize uint) ([]model.Repository, uint64, error)
	Suggest(ctx context.Context, ignoreIds []uint64, count uint64) ([]model.Repository, error)
	Get(ctx context.Context, id uint64, forUpdate bool) (model.Repository, error)
	GetByName(ctx context.Context, name string, repositoryKind model.RepositoryKind) (model.Repository, error)
	Create(ctx context.Context, o model.Repository) (uint64, error)
	Update(ctx context.Context, o model.Repository) error
	DeleteUnused(ctx context.Context) error
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

func (a app) list(ctx context.Context, query string, args ...interface{}) ([]model.Repository, uint64, error) {
	var count uint64
	list := make([]model.Repository, 0)

	scanner := func(rows *sql.Rows) error {
		var item model.Repository
		var rawRepositoryKind string

		if err := rows.Scan(&item.ID, &item.Name, &item.Version, &rawRepositoryKind, &count); err != nil {
			return err
		}

		repositoryKind, err := model.ParseRepositoryKind(rawRepositoryKind)
		if err != nil {
			return err
		}
		item.Kind = repositoryKind

		list = append(list, item)
		return nil
	}

	return list, count, db.List(ctx, a.db, scanner, query, args...)
}

func (a app) get(ctx context.Context, query string, args ...interface{}) (model.Repository, error) {
	var item model.Repository
	var rawRepositoryKind string

	scanner := func(row *sql.Row) error {
		err := row.Scan(&item.ID, &item.Name, &item.Version, &rawRepositoryKind)
		if errors.Is(err, sql.ErrNoRows) {
			item = model.NoneRepository
			return nil
		}

		if err != nil {
			return err
		}

		item.Kind, err = model.ParseRepositoryKind(rawRepositoryKind)
		return err
	}

	return item, db.Get(ctx, a.db, scanner, query, args...)
}

const listQuery = `
SELECT
  id,
  name,
  version,
  kind,
  count(1) OVER() AS full_count
FROM
  ketchup.repository
LIMIT $1
OFFSET $2
`

func (a app) List(ctx context.Context, page, pageSize uint) ([]model.Repository, uint64, error) {
	return a.list(ctx, listQuery, pageSize, (page-1)*pageSize)
}

const suggestQuery = `
SELECT
  id,
  name,
  version,
  kind,
  (
    SELECT
      COUNT(1)
    FROM
      ketchup.ketchup
    WHERE
      repository_id = id
  ) AS count
FROM
  ketchup.repository
WHERE
  id != ALL($2)
ORDER BY
  count DESC
LIMIT $1
`

func (a app) Suggest(ctx context.Context, ignoreIds []uint64, count uint64) ([]model.Repository, error) {
	list, _, err := a.list(ctx, suggestQuery, count, pq.Array(ignoreIds))
	return list, err
}

const getQuery = `
SELECT
  id,
  name,
  version,
  kind
FROM
  ketchup.repository
WHERE
  id = $1
`

func (a app) Get(ctx context.Context, id uint64, forUpdate bool) (model.Repository, error) {
	query := getQuery
	if forUpdate {
		query += " FOR UPDATE"
	}

	return a.get(ctx, query, id)
}

const getByNameQuery = `
SELECT
  id,
  name,
  version,
  kind
FROM
  ketchup.repository
WHERE
  name = $1
  AND kind = $2
`

func (a app) GetByName(ctx context.Context, name string, repositoryKind model.RepositoryKind) (model.Repository, error) {
	return a.get(ctx, getByNameQuery, strings.ToLower(name), repositoryKind.String())
}

const insertLock = `
LOCK ketchup.repository IN SHARE ROW EXCLUSIVE MODE
`

const insertQuery = `
INSERT INTO
  ketchup.repository
(
  name,
  version,
  kind
) VALUES (
  $1,
  $2,
  $3
) RETURNING id
`

func (a app) Create(ctx context.Context, o model.Repository) (uint64, error) {
	if err := db.Exec(ctx, insertLock); err != nil {
		return 0, err
	}

	item, err := a.GetByName(ctx, o.Name, o.Kind)
	if err != nil {
		return 0, err
	}

	if item != model.NoneRepository {
		return item.ID, nil
	}

	return db.Create(ctx, insertQuery, strings.ToLower(o.Name), o.Version, o.Kind.String())
}

const updateRepositoryQuery = `
UPDATE
  ketchup.repository
SET
  version = $2
WHERE
  id = $1
`

func (a app) Update(ctx context.Context, o model.Repository) error {
	return db.Exec(ctx, updateRepositoryQuery, o.ID, o.Version)
}

const deleteQuery = `
DELETE FROM
  ketchup.repository
WHERE
  id NOT IN (
    SELECT
      DISTINCT repository_id
    FROM
      ketchup.ketchup
  )
`

func (a app) DeleteUnused(ctx context.Context) error {
	return db.Exec(ctx, deleteQuery)
}
