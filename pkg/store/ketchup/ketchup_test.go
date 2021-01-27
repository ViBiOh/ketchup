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
			[]model.Ketchup{
				{
					Pattern:    model.DefaultPattern,
					Version:    "0.9.0",
					Repository: model.NewRepository(1, model.Github, "vibioh/ketchup").AddVersion(model.DefaultPattern, "1.0.0"),
					User: model.User{
						ID:    3,
						Email: "nobody@localhost",
					},
				},
				{
					Pattern:    model.DefaultPattern,
					Version:    "1.0.0",
					Repository: model.NewRepository(2, model.Helm, "vibioh/viws").AddVersion(model.DefaultPattern, "1.0.0"),
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
			[]model.Ketchup{},
			0,
			sqlmock.ErrCancelled,
		},
		{
			"invalid rows",
			args{
				page:     1,
				pageSize: 20,
			},
			[]model.Ketchup{},
			0,
			errors.New("converting driver.Value type string (\"a\") to a uint64: invalid syntax"),
		},
		{
			"invalid kind",
			args{
				page:     1,
				pageSize: 20,
			},
			[]model.Ketchup{},
			1,
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

			rows := sqlmock.NewRows([]string{"pattern", "version", "repository_id", "name", "kind", "repository_version", "full_count"})
			expectedQuery := mock.ExpectQuery("SELECT k.pattern, k.version, k.repository_id, r.name, r.kind, rv.version, .+ AS full_count FROM ketchup.ketchup k, ketchup.repository r, ketchup.repository_version rv WHERE user_id = .+ AND k.repository_id = r.id AND rv.repository_id = r.id AND rv.pattern = k.pattern").WithArgs(20, 0, 3).WillReturnRows(rows)

			switch tc.intention {
			case "simple":
				rows.AddRow(model.DefaultPattern, "0.9.0", 1, "vibioh/ketchup", "github", "1.0.0", 2).AddRow(model.DefaultPattern, "1.0.0", 2, "vibioh/viws", "helm", "1.0.0", 2)

			case "timeout":
				savedSQLTimeout := db.SQLTimeout
				db.SQLTimeout = time.Second
				defer func() {
					db.SQLTimeout = savedSQLTimeout
				}()
				expectedQuery.WillDelayFor(db.SQLTimeout * 2)

			case "invalid rows":
				rows.AddRow(model.DefaultPattern, "0.9.0", "a", "vibioh/ketchup", "github", "0.9.0", 2)

			case "invalid kind":
				rows.AddRow(model.DefaultPattern, "1.0.0", 2, "vibioh/viws", "wrong", "1.0.0", 1)
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
					Repository: model.NewRepository(1, model.Github, ""),
					User: model.User{
						ID:    1,
						Email: "nobody@localhost",
					},
				},
				{
					Pattern:    model.DefaultPattern,
					Version:    "1.0.0",
					Repository: model.NewRepository(2, model.Github, ""),
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
			mockDb, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("unable to create mock database: %s", err)
			}
			defer mockDb.Close()

			rows := sqlmock.NewRows([]string{"pattern", "version", "repository_id", "user_id", "email"})
			expectedQuery := mock.ExpectQuery("SELECT k.pattern, k.version, k.repository_id, k.user_id, u.email FROM ketchup.ketchup k, ketchup.user u WHERE repository_id = ANY .+ AND k.user_id = u.id").WithArgs(pq.Array(tc.args.ids)).WillReturnRows(rows)

			switch tc.intention {
			case "simple":
				rows.AddRow(model.DefaultPattern, "0.9.0", 1, 1, "nobody@localhost").AddRow(model.DefaultPattern, "1.0.0", 2, 2, "guest@domain")

			case "timeout":
				savedSQLTimeout := db.SQLTimeout
				db.SQLTimeout = time.Second
				defer func() {
					db.SQLTimeout = savedSQLTimeout
				}()
				expectedQuery.WillDelayFor(db.SQLTimeout * 2)

			case "invalid rows":
				rows.AddRow(model.DefaultPattern, "0.9.0", "a", 1, "nobody@localhost")
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
			"SELECT pattern, version, repository_id, user_id FROM ketchup.ketchup WHERE repository_id = .+ AND user_id = .+",
			model.Ketchup{
				Pattern:    model.DefaultPattern,
				Version:    "0.9.0",
				Repository: model.NewRepository(1, model.Github, ""),
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
			"SELECT pattern, version, repository_id, user_id FROM ketchup.ketchup WHERE repository_id = .+ AND user_id = .+",
			model.NoneKetchup,
			nil,
		},
		{
			"for update",
			args{
				id:        1,
				forUpdate: true,
			},
			"SELECT pattern, version, repository_id, user_id FROM ketchup.ketchup WHERE repository_id = .+ AND user_id = .+ FOR UPDATE",
			model.Ketchup{
				Pattern:    model.DefaultPattern,
				Version:    "0.9.0",
				Repository: model.NewRepository(1, model.Github, ""),
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

			rows := sqlmock.NewRows([]string{"patter", "version", "repository_id", "user_id"})
			mock.ExpectQuery(tc.expectSQL).WithArgs(1, 3).WillReturnRows(rows)

			switch tc.intention {
			case "for update":
				fallthrough

			case "simple":
				rows.AddRow(model.DefaultPattern, "0.9.0", 1, 3)
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
					Pattern:    model.DefaultPattern,
					Version:    "0.9.0",
					Repository: model.NewRepository(1, model.Github, ""),
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

			mock.ExpectQuery("INSERT INTO ketchup.ketchup").WithArgs(model.DefaultPattern, "0.9.0", 1, 3).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

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
					Pattern:    model.DefaultPattern,
					Version:    "0.9.0",
					Repository: model.NewRepository(1, model.Github, ""),
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

			mock.ExpectExec("UPDATE ketchup.ketchup SET pattern = .+, version = .+").WithArgs(1, 3, model.DefaultPattern, "0.9.0").WillReturnResult(sqlmock.NewResult(0, 1))

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
					Repository: model.NewRepository(1, model.Github, ""),
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

			mock.ExpectExec("DELETE FROM ketchup.ketchup").WithArgs(1, 3).WillReturnResult(sqlmock.NewResult(0, 1))

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
