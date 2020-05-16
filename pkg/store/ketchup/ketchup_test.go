package ketchup

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/ViBiOh/httputils/v3/pkg/db"
	"github.com/ViBiOh/ketchup/pkg/model"
	"github.com/lib/pq"
)

var (
	testCtx = model.StoreUser(context.Background(), model.User{ID: 3, Email: "nobody@localhost"})
)

func TestList(t *testing.T) {
	type args struct {
		page     uint
		pageSize uint
	}

	var cases = []struct {
		intention string
		args      args
		expectSQL string
		want      []model.Ketchup
		wantCount uint64
		wantErr   error
	}{
		{
			"simple",
			args{
				page:     1,
				pageSize: 20,
			},
			"SELECT k.version, k.repository_id, r.name, r.version, .+ AS full_count FROM ketchup k, repository r WHERE user_id = .+ AND repository_id = id",
			[]model.Ketchup{
				{
					Version: "0.9.0",
					Repository: model.Repository{
						ID:      1,
						Name:    "vibioh/ketchup",
						Version: "1.0.0",
					},
					User: model.User{
						ID:    3,
						Email: "nobody@localhost",
					},
				},
				{
					Version: "1.0.0",
					Repository: model.Repository{
						ID:      2,
						Name:    "vibioh/viws",
						Version: "1.2.3",
					},
					User: model.User{
						ID:    3,
						Email: "nobody@localhost",
					},
				},
			},
			2,
			nil,
		},
		{
			"timeout",
			args{
				page:     1,
				pageSize: 20,
			},
			"SELECT k.version, k.repository_id, r.name, r.version, .+ AS full_count FROM ketchup k, repository r WHERE user_id = .+ AND repository_id = id",
			nil,
			0,
			sqlmock.ErrCancelled,
		},
		{
			"invalid rows",
			args{
				page:     1,
				pageSize: 20,
			},
			"SELECT k.version, k.repository_id, r.name, r.version, .+ AS full_count FROM ketchup k, repository r WHERE user_id = .+ AND repository_id = id",
			nil,
			0,
			errors.New("converting driver.Value type string (\"a\") to a uint64: invalid syntax"),
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			mockDb, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("unable to create mock database: %s", err)
			}
			defer mockDb.Close()

			rows := sqlmock.NewRows([]string{"version", "repository_id", "name", "version", "full_count"})
			expectedQuery := mock.ExpectQuery(tc.expectSQL).WithArgs(20, 0, 3).WillReturnRows(rows)

			if tc.intention != "invalid rows" {
				rows.AddRow("0.9.0", 1, "vibioh/ketchup", "1.0.0", 2).AddRow("1.0.0", 2, "vibioh/viws", "1.2.3", 2)
			} else {
				rows.AddRow("0.9.0", "a", "vibioh/ketchup", "1.0.0", 2)
			}

			if tc.intention == "timeout" {
				savedSQLTimeout := db.SQLTimeout
				db.SQLTimeout = time.Second
				defer func() {
					db.SQLTimeout = savedSQLTimeout
				}()

				expectedQuery.WillDelayFor(db.SQLTimeout * 2)
			}

			got, gotCount, gotErr := New(mockDb).List(testCtx, tc.args.page, tc.args.pageSize)
			failed := false

			if tc.wantErr == nil && gotErr != nil {
				failed = true
			} else if tc.wantErr != nil && gotErr == nil {
				failed = true
			} else if tc.wantErr != nil && !strings.Contains(gotErr.Error(), tc.wantErr.Error()) {
				failed = true
			} else if !reflect.DeepEqual(got, tc.want) {
				failed = true
			} else if gotCount != tc.wantCount {
				failed = true
			}

			if failed {
				t.Errorf("List() = (%+v, %d, `%s`), want (%+v, %d, `%s`)", got, gotCount, gotErr, tc.want, tc.wantCount, tc.wantErr)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("sqlmock unfilled expectations: %s", err)
			}
		})
	}
}

