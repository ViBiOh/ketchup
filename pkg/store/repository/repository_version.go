package repository

import (
	"context"
	"database/sql"
	"fmt"
	"sort"

	"github.com/ViBiOh/httputils/v3/pkg/db"
	"github.com/ViBiOh/ketchup/pkg/model"
	"github.com/lib/pq"
)

const listRepositoryVersionsForIDsQuery = `
SELECT
  repository_id,
  pattern,
  version
FROM
  ketchup.repository_version
WHERE
  repository_id = ANY($1)
ORDER BY
  repository_id ASC,
  pattern ASC
`

func (a app) enrichRepositoriesVersions(ctx context.Context, repositories []model.Repository) error {
	if len(repositories) == 0 {
		return nil
	}

	ids := make([]uint64, len(repositories))
	for index, repository := range repositories {
		ids[index] = repository.ID
	}

	sort.Sort(model.RepositoryByID(repositories))

	index := 0
	scanner := func(rows *sql.Rows) error {
		var repositoryID uint64
		var pattern, version string

		if err := rows.Scan(&repositoryID, &pattern, &version); err != nil {
			return err
		}

		for ; repositoryID != repositories[index].ID; index++ {
		}

		repositories[index].Versions[pattern] = version

		return nil
	}

	return db.List(ctx, a.db, scanner, listRepositoryVersionsForIDsQuery, pq.Array(ids))
}

const listRepositoryVersionQuery = `
SELECT
  pattern,
  version
FROM
  ketchup.repository_version
WHERE
  repository_id = $1
`

const createRepositoryVersionQuery = `
INSERT INTO
  ketchup.repository_version
(
  repository_id,
  pattern,
  version
) VALUES (
  $1,
  $2,
  $3
)
`

const updateRepositoryVersionQuery = `
UPDATE
  ketchup.repository_version
SET
  version = $3
WHERE
  repository_id = $1
  AND pattern = $2
`

const deleteRepositoryVersionQuery = `
DELETE FROM
  ketchup.repository_version
WHERE
  repository_id = $1
  AND pattern = $2
`

func (a app) UpdateVersions(ctx context.Context, o model.Repository) error {
	patterns := make(map[string]bool)

	scanner := func(rows *sql.Rows) error {
		var pattern string
		var version string

		if err := rows.Scan(&pattern, &version); err != nil {
			return err
		}

		patterns[pattern] = true

		if repositoryVersion, ok := o.Versions[pattern]; ok {
			if repositoryVersion != version {
				if err := db.Exec(ctx, updateRepositoryVersionQuery, o.ID, pattern, version); err != nil {
					return fmt.Errorf("unable to update repository version: %w", err)
				}
			}
		} else if err := db.Exec(ctx, deleteRepositoryVersionQuery, o.ID, pattern); err != nil {
			return fmt.Errorf("unable to delete repository version: %w", err)
		}

		return nil
	}

	err := db.List(ctx, a.db, scanner, listRepositoryVersionQuery, o.ID)
	if err != nil {
		return err
	}

	for pattern, version := range o.Versions {
		if !patterns[pattern] {
			if err := db.Exec(ctx, createRepositoryVersionQuery, o.ID, pattern, version); err != nil {
				return fmt.Errorf("unable to create repository version: %w", err)
			}
		}
	}

	return nil
}
