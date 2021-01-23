package repository

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

func TestList(t *testing.T) {
	type args struct {
		page     uint
		pageSize uint
	}

	var cases = []struct {
		intention string
		args      args
		want      []model.Repository
		wantCount uint64
		wantErr   error
	}{
		{
			"success",
			args{
				page:     1,
				pageSize: 20,
			},
			[]model.Repository{
				{
					ID:   1,
					Name: "vibioh/ketchup",
					Versions: map[string]string{
						model.DefaultPattern: "1.0.0",
					},
				},
				{
					ID:   2,
					Name: "vibioh/viws",
					Versions: map[string]string{
						model.DefaultPattern: "1.2.3",
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
			[]model.Repository{},
			0,
			sqlmock.ErrCancelled,
		},
		{
			"invalid rows",
			args{
				page:     1,
				pageSize: 20,
			},
			[]model.Repository{},
			0,
			errors.New("converting driver.Value type string (\"a\") to a uint64: invalid syntax"),
		},
		{
			"invalid kind",
			args{
				page:     1,
				pageSize: 20,
			},
			[]model.Repository{},
			0,
			errors.New("invalid value `wrong` for repository kind"),
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			mockDb, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("unable to create mock database: %s", err)
			}
			defer mockDb.Close()

			rows := sqlmock.NewRows([]string{"id", "name", "kind", "full_count"})
			expectedQuery := mock.ExpectQuery("SELECT id, name, kind, .+ AS full_count FROM ketchup.repository").WithArgs(20, 0).WillReturnRows(rows)

			switch tc.intention {
			case "success":
				mock.ExpectQuery(
					"SELECT repository_id, pattern, version FROM ketchup.repository_version WHERE repository_id = ANY",
				).WithArgs(
					pq.Array([]uint64{1, 2}),
				).WillReturnRows(
					sqlmock.NewRows([]string{"repository_id", "pattern", "version"}).AddRow(1, model.DefaultPattern, "1.0.0").AddRow(2, model.DefaultPattern, "1.2.3"),
				)
				rows.AddRow(1, "vibioh/ketchup", "github", 2).AddRow(2, "vibioh/viws", "github", 2)

			case "timeout":
				savedSQLTimeout := db.SQLTimeout
				db.SQLTimeout = time.Second
				defer func() {
					db.SQLTimeout = savedSQLTimeout
				}()
				expectedQuery.WillDelayFor(db.SQLTimeout * 2)

			case "invalid rows":
				rows.AddRow("a", "vibioh/ketchup", "github", 2)

			case "invalid kind":
				rows.AddRow(1, "vibioh/viws", "wrong", 1)
			}

			got, gotCount, gotErr := New(mockDb).List(context.Background(), tc.args.page, tc.args.pageSize)
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

func TestSuggest(t *testing.T) {
	type args struct {
		ignoreIds []uint64
		count     uint64
	}

	var cases = []struct {
		intention string
		args      args
		want      []model.Repository
		wantErr   error
	}{
		{
			"simple",
			args{
				ignoreIds: []uint64{8000},
				count:     2,
			},
			[]model.Repository{
				{
					ID:   1,
					Name: "vibioh/ketchup",
					Versions: map[string]string{
						model.DefaultPattern: "1.0.0",
					},
				},
				{
					ID:   2,
					Name: "vibioh/viws",
					Versions: map[string]string{
						model.DefaultPattern: "1.2.3",
					},
					Kind: model.Helm,
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

			rows := sqlmock.NewRows([]string{"id", "name", "kind", "full_count"})
			mock.ExpectQuery("SELECT id, name, kind, \\( SELECT COUNT\\(1\\) FROM ketchup.ketchup WHERE repository_id = id \\) AS count FROM ketchup.repository").WithArgs(tc.args.count, pq.Array(tc.args.ignoreIds)).WillReturnRows(rows)
			rows.AddRow(1, "vibioh/ketchup", "github", 2).AddRow(2, "vibioh/viws", "helm", 2)

			mock.ExpectQuery(
				"SELECT repository_id, pattern, version FROM ketchup.repository_version WHERE repository_id = ANY",
			).WithArgs(
				pq.Array([]uint64{1, 2}),
			).WillReturnRows(
				sqlmock.NewRows([]string{"repository_id", "pattern", "version"}).AddRow(1, model.DefaultPattern, "1.0.0").AddRow(2, model.DefaultPattern, "1.2.3"),
			)

			got, gotErr := New(mockDb).Suggest(context.Background(), tc.args.ignoreIds, tc.args.count)
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
				t.Errorf("Suggest() = (%+v, `%s`), want (%+v, `%s`)", got, gotErr, tc.want, tc.wantErr)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("sqlmock unfilled expectations: %s", err)
			}
		})
	}
}

func TestGet(t *testing.T) {
	type args struct {
		id        uint64
		forUpdate bool
	}

	var cases = []struct {
		intention string
		args      args
		expectSQL string
		want      model.Repository
		wantErr   error
	}{
		{
			"simple",
			args{
				id: 1,
			},
			"SELECT id, name, kind FROM ketchup.repository WHERE id =",
			model.Repository{
				ID:   1,
				Name: "vibioh/ketchup",
				Versions: map[string]string{
					model.DefaultPattern: "1.0.0",
				},
				Kind: model.Helm,
			},
			nil,
		},
		{
			"for update",
			args{
				id:        1,
				forUpdate: true,
			},
			"SELECT id, name, kind FROM ketchup.repository WHERE id = .+ FOR UPDATE",
			model.Repository{
				ID:   1,
				Name: "vibioh/ketchup",
				Versions: map[string]string{
					model.DefaultPattern: "1.0.0",
				},
				Kind: model.Helm,
			},
			nil,
		},
		{
			"no rows",
			args{
				id: 1,
			},
			"SELECT id, name, kind FROM ketchup.repository WHERE id =",
			model.NoneRepository,
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

			rows := sqlmock.NewRows([]string{"id", "name", "kind"})
			mock.ExpectQuery(tc.expectSQL).WithArgs(1).WillReturnRows(rows)

			switch tc.intention {
			case "simple":
				fallthrough

			case "for update":
				rows.AddRow(1, "vibioh/ketchup", "helm")

				mock.ExpectQuery(
					"SELECT repository_id, pattern, version FROM ketchup.repository_version WHERE repository_id = ANY",
				).WithArgs(
					pq.Array([]uint64{1}),
				).WillReturnRows(
					sqlmock.NewRows([]string{"repository_id", "pattern", "version"}).AddRow(1, model.DefaultPattern, "1.0.0"),
				)
			}

			got, gotErr := New(mockDb).Get(context.Background(), tc.args.id, tc.args.forUpdate)

			failed := false

			if !errors.Is(gotErr, tc.wantErr) {
				failed = true
			} else if !reflect.DeepEqual(got, tc.want) {
				failed = true
			}

			if failed {
				t.Errorf("Get() = (%+v, `%s`), want (%+v, `%s`)", got, gotErr, tc.want, tc.wantErr)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("sqlmock unfilled expectations: %s", err)
			}
		})
	}
}

func TestGetByName(t *testing.T) {
	type args struct {
		name           string
		repositoryKind model.RepositoryKind
	}

	var cases = []struct {
		intention string
		args      args
		want      model.Repository
		wantErr   error
	}{
		{
			"simple",
			args{
				name: "vibioh/ketchup",
			},
			model.Repository{
				ID:   1,
				Name: "vibioh/ketchup",
				Versions: map[string]string{
					model.DefaultPattern: "1.0.0",
				},
				Kind: model.Github,
			},
			nil,
		},
		{
			"no rows",
			args{
				name: "vibioh/ketchup",
			},
			model.NoneRepository,
			nil,
		},
		{
			"invalid kind",
			args{
				name: "vibioh/ketchup",
			},
			model.NoneRepository,
			errors.New("invalid value `wrong` for repository kind"),
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			mockDb, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("unable to create mock database: %s", err)
			}
			defer mockDb.Close()

			rows := sqlmock.NewRows([]string{"id", "name", "kind"})
			mock.ExpectQuery("SELECT id, name, kind FROM ketchup.repository WHERE name =").WithArgs("vibioh/ketchup", "github").WillReturnRows(rows)

			switch tc.intention {
			case "simple":
				rows.AddRow(1, "vibioh/ketchup", "github")
				mock.ExpectQuery(
					"SELECT repository_id, pattern, version FROM ketchup.repository_version WHERE repository_id = ANY",
				).WithArgs(
					pq.Array([]uint64{1}),
				).WillReturnRows(
					sqlmock.NewRows([]string{"repository_id", "pattern", "version"}).AddRow(1, model.DefaultPattern, "1.0.0"),
				)

			case "invalid kind":
				rows.AddRow(1, "vibioh/ketchup", "wrong")
			}

			got, gotErr := New(mockDb).GetByName(context.Background(), tc.args.name, tc.args.repositoryKind)

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
				t.Errorf("GetByName() = (%+v, `%s`), want (%+v, `%s`)", got, gotErr, tc.want, tc.wantErr)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("sqlmock unfilled expectations: %s", err)
			}
		})
	}
}

func TestCreate(t *testing.T) {
	type args struct {
		o model.Repository
	}

	var cases = []struct {
		intention string
		args      args
		want      uint64
		wantErr   error
	}{
		{
			"error lock",
			args{
				o: model.Repository{
					Name: "vibioh/ketchup",
					Versions: map[string]string{
						model.DefaultPattern: "1.0.0",
					},
					Kind: model.Github,
				},
			},
			0,
			errors.New("unable to obtain lock"),
		},
		{
			"error get",
			args{
				o: model.Repository{
					Name: "vibioh/ketchup",
					Versions: map[string]string{
						model.DefaultPattern: "1.0.0",
					},
					Kind: model.Github,
				},
			},
			0,
			errors.New("unable to read"),
		},
		{
			"found get",
			args{
				o: model.Repository{
					Name: "vibioh/ketchup",
					Versions: map[string]string{
						model.DefaultPattern: "1.0.0",
					},
					Kind: model.Github,
				},
			},
			0,
			errors.New("repository already exists with name"),
		},
		{
			"success",
			args{
				o: model.Repository{
					Name: "vibioh/ketchup",
					Versions: map[string]string{
						model.DefaultPattern: "1.0.0",
					},
					Kind: model.Github,
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
			tx, err := mockDb.Begin()
			if err != nil {
				t.Errorf("unable to create tx: %s", err)
			}
			ctx = db.StoreTx(ctx, tx)

			lockQuery := mock.ExpectExec("LOCK ketchup.repository IN SHARE ROW EXCLUSIVE MODE")

			switch tc.intention {
			case "error lock":
				lockQuery.WillReturnError(errors.New("unable to obtain lock"))
			case "error get":
				lockQuery.WillReturnResult(sqlmock.NewResult(0, 0))
				mock.ExpectQuery("SELECT id, name, kind FROM ketchup.repository WHERE name = .+ AND kind = .+").WithArgs("vibioh/ketchup", tc.args.o.Kind.String()).WillReturnError(errors.New("unable to read"))
			case "found get":
				lockQuery.WillReturnResult(sqlmock.NewResult(0, 0))
				mock.ExpectQuery("SELECT id, name, kind FROM ketchup.repository WHERE name = .+ AND kind = .+").WithArgs("vibioh/ketchup", "github").WillReturnRows(sqlmock.NewRows([]string{"id", "name", "kind"}).AddRow(1, "vibioh/ketchup", "github"))
				mock.ExpectQuery(
					"SELECT repository_id, pattern, version FROM ketchup.repository_version WHERE repository_id = ANY",
				).WithArgs(
					pq.Array([]uint64{1}),
				).WillReturnRows(
					sqlmock.NewRows([]string{"repository_id", "pattern", "version"}).AddRow(1, model.DefaultPattern, "1.0.0"),
				)
			case "success":
				lockQuery.WillReturnResult(sqlmock.NewResult(0, 0))
				mock.ExpectQuery("SELECT id, name, kind FROM ketchup.repository WHERE name = .+ AND kind = .+").WithArgs("vibioh/ketchup", tc.args.o.Kind.String()).WillReturnRows(sqlmock.NewRows([]string{"id", "name", "kind"}))
				mock.ExpectQuery("INSERT INTO ketchup.repository").WithArgs("vibioh/ketchup", "github").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
				mock.ExpectQuery("SELECT pattern, version FROM ketchup.repository_version WHERE repository_id = .+").WithArgs(1).WillReturnRows(sqlmock.NewRows([]string{"pattern", "version"}))
				mock.ExpectExec("INSERT INTO ketchup.repository_version").WithArgs(1, model.DefaultPattern, "1.0.0").WillReturnResult(sqlmock.NewResult(0, 1))
			}

			got, gotErr := New(mockDb).Create(ctx, tc.args.o)

			failed := false

			if tc.wantErr == nil && gotErr != nil {
				failed = true
			} else if tc.wantErr != nil && gotErr == nil {
				failed = true
			} else if tc.wantErr != nil && !strings.Contains(gotErr.Error(), tc.wantErr.Error()) {
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

func TestDeleteUnused(t *testing.T) {
	var cases = []struct {
		intention string
		wantErr   error
	}{
		{
			"simple",
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

			mock.ExpectExec("DELETE FROM ketchup.repository WHERE id NOT IN").WillReturnResult(sqlmock.NewResult(0, 1))

			gotErr := New(mockDb).DeleteUnused(ctx)

			failed := false

			if !errors.Is(gotErr, tc.wantErr) {
				failed = true
			}

			if failed {
				t.Errorf("DeleteUnused() = `%s`, want `%s`", gotErr, tc.wantErr)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("sqlmock unfilled expectations: %s", err)
			}
		})
	}
}