func TestListByRepositoriesID(t *testing.T) {
	type args struct {
		ids []uint64
	}

	var cases = []struct {
		intention string
		args      args
		expectSQL string
		want      []model.Ketchup
		wantErr   error
	}{
		{
			"simple",
			args{
				ids: []uint64{1, 2},
			},
			"SELECT k.version, k.repository_id, k.user_id, u.email FROM ketchup k, \"user\" u WHERE repository_id = ANY .+ AND k.user_id = u.id",
			[]model.Ketchup{
				{
					Version: "0.9.0",
					Repository: model.Repository{
						ID: 1,
					},
					User: model.User{
						ID:    1,
						Email: "nobody@localhost",
					},
				},
				{
					Version: "1.0.0",
					Repository: model.Repository{
						ID: 2,
					},
					User: model.User{
						ID:    2,
						Email: "guest@domain",
					},
				},
			},
			nil,
		},
		{
			"timeout",
			args{
				ids: []uint64{1, 2},
			},
			"SELECT k.version, k.repository_id, k.user_id, u.email FROM ketchup k, \"user\" u WHERE repository_id = ANY .+ AND k.user_id = u.id",
			nil,
			sqlmock.ErrCancelled,
		},
		{
			"invalid rows",
			args{
				ids: []uint64{1, 2},
			},
			"SELECT k.version, k.repository_id, k.user_id, u.email FROM ketchup k, \"user\" u WHERE repository_id = ANY .+ AND k.user_id = u.id",
			nil,
			errors.New("converting driver.Value type string (\"a\") to a uint64: invalid syntax"),
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			mockDb, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("unable to create mock database: %s", err)
			}
			defer mockDb.Close()

			rows := sqlmock.NewRows([]string{"version", "repository_id", "user_id", "email"})
			expectedQuery := mock.ExpectQuery(tc.expectSQL).WithArgs(pq.Array(tc.args.ids)).WillReturnRows(rows)

			if tc.intention != "invalid rows" {
				rows.AddRow("0.9.0", 1, 1, "nobody@localhost").AddRow("1.0.0", 2, 2, "guest@domain")
			} else {
				rows.AddRow("0.9.0", "a", 1, "nobody@localhost")
			}

			if tc.intention == "timeout" {
				savedSQLTimeout := db.SQLTimeout
				db.SQLTimeout = time.Second
				defer func() {
					db.SQLTimeout = savedSQLTimeout
				}()

				expectedQuery.WillDelayFor(db.SQLTimeout * 2)
			}

			got, gotErr := New(mockDb).ListByRepositoriesID(testCtx, tc.args.ids)
			failed := false

			if tc.wantErr == nil && gotErr != nil {
				failed = true
			} else if tc.wantErr != nil && gotErr == nil {
				failed = true
			} else if tc.wantErr != nil && !strings.Contains(gotErr.Error(), tc.wantErr.Error()) {
				failed = true
			} else if !reflect.DeepEqual(got, tc.want) {
				failed = true
			}

			if failed {
				t.Errorf("ListByRepositoriesID() = (%+v, `%s`), want (%+v, `%s`)", got, gotErr, tc.want, tc.wantErr)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("sqlmock unfilled expectations: %s", err)
			}
		})
	}
}

func TestGetByRepositoryID(t *testing.T) {
	type args struct {
		id        uint64
		forUpdate bool
	}

	var cases = []struct {
		intention string
		args      args
		expectSQL string
		want      model.Ketchup
		wantErr   error
	}{
		{
			"simple",
			args{
				id: 1,
			},
			"SELECT version, repository_id, user_id FROM ketchup WHERE repository_id = .+ AND user_id = .+",
			model.Ketchup{
				Version: "0.9.0",
				Repository: model.Repository{
					ID: 1,
				},
				User: model.User{
					ID:    3,
					Email: "nobody@localhost",
				},
			},
			nil,
		},
		{
			"no rows",
			args{
				id: 1,
			},
			"SELECT version, repository_id, user_id FROM ketchup WHERE repository_id = .+ AND user_id = .+",
			model.NoneKetchup,
			nil,
		},
		{
			"forUpdate",
			args{
				id:        1,
				forUpdate: true,
			},
			"SELECT version, repository_id, user_id FROM ketchup WHERE repository_id = .+ AND user_id = .+ FOR UPDATE",
			model.Ketchup{
				Version: "0.9.0",
				Repository: model.Repository{
					ID: 1,
				},
				User: model.User{
					ID:    3,
					Email: "nobody@localhost",
				},
			},
			nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			mockDb, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("unable to create mock database: %s", err)
			}
			defer mockDb.Close()

			rows := sqlmock.NewRows([]string{"version", "repository_id", "user_id"})
			mock.ExpectQuery(tc.expectSQL).WithArgs(1, 3).WillReturnRows(rows)

			if tc.intention != "no rows" {
				rows.AddRow("0.9.0", 1, 3)
			}

			got, gotErr := New(mockDb).GetByRepositoryID(testCtx, tc.args.id, tc.args.forUpdate)

			failed := false

			if !errors.Is(gotErr, tc.wantErr) {
				failed = true
			} else if !reflect.DeepEqual(got, tc.want) {
				failed = true
			}

			if failed {
				t.Errorf("GetByRepositoryID() = (%+v, `%s`), want (%+v, `%s`)", got, gotErr, tc.want, tc.wantErr)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("sqlmock unfilled expectations: %s", err)
			}
		})
	}
}

