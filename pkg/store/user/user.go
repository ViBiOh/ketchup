package user

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/ViBiOh/httputils/v3/pkg/db"
	"github.com/ViBiOh/ketchup/pkg/model"
	"github.com/ViBiOh/ketchup/pkg/store"
)

var (
	sortKeyMatcher = regexp.MustCompile(`[A-Za-z0-9]+`)
)

// App of package
type App interface {
	StartAtomic(ctx context.Context) (context.Context, error)
	EndAtomic(ctx context.Context, err error) error

	List(ctx context.Context, page, pageSize uint, sortKey string, sortAsc bool) ([]model.User, uint, error)
	Get(ctx context.Context, id uint64) (model.User, error)
	GetByEmail(ctx context.Context, email string) (model.User, error)
	GetByLoginID(ctx context.Context, loginID uint64) (model.User, error)
	Create(ctx context.Context, o model.User) (uint64, error)
	Update(ctx context.Context, o model.User) error
	Delete(ctx context.Context, o model.User) error
}

type app struct {
	db *sql.DB
}

// New creates new App from Config
func New(db *sql.DB) App {
	return app{
		db: db,
	}
}

func scanItems(rows *sql.Rows) ([]model.User, uint, error) {
	var totalCount uint
	list := make([]model.User, 0)

	for rows.Next() {
		var item model.User

		if err := rows.Scan(&item.ID, &item.Email, &item.Login.ID, &totalCount); err != nil {
			return nil, 0, err
		}

		list = append(list, item)
	}

	return list, totalCount, nil
}

func (a app) StartAtomic(ctx context.Context) (context.Context, error) {
	return store.StartAtomic(ctx, a.db)
}

func (a app) EndAtomic(ctx context.Context, err error) error {
	return store.EndAtomic(ctx, err)
}

const listQuery = `
SELECT
  id,
  email,
  login_id,
  count(1) OVER() AS full_count
FROM
  "user"
ORDER BY %s
LIMIT $1
OFFSET $2
`

func (a app) List(ctx context.Context, page, pageSize uint, sortKey string, sortAsc bool) ([]model.User, uint, error) {
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

	rows, err := a.db.QueryContext(ctx, fmt.Sprintf(listQuery, order), pageSize, offset)
	if err != nil {
		return nil, 0, err
	}

	defer func() {
		err = db.RowsClose(rows, err)
	}()

	return scanItems(rows)
}

const getQuery = `
SELECT
  id,
  email,
  login_id
FROM
  "user"
WHERE
  id = $1
`

func (a app) Get(ctx context.Context, id uint64) (model.User, error) {
	var item model.User
	scanner := func(row db.RowScanner) error {
		err := row.Scan(&item.ID, &item.Email, &item.Login.ID)
		if errors.Is(err, sql.ErrNoRows) {
			item = model.NoneUser
			return nil
		}

		return err
	}

	err := db.GetRow(ctx, a.db, scanner, getQuery, id)
	return item, err
}

const getByEmailQuery = `
SELECT
  id,
  email,
  login_id
FROM
  "user"
WHERE
  email = $1
`

func (a app) GetByEmail(ctx context.Context, email string) (model.User, error) {
	var item model.User
	scanner := func(row db.RowScanner) error {
		err := row.Scan(&item.ID, &item.Email, &item.Login.ID)
		if errors.Is(err, sql.ErrNoRows) {
			item = model.NoneUser
			return nil
		}

		return err
	}

	err := db.GetRow(ctx, a.db, scanner, getByEmailQuery, email)
	return item, err
}

const getByLoginIDQuery = `
SELECT
  id,
  email,
  login_id
FROM
  "user"
WHERE
  login_id = $1
`

func (a app) GetByLoginID(ctx context.Context, loginID uint64) (model.User, error) {
	var item model.User
	scanner := func(row db.RowScanner) error {
		err := row.Scan(&item.ID, &item.Email, &item.Login.ID)
		if errors.Is(err, sql.ErrNoRows) {
			item = model.NoneUser
			return nil
		}

		return err
	}

	err := db.GetRow(ctx, a.db, scanner, getByLoginIDQuery, loginID)
	return item, err
}

const insertQuery = `
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

func (a app) Create(ctx context.Context, o model.User) (uint64, error) {
	return db.Create(ctx, a.db, insertQuery, strings.ToLower(o.Email), o.Login.ID)
}

const updateQuery = `
UPDATE
  "user"
SET
  email = $2
WHERE
  id = $1
`

func (a app) Update(ctx context.Context, o model.User) error {
	return db.Exec(ctx, a.db, updateQuery, o.ID, o.Email)
}

const deleteQuery = `
DELETE FROM
  "user"
WHERE
  id = $1
`

func (a app) Delete(ctx context.Context, o model.User) error {
	return db.Exec(ctx, a.db, deleteQuery, o.ID)
}
