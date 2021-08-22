package user

import (
	"context"
	"database/sql"
	"errors"
	"reflect"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	authModel "github.com/ViBiOh/auth/v2/pkg/model"
	"github.com/ViBiOh/httputils/v4/pkg/db"
	"github.com/ViBiOh/ketchup/pkg/model"
)

var (
	testEmail = "nobody@localhost"
)

func testWithMock(t *testing.T, action func(*sql.DB, sqlmock.Sqlmock)) {
	mockDb, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("unable to create mock database: %s", err)
	}
	defer mockDb.Close()

	action(mockDb, mock)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("sqlmock unfilled expectations: %s", err)
	}
}

func TestGetByEmail(t *testing.T) {
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
			"simple",
			args{
				email: testEmail,
			},
			model.NewUser(1, testEmail, authModel.NewUser(1, "")),
			nil,
		},
		{
			"no rows",
			args{
				email: testEmail,
			},
			model.User{},
			nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			testWithMock(t, func(mockDb *sql.DB, mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "email", "login_id"})
				mock.ExpectQuery("SELECT id, email, login_id FROM ketchup.user").WithArgs(testEmail).WillReturnRows(rows)

				if tc.intention != "no rows" {
					rows.AddRow(1, testEmail, 1)
				}

				got, gotErr := New(db.NewFromSQL(mockDb)).GetByEmail(context.Background(), tc.args.email)

				failed := false

				if !errors.Is(gotErr, tc.wantErr) {
					failed = true
				} else if !reflect.DeepEqual(got, tc.want) {
					failed = true
				}

				if failed {
					t.Errorf("GetByEmail() = (%+v, `%s`), want (%+v, `%s`)", got, gotErr, tc.want, tc.wantErr)
				}
			})
		})
	}
}

func TestGetByLoginID(t *testing.T) {
	type args struct {
		loginID uint64
	}

	var cases = []struct {
		intention string
		args      args
		want      model.User
		wantErr   error
	}{
		{
			"simple",
			args{
				loginID: 2,
			},
			model.NewUser(1, testEmail, authModel.NewUser(2, "")),
			nil,
		},
		{
			"no rows",
			args{
				loginID: 2,
			},
			model.User{},
			nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			testWithMock(t, func(mockDb *sql.DB, mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "email", "login_id"})
				mock.ExpectQuery("SELECT id, email, login_id FROM ketchup.user").WithArgs(2).WillReturnRows(rows)

				if tc.intention != "no rows" {
					rows.AddRow(1, testEmail, 2)
				}

				got, gotErr := New(db.NewFromSQL(mockDb)).GetByLoginID(context.Background(), tc.args.loginID)

				failed := false

				if !errors.Is(gotErr, tc.wantErr) {
					failed = true
				} else if !reflect.DeepEqual(got, tc.want) {
					failed = true
				}

				if failed {
					t.Errorf("GetByLoginID() = (%+v, `%s`), want (%+v, `%s`)", got, gotErr, tc.want, tc.wantErr)
				}
			})
		})
	}
}

func TestCreate(t *testing.T) {
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
			"simple",
			args{
				o: model.NewUser(0, testEmail, authModel.User{
					ID:       1,
					Login:    "vibioh",
					Password: "secret",
				}),
			},
			1,
			nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			testWithMock(t, func(mockDb *sql.DB, mock sqlmock.Sqlmock) {
				ctx := context.Background()

				mock.ExpectBegin()
				if tx, err := mockDb.Begin(); err != nil {
					t.Errorf("unable to create tx: %v", err)
				} else {
					ctx = db.StoreTx(ctx, tx)
				}

				mock.ExpectQuery("INSERT INTO ketchup.user").WithArgs(testEmail, 1).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

				got, gotErr := New(db.NewFromSQL(mockDb)).Create(ctx, tc.args.o)

				failed := false

				if !errors.Is(gotErr, tc.wantErr) {
					failed = true
				} else if got != tc.want {
					failed = true
				}

				if failed {
					t.Errorf("Create() = (%d, `%s`), want (%d, `%s`)", got, gotErr, tc.want, tc.wantErr)
				}
			})
		})
	}
}
