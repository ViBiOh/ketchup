package ketchup

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/ViBiOh/ketchup/pkg/model"
	"github.com/jackc/pgx/v4"
)

// App of package
type App struct {
	db model.Database
}

// New creates new App from Config
func New(db model.Database) App {
	return App{
		db: db,
	}
}

// DoAtomic does an atomic operation
func (a App) DoAtomic(ctx context.Context, action func(context.Context) error) error {
	return a.db.DoAtomic(ctx, action)
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

const listQueryRestart = `
  AND (
    (
      k.repository_id = $%d
      AND k.pattern > $%d
    ) OR (
      k.repository_id > $%d
    )
  )
`

// List ketchups
func (a App) List(ctx context.Context, pageSize uint, last string) ([]model.Ketchup, uint64, error) {
	user := model.ReadUser(ctx)

	var totalCount uint64
	list := make([]model.Ketchup, 0)

	scanner := func(rows pgx.Rows) error {
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

		list = append(list, item.WithID())
		return nil
	}

	var query strings.Builder
	query.WriteString(listQuery)
	queryArgs := []interface{}{
		user.ID,
	}

	if len(last) != 0 {
		parts := strings.Split(last, "|")
		if len(parts) != 2 {
			return nil, 0, fmt.Errorf("invalid last key format: %s", last)
		}

		lastID, err := strconv.ParseUint(parts[0], 10, 64)
		if err != nil {
			return nil, 0, fmt.Errorf("invalid last key id: %d", err)
		}

		value := len(queryArgs)
		query.WriteString(fmt.Sprintf(listQueryRestart, value+1, value+2, value+3))
		queryArgs = append(queryArgs, lastID, parts[1], lastID)
	}

	query.WriteString(" ORDER BY k.repository_id ASC, k.pattern ASC")

	queryArgs = append(queryArgs, pageSize)
	query.WriteString(fmt.Sprintf(" LIMIT $%d", len(queryArgs)))

	return list, totalCount, a.db.List(ctx, scanner, query.String(), queryArgs...)
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

// ListByRepositoriesID lists ketchup by repositories id
func (a App) ListByRepositoriesID(ctx context.Context, ids []uint64, frequency model.KetchupFrequency) ([]model.Ketchup, error) {
	list := make([]model.Ketchup, 0)

	scanner := func(rows pgx.Rows) error {
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

	return list, a.db.List(ctx, scanner, listByRepositoriesIDQuery, ids, strings.ToLower(frequency.String()))
}

const listOutdateByFrequencyQuery = `
SELECT
  rv.pattern,
  rv.version,
  k.frequency,
  r.id,
  r.name,
  r.part,
  r.kind,
  k.user_id,
  u.email
FROM
  ketchup.ketchup AS k
INNER JOIN
  ketchup.repository r ON r.id = k.repository_id
INNER JOIN
  ketchup.repository_version AS rv ON rv.repository_id = r.id
INNER JOIN
  ketchup.user AS u ON u.id = k.user_id
WHERE
  k.version <> rv.version
  AND k.frequency = $1
`

// ListOutdatedByFrequency lists outdated ketchup by frequency id
func (a App) ListOutdatedByFrequency(ctx context.Context, frequency model.KetchupFrequency) ([]model.Ketchup, error) {
	list := make([]model.Ketchup, 0)

	scanner := func(rows pgx.Rows) error {
		var item model.Ketchup
		item.Repository = model.NewRepository(0, 0, "", "")
		var rawKetchupFrequency, rawRepositoryKind string

		if err := rows.Scan(&item.Pattern, &item.Version, &rawKetchupFrequency, &item.Repository.ID, &item.Repository.Name, &item.Repository.Part, &rawRepositoryKind, &item.User.ID, &item.User.Email); err != nil {
			return err
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

		list = append(list, item)
		return nil
	}

	return list, a.db.List(ctx, scanner, listOutdateByFrequencyQuery, strings.ToLower(frequency.String()))
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
  AND k.pattern = $3
  AND k.repository_id = r.id
`

// GetByRepository retrieves ketchup for a repository and patern
func (a App) GetByRepository(ctx context.Context, id uint64, pattern string, forUpdate bool) (model.Ketchup, error) {
	query := getQuery
	if forUpdate {
		query += " FOR UPDATE"
	}

	user := model.ReadUser(ctx)
	item := model.Ketchup{
		User:       user,
		Repository: model.NewGithubRepository(0, ""),
	}

	scanner := func(row pgx.Row) error {
		var rawRepositoryKind, rawKetchupFrequency string

		err := row.Scan(&item.Pattern, &item.Version, &rawKetchupFrequency, &item.Repository.ID, &item.User.ID, &item.Repository.Name, &item.Repository.Part, &rawRepositoryKind)
		if errors.Is(err, pgx.ErrNoRows) {
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

	return item, a.db.Get(ctx, scanner, query, id, user.ID, pattern)
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

// Create a ketchup
func (a App) Create(ctx context.Context, o model.Ketchup) (uint64, error) {
	return a.db.Create(ctx, insertQuery, o.Pattern, o.Version, strings.ToLower(o.Frequency.String()), o.Repository.ID, model.ReadUser(ctx).ID)
}

const updateQuery = `
UPDATE
  ketchup.ketchup
SET
  pattern = $4,
  version = $5,
  frequency = $6
WHERE
  repository_id = $1
  AND user_id = $2
  AND pattern = $3
`

// Update a ketchup
func (a App) Update(ctx context.Context, o model.Ketchup, oldPattern string) error {
	return a.db.Exec(ctx, updateQuery, o.Repository.ID, model.ReadUser(ctx).ID, oldPattern, o.Pattern, o.Version, strings.ToLower(o.Frequency.String()))
}

const updateAllQuery = `
UPDATE
  ketchup.ketchup AS k
SET
  version = rv.version
FROM
  ketchup.repository AS r,
  ketchup.repository_version AS rv
WHERE
  k.repository_id = r.id
  AND k.repository_id = rv.repository_id
  AND k.pattern = rv.pattern
  AND k.version <> rv.version
  AND k.user_id = $1
`

// UpdateAll ketchups
func (a App) UpdateAll(ctx context.Context) error {
	return a.db.Exec(ctx, updateAllQuery, model.ReadUser(ctx).ID)
}

const deleteQuery = `
DELETE FROM
  ketchup.ketchup
WHERE
  repository_id = $1
  AND user_id = $2
  AND pattern = $3
`

// Delete a ketchup
func (a App) Delete(ctx context.Context, o model.Ketchup) error {
	return a.db.Exec(ctx, deleteQuery, o.Repository.ID, model.ReadUser(ctx).ID, o.Pattern)
}
