package ketchup

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/ViBiOh/httputils/v4/pkg/db"
	"github.com/ViBiOh/ketchup/pkg/model"
	"github.com/ViBiOh/ketchup/pkg/store"
	"github.com/lib/pq"
)

// App of package
type App interface {
	DoAtomic(ctx context.Context, action func(context.Context) error) error
	List(ctx context.Context, page, pageSize uint) ([]model.Ketchup, uint64, error)
	ListByRepositoriesID(ctx context.Context, ids []uint64, frequency model.KetchupFrequency) ([]model.Ketchup, error)
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
  k.pattern,
  k.version,
  k.frequency,
  k.repository_id,
  r.name,
  r.part,
  r.kind,
  rv.version,
  count(1) OVER() AS full_count
FROM
  ketchup.ketchup k,
  ketchup.repository r,
  ketchup.repository_version rv
WHERE
  user_id = $1
  AND k.repository_id = r.id
  AND rv.repository_id = r.id
  AND rv.pattern = k.pattern
`

func (a app) List(ctx context.Context, page, pageSize uint) ([]model.Ketchup, uint64, error) {
	user := model.ReadUser(ctx)

	var totalCount uint64
	list := make([]model.Ketchup, 0)

	scanner := func(rows *sql.Rows) error {
		item := model.NewKetchup("", "", model.Daily, model.NewRepository(0, 0, "", ""))
		item.User = user
		var rawRepositoryKind string
		var rawKetchupFrequency string
		var repositoryVersion string

		if err := rows.Scan(&item.Pattern, &item.Version, &rawKetchupFrequency, &item.Repository.ID, &item.Repository.Name, &item.Repository.Part, &rawRepositoryKind, &repositoryVersion, &totalCount); err != nil {
			return err
		}

		item.Repository.AddVersion(item.Pattern, repositoryVersion)

		ketchupFrequency, err := model.ParseKetchupFrequency(rawKetchupFrequency)
		if err != nil {
			return err
		}
		item.Frequency = ketchupFrequency

		repositoryKind, err := model.ParseRepositoryKind(rawRepositoryKind)
		if err != nil {
			return err
		}
		item.Repository.Kind = repositoryKind

		list = append(list, item)
		return nil
	}

	var query strings.Builder
	query.WriteString(listQuery)
	queryArgs := []interface{}{
		user.ID,
	}

	queryArgs = append(queryArgs, store.AddPagination(&query, len(queryArgs), page, pageSize)...)

	return list, totalCount, db.List(ctx, a.db, scanner, query.String(), queryArgs...)
}

const listByRepositoriesIDQuery = `
SELECT
  k.pattern,
  k.version,
  k.frequency,
  k.repository_id,
  k.user_id,
  u.email
FROM
  ketchup.ketchup k,
  ketchup.user u
WHERE
  repository_id = ANY ($1)
  AND k.user_id = u.id
  AND k.frequency = $2
`

func (a app) ListByRepositoriesID(ctx context.Context, ids []uint64, frequency model.KetchupFrequency) ([]model.Ketchup, error) {
	list := make([]model.Ketchup, 0)

	scanner := func(rows *sql.Rows) error {
		var item model.Ketchup
		item.Repository = model.NewRepository(0, 0, "", "")
		var rawKetchupFrequency string

		if err := rows.Scan(&item.Pattern, &item.Version, &rawKetchupFrequency, &item.Repository.ID, &item.User.ID, &item.User.Email); err != nil {
			return err
		}

		ketchupFrequency, err := model.ParseKetchupFrequency(rawKetchupFrequency)
		if err != nil {
			return err
		}
		item.Frequency = ketchupFrequency

		list = append(list, item)
		return nil
	}

	return list, db.List(ctx, a.db, scanner, listByRepositoriesIDQuery, pq.Array(ids), strings.ToLower(frequency.String()))
}

const getQuery = `
SELECT
  k.pattern,
  k.version,
  k.frequency,
  k.repository_id,
  k.user_id,
  r.name,
  r.part,
  r.kind
FROM
  ketchup.ketchup k,
  ketchup.repository r
WHERE
  k.repository_id = $1
  AND k.user_id = $2
  AND k.repository_id = r.id
`

func (a app) GetByRepositoryID(ctx context.Context, id uint64, forUpdate bool) (model.Ketchup, error) {
	query := getQuery
	if forUpdate {
		query += " FOR UPDATE"
	}

	user := model.ReadUser(ctx)
	item := model.Ketchup{
		User:       user,
		Repository: model.NewGithubRepository(0, ""),
	}

	scanner := func(row *sql.Row) error {
		var rawRepositoryKind string
		var rawKetchupFrequency string

		err := row.Scan(&item.Pattern, &item.Version, &rawKetchupFrequency, &item.Repository.ID, &item.User.ID, &item.Repository.Name, &item.Repository.Part, &rawRepositoryKind)
		if errors.Is(err, sql.ErrNoRows) {
			item = model.NoneKetchup
			return nil
		}

		ketchupFrequency, err := model.ParseKetchupFrequency(rawKetchupFrequency)
		if err != nil {
			return err
		}
		item.Frequency = ketchupFrequency

		repositoryKind, err := model.ParseRepositoryKind(rawRepositoryKind)
		if err != nil {
			return err
		}
		item.Repository.Kind = repositoryKind

		return err
	}

	return item, db.Get(ctx, a.db, scanner, query, id, user.ID)
}

const insertQuery = `
INSERT INTO
  ketchup.ketchup
(
  pattern,
  version,
  frequency,
  repository_id,
  user_id
) VALUES (
  $1,
  $2,
  $3,
  $4,
  $5
) RETURNING 1
`

func (a app) Create(ctx context.Context, o model.Ketchup) (uint64, error) {
	return db.Create(ctx, insertQuery, o.Pattern, o.Version, strings.ToLower(o.Frequency.String()), o.Repository.ID, model.ReadUser(ctx).ID)
}

const updateQuery = `
UPDATE
  ketchup.ketchup
SET
  pattern = $3,
  version = $4,
  frequency = $5
WHERE
  repository_id = $1
  AND user_id = $2
`

func (a app) Update(ctx context.Context, o model.Ketchup) error {
	return db.Exec(ctx, updateQuery, o.Repository.ID, model.ReadUser(ctx).ID, o.Pattern, o.Version, strings.ToLower(o.Frequency.String()))
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
