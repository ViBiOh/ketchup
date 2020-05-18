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
		want      []model.Repository
		wantCount uint64
		wantErr   error
	}{
		{
			"simple",
			args{
				page:     1,
				pageSize: 20,
			},
			"SELECT id, name, version, .+ AS full_count FROM repository",
			[]model.Repository{
				{
					ID:      1,
					Name:    "vibioh/ketchup",
					Version: "1.0.0",
				},
				{
					ID:      2,
					Name:    "vibioh/viws",
					Version: "1.2.3",
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
			"SELECT id, name, version, .+ AS full_count FROM repository",
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
			"SELECT id, name, version, .+ AS full_count FROM repository",
			[]model.Repository{},
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

			rows := sqlmock.NewRows([]string{"id", "email", "login_id", "full_count"})
			expectedQuery := mock.ExpectQuery(tc.expectSQL).WithArgs(20, 0).WillReturnRows(rows)

			if tc.intention != "invalid rows" {
				rows.AddRow(1, "vibioh/ketchup", "1.0.0", 2).AddRow(2, "vibioh/viws", "1.2.3", 2)
			} else {
				rows.AddRow("a", "vibioh/ketchup", "1.0.0", 2)
			}

			if tc.intention == "timeout" {
				savedSQLTimeout := db.SQLTimeout
				db.SQLTimeout = time.Second
				defer func() {
					db.SQLTimeout = savedSQLTimeout
				}()

				expectedQuery.WillDelayFor(db.SQLTimeout * 2)
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
			"SELECT id, name, version FROM repository WHERE id =",
			model.Repository{
				ID:      1,
				Name:    "vibioh/ketchup",
				Version: "1.0.0",
			},
			nil,
		},
		{
			"for update",
			args{
				id:        1,
				forUpdate: true,
			},
			"SELECT id, name, version FROM repository WHERE id = .+ FOR UPDATE",
			model.Repository{
				ID:      1,
				Name:    "vibioh/ketchup",
				Version: "1.0.0",
			},
			nil,
		},
		{
			"no rows",
			args{
				id: 1,
			},
			"SELECT id, name, version FROM repository WHERE id =",
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

			rows := sqlmock.NewRows([]string{"id", "email", "login_id"})
			mock.ExpectQuery(tc.expectSQL).WithArgs(1).WillReturnRows(rows)
			if tc.intention != "no rows" {
				rows.AddRow(1, "vibioh/ketchup", "1.0.0")
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
		name string
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
				ID:      1,
				Name:    "vibioh/ketchup",
				Version: "1.0.0",
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
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			mockDb, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("unable to create mock database: %s", err)
			}
			defer mockDb.Close()

			rows := sqlmock.NewRows([]string{"id", "email", "login_id"})
			mock.ExpectQuery("SELECT id, name, version FROM repository WHERE name =").WithArgs("vibioh/ketchup").WillReturnRows(rows)

			if tc.intention != "no rows" {
				rows.AddRow(1, "vibioh/ketchup", "1.0.0")
			}

			got, gotErr := New(mockDb).GetByName(context.Background(), tc.args.name)

			failed := false

			if !errors.Is(gotErr, tc.wantErr) {
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
					Name:    "vibioh/ketchup",
					Version: "1.0.0",
				},
			},
			0,
			errors.New("unable to obtain lock"),
		},
		{
			"error get",
			args{
				o: model.Repository{
					Name:    "vibioh/ketchup",
					Version: "1.0.0",
				},
			},
			0,
			errors.New("unable to read"),
		},
		{
			"found get",
			args{
				o: model.Repository{
					Name:    "vibioh/ketchup",
					Version: "1.0.0",
				},
			},
			1,
			nil,
		},
		{
			"no rows",
			args{
				o: model.Repository{
					Name:    "vibioh/ketchup",
					Version: "1.0.0",
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

			lockQuery := mock.ExpectExec("LOCK repository IN SHARE ROW EXCLUSIVE MODE")
			if tc.intention == "error lock" {
				lockQuery.WillReturnError(errors.New("unable to obtain lock"))
			} else {
				lockQuery.WillReturnResult(sqlmock.NewResult(0, 0))

				getQuery := mock.ExpectQuery("SELECT id, name, version FROM repository WHERE name =").WithArgs("vibioh/ketchup")

				if tc.intention == "error get" {
					getQuery.WillReturnError(errors.New("unable to read"))
				} else {
					rows := sqlmock.NewRows([]string{"id", "email", "login_id"})
					getQuery.WillReturnRows(rows)

					if tc.intention == "no rows" {
						mock.ExpectQuery("INSERT INTO repository").WithArgs("vibioh/ketchup", "1.0.0").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
					} else {
						rows.AddRow(1, "vibioh/ketchup", "1.0.0")
					}
				}

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

func TestUpdate(t *testing.T) {
	type args struct {
		o model.Repository
	}

	var cases = []struct {
		intention string
		args      args
		wantErr   error
	}{
		{
			"simple",
			args{
				o: model.Repository{
					ID:      1,
					Name:    "vibioh/ketchup",
					Version: "1.0.0",
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

			ctx := context.Background()
			mock.ExpectBegin()
			tx, err := mockDb.Begin()
			if err != nil {
				t.Errorf("unable to create tx: %s", err)
			}
			ctx = db.StoreTx(ctx, tx)

			mock.ExpectExec("UPDATE repository SET version").WithArgs(1, "1.0.0").WillReturnResult(sqlmock.NewResult(0, 1))

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

			mock.ExpectExec("DELETE FROM repository WHERE id NOT IN").WillReturnResult(sqlmock.NewResult(0, 1))

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
