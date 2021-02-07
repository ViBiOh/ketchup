package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
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
	UpdateVersions(ctx context.Context, o model.Repository) error
	DeleteUnused(ctx context.Context) error
	DeleteUnusedVersions(ctx context.Context) error
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
		var rawRepositoryKind string
		item := model.NewRepository(0, 0, "")

		if err := rows.Scan(&item.ID, &item.Name, &rawRepositoryKind, &count); err != nil {
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

	err := db.List(ctx, a.db, scanner, query, args...)
	if err != nil {
		return list, 0, err
	}

	return list, count, a.enrichRepositoriesVersions(ctx, list)
}

func (a app) get(ctx context.Context, query string, args ...interface{}) (model.Repository, error) {
	var rawRepositoryKind string
	item := model.NewRepository(0, 0, "")

	scanner := func(row *sql.Row) error {
		err := row.Scan(&item.ID, &item.Name, &rawRepositoryKind)
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

	if err := db.Get(ctx, a.db, scanner, query, args...); err != nil {
		return model.NoneRepository, err
	}

	if item.ID == 0 {
		return item, nil
	}

	return item, a.enrichRepositoriesVersions(ctx, []model.Repository{item})
}

const listQuery = `
SELECT
  id,
  name,
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
  kind
) VALUES (
  $1,
  $2
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

	if item.ID != 0 {
		return 0, fmt.Errorf("%s repository already exists with name `%s`", o.Kind.String(), o.Name)
	}

	id, err := db.Create(ctx, insertQuery, strings.ToLower(o.Name), o.Kind.String())
	if err != nil {
		return 0, err
	}

	o.ID = id
	return id, a.UpdateVersions(ctx, o)
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

const deleteVersionsQuery = `
DELETE FROM
  ketchup.repository_version r
WHERE NOT EXISTS (
    SELECT
    FROM
      ketchup.ketchup k
    WHERE
      r.repository_id = k.repository_id
      AND r.pattern = k.pattern
  )
`

func (a app) DeleteUnusedVersions(ctx context.Context) error {
	return db.Exec(ctx, deleteVersionsQuery)
}
