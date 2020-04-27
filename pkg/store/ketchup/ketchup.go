package ketchup

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"

	authModel "github.com/ViBiOh/auth/v2/pkg/model"
	"github.com/ViBiOh/httputils/v3/pkg/db"
	"github.com/ViBiOh/ketchup/pkg/model"
	"github.com/ViBiOh/ketchup/pkg/store"
)

var (
	sortKeyMatcher                    = regexp.MustCompile(`[A-Za-z0-9]+`)
	_              store.KetchupStore = app{}
)

// App of package
type App interface {
	StartAtomic(ctx context.Context) (context.Context, error)
	EndAtomic(ctx context.Context, err error) error

	List(ctx context.Context, page, pageSize uint, sortKey string, sortAsc bool) ([]model.Ketchup, uint, error)
	Get(ctx context.Context, id uint64) (model.Ketchup, error)
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

func scanItem(row db.RowScanner) (model.Ketchup, error) {
	var item model.Ketchup

	if err := row.Scan(&item.ID, &item.Version, &item.Repository.ID); err != nil {
		if err == sql.ErrNoRows {
			return model.NoneKetchup, nil
		}

		return model.NoneKetchup, err
	}

	return item, nil
}

func scanItems(rows *sql.Rows) ([]model.Ketchup, uint, error) {
	var totalCount uint
	list := make([]model.Ketchup, 0)

	for rows.Next() {
		var item model.Ketchup

		if err := rows.Scan(&item.ID, &item.Version, &item.Repository.ID, &totalCount); err != nil {
			return nil, 0, err
		}

		list = append(list, item)
	}

	return list, totalCount, nil
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
  version,
  repository_id,
  count(1) OVER() AS full_count
FROM
  ketchup
WHERE
  user_id = $3
ORDER BY %s
LIMIT $1
OFFSET $2
`

func (a app) List(ctx context.Context, page, pageSize uint, sortKey string, sortAsc bool) ([]model.Ketchup, uint, error) {
	order := "creation_date DESC"

	if sortKeyMatcher.MatchString(sortKey) {
		order = sortKey

		if !sortAsc {
			order += " DESC"
		}
	}

	offset := (page - 1) * pageSize

	ctx, cancel := context.WithTimeout(ctx, db.SQLTimeout)
	defer cancel()

	rows, err := a.db.QueryContext(ctx, fmt.Sprintf(listQuery, order), pageSize, offset, authModel.ReadUser(ctx).ID)
	if err != nil {
		return nil, 0, err
	}

	defer func() {
		err = db.RowsClose(rows, err)
	}()

	return scanItems(rows)
}

const getQuery = `
SELECT
  id,
  version,
  repository_id
FROM
  ketchup
WHERE
  id = $1
`

func (a app) Get(ctx context.Context, id uint64) (model.Ketchup, error) {
	return scanItem(db.GetRow(ctx, a.db, getQuery, id))
}

const insertQuery = `
INSERT INTO
  ketchup
(
  version,
  repository_id,
  user_id
) VALUES (
  $1,
  $2,
  $3
) RETURNING id
`

func (a app) Create(ctx context.Context, o model.Ketchup) (uint64, error) {
	return db.Create(ctx, a.db, insertQuery, o.Version, o.Repository.ID, authModel.ReadUser(ctx).ID)
}

const updateQuery = `
UPDATE
  ketchup
SET
  version = $2
WHERE
  id = $1
`

func (a app) Update(ctx context.Context, o model.Ketchup) error {
	return db.Exec(ctx, a.db, updateQuery, o.ID, o.Version)
}

const deleteQuery = `
DELETE FROM
  ketchup
WHERE
  id = $1
`

func (a app) Delete(ctx context.Context, o model.Ketchup) error {
	return db.Exec(ctx, a.db, deleteQuery, o.ID)
}
