package repository

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"strings"

	"github.com/ViBiOh/httputils/v3/pkg/db"
	"github.com/ViBiOh/ketchup/pkg/model"
	"github.com/ViBiOh/ketchup/pkg/store"
)

var (
	sortKeyMatcher = regexp.MustCompile(`[A-Za-z0-9]+`)
)

// App of package
type App interface {
	StartAtomic(ctx context.Context) (context.Context, error)
	EndAtomic(ctx context.Context, err error) error

	List(ctx context.Context, page, pageSize uint) ([]model.Repository, uint, error)
	Get(ctx context.Context, id uint64) (model.Repository, error)
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

func (a app) StartAtomic(ctx context.Context) (context.Context, error) {
	return store.StartAtomic(ctx, a.db)
}

func (a app) EndAtomic(ctx context.Context, err error) error {
	return store.EndAtomic(ctx, err)
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

func (a app) List(ctx context.Context, page, pageSize uint) ([]model.Repository, uint, error) {
	ctx, cancel := context.WithTimeout(ctx, db.SQLTimeout)
	defer cancel()

	rows, err := a.db.QueryContext(ctx, listQuery, pageSize, (page-1)*pageSize)
	if err != nil {
		return nil, 0, err
	}

	defer func() {
		err = db.RowsClose(rows, err)
	}()

	var totalCount uint
	list := make([]model.Repository, 0)

	for rows.Next() {
		var item model.Repository

		if err := rows.Scan(&item.ID, &item.Name, &item.Version, &totalCount); err != nil {
			return nil, 0, err
		}

		list = append(list, item)
	}

	return list, totalCount, nil
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

func (a app) Get(ctx context.Context, id uint64) (model.Repository, error) {
	var item model.Repository
	scanner := func(row db.RowScanner) error {
		err := row.Scan(&item.ID, &item.Name, &item.Version)
		if errors.Is(err, sql.ErrNoRows) {
			item = model.NoneRepository
			return nil
		}

		return err
	}

	err := db.GetRow(ctx, a.db, scanner, getQuery, id)
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
	scanner := func(row db.RowScanner) error {
		err := row.Scan(&item.ID, &item.Name, &item.Version)
		if errors.Is(err, sql.ErrNoRows) {
			item = model.NoneRepository
			return nil
		}

		return err
	}

	err := db.GetRow(ctx, a.db, scanner, getByNameQuery, name)
	return item, err
}

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
	return db.Create(ctx, a.db, insertQuery, strings.ToLower(o.Name), o.Version)
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
	return db.Exec(ctx, a.db, updateRepositoryQuery, o.ID, o.Version)
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
	return db.Exec(ctx, a.db, deleteQuery)
}
