package ketchup

import (
	"context"
	"database/sql"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	authModel "github.com/ViBiOh/auth/v2/pkg/model"
	"github.com/ViBiOh/httputils/v4/pkg/db"
	"github.com/ViBiOh/ketchup/pkg/model"
	"github.com/lib/pq"
)

var (
	testCtx = model.StoreUser(context.Background(), model.NewUser(3, testEmail, authModel.NewUser(0, "")))

	testEmail       = "nobody@localhost"
	repositoryName  = "vibioh/ketchup"
	viwsRepository  = "vibioh/viws"
	chartRepository = "https://charts.vibioh.fr"

	repositoryVersion = "1.0.0"
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

func testWithTransaction(t *testing.T, mockDb *sql.DB, mock sqlmock.Sqlmock) context.Context {
	mock.ExpectBegin()
	tx, err := mockDb.Begin()
	if err != nil {
		t.Errorf("unable to create tx: %s", err)
	}

	return db.StoreTx(testCtx, tx)
}

func TestList(t *testing.T) {
	type args struct {
		pageSize uint
	}

	var cases = []struct {
		intention string
		args      args
		want      []model.Ketchup
		wantCount uint64
		wantErr   error
	}{
		{
			"simple",
			args{
				pageSize: 20,
			},
			[]model.Ketchup{
				{
					Pattern:    model.DefaultPattern,
					Version:    "0.9.0",
					Frequency:  model.Daily,
					Repository: model.NewGithubRepository(1, repositoryName).AddVersion(model.DefaultPattern, repositoryVersion),
					User:       model.NewUser(3, testEmail, authModel.NewUser(0, "")),
				},
				{
					Pattern:    model.DefaultPattern,
					Version:    repositoryVersion,
					Frequency:  model.Daily,
					Repository: model.NewHelmRepository(2, chartRepository, "app").AddVersion(model.DefaultPattern, repositoryVersion),
					User:       model.NewUser(3, testEmail, authModel.NewUser(0, "")),
				},
			},
			2,
			nil,
		},
		{
			"timeout",
			args{
				pageSize: 20,
			},
			[]model.Ketchup{},
			0,
			sqlmock.ErrCancelled,
		},
		{
			"invalid rows",
			args{
				pageSize: 20,
			},
			[]model.Ketchup{},
			0,
			errors.New("converting driver.Value type string (\"a\") to a uint64: invalid syntax"),
		},
		{
			"invalid kind",
			args{
				pageSize: 20,
			},
			[]model.Ketchup{},
			1,
			errors.New("invalid value `wrong` for repository kind"),
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			testWithMock(t, func(mockDb *sql.DB, mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"pattern", "version", "frequency", "repository_id", "name", "part", "kind", "repository_version", "full_count"})
				expectedQuery := mock.ExpectQuery("SELECT k.pattern, k.version, k.frequency, k.repository_id, r.name, r.part, r.kind, rv.version, .+ AS full_count FROM ketchup.ketchup k, ketchup.repository r, ketchup.repository_version rv WHERE user_id = .+ AND k.repository_id = r.id AND rv.repository_id = r.id AND rv.pattern = k.pattern ORDER BY k.repository_id ASC, k.pattern ASC").WithArgs(3, 20).WillReturnRows(rows)

				switch tc.intention {
				case "simple":
					rows.AddRow(model.DefaultPattern, "0.9.0", "Daily", 1, repositoryName, "", "github", repositoryVersion, 2).AddRow(model.DefaultPattern, repositoryVersion, "Daily", 2, chartRepository, "app", "helm", repositoryVersion, 2)

				case "timeout":
					savedSQLTimeout := db.SQLTimeout
					db.SQLTimeout = time.Second
					defer func() {
						db.SQLTimeout = savedSQLTimeout
					}()
					expectedQuery.WillDelayFor(db.SQLTimeout * 2)

				case "invalid rows":
					rows.AddRow(model.DefaultPattern, "0.9.0", "Daily", "a", repositoryName, "", "github", "0.9.0", 2)

				case "invalid kind":
					rows.AddRow(model.DefaultPattern, repositoryVersion, "Daily", 2, viwsRepository, "", "wrong", repositoryVersion, 1)
				}

				got, gotCount, gotErr := New(mockDb).List(testCtx, tc.args.pageSize, "")
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
			})
		})
	}
}

