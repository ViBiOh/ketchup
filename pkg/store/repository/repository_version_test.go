package repository

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/ViBiOh/httputils/v4/pkg/db"
	"github.com/ViBiOh/ketchup/pkg/model"
)

var (
	errFailed = errors.New("failed")
)

func TestEnrichRepositoriesVersions(t *testing.T) {
	type args struct {
		repositories []model.Repository
	}

	var cases = []struct {
		intention string
		args      args
		want      []model.Repository
		wantErr   error
	}{
		{
			"empty",
			args{
				repositories: nil,
			},
			nil,
			nil,
		},
		{
			"invalid rows",
			args{
				repositories: []model.Repository{
					model.NewRepository(1, model.Github, ""),
				},
			},
			[]model.Repository{
				model.NewRepository(1, model.Github, ""),
			},
			errors.New("type string (\"a\") to a uint64"),
		},
		{
			"invalid rows",
			args{
				repositories: []model.Repository{
					model.NewRepository(1, model.Github, ""),
				},
			},
			[]model.Repository{
				model.NewRepository(1, model.Github, ""),
			},
			errors.New("type string (\"a\") to a uint64"),
		},
		{
			"sequential",
			args{
				repositories: []model.Repository{
					model.NewRepository(2, model.Github, ""),
					model.NewRepository(1, model.Github, ""),
				},
			},
			[]model.Repository{
				model.NewRepository(1, model.Github, "").AddVersion(model.DefaultPattern, "1.0.0"),
				model.NewRepository(2, model.Github, "").AddVersion(model.DefaultPattern, "1.1.0"),
			},
			nil,
		},
		{
			"gap",
			args{
				repositories: []model.Repository{
					{
						ID:       2,
						Versions: make(map[string]string),
					},
					{
						ID:       3,
						Versions: make(map[string]string),
					},
					{
						ID:       1,
						Versions: make(map[string]string),
					},
				},
			},
			[]model.Repository{
				model.NewRepository(1, model.Github, "").AddVersion(model.DefaultPattern, "1.0.0").AddVersion("beta", "1.0.0-beta"),
				model.NewRepository(2, model.Github, ""),
				model.NewRepository(3, model.Github, "").AddVersion(model.DefaultPattern, "1.1.0").AddVersion("alpha", "2.0.0"),
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

			rows := sqlmock.NewRows([]string{"repository_id", "pattern", "version"})
			mock.ExpectQuery("SELECT repository_id, pattern, version FROM ketchup.repository_version WHERE repository_id = ANY").WillReturnRows(rows)

			switch tc.intention {
			case "invalid rows":
				rows.AddRow("a", model.DefaultPattern, "1.0.0")

			case "sequential":
				rows.AddRow(1, model.DefaultPattern, "1.0.0")
				rows.AddRow(2, model.DefaultPattern, "1.1.0")

			case "gap":
				rows.AddRow(1, "beta", "1.0.0-beta")
				rows.AddRow(1, model.DefaultPattern, "1.0.0")
				rows.AddRow(3, "alpha", "2.0.0")
				rows.AddRow(3, model.DefaultPattern, "1.1.0")
			}

			gotErr := app{db: mockDb}.enrichRepositoriesVersions(context.Background(), tc.args.repositories)

			failed := false

			if tc.wantErr == nil && gotErr != nil {
				failed = true
			} else if tc.wantErr != nil && gotErr == nil {
				failed = true
			} else if tc.wantErr != nil && !strings.Contains(gotErr.Error(), tc.wantErr.Error()) {
				failed = true
			} else if !reflect.DeepEqual(tc.args.repositories, tc.want) {
				failed = true
			}

			if failed {
				t.Errorf("enrichRepositoriesVersions() = (%+v, `%s`), want (%+v, `%s`)", tc.args.repositories, gotErr, tc.want, tc.wantErr)
			}
		})
	}
}

func TestUpdateVersions(t *testing.T) {
	type args struct {
		o model.Repository
	}

	var cases = []struct {
		intention string
		args      args
		wantErr   error
	}{
		{
			"no version",
			args{
				o: model.NoneRepository,
			},
			nil,
		},
		{
			"create error",
			args{
				o: model.NewRepository(0, model.Github, "").AddVersion(model.DefaultPattern, "1.0.0"),
			},
			errFailed,
		},
		{
			"create",
			args{
				o: model.NewRepository(0, model.Github, "").AddVersion(model.DefaultPattern, "1.0.0"),
			},
			nil,
		},
		{
			"no update",
			args{
				o: model.NewRepository(0, model.Github, "").AddVersion(model.DefaultPattern, "1.0.0"),
			},
			nil,
		},
		{
			"update error",
			args{
				o: model.NewRepository(0, model.Github, "").AddVersion(model.DefaultPattern, "1.0.0"),
			},
			errFailed,
		},
		{
			"update",
			args{
				o: model.NewRepository(0, model.Github, "").AddVersion(model.DefaultPattern, "1.0.0"),
			},
			nil,
		},
		{
			"delete error",
			args{
				o: model.NewRepository(0, model.Github, ""),
			},
			errFailed,
		},
		{
			"delete",
			args{
				o: model.NewRepository(0, model.Github, ""),
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

			ctx := context.Background()
			mock.ExpectBegin()
			tx, err := mockDb.Begin()
			if err != nil {
				t.Errorf("unable to create tx: %s", err)
			}
			ctx = db.StoreTx(ctx, tx)

			rows := sqlmock.NewRows([]string{"pattern", "version"})
			mock.ExpectQuery("SELECT pattern, version FROM ketchup.repository_version WHERE repository_id =").WillReturnRows(rows)

			switch tc.intention {
			case "create error":
				mock.ExpectExec("INSERT INTO ketchup.repository_version").WillReturnError(errFailed)

			case "create":
				mock.ExpectExec("INSERT INTO ketchup.repository_version").WillReturnResult(sqlmock.NewResult(0, 1))

			case "no update":
				rows.AddRow(model.DefaultPattern, "1.0.0")

			case "update error":
				rows.AddRow(model.DefaultPattern, "0.9.0")
				mock.ExpectExec("UPDATE ketchup.repository_version SET version =").WillReturnError(errFailed)

			case "update":
				rows.AddRow(model.DefaultPattern, "0.9.0")
				mock.ExpectExec("UPDATE ketchup.repository_version SET version =").WillReturnResult(sqlmock.NewResult(0, 1))

			case "delete error":
				rows.AddRow(model.DefaultPattern, "1.0.0")
				mock.ExpectExec("DELETE FROM ketchup.repository_version WHERE repository_id =").WillReturnError(errFailed)

			case "delete":
				rows.AddRow(model.DefaultPattern, "1.0.0")
				mock.ExpectExec("DELETE FROM ketchup.repository_version WHERE repository_id =").WillReturnResult(sqlmock.NewResult(0, 1))
			}

			gotErr := New(mockDb).UpdateVersions(ctx, tc.args.o)

			failed := false

			if !errors.Is(gotErr, tc.wantErr) {
				failed = true
			}

			if failed {
				t.Errorf("UpdateVersions() = `%s`, want `%s`", gotErr, tc.wantErr)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("sqlmock unfilled expectations: %s", err)
			}
		})
	}
}
