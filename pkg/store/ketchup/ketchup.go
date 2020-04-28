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
	"github.com/lib/pq"
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
	ListByRepositoriesID(ctx context.Context, ids []uint64) ([]model.Ketchup, error)
	GetByRepositoryID(ctx context.Context, id uint64) (model.Ketchup, error)
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

	if err := row.Scan(&item.Version, &item.Repository.ID, &item.User.ID); err != nil {
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

		if err := rows.Scan(&item.Version, &item.Repository.ID, &item.User.ID, &totalCount); err != nil {
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
  version,
  repository_id,
  user_id,
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

const listByRepositoriesIDQuery = `
SELECT
  version,
  repository_id,
  user_id
FROM
  ketchup
WHERE
  repository_id = ANY($1)
`

func (a app) ListByRepositoriesID(ctx context.Context, ids []uint64) ([]model.Ketchup, error) {
	ctx, cancel := context.WithTimeout(ctx, db.SQLTimeout)
	defer cancel()

	rows, err := a.db.QueryContext(ctx, listByRepositoriesIDQuery, pq.Array(ids))
	if err != nil {
		return nil, err
	}

	defer func() {
		err = db.RowsClose(rows, err)
	}()

	list := make([]model.Ketchup, 0)

	for rows.Next() {
		item, err := scanItem(rows)
		if err != nil {
			return nil, err
		}

		list = append(list, item)
	}

	return list, nil
}

const getQuery = `
SELECT
  version,
  repository_id,
  user_id
FROM
  ketchup
WHERE
  repository_id = $1
  AND user_id = $2
`

func (a app) GetByRepositoryID(ctx context.Context, id uint64) (model.Ketchup, error) {
	return scanItem(db.GetRow(ctx, a.db, getQuery, id, authModel.ReadUser(ctx).ID))
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
) RETURNING 1
`

func (a app) Create(ctx context.Context, o model.Ketchup) (uint64, error) {
	return db.Create(ctx, a.db, insertQuery, o.Version, o.Repository.ID, authModel.ReadUser(ctx).ID)
}

const updateQuery = `
UPDATE
  ketchup
SET
  version = $3
WHERE
  repository_id = $1
  AND user_id = $2
`

func (a app) Update(ctx context.Context, o model.Ketchup) error {
	return db.Exec(ctx, a.db, updateQuery, o.Repository.ID, authModel.ReadUser(ctx).ID, o.Version)
}

const deleteQuery = `
DELETE FROM
  ketchup
WHERE
  repository_id = $1
  AND user_id = $2
`

func (a app) Delete(ctx context.Context, o model.Ketchup) error {
	return db.Exec(ctx, a.db, deleteQuery, o.Repository.ID, authModel.ReadUser(ctx).ID)
}
