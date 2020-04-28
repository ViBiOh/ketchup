package ketchup

import (
	"context"
	"database/sql"

	"github.com/ViBiOh/httputils/v3/pkg/db"
	"github.com/ViBiOh/ketchup/pkg/model"
	"github.com/ViBiOh/ketchup/pkg/store"
	"github.com/lib/pq"
)

// App of package
type App interface {
	StartAtomic(ctx context.Context) (context.Context, error)
	EndAtomic(ctx context.Context, err error) error

	List(ctx context.Context, page, pageSize uint) ([]model.Ketchup, uint, error)
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

func (a app) StartAtomic(ctx context.Context) (context.Context, error) {
	return store.StartAtomic(ctx, a.db)
}

func (a app) EndAtomic(ctx context.Context, err error) error {
	return store.EndAtomic(ctx, err)
}

const listQuery = `
SELECT
  k.version,
  k.repository_id,
  r.name,
  r.version,
  count(1) OVER() AS full_count
FROM
  ketchup k,
  repository r
WHERE
  user_id = $3
  AND repository_id = id
ORDER BY r.name
LIMIT $1
OFFSET $2
`

func (a app) List(ctx context.Context, page, pageSize uint) ([]model.Ketchup, uint, error) {
	user := model.ReadUser(ctx)

	ctx, cancel := context.WithTimeout(ctx, db.SQLTimeout)
	defer cancel()

	rows, err := a.db.QueryContext(ctx, listQuery, pageSize, (page-1)*pageSize, user.ID)
	if err != nil {
		return nil, 0, err
	}

	defer func() {
		err = db.RowsClose(rows, err)
	}()

	var totalCount uint
	list := make([]model.Ketchup, 0)

	for rows.Next() {
		item := model.Ketchup{
			User: user,
		}

		if err := rows.Scan(&item.Version, &item.Repository.ID, &item.Repository.Name, &item.Repository.Version, &totalCount); err != nil {
			return nil, 0, err
		}

		list = append(list, item)
	}

	return list, totalCount, nil
}

const listByRepositoriesIDQuery = `
SELECT
  k.version,
  k.repository_id,
  k.user_id,
  u.email
FROM
  ketchup k,
  "user" u
WHERE
  repository_id = ANY($1)
  AND k.user_id = u.id
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
		var item model.Ketchup

		if err := rows.Scan(&item.Version, &item.Repository.ID, &item.User.ID, &item.User.Email); err != nil {
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

func (a app) GetByRepositoryID(ctx context.Context, id uint64, forUpdate bool) (model.Ketchup, error) {
	query := getQuery
	if forUpdate {
		query += " FOR UPDATE"
	}

	user := model.ReadUser(ctx)
	item := model.Ketchup{
		User: user,
	}

	if err := db.GetRow(ctx, a.db, query, id, user.ID).Scan(&item.Version, &item.Repository.ID, &item.User.ID); err != nil {
		if err == sql.ErrNoRows {
			return model.NoneKetchup, nil
		}

		return model.NoneKetchup, err
	}

	return item, nil
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
	return db.Create(ctx, a.db, insertQuery, o.Version, o.Repository.ID, model.ReadUser(ctx).ID)
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
	return db.Exec(ctx, a.db, updateQuery, o.Repository.ID, model.ReadUser(ctx).ID, o.Version)
}

const deleteQuery = `
DELETE FROM
  ketchup
WHERE
  repository_id = $1
  AND user_id = $2
`

func (a app) Delete(ctx context.Context, o model.Ketchup) error {
	return db.Exec(ctx, a.db, deleteQuery, o.Repository.ID, model.ReadUser(ctx).ID)
}