func TestCreate(t *testing.T) {
	type args struct {
		o model.Ketchup
	}

	var cases = []struct {
		intention string
		args      args
		want      uint64
		wantErr   error
	}{
		{
			"simple",
			args{
				o: model.Ketchup{
					Version: "0.9.0",
					Repository: model.Repository{
						ID: 1,
					},
				},
			},
			1,
			nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			mockDb, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("unable to create mock database: %s", err)
			}
			defer mockDb.Close()

			mock.ExpectBegin()
			tx, err := mockDb.Begin()
			if err != nil {
				t.Errorf("unable to create tx: %s", err)
			}
			ctx := db.StoreTx(testCtx, tx)

			mock.ExpectQuery("INSERT INTO ketchup").WithArgs("0.9.0", 1, 3).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

			got, gotErr := New(mockDb).Create(ctx, tc.args.o)

			failed := false

			if !errors.Is(gotErr, tc.wantErr) {
				failed = true
			} else if got != tc.want {
				failed = true
			}

			if failed {
				t.Errorf("Create() = (%d, `%s`), want (%d, `%s`)", got, gotErr, tc.want, tc.wantErr)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("sqlmock unfilled expectations: %s", err)
			}
		})
	}
}

func TestUpdate(t *testing.T) {
	type args struct {
		o model.Ketchup
	}

	var cases = []struct {
		intention string
		args      args
		wantErr   error
	}{
		{
			"simple",
			args{
				o: model.Ketchup{
					Version: "0.9.0",
					Repository: model.Repository{
						ID: 1,
					},
				},
			},
			nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			mockDb, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("unable to create mock database: %s", err)
			}
			defer mockDb.Close()

			mock.ExpectBegin()
			tx, err := mockDb.Begin()
			if err != nil {
				t.Errorf("unable to create tx: %s", err)
			}
			ctx := db.StoreTx(testCtx, tx)

			mock.ExpectExec("UPDATE ketchup SET version").WithArgs(1, 3, "0.9.0").WillReturnResult(sqlmock.NewResult(0, 1))

			gotErr := New(mockDb).Update(ctx, tc.args.o)

			failed := false

			if !errors.Is(gotErr, tc.wantErr) {
				failed = true
			}

			if failed {
				t.Errorf("Update() = `%s`, want `%s`", gotErr, tc.wantErr)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("sqlmock unfilled expectations: %s", err)
			}
		})
	}
}

func TestDelete(t *testing.T) {
	type args struct {
		o model.Ketchup
	}

	var cases = []struct {
		intention string
		args      args
		wantErr   error
	}{
		{
			"simple",
			args{
				o: model.Ketchup{
					Repository: model.Repository{
						ID: 1,
					},
				},
			},
			nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			mockDb, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("unable to create mock database: %s", err)
			}
			defer mockDb.Close()

			mock.ExpectBegin()
			tx, err := mockDb.Begin()
			if err != nil {
				t.Errorf("unable to create tx: %s", err)
			}
			ctx := db.StoreTx(testCtx, tx)

			mock.ExpectExec("DELETE FROM ketchup").WithArgs(1, 3).WillReturnResult(sqlmock.NewResult(0, 1))

			gotErr := New(mockDb).Delete(ctx, tc.args.o)

			failed := false

			if !errors.Is(gotErr, tc.wantErr) {
				failed = true
			}

			if failed {
				t.Errorf("Delete() = `%s`, want `%s`", gotErr, tc.wantErr)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("sqlmock unfilled expectations: %s", err)
			}
		})
	}
}
