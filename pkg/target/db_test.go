package target

import (
	"database/sql"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestListTargets(t *testing.T) {
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
		want      []Target
		wantCount uint
		wantErr   error
	}{
		{
			"simple",
			args{
				page:     1,
				pageSize: 20,
			},
			"SELECT id, repository, current_version, latest_version, count\\(id\\) OVER\\(\\) AS full_count FROM target ORDER BY creation_date DESC",
			[]Target{
				{
					ID:             8000,
					Repository:     "vibioh/ketchup",
					CurrentVersion: "1.0.0",
					LatestVersion:  "1.0.1",
				},
				{
					ID:             8001,
					Repository:     "vibioh/viws",
					CurrentVersion: "2.0.0",
					LatestVersion:  "2.0.1",
				},
			},
			2,
			nil,
		},
		{
			"order",
			args{
				page:     1,
				pageSize: 20,
				sortKey:  "repository",
				sortAsc:  true,
			},
			"SELECT id, repository, current_version, latest_version, count\\(id\\) OVER\\(\\) AS full_count FROM target ORDER BY repository",
			[]Target{
				{
					ID:             8000,
					Repository:     "vibioh/ketchup",
					CurrentVersion: "1.0.0",
					LatestVersion:  "1.0.1",
				},
				{
					ID:             8001,
					Repository:     "vibioh/viws",
					CurrentVersion: "2.0.0",
					LatestVersion:  "2.0.1",
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
				sortKey:  "repository",
				sortAsc:  false,
			},
			"SELECT id, repository, current_version, latest_version, count\\(id\\) OVER\\(\\) AS full_count FROM target ORDER BY repository DESC",
			[]Target{
				{
					ID:             8000,
					Repository:     "vibioh/ketchup",
					CurrentVersion: "1.0.0",
					LatestVersion:  "1.0.1",
				},
				{
					ID:             8001,
					Repository:     "vibioh/viws",
					CurrentVersion: "2.0.0",
					LatestVersion:  "2.0.1",
				},
			},
			2,
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

			expectedQuery := mock.ExpectQuery(tc.expectSQL).WithArgs(20, 0).WillReturnRows(sqlmock.NewRows([]string{"id", "repository", "current_version", "latest_version", "full_count"}).AddRow(8000, "vibioh/ketchup", "1.0.0", "1.0.1", 2).AddRow(8001, "vibioh/viws", "2.0.0", "2.0.1", 2))

			if tc.intention == "timeout" {
				savedSQLTimeout := sqlTimeout
				sqlTimeout = time.Second
				defer func() {
					sqlTimeout = savedSQLTimeout
				}()

				expectedQuery.WillDelayFor(sqlTimeout * 2)
			}

			got, gotCount, gotErr := app{db: db}.listTargets(tc.args.page, tc.args.pageSize, tc.args.sortKey, tc.args.sortAsc)
			failed := false

			if tc.wantErr != nil && !errors.Is(gotErr, tc.wantErr) {
				failed = true
			} else if !reflect.DeepEqual(got, tc.want) {
				failed = true
			} else if gotCount != tc.wantCount {
				failed = true
			}

			if failed {
				t.Errorf("listTargets() = (%v, %d, `%s`), want (%v, %d, `%s`)", got, gotCount, gotErr, tc.want, tc.wantCount, tc.wantErr)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("sqlmock unfilled expectations: %s", err)
			}
		})
	}
}

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

func TestSaveTarget(t *testing.T) {
	type args struct {
		o  Target
		tx bool
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
				o: Target{
					Repository:     "vibioh/ketchup",
					CurrentVersion: "1.0.0",
					LatestVersion:  "1.0.1",
				},
				tx: false,
			},
			8000,
			nil,
		},
		{
			"timeout",
			args{
				o: Target{
					Repository:     "vibioh/ketchup",
					CurrentVersion: "1.0.0",
					LatestVersion:  "1.0.1",
				},
				tx: false,
			},
			0,
			sqlmock.ErrCancelled,
		},
		{
			"with tx",
			args{
				o: Target{
					Repository:     "vibioh/ketchup",
					CurrentVersion: "1.0.0",
					LatestVersion:  "1.0.1",
				},
				tx: true,
			},
			8000,
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

			expectedQuery := mock.ExpectQuery("INSERT INTO target").WithArgs("vibioh/ketchup", "1.0.0", "1.0.1").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(8000))

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

			got, gotErr := app{db: db}.saveTarget(tc.args.o, tx)

			failed := false

			if tc.wantErr != nil && !errors.Is(gotErr, tc.wantErr) {
				failed = true
			} else if got != tc.want {
				failed = true
			}

			if failed {
				t.Errorf("saveTarget() = (%d, `%s`), want (%d, `%s`)", got, gotErr, tc.want, tc.wantErr)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("sqlmock unfilled expectations: %s", err)
			}
		})
	}
}

func TestUpdateTarget(t *testing.T) {
	type args struct {
		o  Target
		tx bool
	}

	var cases = []struct {
		intention string
		args      args
		want      uint64
		wantErr   error
	}{
		{
			"update",
			args{
				o: Target{
					ID:             8000,
					Repository:     "vibioh/ketchup",
					CurrentVersion: "1.0.0",
					LatestVersion:  "1.0.1",
				},
				tx: false,
			},
			0,
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

			expectedQuery := mock.ExpectExec("UPDATE target SET").WithArgs(8000, "1.0.0", "1.0.1").WillReturnResult(sqlmock.NewResult(0, 1))

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

			got, gotErr := app{db: db}.saveTarget(tc.args.o, tx)

			failed := false

			if tc.wantErr != nil && !errors.Is(gotErr, tc.wantErr) {
				failed = true
			} else if got != tc.want {
				failed = true
			}

			if failed {
				t.Errorf("saveTarget() = (%d, `%s`), want (%d, `%s`)", got, gotErr, tc.want, tc.wantErr)
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
