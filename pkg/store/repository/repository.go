package repository

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/ViBiOh/ketchup/pkg/model"
	"github.com/jackc/pgx/v5"
)

type App struct {
	db model.Database
}

func New(db model.Database) App {
	return App{
		db: db,
	}
}

func (a App) DoAtomic(ctx context.Context, action func(context.Context) error) error {
	return a.db.DoAtomic(ctx, action)
}

func (a App) list(ctx context.Context, query string, args ...any) ([]model.Repository, uint64, error) {
	var count uint64
	var list []model.Repository

	scanner := func(rows pgx.Rows) error {
		var rawRepositoryKind string
		item := model.NewEmptyRepository()

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

func (a App) get(ctx context.Context, query string, args ...any) (model.Repository, error) {
	var rawRepositoryKind string
	item := model.NewEmptyRepository()

	scanner := func(row pgx.Row) error {
		err := row.Scan(&item.ID, &rawRepositoryKind, &item.Name, &item.Part)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil
			}

			return err
		}

		item.Kind, err = model.ParseRepositoryKind(rawRepositoryKind)

		return err
	}

	if err := a.db.Get(ctx, scanner, query, args...); err != nil {
		return item, err
	}

	if item.IsZero() {
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

func (a App) List(ctx context.Context, pageSize uint, last string) ([]model.Repository, uint64, error) {
	var query strings.Builder
	query.WriteString(listQuery)
	var queryArgs []any

	if len(last) != 0 {
		lastID, err := strconv.ParseUint(last, 10, 64)
		if err != nil {
			return nil, 0, fmt.Errorf("invalid last key: %w", err)
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

func (a App) ListByKinds(ctx context.Context, pageSize uint, last string, kinds ...model.RepositoryKind) ([]model.Repository, uint64, error) {
	var query strings.Builder
	query.WriteString(listByKindsQuery)
	var queryArgs []any

	kindsValue := make([]string, len(kinds))
	for i, kind := range kinds {
		kindsValue[i] = kind.String()
	}

	queryArgs = append(queryArgs, kindsValue)

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

func (a App) Suggest(ctx context.Context, ignoreIds []model.Identifier, count uint64) ([]model.Repository, error) {
	list, _, err := a.list(ctx, suggestQuery, count, ignoreIds)
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

func (a App) Get(ctx context.Context, id model.Identifier, forUpdate bool) (model.Repository, error) {
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

func (a App) GetByName(ctx context.Context, repositoryKind model.RepositoryKind, name, part string) (model.Repository, error) {
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

func (a App) Create(ctx context.Context, o model.Repository) (model.Identifier, error) {
	if err := a.db.Exec(ctx, insertLock); err != nil {
		return 0, err
	}

	item, err := a.GetByName(ctx, o.Kind, o.Name, o.Part)
	if err != nil {
		return 0, err
	}

	if !item.ID.IsZero() {
		return 0, fmt.Errorf("%s repository already exists with name=%s part=%s", o.Kind.String(), o.Name, o.Part)
	}

	id, err := a.db.Create(ctx, insertQuery, o.Kind.String(), strings.ToLower(o.Name), strings.ToLower(o.Part))
	if err != nil {
		return 0, err
	}

	o.ID = model.Identifier(id)
	return o.ID, a.UpdateVersions(ctx, o)
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

func (a App) DeleteUnused(ctx context.Context) error {
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

func (a App) DeleteUnusedVersions(ctx context.Context) error {
	return a.db.Exec(ctx, deleteVersionsQuery)
}
