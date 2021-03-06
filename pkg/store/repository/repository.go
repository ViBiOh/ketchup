package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/ViBiOh/httputils/v4/pkg/db"
	"github.com/ViBiOh/ketchup/pkg/model"
	"github.com/lib/pq"
)

// App of package
type App interface {
	DoAtomic(ctx context.Context, action func(context.Context) error) error
	List(ctx context.Context, pageSize uint, last string) ([]model.Repository, uint64, error)
	ListByKinds(ctx context.Context, pageSize uint, last string, kinds ...model.RepositoryKind) ([]model.Repository, uint64, error)
	Suggest(ctx context.Context, ignoreIds []uint64, count uint64) ([]model.Repository, error)
	Get(ctx context.Context, id uint64, forUpdate bool) (model.Repository, error)
	GetByName(ctx context.Context, repositoryKind model.RepositoryKind, name, part string) (model.Repository, error)
	Create(ctx context.Context, o model.Repository) (uint64, error)
	UpdateVersions(ctx context.Context, o model.Repository) error
	DeleteUnused(ctx context.Context) error
	DeleteUnusedVersions(ctx context.Context) error
}

type app struct {
	db db.App
}

// New creates new App from Config
func New(db db.App) App {
	return app{
		db: db,
	}
}

func (a app) DoAtomic(ctx context.Context, action func(context.Context) error) error {
	return a.db.DoAtomic(ctx, action)
}

func (a app) list(ctx context.Context, query string, args ...interface{}) ([]model.Repository, uint64, error) {
	var count uint64
	list := make([]model.Repository, 0)

	scanner := func(rows *sql.Rows) error {
		var rawRepositoryKind string
		item := model.Repository{
			Versions: make(map[string]string),
		}

		if err := rows.Scan(&item.ID, &rawRepositoryKind, &item.Name, &item.Part, &count); err != nil {
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

	err := a.db.List(ctx, scanner, query, args...)
	if err != nil {
		return list, 0, err
	}

	return list, count, a.enrichRepositoriesVersions(ctx, list)
}

func (a app) get(ctx context.Context, query string, args ...interface{}) (model.Repository, error) {
	var rawRepositoryKind string
	item := model.Repository{
		Versions: make(map[string]string),
	}

	scanner := func(row *sql.Row) error {
		err := row.Scan(&item.ID, &rawRepositoryKind, &item.Name, &item.Part)
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

	if err := a.db.Get(ctx, scanner, query, args...); err != nil {
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
  kind,
  name,
  part,
  count(1) OVER() AS full_count
FROM
  ketchup.repository
WHERE
  TRUE
`

func (a app) List(ctx context.Context, pageSize uint, last string) ([]model.Repository, uint64, error) {
	var query strings.Builder
	query.WriteString(listQuery)
	var queryArgs []interface{}

	if len(last) != 0 {
		lastID, err := strconv.ParseUint(last, 10, 64)
		if err != nil {
			return nil, 0, fmt.Errorf("invalid last key: %s", err)
		}

		queryArgs = append(queryArgs, lastID)
		query.WriteString(fmt.Sprintf(" AND id > $%d", len(queryArgs)))
	}

	query.WriteString(" ORDER BY id ASC")

	queryArgs = append(queryArgs, pageSize)
	query.WriteString(fmt.Sprintf(" LIMIT $%d", len(queryArgs)))

	return a.list(ctx, query.String(), queryArgs...)
}

const listByKindsQuery = `
SELECT
  id,
  kind,
  name,
  part,
  count(1) OVER() AS full_count
FROM
  ketchup.repository
WHERE
  kind = ANY($1)
`

const listByKindRestartQuery = `
  AND (
    (
      name = $%d AND part > $%d
    ) OR (
      name > $%d
    )
  )
`

func (a app) ListByKinds(ctx context.Context, pageSize uint, last string, kinds ...model.RepositoryKind) ([]model.Repository, uint64, error) {
	var query strings.Builder
	query.WriteString(listByKindsQuery)
	var queryArgs []interface{}

	kindsValue := make([]string, len(kinds))
	for i, kind := range kinds {
		kindsValue[i] = kind.String()
	}

	queryArgs = append(queryArgs, pq.Array(kindsValue))

	if len(last) != 0 {
		parts := strings.Split(last, "|")
		if len(parts) != 2 {
			return nil, 0, errors.New("invalid last key format")
		}

		queryIndex := len(queryArgs)
		query.WriteString(fmt.Sprintf(listByKindRestartQuery, queryIndex+1, queryIndex+2, queryIndex+3))
		queryArgs = append(queryArgs, parts[0], parts[1], parts[0])
	}

	query.WriteString(" ORDER BY name ASC, part ASC")

	queryArgs = append(queryArgs, pageSize)
	query.WriteString(fmt.Sprintf(" LIMIT $%d", len(queryArgs)))

	list, count, err := a.list(ctx, query.String(), queryArgs...)

	return list, count, err
}

const suggestQuery = `
SELECT
  id,
  kind,
  name,
  part,
  (
    SELECT
      COUNT(1)
    FROM
      ketchup.ketchup
    WHERE
      repository_id = id
      AND pattern = 'stable'
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
  kind,
  name,
  part
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
  kind,
  name,
  part
FROM
  ketchup.repository
WHERE
  kind = $1
  AND name = $2
  AND part = $3
`

func (a app) GetByName(ctx context.Context, repositoryKind model.RepositoryKind, name, part string) (model.Repository, error) {
	return a.get(ctx, getByNameQuery, repositoryKind.String(), strings.ToLower(name), strings.ToLower(part))
}

const insertLock = `
LOCK ketchup.repository IN SHARE ROW EXCLUSIVE MODE
`

const insertQuery = `
INSERT INTO
  ketchup.repository
(
  kind,
  name,
  part
) VALUES (
  $1,
  $2,
  $3
) RETURNING id
`

func (a app) Create(ctx context.Context, o model.Repository) (uint64, error) {
	if err := a.db.Exec(ctx, insertLock); err != nil {
		return 0, err
	}

	item, err := a.GetByName(ctx, o.Kind, o.Name, o.Part)
	if err != nil {
		return 0, err
	}

	if item.ID != 0 {
		return 0, fmt.Errorf("%s repository already exists with name=%s part=%s", o.Kind.String(), o.Name, o.Part)
	}

	id, err := a.db.Create(ctx, insertQuery, o.Kind.String(), strings.ToLower(o.Name), strings.ToLower(o.Part))
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
	return a.db.Exec(ctx, deleteQuery)
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
	return a.db.Exec(ctx, deleteVersionsQuery)
}
