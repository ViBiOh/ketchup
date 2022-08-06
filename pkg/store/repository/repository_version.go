package repository

import (
	"context"
	"fmt"

	"github.com/ViBiOh/ketchup/pkg/model"
	"github.com/jackc/pgx/v4"
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

func (a App) enrichRepositoriesVersions(ctx context.Context, repositories []model.Repository) error {
	if len(repositories) == 0 {
		return nil
	}

	ids := make([]model.Identifier, len(repositories))
	for index, repository := range repositories {
		ids[index] = repository.ID
	}

	var repository model.Repository

	scanner := func(rows pgx.Rows) error {
		var repositoryID model.Identifier
		var pattern, version string

		if err := rows.Scan(&repositoryID, &pattern, &version); err != nil {
			return err
		}

		if repository.ID != repositoryID {
			repository = findRepository(repositories, repositoryID)
		}

		repository.Versions[pattern] = version

		return nil
	}

	return a.db.List(ctx, scanner, listRepositoryVersionsForIDsQuery, ids)
}

func findRepository(repositories []model.Repository, id model.Identifier) model.Repository {
	for _, repo := range repositories {
		if repo.ID == id {
			return repo
		}
	}

	return model.NewEmptyRepository()
}

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

// UpdateVersions of a repository
func (a App) UpdateVersions(ctx context.Context, o model.Repository) error {
	patterns, err := a.getRepositoryVersions(ctx, o)
	if err != nil {
		return fmt.Errorf("unable to fetch repository versions: %w", err)
	}

	for pattern, version := range patterns {
		repositoryVersion, ok := o.Versions[pattern]
		if !ok {
			if err := a.db.One(ctx, deleteRepositoryVersionQuery, o.ID, pattern); err != nil {
				return fmt.Errorf("unable to delete repository version: %w", err)
			}
			continue
		}

		if repositoryVersion == version {
			continue
		}

		if err := a.db.One(ctx, updateRepositoryVersionQuery, o.ID, pattern, repositoryVersion); err != nil {
			return fmt.Errorf("unable to update repository version: %w", err)
		}
	}

	for pattern, version := range o.Versions {
		if _, ok := patterns[pattern]; ok {
			continue
		}

		if err := a.db.One(ctx, createRepositoryVersionQuery, o.ID, pattern, version); err != nil {
			return fmt.Errorf("unable to create repository version: %w", err)
		}
	}

	return nil
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

func (a App) getRepositoryVersions(ctx context.Context, o model.Repository) (map[string]string, error) {
	patterns := make(map[string]string)

	scanner := func(rows pgx.Rows) error {
		var pattern string
		var version string

		if err := rows.Scan(&pattern, &version); err != nil {
			return err
		}

		patterns[pattern] = version
		return nil
	}

	err := a.db.List(ctx, scanner, listRepositoryVersionQuery, o.ID)
	if err != nil {
		return nil, fmt.Errorf("unable to list repository version: %w", err)
	}

	return patterns, err
}