func TestListByRepositoriesID(t *testing.T) {
	type args struct {
		ids       []uint64
		frequency model.KetchupFrequency
	}

	var cases = []struct {
		intention string
		args      args
		want      []model.Ketchup
		wantErr   error
	}{
		{
			"simple",
			args{
				ids: []uint64{1, 2},
			},
			[]model.Ketchup{
				{
					Pattern:    model.DefaultPattern,
					Version:    "0.9.0",
					Frequency:  model.Daily,
					Repository: model.NewGithubRepository(1, ""),
					User:       model.NewUser(1, testEmail, authModel.NewUser(0, "")),
				},
				{
					Pattern:    model.DefaultPattern,
					Version:    repositoryVersion,
					Frequency:  model.Daily,
					Repository: model.NewGithubRepository(2, ""),
					User:       model.NewUser(2, "guest@domain", authModel.NewUser(0, "")),
				},
			},
			nil,
		},
		{
			"timeout",
			args{
				ids:       []uint64{1, 2},
				frequency: model.Daily,
			},
			make([]model.Ketchup, 0),
			sqlmock.ErrCancelled,
		},
		{
			"invalid rows",
			args{
				ids: []uint64{1, 2},
			},
			make([]model.Ketchup, 0),
			errors.New("converting driver.Value type string (\"a\") to a uint64: invalid syntax"),
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			testWithMock(t, func(mockDb *sql.DB, mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"pattern", "version", "frequency", "repository_id", "user_id", "email"})
				expectedQuery := mock.ExpectQuery("SELECT k.pattern, k.version, k.frequency, k.repository_id, k.user_id, u.email FROM ketchup.ketchup k, ketchup.user u WHERE repository_id = ANY .+ AND k.user_id = u.id AND k.frequency = .+").WithArgs(pq.Array(tc.args.ids), strings.ToLower(tc.args.frequency.String())).WillReturnRows(rows)

				switch tc.intention {
				case "simple":
					rows.AddRow(model.DefaultPattern, "0.9.0", "Daily", 1, 1, testEmail).AddRow(model.DefaultPattern, repositoryVersion, "Daily", 2, 2, "guest@domain")

				case "timeout":
					savedSQLTimeout := db.SQLTimeout
					db.SQLTimeout = time.Second
					defer func() {
						db.SQLTimeout = savedSQLTimeout
					}()
					expectedQuery.WillDelayFor(db.SQLTimeout * 2)

				case "invalid rows":
					rows.AddRow(model.DefaultPattern, "0.9.0", "Daily", "a", 1, testEmail)
				}

				got, gotErr := New(mockDb).ListByRepositoriesID(testCtx, tc.args.ids, tc.args.frequency)
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
			})
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
			"SELECT k.pattern, k.version, k.frequency, k.repository_id, k.user_id, r.name, r.part, r.kind FROM ketchup.ketchup k, ketchup.repository r WHERE k.repository_id = .+ AND k.user_id = .+ AND k.repository_id = r.id",
			model.Ketchup{
				Pattern:    model.DefaultPattern,
				Version:    "0.9.0",
				Frequency:  model.Daily,
				Repository: model.NewGithubRepository(1, repositoryName),
				User:       model.NewUser(3, testEmail, authModel.NewUser(0, "")),
			},
			nil,
		},
		{
			"no rows",
			args{
				id: 1,
			},
			"SELECT k.pattern, k.version, k.frequency, k.repository_id, k.user_id, r.name, r.part, r.kind FROM ketchup.ketchup k, ketchup.repository r WHERE k.repository_id = .+ AND k.user_id = .+ AND k.repository_id = r.id",
			model.NoneKetchup,
			nil,
		},
		{
			"for update",
			args{
				id:        1,
				forUpdate: true,
			},
			"SELECT k.pattern, k.version, k.frequency, k.repository_id, k.user_id, r.name, r.part, r.kind FROM ketchup.ketchup k, ketchup.repository r WHERE k.repository_id = .+ AND k.user_id = .+ AND k.repository_id = r.id FOR UPDATE",
			model.Ketchup{
				Pattern:    model.DefaultPattern,
				Version:    "0.9.0",
				Frequency:  model.Daily,
				Repository: model.NewGithubRepository(1, repositoryName),
				User:       model.NewUser(3, testEmail, authModel.NewUser(0, "")),
			},
			nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			testWithMock(t, func(mockDb *sql.DB, mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"patter", "version", "frequency", "repository_id", "user_id", "name", "part", "kind"})
				mock.ExpectQuery(tc.expectSQL).WithArgs(1, 3).WillReturnRows(rows)

				switch tc.intention {
				case "for update":
					fallthrough

				case "simple":
					rows.AddRow(model.DefaultPattern, "0.9.0", "Daily", 1, 3, repositoryName, "", "github")
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
			})
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
					Pattern:    model.DefaultPattern,
					Version:    "0.9.0",
					Frequency:  model.Daily,
					Repository: model.NewGithubRepository(1, ""),
				},
			},
			1,
			nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			testWithMock(t, func(mockDb *sql.DB, mock sqlmock.Sqlmock) {
				ctx := testWithTransaction(t, mockDb, mock)

				mock.ExpectQuery("INSERT INTO ketchup.ketchup").WithArgs(model.DefaultPattern, "0.9.0", "daily", 1, 3).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

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
			})
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
					Pattern:    model.DefaultPattern,
					Version:    "0.9.0",
					Frequency:  model.Daily,
					Repository: model.NewGithubRepository(1, ""),
				},
			},
			nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			testWithMock(t, func(mockDb *sql.DB, mock sqlmock.Sqlmock) {
				ctx := testWithTransaction(t, mockDb, mock)

				mock.ExpectExec("UPDATE ketchup.ketchup SET pattern = .+, version = .+").WithArgs(1, 3, model.DefaultPattern, "0.9.0", "daily").WillReturnResult(sqlmock.NewResult(0, 1))

				gotErr := New(mockDb).Update(ctx, tc.args.o)

				failed := false

				if !errors.Is(gotErr, tc.wantErr) {
					failed = true
				}

				if failed {
					t.Errorf("Update() = `%s`, want `%s`", gotErr, tc.wantErr)
				}
			})
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
					Repository: model.NewGithubRepository(1, ""),
				},
			},
			nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			testWithMock(t, func(mockDb *sql.DB, mock sqlmock.Sqlmock) {
				ctx := testWithTransaction(t, mockDb, mock)

				mock.ExpectExec("DELETE FROM ketchup.ketchup").WithArgs(1, 3).WillReturnResult(sqlmock.NewResult(0, 1))

				gotErr := New(mockDb).Delete(ctx, tc.args.o)

				failed := false

				if !errors.Is(gotErr, tc.wantErr) {
					failed = true
				}

				if failed {
					t.Errorf("Delete() = `%s`, want `%s`", gotErr, tc.wantErr)
				}
			})
		})
	}
}
