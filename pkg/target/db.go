package target

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/ViBiOh/httputils/v3/pkg/db"
	"github.com/ViBiOh/ketchup/pkg/model"
)

func scanTarget(row model.RowScanner) (Target, error) {
	var target Target

	err := row.Scan(&target.ID, &target.Owner, &target.Repository, &target.Version)
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

		if err := rows.Scan(&target.ID, &target.Owner, &target.Repository, &target.Version, &totalCount); err != nil {
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
  owner,
  repository,
  version,
  count(id) OVER() AS full_count
FROM
  target
ORDER BY $3
LIMIT $1
OFFSET $2
`

// FindTargetsByIds finds Targets by ids
func (a app) listTargets(page, pageSize uint, sortKey string, sortAsc bool) ([]Target, uint, error) {
	order := "creation_date DESC"

	if sortKey != "" {
		order = sortKey
	}
	if !sortAsc {
		order = fmt.Sprintf("%s DESC", order)
	}

	offset := (page - 1) * pageSize

	rows, err := a.db.Query(listQuery, pageSize, offset, order)
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
  owner,
  repository,
  version
FROM
  target
WHERE
  id = $1
`

func (a app) getTargetByID(id uint64) (Target, error) {
	return scanTarget(a.db.QueryRow(getByIDQuery, id))
}

const insertQuery = `
INSERT INTO
  target
(
  owner,
  repository,
  version
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
  version = $2
WHERE
  id = $1
`

func (a app) saveTarget(o *Target, tx *sql.Tx) (err error) {
	if o == nil {
		err = errors.New("cannot save nil Target")
	}

	var usedTx *sql.Tx
	if usedTx, err = db.GetTx(a.db, tx); err != nil {
		return
	}

	if usedTx != tx {
		defer func() {
			err = db.EndTx(usedTx, err)
		}()
	}

	var newID uint64

	if o.ID != 0 {
		_, err = usedTx.Exec(updateQuery, o.ID, o.Version)
	} else if insertErr := usedTx.QueryRow(insertQuery, o.Owner, o.Repository, o.Version).Scan(&newID); insertErr == nil {
		o.ID = newID
	} else {
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

	_, err = usedTx.Exec(deleteQuery, o.ID)
	return
}
