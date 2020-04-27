package store

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	authModel "github.com/ViBiOh/auth/v2/pkg/model"
	"github.com/ViBiOh/httputils/v3/pkg/db"
	"github.com/ViBiOh/ketchup/pkg/model"
)

func TestListUsers(t *testing.T) {
	type args struct {
		page     uint
		pageSize uint
		sortKey  string
		sortAsc  bool
	}

	var cases = []struct {
		intention string
		args      args
		expectSQL string
		want      []model.User
		wantCount uint
		wantErr   error
	}{
		{
			"simple",
			args{
				page:     1,
				pageSize: 20,
			},
			"SELECT id, email, .+ AS full_count FROM \"user\" ORDER BY creation_date DESC",
			[]model.User{
				{
					ID:    1,
					Email: "nobody@localhost",
				},
				{
					ID:    2,
					Email: "guest@internet",
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
			"SELECT id, email, .+ AS full_count FROM \"user\" ORDER BY creation_date DESC",
			nil,
			0,
			sqlmock.ErrCancelled,
		},
		{
			"order",
			args{
				page:     1,
				pageSize: 20,
				sortKey:  "email",
				sortAsc:  true,
			},
			"SELECT id, email, .+ AS full_count FROM \"user\" ORDER BY email",
			[]model.User{
				{
					ID:    1,
					Email: "nobody@localhost",
				},
				{
					ID:    2,
					Email: "guest@internet",
				},
			},
			2,
			nil,
		},
		{
			"descending",
			args{
				page:     1,
				pageSize: 20,
				sortKey:  "email",
				sortAsc:  false,
			},
			"SELECT id, email, .+ AS full_count FROM \"user\" ORDER BY email DESC",
			[]model.User{
				{
					ID:    1,
					Email: "nobody@localhost",
				},
				{
					ID:    2,
					Email: "guest@internet",
				},
			},
			2,
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

			expectedQuery := mock.ExpectQuery(tc.expectSQL).WithArgs(20, 0).WillReturnRows(sqlmock.NewRows([]string{"id", "email", "full_count"}).AddRow(1, "nobody@localhost", 2).AddRow(2, "guest@internet", 2))

			if tc.intention == "timeout" {
				savedSQLTimeout := db.SQLTimeout
				db.SQLTimeout = time.Second
				defer func() {
					db.SQLTimeout = savedSQLTimeout
				}()

				expectedQuery.WillDelayFor(db.SQLTimeout * 2)
			}

			got, gotCount, gotErr := New(mockDb).ListUsers(context.Background(), tc.args.page, tc.args.pageSize, tc.args.sortKey, tc.args.sortAsc)
			failed := false

			if tc.wantErr == nil && gotErr != nil {
				failed = true
			} else if tc.wantErr != nil && !errors.Is(gotErr, tc.wantErr) {
				failed = true
			} else if !reflect.DeepEqual(got, tc.want) {
				failed = true
			} else if gotCount != tc.wantCount {
				failed = true
			}

			if failed {
				t.Errorf("ListUsers() = (%+v, %d, `%s`), want (%+v, %d, `%s`)", got, gotCount, gotErr, tc.want, tc.wantCount, tc.wantErr)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("sqlmock unfilled expectations: %s", err)
			}
		})
	}
}

func TestGetUser(t *testing.T) {
	type args struct {
		id uint64
	}

	var cases = []struct {
		intention string
		args      args
		want      model.User
		wantErr   error
	}{
		{
			"create",
			args{
				id: 1,
			},
			model.User{
				ID:    1,
				Email: "nobody@localhost",
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

			mock.ExpectQuery("SELECT id, email FROM \"user\"").WithArgs(1).WillReturnRows(sqlmock.NewRows([]string{"id", "email"}).AddRow(1, "nobody@localhost"))

			got, gotErr := New(mockDb).GetUser(context.Background(), tc.args.id)

			failed := false

			if tc.wantErr == nil && gotErr != nil {
				failed = true
			} else if tc.wantErr != nil && !errors.Is(gotErr, tc.wantErr) {
				failed = true
			} else if !reflect.DeepEqual(got, tc.want) {
				failed = true
			}

			if failed {
				t.Errorf("GetUser() = (%+v, `%s`), want (%+v, `%s`)", got, gotErr, tc.want, tc.wantErr)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("sqlmock unfilled expectations: %s", err)
			}
		})
	}
}

func TestGetUserByEmail(t *testing.T) {
	type args struct {
		email string
	}

	var cases = []struct {
		intention string
		args      args
		want      model.User
		wantErr   error
	}{
		{
			"create",
			args{
				email: "nobody@localhost",
			},
			model.User{
				ID:    1,
				Email: "nobody@localhost",
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

			mock.ExpectQuery("SELECT id, email FROM \"user\"").WithArgs("nobody@localhost").WillReturnRows(sqlmock.NewRows([]string{"id", "email"}).AddRow(1, "nobody@localhost"))

			got, gotErr := New(mockDb).GetUserByEmail(context.Background(), tc.args.email)

			failed := false

			if tc.wantErr == nil && gotErr != nil {
				failed = true
			} else if tc.wantErr != nil && !errors.Is(gotErr, tc.wantErr) {
				failed = true
			} else if !reflect.DeepEqual(got, tc.want) {
				failed = true
			}

			if failed {
				t.Errorf("GetUserByEmail() = (%+v, `%s`), want (%+v, `%s`)", got, gotErr, tc.want, tc.wantErr)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("sqlmock unfilled expectations: %s", err)
			}
		})
	}
}

func TestCreateUser(t *testing.T) {
	type args struct {
		o model.User
	}

	var cases = []struct {
		intention string
		args      args
		want      uint64
		wantErr   error
	}{
		{
			"create",
			args{
				o: model.User{
					Email: "nobody@localhost",
					Login: authModel.User{
						ID:       1,
						Login:    "vibioh",
						Password: "secret",
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
			mock.ExpectQuery("INSERT INTO \"user\"").WithArgs("nobody@localhost", 1).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
			mock.ExpectCommit()

			got, gotErr := New(mockDb).CreateUser(context.Background(), tc.args.o)

			failed := false

			if tc.wantErr == nil && gotErr != nil {
				failed = true
			} else if tc.wantErr != nil && !errors.Is(gotErr, tc.wantErr) {
				failed = true
			} else if got != tc.want {
				failed = true
			}

			if failed {
				t.Errorf("CreateUser() = (%d, `%s`), want (%d, `%s`)", got, gotErr, tc.want, tc.wantErr)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("sqlmock unfilled expectations: %s", err)
			}
		})
	}
}

func TestUpdateUser(t *testing.T) {
	type args struct {
		o model.User
	}

	var cases = []struct {
		intention string
		args      args
		wantErr   error
	}{
		{
			"update",
			args{
				o: model.User{
					ID:    1,
					Email: "nobody@internet",
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
			mock.ExpectExec("UPDATE \"user\" SET email").WithArgs(1, "nobody@internet").WillReturnResult(sqlmock.NewResult(0, 1))
			mock.ExpectCommit()

			gotErr := New(mockDb).UpdateUser(context.Background(), tc.args.o)

			failed := false

			if tc.wantErr == nil && gotErr != nil {
				failed = true
			} else if tc.wantErr != nil && !errors.Is(gotErr, tc.wantErr) {
				failed = true
			}

			if failed {
				t.Errorf("UpdateUser() = `%s`, want `%s`", gotErr, tc.wantErr)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("sqlmock unfilled expectations: %s", err)
			}
		})
	}
}

func TestDeleteUser(t *testing.T) {
	type args struct {
		o model.User
	}

	var cases = []struct {
		intention string
		args      args
		wantErr   error
	}{
		{
			"update",
			args{
				o: model.User{
					ID: 1,
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
			mock.ExpectExec("DELETE FROM \"user\"").WithArgs(1).WillReturnResult(sqlmock.NewResult(0, 1))
			mock.ExpectCommit()

			gotErr := New(mockDb).DeleteUser(context.Background(), tc.args.o)

			failed := false

			if tc.wantErr == nil && gotErr != nil {
				failed = true
			} else if tc.wantErr != nil && !errors.Is(gotErr, tc.wantErr) {
				failed = true
			}

			if failed {
				t.Errorf("DeleteUser() = `%s`, want `%s`", gotErr, tc.wantErr)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("sqlmock unfilled expectations: %s", err)
			}
		})
	}
}
