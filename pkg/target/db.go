package target

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"time"

	"github.com/ViBiOh/httputils/v3/pkg/db"
	"github.com/ViBiOh/ketchup/pkg/model"
)

var (
	sortKeyMatcher = regexp.MustCompile(`[A-Za-z0-9]+`)
	sqlTimeout     = time.Second * 10
)

func scanTarget(row model.RowScanner) (Target, error) {
	var target Target

	err := row.Scan(&target.ID, &target.Repository, &target.CurrentVersion, &target.LatestVersion)
	if err != nil {
		if err == sql.ErrNoRows {
			return target, err
		}

		return target, err
	}

	return target, nil
}

func scanTargets(rows *sql.Rows) ([]Target, uint, error) {
	var totalCount uint
	list := make([]Target, 0)

	for rows.Next() {
		var target Target

		if err := rows.Scan(&target.ID, &target.Repository, &target.CurrentVersion, &target.LatestVersion, &totalCount); err != nil {
			if err == sql.ErrNoRows {
				return nil, 0, err
			}

			return nil, 0, err
		}

		list = append(list, target)
	}

	return list, totalCount, nil
}

const listQuery = `
SELECT
  id,
  repository,
  current_version,
  latest_version,
  count(id) OVER() AS full_count
FROM
  target
ORDER BY %s
LIMIT $1
OFFSET $2
`

// FindTargetsByIds finds Targets by ids
func (a app) listTargets(page, pageSize uint, sortKey string, sortAsc bool) ([]Target, uint, error) {
	order := "creation_date DESC"

	if sortKeyMatcher.MatchString(sortKey) {
		order = sortKey

		if !sortAsc {
			order += " DESC"
		}
	}

	offset := (page - 1) * pageSize

	ctx, cancel := context.WithTimeout(context.Background(), sqlTimeout)
	defer cancel()

	rows, err := a.db.QueryContext(ctx, fmt.Sprintf(listQuery, order), pageSize, offset)
	if err != nil {
		return nil, 0, err
	}

	defer func() {
		err = db.RowsClose(rows, err)
	}()

	return scanTargets(rows)
}

const getByIDQuery = `
SELECT
  id,
  repository,
  current_version,
  latest_version
FROM
  target
WHERE
  id = $1
`

func (a app) getTargetByID(id uint64) (Target, error) {
	ctx, cancel := context.WithTimeout(context.Background(), sqlTimeout)
	defer cancel()

	return scanTarget(a.db.QueryRowContext(ctx, getByIDQuery, id))
}

const getByRepositoryQuery = `
SELECT
  id,
  repository,
  current_version,
  latest_version
FROM
  target
WHERE
  repository = $1
`

func (a app) getTargetByRepository(repository string) (Target, error) {
	ctx, cancel := context.WithTimeout(context.Background(), sqlTimeout)
	defer cancel()

	return scanTarget(a.db.QueryRowContext(ctx, getByRepositoryQuery, repository))
}

const insertQuery = `
INSERT INTO
  target
(
  repository,
  current_version,
  latest_version
) VALUES (
  $1,
  $2,
  $3
)
RETURNING id
`

const updateQuery = `
UPDATE
  target
SET
  current_version = $2,
  latest_version = $3
WHERE
  id = $1
`

func (a app) saveTarget(o Target, tx *sql.Tx) (newID uint64, err error) {
	var usedTx *sql.Tx
	if usedTx, err = db.GetTx(a.db, tx); err != nil {
		return
	}

	if usedTx != tx {
		defer func() {
			err = db.EndTx(usedTx, err)
		}()
	}

	ctx, cancel := context.WithTimeout(context.Background(), sqlTimeout)
	defer cancel()

	if o.ID != 0 {
		_, err = usedTx.ExecContext(ctx, updateQuery, o.ID, o.CurrentVersion, o.LatestVersion)
	} else if insertErr := usedTx.QueryRowContext(ctx, insertQuery, o.Repository, o.CurrentVersion, o.LatestVersion).Scan(&newID); insertErr != nil {
		err = insertErr
	}

	return
}

const deleteQuery = `
DELETE FROM
  target
WHERE
  id = $1
`

func (a app) deleteTarget(o Target, tx *sql.Tx) (err error) {
	var usedTx *sql.Tx
	if usedTx, err = db.GetTx(a.db, tx); err != nil {
		return
	}

	if usedTx != tx {
		defer func() {
			err = db.EndTx(usedTx, err)
		}()
	}

	ctx, cancel := context.WithTimeout(context.Background(), sqlTimeout)
	defer cancel()

	_, err = usedTx.ExecContext(ctx, deleteQuery, o.ID)
	return
}
