package ketchup

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

func TestList(t *testing.T) {
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
		want      []model.Ketchup
		wantCount uint
		wantErr   error
	}{
		{
			"simple",
			args{
				page:     1,
				pageSize: 20,
			},
			"SELECT version, repository_id, user_id, .+ AS full_count FROM ketchup WHERE user_id = .+ ORDER BY creation_date DESC",
			[]model.Ketchup{
				{
					Version: "0.9.0",
					Repository: model.Repository{
						ID: 1,
					},
					User: model.User{
						ID: 3,
					},
				},
				{
					Version: "1.0.0",
					Repository: model.Repository{
						ID: 2,
					},
					User: model.User{
						ID: 3,
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
			"SELECT version, repository_id, user_id, .+ AS full_count FROM ketchup WHERE user_id = .+ ORDER BY creation_date DESC",
			nil,
			0,
			sqlmock.ErrCancelled,
		},
		{
			"order",
			args{
				page:     1,
				pageSize: 20,
				sortKey:  "version",
				sortAsc:  true,
			},
			"SELECT version, repository_id, user_id, .+ AS full_count FROM ketchup WHERE user_id = .+ ORDER BY version",
			[]model.Ketchup{
				{
					Version: "0.9.0",
					Repository: model.Repository{
						ID: 1,
					},
					User: model.User{
						ID: 3,
					},
				},
				{
					Version: "1.0.0",
					Repository: model.Repository{
						ID: 2,
					},
					User: model.User{
						ID: 3,
					},
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
				sortKey:  "version",
				sortAsc:  false,
			},
			"SELECT version, repository_id, user_id, .+ AS full_count FROM ketchup WHERE user_id = .+ ORDER BY version DESC",
			[]model.Ketchup{
				{
					Version: "0.9.0",
					Repository: model.Repository{
						ID: 1,
					},
					User: model.User{
						ID: 3,
					},
				},
				{
					Version: "1.0.0",
					Repository: model.Repository{
						ID: 2,
					},
					User: model.User{
						ID: 3,
					},
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

			expectedQuery := mock.ExpectQuery(tc.expectSQL).WithArgs(20, 0, 3).WillReturnRows(sqlmock.NewRows([]string{"version", "repository_id", "user_id", "full_count"}).AddRow("0.9.0", 1, 3, 2).AddRow("1.0.0", 2, 3, 2))

			if tc.intention == "timeout" {
				savedSQLTimeout := db.SQLTimeout
				db.SQLTimeout = time.Second
				defer func() {
					db.SQLTimeout = savedSQLTimeout
				}()

				expectedQuery.WillDelayFor(db.SQLTimeout * 2)
			}

			got, gotCount, gotErr := New(mockDb).List(authModel.StoreUser(context.Background(), authModel.NewUser(3, "vibioh")), tc.args.page, tc.args.pageSize, tc.args.sortKey, tc.args.sortAsc)
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
				t.Errorf("List() = (%+v, %d, `%s`), want (%+v, %d, `%s`)", got, gotCount, gotErr, tc.want, tc.wantCount, tc.wantErr)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("sqlmock unfilled expectations: %s", err)
			}
		})
	}
}

func TestGetByRepositoryID(t *testing.T) {
	type args struct {
		id uint64
	}

	var cases = []struct {
		intention string
		args      args
		want      model.Ketchup
		wantErr   error
	}{
		{
			"simple",
			args{
				id: 1,
			},
			model.Ketchup{
				Version: "0.9.0",
				Repository: model.Repository{
					ID: 1,
				},
				User: model.User{
					ID: 3,
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

			mock.ExpectQuery("SELECT version, repository_id, user_id FROM ketchup").WithArgs(1, 3).WillReturnRows(sqlmock.NewRows([]string{"email", "repository_id", "user_id"}).AddRow("0.9.0", 1, 3))

			got, gotErr := New(mockDb).GetByRepositoryID(authModel.StoreUser(context.Background(), authModel.NewUser(3, "vibioh")), tc.args.id)

			failed := false

			if tc.wantErr == nil && gotErr != nil {
				failed = true
			} else if tc.wantErr != nil && !errors.Is(gotErr, tc.wantErr) {
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
					Version: "0.9.0",
					Repository: model.Repository{
						ID: 1,
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
			mock.ExpectQuery("INSERT INTO ketchup").WithArgs("0.9.0", 1, 3).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
			mock.ExpectCommit()

			got, gotErr := New(mockDb).Create(authModel.StoreUser(context.Background(), authModel.NewUser(3, "vibioh")), tc.args.o)

			failed := false

			if tc.wantErr == nil && gotErr != nil {
				failed = true
			} else if tc.wantErr != nil && !errors.Is(gotErr, tc.wantErr) {
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
					Version: "0.9.0",
					Repository: model.Repository{
						ID: 1,
					},
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
			mock.ExpectExec("UPDATE ketchup SET version").WithArgs(1, 3, "0.9.0").WillReturnResult(sqlmock.NewResult(0, 1))
			mock.ExpectCommit()

			gotErr := New(mockDb).Update(authModel.StoreUser(context.Background(), authModel.NewUser(3, "vibioh")), tc.args.o)

			failed := false

			if tc.wantErr == nil && gotErr != nil {
				failed = true
			} else if tc.wantErr != nil && !errors.Is(gotErr, tc.wantErr) {
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
					Repository: model.Repository{
						ID: 1,
					},
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
			mock.ExpectExec("DELETE FROM ketchup").WithArgs(1, 3).WillReturnResult(sqlmock.NewResult(0, 1))
			mock.ExpectCommit()

			gotErr := New(mockDb).Delete(authModel.StoreUser(context.Background(), authModel.NewUser(3, "vibioh")), tc.args.o)

			failed := false

			if tc.wantErr == nil && gotErr != nil {
				failed = true
			} else if tc.wantErr != nil && !errors.Is(gotErr, tc.wantErr) {
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
