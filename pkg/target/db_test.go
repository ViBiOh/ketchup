package target

import (
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestGetTargetByID(t *testing.T) {
	type args struct {
		id uint64
	}

	var cases = []struct {
		intention string
		args      args
		want      Target
		wantErr   error
	}{
		{
			"simple",
			args{
				id: 8000,
			},
			Target{
				ID:             8000,
				Repository:     "vibioh/ketchup",
				CurrentVersion: "1.0.0",
				LatestVersion:  "1.0.1",
			},
			nil,
		},
		{
			"timeout",
			args{
				id: 8000,
			},
			Target{},
			sqlmock.ErrCancelled,
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("unable to create mock database: %s", err)
			}
			defer db.Close()

			expectedQuery := mock.ExpectQuery("SELECT id, repository, current_version, latest_version FROM target WHERE id = (.+)").WillReturnRows(sqlmock.NewRows([]string{"id", "repository", "current_version", "latest_version"}).AddRow(8000, "vibioh/ketchup", "1.0.0", "1.0.1"))

			if tc.intention == "timeout" {
				savedSQLTimeout := sqlTimeout
				sqlTimeout = time.Second
				defer func() {
					sqlTimeout = savedSQLTimeout
				}()

				expectedQuery.WillDelayFor(sqlTimeout * 2)
			}

			got, gotErr := app{db: db}.getTargetByID(tc.args.id)
			failed := false

			if tc.wantErr != nil && !errors.Is(gotErr, tc.wantErr) {
				failed = true
			} else if got != tc.want {
				failed = true
			}

			if failed {
				t.Errorf("getTargetByID() = (%v, `%s`), want (%v, `%s`)", got, gotErr, tc.want, tc.wantErr)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("sqlmock unfilled expectations: %s", err)
			}
		})
	}
}

func TestGetTargetByRepository(t *testing.T) {
	type args struct {
		repository string
	}

	var cases = []struct {
		intention string
		args      args
		want      Target
		wantErr   error
	}{
		{
			"simple",
			args{
				repository: "vibioh/ketchup",
			},
			Target{
				ID:             8000,
				Repository:     "vibioh/ketchup",
				CurrentVersion: "1.0.0",
				LatestVersion:  "1.0.1",
			},
			nil,
		},
		{
			"timeout",
			args{
				repository: "vibioh/ketchup",
			},
			Target{},
			sqlmock.ErrCancelled,
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("unable to create mock database: %s", err)
			}
			defer db.Close()

			expectedQuery := mock.ExpectQuery("SELECT id, repository, current_version, latest_version FROM target WHERE repository = (.+)").WillReturnRows(sqlmock.NewRows([]string{"id", "repository", "current_version", "latest_version"}).AddRow(8000, "vibioh/ketchup", "1.0.0", "1.0.1"))

			if tc.intention == "timeout" {
				savedSQLTimeout := sqlTimeout
				sqlTimeout = time.Second
				defer func() {
					sqlTimeout = savedSQLTimeout
				}()

				expectedQuery.WillDelayFor(sqlTimeout * 2)
			}

			got, gotErr := app{db: db}.getTargetByRepository(tc.args.repository)
			failed := false

			if tc.wantErr != nil && !errors.Is(gotErr, tc.wantErr) {
				failed = true
			} else if got != tc.want {
				failed = true
			}

			if failed {
				t.Errorf("getTargetByRepository() = (%v, `%s`), want (%v, `%s`)", got, gotErr, tc.want, tc.wantErr)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("sqlmock unfilled expectations: %s", err)
			}
		})
	}
}

func TestDeleteTarget(t *testing.T) {
	type args struct {
		o  Target
		tx bool
	}

	var cases = []struct {
		intention string
		args      args
		wantErr   error
	}{
		{
			"simple",
			args{
				o: Target{
					ID:             8000,
					Repository:     "vibioh/ketchup",
					CurrentVersion: "1.0.0",
					LatestVersion:  "1.0.1",
				},
				tx: false,
			},
			nil,
		},
		{
			"timeout",
			args{
				o: Target{
					ID:             8000,
					Repository:     "vibioh/ketchup",
					CurrentVersion: "1.0.0",
					LatestVersion:  "1.0.1",
				},
				tx: false,
			},
			sqlmock.ErrCancelled,
		},
		{
			"with tx",
			args{
				o: Target{
					ID:             8000,
					Repository:     "vibioh/ketchup",
					CurrentVersion: "1.0.0",
					LatestVersion:  "1.0.1",
				},
				tx: true,
			},
			nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("unable to create mock database: %s", err)
			}
			defer db.Close()

			var tx *sql.Tx
			if tc.args.tx {
				mock.ExpectBegin()
				dbTx, err := db.Begin()

				if err != nil {
					t.Errorf("unable to create tx: %v", err)
				}
				tx = dbTx
			}

			if !tc.args.tx {
				mock.ExpectBegin()
			}

			expectedQuery := mock.ExpectExec("DELETE FROM target WHERE id = (.+)").WillReturnResult(sqlmock.NewResult(0, 1))

			if !tc.args.tx {
				if tc.wantErr != nil {
					mock.ExpectRollback()
				} else {
					mock.ExpectCommit()
				}
			}

			if tc.intention == "timeout" {
				savedSQLTimeout := sqlTimeout
				sqlTimeout = time.Second
				defer func() {
					sqlTimeout = savedSQLTimeout
				}()

				expectedQuery.WillDelayFor(sqlTimeout * 2)
			}

			gotErr := app{db: db}.deleteTarget(tc.args.o, tx)

			failed := false

			if tc.wantErr != nil && !errors.Is(gotErr, tc.wantErr) {
				failed = true
			}

			if failed {
				t.Errorf("deleteTarget() = `%s`, want `%s`", gotErr, tc.wantErr)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("sqlmock unfilled expectations: %s", err)
			}
		})
	}
}
