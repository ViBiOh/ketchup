package store

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/ViBiOh/httputils/v3/pkg/db"
	"github.com/ViBiOh/ketchup/pkg/model"
)

func scanUser(row db.RowScanner) (model.User, error) {
	var user model.User

	if err := row.Scan(&user.ID, &user.Email); err != nil {
		if err == sql.ErrNoRows {
			return model.NoneUser, nil
		}

		return model.NoneUser, err
	}

	return user, nil
}

func scanUsers(rows *sql.Rows) ([]model.User, uint, error) {
	var totalCount uint
	list := make([]model.User, 0)

	for rows.Next() {
		var user model.User

		if err := rows.Scan(&user.ID, &user.Email, &totalCount); err != nil {
			return nil, 0, err
		}

		list = append(list, user)
	}

	return list, totalCount, nil
}

const listUsersQuery = `
SELECT
  id,
  email,
  count(1) OVER() AS full_count
FROM
  "user"
ORDER BY %s
LIMIT $1
OFFSET $2
`

func (a app) ListUsers(ctx context.Context, page, pageSize uint, sortKey string, sortAsc bool) ([]model.User, uint, error) {
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

	rows, err := a.db.QueryContext(ctx, fmt.Sprintf(listUsersQuery, order), pageSize, offset)
	if err != nil {
		return nil, 0, err
	}

	defer func() {
		err = db.RowsClose(rows, err)
	}()

	return scanUsers(rows)
}

const getUserByIDQuery = `
SELECT
  id,
  email
FROM
  "user"
WHERE
  id = $1
`

func (a app) GetUser(ctx context.Context, id uint64) (model.User, error) {
	return scanUser(db.GetRow(ctx, a.db, getUserByIDQuery, id))
}

const getUserByEmailQuery = `
SELECT
  id,
  email
FROM
  "user"
WHERE
  email = $1
`

func (a app) GetUserByEmail(ctx context.Context, email string) (model.User, error) {
	return scanUser(db.GetRow(ctx, a.db, getUserByEmailQuery, email))
}

const insertUserQuery = `
INSERT INTO
  "user"
(
  email,
  login_id
) VALUES (
  $1,
  $2
) RETURNING id
`

func (a app) CreateUser(ctx context.Context, o model.User) (uint64, error) {
	return db.Create(ctx, a.db, insertUserQuery, o.Email, o.Login.ID)
}

const updateUserQuery = `
UPDATE
  "user"
SET
  email = $2
WHERE
  id = $1
`

func (a app) UpdateUser(ctx context.Context, o model.User) error {
	return db.Exec(ctx, a.db, updateUserQuery, o.ID, o.Email)
}

const deleteUserQuery = `
DELETE FROM
  "user"
WHERE
  id = $1
`

func (a app) DeleteUser(ctx context.Context, o model.User) error {
	return db.Exec(ctx, a.db, deleteUserQuery, o.ID)
}
