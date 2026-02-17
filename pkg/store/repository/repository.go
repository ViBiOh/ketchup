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

type Service struct {
	db model.Database
}

func New(db model.Database) Service {
	return Service{
		db: db,
	}
}

func (s Service) DoAtomic(ctx context.Context, action func(context.Context) error) error {
	return s.db.DoAtomic(ctx, action)
}

func (s Service) list(ctx context.Context, query string, args ...any) ([]model.Repository, error) {
	var list []model.Repository

	scanner := func(rows pgx.Rows) error {
		var rawRepositoryKind string
		item := model.NewEmptyRepository()

		if err := rows.Scan(&item.ID, &rawRepositoryKind, &item.Name, &item.Part); err != nil {
			return err
		}

		repositoryKind, err := model.ParseRepositoryKind(rawRepositoryKind)
		if err != nil {
			return fmt.Errorf("parse kind `%s`: %w", rawRepositoryKind, err)
		}
		item.Kind = repositoryKind

		list = append(list, item)
		return nil
	}

	err := s.db.List(ctx, scanner, query, args...)
	if err != nil {
		return list, err
	}

	return list, s.enrichRepositoriesVersions(ctx, list)
}

func (s Service) get(ctx context.Context, query string, args ...any) (model.Repository, error) {
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

	if err := s.db.Get(ctx, scanner, query, args...); err != nil {
		return item, err
	}

	if item.IsZero() {
		return item, nil
	}

	return item, s.enrichRepositoriesVersions(ctx, []model.Repository{item})
}

const listQuery = `
SELECT
  id,
  kind,
  name,
  part
FROM
  ketchup.repository
WHERE
  TRUE
`

func (s Service) List(ctx context.Context, pageSize uint, last string) ([]model.Repository, error) {
	var query strings.Builder
	query.WriteString(listQuery)
	var queryArgs []any

	if len(last) != 0 {
		lastID, err := strconv.ParseUint(last, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid last key: %w", err)
		}

		queryArgs = append(queryArgs, lastID)
		query.WriteString(fmt.Sprintf(" AND id > $%d", len(queryArgs)))
	}

	query.WriteString(" ORDER BY id ASC")

	queryArgs = append(queryArgs, pageSize)
	query.WriteString(fmt.Sprintf(" LIMIT $%d", len(queryArgs)))

	return s.list(ctx, query.String(), queryArgs...)
}

const listByKindsQuery = `
SELECT
  id,
  kind,
  name,
  part
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

func (s Service) ListByKinds(ctx context.Context, pageSize uint, last string, kinds ...model.RepositoryKind) ([]model.Repository, error) {
	var query strings.Builder
	query.WriteString(listByKindsQuery)
	var queryArgs []any

	kindsValue := make([]string, len(kinds))
	for i, kind := range kinds {
		kindsValue[i] = strings.ToLower(kind.String())
	}

	queryArgs = append(queryArgs, kindsValue)

	if len(last) != 0 {
		parts := strings.Split(last, "|")
		if len(parts) != 2 {
			return nil, errors.New("invalid last key format")
		}

		queryIndex := len(queryArgs)
		query.WriteString(fmt.Sprintf(listByKindRestartQuery, queryIndex+1, queryIndex+2, queryIndex+3))
		queryArgs = append(queryArgs, parts[0], parts[1], parts[0])
	}

	query.WriteString(" ORDER BY name ASC, part ASC")

	queryArgs = append(queryArgs, pageSize)
	query.WriteString(fmt.Sprintf(" LIMIT $%d", len(queryArgs)))

	list, err := s.list(ctx, query.String(), queryArgs...)

	return list, err
}

const suggestQuery = `
WITH ranked_repositories AS (
  SELECT
    r.id,
    r.kind,
    r.name,
    r.part,
    COUNT(k.user_id) as stable_count,
    ROW_NUMBER() OVER (PARTITION BY r.kind ORDER BY COUNT(k.user_id) DESC) as rank
  FROM
    ketchup.repository r
  JOIN
    ketchup.ketchup k ON r.id = k.repository_id
  WHERE
    k.pattern = 'stable'
    AND k.repository_id != ALL($2)
  GROUP BY
    r.id, r.kind, r.name, r.part
)
SELECT
  id,
  kind,
  name,
  part
FROM
  ranked_repositories
WHERE
  rank = 1
LIMIT $1
`

func (s Service) Suggest(ctx context.Context, ignoreIds []model.Identifier, count uint64) ([]model.Repository, error) {
	return s.list(ctx, suggestQuery, count, ignoreIds)
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

func (s Service) Get(ctx context.Context, id model.Identifier, forUpdate bool) (model.Repository, error) {
	query := getQuery
	if forUpdate {
		query += " FOR UPDATE"
	}

	return s.get(ctx, query, id)
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

func (s Service) GetByName(ctx context.Context, repositoryKind model.RepositoryKind, name, part string) (model.Repository, error) {
	return s.get(ctx, getByNameQuery, strings.ToLower(repositoryKind.String()), strings.ToLower(name), strings.ToLower(part))
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

func (s Service) Create(ctx context.Context, o model.Repository) (model.Identifier, error) {
	if err := s.db.Exec(ctx, insertLock); err != nil {
		return 0, err
	}

	item, err := s.GetByName(ctx, o.Kind, o.Name, o.Part)
	if err != nil {
		return 0, err
	}

	if !item.ID.IsZero() {
		return 0, fmt.Errorf("%s repository already exists with name=%s part=%s", o.Kind.String(), o.Name, o.Part)
	}

	id, err := s.db.Create(ctx, insertQuery, strings.ToLower(o.Kind.String()), strings.ToLower(o.Name), strings.ToLower(o.Part))
	if err != nil {
		return 0, err
	}

	o.ID = model.Identifier(id)
	return o.ID, s.UpdateVersions(ctx, o)
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

func (s Service) DeleteUnused(ctx context.Context) error {
	return s.db.Exec(ctx, deleteQuery)
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

func (s Service) DeleteUnusedVersions(ctx context.Context) error {
	return s.db.Exec(ctx, deleteVersionsQuery)
}
