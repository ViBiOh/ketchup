package repository

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"strings"

	"github.com/ViBiOh/httputils/v3/pkg/db"
	"github.com/ViBiOh/ketchup/pkg/model"
)

var (
	sortKeyMatcher = regexp.MustCompile(`[A-Za-z0-9]+`)
)

// App of package
type App interface {
	DoAtomic(ctx context.Context, action func(context.Context) error) error

	List(ctx context.Context, page, pageSize uint) ([]model.Repository, uint64, error)
	Get(ctx context.Context, id uint64, forUpdate bool) (model.Repository, error)
	GetByName(ctx context.Context, name string) (model.Repository, error)
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

const listQuery = `
SELECT
  id,
  name,
  version,
  count(1) OVER() AS full_count
FROM
  repository
LIMIT $1
OFFSET $2
`

func (a app) List(ctx context.Context, page, pageSize uint) ([]model.Repository, uint64, error) {
	var totalCount uint64
	list := make([]model.Repository, 0)

	scanner := func(rows *sql.Rows) error {
		var item model.Repository
		if err := rows.Scan(&item.ID, &item.Name, &item.Version, &totalCount); err != nil {
			return err
		}

		list = append(list, item)
		return nil
	}

	err := db.List(ctx, a.db, scanner, listQuery, pageSize, (page-1)*pageSize)

	return list, totalCount, err
}

const getQuery = `
SELECT
  id,
  name,
  version
FROM
  repository
WHERE
  id = $1
`

func (a app) Get(ctx context.Context, id uint64, forUpdate bool) (model.Repository, error) {
	query := getQuery
	if forUpdate {
		query += " FOR UPDATE"
	}

	var item model.Repository
	scanner := func(row *sql.Row) error {
		err := row.Scan(&item.ID, &item.Name, &item.Version)
		if errors.Is(err, sql.ErrNoRows) {
			item = model.NoneRepository
			return nil
		}

		return err
	}

	err := db.Get(ctx, a.db, scanner, query, id)
	return item, err
}

const getByNameQuery = `
SELECT
  id,
  name,
  version
FROM
  repository
WHERE
  name = $1
`

func (a app) GetByName(ctx context.Context, name string) (model.Repository, error) {
	var item model.Repository
	scanner := func(row *sql.Row) error {
		err := row.Scan(&item.ID, &item.Name, &item.Version)
		if errors.Is(err, sql.ErrNoRows) {
			item = model.NoneRepository
			return nil
		}

		return err
	}

	err := db.Get(ctx, a.db, scanner, getByNameQuery, strings.ToLower(name))
	return item, err
}

const insertLock = `
LOCK repository IN SHARE ROW EXCLUSIVE MODE
`

const insertQuery = `
INSERT INTO
  repository
(
  name,
  version
) VALUES (
  $1,
  $2
) RETURNING id
`

func (a app) Create(ctx context.Context, o model.Repository) (uint64, error) {
	if err := db.Exec(ctx, insertLock); err != nil {
		return 0, err
	}

	item, err := a.GetByName(ctx, o.Name)
	if err != nil {
		return 0, err
	}

	if item != model.NoneRepository {
		return item.ID, nil
	}

	return db.Create(ctx, insertQuery, strings.ToLower(o.Name), o.Version)
}

const updateRepositoryQuery = `
UPDATE
  repository
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
  repository
WHERE
  id NOT IN (
    SELECT
      DISTINCT repository_id
    FROM
      ketchup
  )
`

func (a app) DeleteUnused(ctx context.Context) error {
	return db.Exec(ctx, deleteQuery)
}
