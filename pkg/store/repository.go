package store

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/ViBiOh/httputils/v3/pkg/db"
	"github.com/ViBiOh/ketchup/pkg/model"
)

func scanRepository(row db.RowScanner) (model.Repository, error) {
	var repository model.Repository

	if err := row.Scan(&repository.ID, &repository.Name, &repository.Version); err != nil {
		if err == sql.ErrNoRows {
			return model.NoneRepository, nil
		}

		return model.NoneRepository, err
	}

	return repository, nil
}

func scanRepositories(rows *sql.Rows) ([]model.Repository, uint, error) {
	var totalCount uint
	list := make([]model.Repository, 0)

	for rows.Next() {
		var repository model.Repository

		if err := rows.Scan(&repository.ID, &repository.Name, &repository.Version, &totalCount); err != nil {
			return nil, 0, err
		}

		list = append(list, repository)
	}

	return list, totalCount, nil
}

const listRepositoriesQuery = `
SELECT
  id,
  name,
  version,
  count(1) OVER() AS full_count
FROM
  repository
ORDER BY %s
LIMIT $1
OFFSET $2
`

func (a app) ListRepositories(ctx context.Context, page, pageSize uint, sortKey string, sortAsc bool) ([]model.Repository, uint, error) {
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

	rows, err := a.db.QueryContext(ctx, fmt.Sprintf(listRepositoriesQuery, order), pageSize, offset)
	if err != nil {
		return nil, 0, err
	}

	defer func() {
		err = db.RowsClose(rows, err)
	}()

	return scanRepositories(rows)
}

const getRepositoryByIDQuery = `
SELECT
  id,
  name,
  version
FROM
  repository
WHERE
  id = $1
`

func (a app) GetRepository(ctx context.Context, id uint64) (model.Repository, error) {
	return scanRepository(db.GetRow(ctx, a.db, getRepositoryByIDQuery, id))
}

const insertRepositoryQuery = `
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

func (a app) CreateRepository(ctx context.Context, o model.Repository) (uint64, error) {
	return db.Create(ctx, a.db, insertRepositoryQuery, o.Name, o.Version)
}

const updateRepositoryQuery = `
UPDATE
  repository
SET
  version = $2
WHERE
  id = $1
`

func (a app) UpdateRepository(ctx context.Context, o model.Repository) error {
	return db.Exec(ctx, a.db, updateRepositoryQuery, o.ID, o.Version)
}

const deleteRepositoryQuery = `
DELETE FROM
  repository
WHERE
  id = $1
`

func (a app) DeleteRepository(ctx context.Context, o model.Repository) error {
	return db.Exec(ctx, a.db, deleteRepositoryQuery, o.ID)
}
