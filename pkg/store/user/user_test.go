package user

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	authModel "github.com/ViBiOh/auth/v2/pkg/model"
	"github.com/ViBiOh/httputils/v3/pkg/db"
	"github.com/ViBiOh/ketchup/pkg/model"
)

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
				email: "nobody@localhost",
			},
			model.User{
				ID:    1,
				Email: "nobody@localhost",
				Login: authModel.NewUser(1, ""),
			},
			nil,
		},
		{
			"no rows",
			args{
				email: "nobody@localhost",
			},
			model.NoneUser,
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

			rows := sqlmock.NewRows([]string{"id", "email", "login_id"})
			mock.ExpectQuery("SELECT id, email, login_id FROM \"user\"").WithArgs("nobody@localhost").WillReturnRows(rows)

			if tc.intention != "no rows" {
				rows.AddRow(1, "nobody@localhost", 1)
			}

			got, gotErr := New(mockDb).GetByEmail(context.Background(), tc.args.email)

			failed := false

			if !errors.Is(gotErr, tc.wantErr) {
				failed = true
			} else if !reflect.DeepEqual(got, tc.want) {
				failed = true
			}

			if failed {
				t.Errorf("GetByEmail() = (%+v, `%s`), want (%+v, `%s`)", got, gotErr, tc.want, tc.wantErr)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("sqlmock unfilled expectations: %s", err)
			}
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
			model.User{
				ID:    1,
				Email: "nobody@localhost",
				Login: authModel.NewUser(2, ""),
			},
			nil,
		},
		{
			"no rows",
			args{
				loginID: 2,
			},
			model.NoneUser,
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

			rows := sqlmock.NewRows([]string{"id", "email", "login_id"})
			mock.ExpectQuery("SELECT id, email, login_id FROM \"user\"").WithArgs(2).WillReturnRows(rows)

			if tc.intention != "no rows" {
				rows.AddRow(1, "nobody@localhost", 2)
			}

			got, gotErr := New(mockDb).GetByLoginID(context.Background(), tc.args.loginID)

			failed := false

			if !errors.Is(gotErr, tc.wantErr) {
				failed = true
			} else if !reflect.DeepEqual(got, tc.want) {
				failed = true
			}

			if failed {
				t.Errorf("GetByLoginID() = (%+v, `%s`), want (%+v, `%s`)", got, gotErr, tc.want, tc.wantErr)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("sqlmock unfilled expectations: %s", err)
			}
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

			ctx := context.Background()

			mock.ExpectBegin()
			if tx, err := mockDb.Begin(); err != nil {
				t.Errorf("unable to create tx: %v", err)
			} else {
				ctx = db.StoreTx(ctx, tx)
			}

			mock.ExpectQuery("INSERT INTO \"user\"").WithArgs("nobody@localhost", 1).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

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
