package repository

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/ViBiOh/ketchup/pkg/mocks"
	"github.com/ViBiOh/ketchup/pkg/model"
	"github.com/golang/mock/gomock"
	"github.com/jackc/pgx/v4"
)

var (
	ketchupRepository = "vibioh/ketchup"
	viwsRepository    = "vibioh/viws"
	chartRepository   = "https://charts.vibioh.fr"
)

func TestList(t *testing.T) {
	type args struct {
		pageSize uint
		last     string
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
				pageSize: 20,
			},
			[]model.Repository{
				model.NewGithubRepository(1, ketchupRepository).AddVersion(model.DefaultPattern, "1.0.0"),
				model.NewGithubRepository(2, viwsRepository).AddVersion(model.DefaultPattern, "1.2.3"),
			},
			2,
			nil,
		},
		{
			"error",
			args{
				pageSize: 20,
			},
			[]model.Repository{},
			0,
			errors.New("timeout"),
		},
		{
			"invalid kind",
			args{
				pageSize: 20,
			},
			[]model.Repository{},
			0,
			errors.New("invalid value `wrong` for repository kind"),
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockDatabase := mocks.NewDatabase(ctrl)

			instance := App{db: mockDatabase}

			switch tc.intention {
			case "success":
				mockRows := mocks.NewRows(ctrl)
				mockRows.EXPECT().Scan(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(pointers ...interface{}) error {
					*pointers[0].(*uint64) = 1
					*pointers[1].(*string) = "github"
					*pointers[2].(*string) = ketchupRepository
					*pointers[3].(*string) = ""
					*pointers[4].(*uint64) = 2

					return nil
				})
				mockRows.EXPECT().Scan(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(pointers ...interface{}) error {
					*pointers[0].(*uint64) = 2
					*pointers[1].(*string) = "github"
					*pointers[2].(*string) = viwsRepository
					*pointers[3].(*string) = ""
					*pointers[4].(*uint64) = 2

					return nil
				})
				dummyFn := func(_ context.Context, scanner func(pgx.Rows) error, _ string, _ ...interface{}) error {
					if err := scanner(mockRows); err != nil {
						return err
					}
					return scanner(mockRows)
				}
				mockDatabase.EXPECT().List(gomock.Any(), gomock.Any(), gomock.Any(), uint(20)).DoAndReturn(dummyFn)

				enrichRows := mocks.NewRows(ctrl)
				enrichRows.EXPECT().Scan(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(pointers ...interface{}) error {
					*pointers[0].(*uint64) = 1
					*pointers[1].(*string) = model.DefaultPattern
					*pointers[2].(*string) = "1.0.0"

					return nil
				})
				enrichRows.EXPECT().Scan(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(pointers ...interface{}) error {
					*pointers[0].(*uint64) = 2
					*pointers[1].(*string) = model.DefaultPattern
					*pointers[2].(*string) = "1.2.3"

					return nil
				})
				enrichFn := func(_ context.Context, scanner func(pgx.Rows) error, _ string, _ ...interface{}) error {
					if err := scanner(enrichRows); err != nil {
						return err
					}
					return scanner(enrichRows)
				}
				mockDatabase.EXPECT().List(gomock.Any(), gomock.Any(), gomock.Any(), []uint64{1, 2}).DoAndReturn(enrichFn)
			case "error":
				mockDatabase.EXPECT().List(gomock.Any(), gomock.Any(), gomock.Any(), uint(20)).Return(errors.New("timeout"))
			case "invalid kind":
				mockRows := mocks.NewRows(ctrl)
				mockRows.EXPECT().Scan(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(pointers ...interface{}) error {
					*pointers[0].(*uint64) = 1
					*pointers[1].(*string) = "wrong"
					*pointers[2].(*string) = ketchupRepository
					*pointers[3].(*string) = ""
					*pointers[4].(*uint64) = 2

					return nil
				})
				dummyFn := func(_ context.Context, scanner func(pgx.Rows) error, _ string, _ ...interface{}) error {
					if err := scanner(mockRows); err != nil {
						return err
					}
					return scanner(mockRows)
				}
				mockDatabase.EXPECT().List(gomock.Any(), gomock.Any(), gomock.Any(), uint(20)).DoAndReturn(dummyFn)
			}

			got, gotCount, gotErr := instance.List(context.Background(), tc.args.pageSize, tc.args.last)
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
				model.NewGithubRepository(1, ketchupRepository).AddVersion(model.DefaultPattern, "1.0.0"),
				model.NewHelmRepository(2, chartRepository, "app").AddVersion(model.DefaultPattern, "1.2.3"),
			},
			nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockDatabase := mocks.NewDatabase(ctrl)

			instance := App{db: mockDatabase}

			switch tc.intention {
			case "simple":
				mockRows := mocks.NewRows(ctrl)
				mockRows.EXPECT().Scan(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(pointers ...interface{}) error {
					*pointers[0].(*uint64) = 1
					*pointers[1].(*string) = "github"
					*pointers[2].(*string) = ketchupRepository
					*pointers[3].(*string) = ""
					*pointers[4].(*uint64) = 2

					return nil
				})
				mockRows.EXPECT().Scan(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(pointers ...interface{}) error {
					*pointers[0].(*uint64) = 2
					*pointers[1].(*string) = "helm"
					*pointers[2].(*string) = chartRepository
					*pointers[3].(*string) = "app"
					*pointers[4].(*uint64) = 2

					return nil
				})
				dummyFn := func(_ context.Context, scanner func(pgx.Rows) error, _ string, _ ...interface{}) error {
					if err := scanner(mockRows); err != nil {
						return err
					}
					return scanner(mockRows)
				}
				mockDatabase.EXPECT().List(gomock.Any(), gomock.Any(), gomock.Any(), uint64(2), []uint64{8000}).DoAndReturn(dummyFn)

				enrichRows := mocks.NewRows(ctrl)
				enrichRows.EXPECT().Scan(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(pointers ...interface{}) error {
					*pointers[0].(*uint64) = 1
					*pointers[1].(*string) = model.DefaultPattern
					*pointers[2].(*string) = "1.0.0"

					return nil
				})
				enrichRows.EXPECT().Scan(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(pointers ...interface{}) error {
					*pointers[0].(*uint64) = 2
					*pointers[1].(*string) = model.DefaultPattern
					*pointers[2].(*string) = "1.2.3"

					return nil
				})
				enrichFn := func(_ context.Context, scanner func(pgx.Rows) error, _ string, _ ...interface{}) error {
					if err := scanner(enrichRows); err != nil {
						return err
					}
					return scanner(enrichRows)
				}
				mockDatabase.EXPECT().List(gomock.Any(), gomock.Any(), gomock.Any(), []uint64{1, 2}).DoAndReturn(enrichFn)
			}

			got, gotErr := instance.Suggest(context.Background(), tc.args.ignoreIds, tc.args.count)
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
			"SELECT id, kind, name, part FROM ketchup.repository WHERE id =",
			model.NewHelmRepository(1, chartRepository, "app").AddVersion(model.DefaultPattern, "1.0.0"),
			nil,
		},
		{
			"no rows",
			args{
				id:        1,
				forUpdate: true,
			},
			"SELECT id, kind, name, part FROM ketchup.repository WHERE id =",
			model.Repository{},
			nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockDatabase := mocks.NewDatabase(ctrl)

			instance := App{db: mockDatabase}

			switch tc.intention {
			case "simple":
				mockRow := mocks.NewRow(ctrl)
				mockRow.EXPECT().Scan(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(pointers ...interface{}) error {
					*pointers[0].(*uint64) = 1
					*pointers[1].(*string) = "helm"
					*pointers[2].(*string) = chartRepository
					*pointers[3].(*string) = "app"

					return nil
				})
				dummyFn := func(_ context.Context, scanner func(pgx.Row) error, _ string, _ ...interface{}) error {
					return scanner(mockRow)
				}
				mockDatabase.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any(), uint64(1)).DoAndReturn(dummyFn)

				enrichRows := mocks.NewRows(ctrl)
				enrichRows.EXPECT().Scan(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(pointers ...interface{}) error {
					*pointers[0].(*uint64) = 1
					*pointers[1].(*string) = model.DefaultPattern
					*pointers[2].(*string) = "1.0.0"

					return nil
				})
				enrichFn := func(_ context.Context, scanner func(pgx.Rows) error, _ string, _ ...interface{}) error {
					return scanner(enrichRows)
				}
				mockDatabase.EXPECT().List(gomock.Any(), gomock.Any(), gomock.Any(), []uint64{1}).DoAndReturn(enrichFn)
			case "no rows":
				mockRow := mocks.NewRow(ctrl)
				mockRow.EXPECT().Scan(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(pointers ...interface{}) error {
					return pgx.ErrNoRows
				})
				dummyFn := func(_ context.Context, scanner func(pgx.Row) error, _ string, _ ...interface{}) error {
					return scanner(mockRow)
				}
				mockDatabase.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any(), uint64(1)).DoAndReturn(dummyFn)
			}

			got, gotErr := instance.Get(context.Background(), tc.args.id, tc.args.forUpdate)

			failed := false

			if !errors.Is(gotErr, tc.wantErr) {
				failed = true
			} else if !reflect.DeepEqual(got, tc.want) {
				failed = true
			}

			if failed {
				t.Errorf("Get() = (%+v, `%s`), want (%+v, `%s`)", got, gotErr, tc.want, tc.wantErr)
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
				o: model.NewGithubRepository(0, ketchupRepository).AddVersion(model.DefaultPattern, "1.0.0"),
			},
			0,
			errors.New("unable to obtain lock"),
		},
		{
			"error get",
			args{
				o: model.NewGithubRepository(0, ketchupRepository).AddVersion(model.DefaultPattern, "1.0.0"),
			},
			0,
			errors.New("unable to read"),
		},
		{
			"found get",
			args{
				o: model.NewGithubRepository(0, ketchupRepository).AddVersion(model.DefaultPattern, "1.0.0"),
			},
			0,
			errors.New("repository already exists with name"),
		},
		{
			"error create",
			args{
				o: model.NewGithubRepository(0, ketchupRepository).AddVersion(model.DefaultPattern, "1.0.0"),
			},
			0,
			errors.New("timeout"),
		},
		{
			"success",
			args{
				o: model.NewGithubRepository(0, ketchupRepository).AddVersion(model.DefaultPattern, "1.0.0"),
			},
			1,
			nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockDatabase := mocks.NewDatabase(ctrl)

			instance := App{db: mockDatabase}

			switch tc.intention {
			case "error lock":
				mockDatabase.EXPECT().Exec(gomock.Any(), gomock.Any()).Return(errors.New("unable to obtain lock"))
			case "error get":
				mockDatabase.EXPECT().Exec(gomock.Any(), gomock.Any()).Return(nil)
				mockDatabase.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any(), "github", ketchupRepository, "").Return(errors.New("unable to read"))
			case "found get":
				mockDatabase.EXPECT().Exec(gomock.Any(), gomock.Any()).Return(nil)

				mockRow := mocks.NewRow(ctrl)
				mockRow.EXPECT().Scan(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(pointers ...interface{}) error {
					*pointers[0].(*uint64) = 1
					*pointers[1].(*string) = "github"
					*pointers[2].(*string) = ketchupRepository
					*pointers[3].(*string) = ""

					return nil
				})
				dummyFn := func(_ context.Context, scanner func(pgx.Row) error, _ string, _ ...interface{}) error {
					return scanner(mockRow)
				}
				mockDatabase.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any(), "github", ketchupRepository, "").DoAndReturn(dummyFn)

				enrichRows := mocks.NewRows(ctrl)
				enrichRows.EXPECT().Scan(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(pointers ...interface{}) error {
					*pointers[0].(*uint64) = 1
					*pointers[1].(*string) = model.DefaultPattern
					*pointers[2].(*string) = "1.0.0"

					return nil
				})
				enrichFn := func(_ context.Context, scanner func(pgx.Rows) error, _ string, _ ...interface{}) error {
					return scanner(enrichRows)
				}
				mockDatabase.EXPECT().List(gomock.Any(), gomock.Any(), gomock.Any(), []uint64{1}).DoAndReturn(enrichFn)
			case "error create":
				mockDatabase.EXPECT().Exec(gomock.Any(), gomock.Any()).Return(nil)
				mockDatabase.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any(), "github", ketchupRepository, "").Return(nil)
				mockDatabase.EXPECT().Create(gomock.Any(), gomock.Any(), "github", ketchupRepository, "").Return(uint64(0), errors.New("timeout"))
			case "success":
				mockDatabase.EXPECT().Exec(gomock.Any(), gomock.Any()).Return(nil)
				mockDatabase.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any(), "github", ketchupRepository, "").Return(nil)
				mockDatabase.EXPECT().Create(gomock.Any(), gomock.Any(), "github", ketchupRepository, "").Return(uint64(1), nil)
				mockDatabase.EXPECT().List(gomock.Any(), gomock.Any(), gomock.Any(), uint64(1)).Return(nil)
				mockDatabase.EXPECT().Exec(gomock.Any(), gomock.Any(), uint64(1), model.DefaultPattern, "1.0.0").Return(nil)
			}

			got, gotErr := instance.Create(context.Background(), tc.args.o)

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
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockDatabase := mocks.NewDatabase(ctrl)

			instance := App{db: mockDatabase}

			switch tc.intention {
			case "simple":
				mockDatabase.EXPECT().Exec(gomock.Any(), gomock.Any()).Return(nil)
			}

			gotErr := instance.DeleteUnused(context.Background())

			failed := false

			if !errors.Is(gotErr, tc.wantErr) {
				failed = true
			}

			if failed {
				t.Errorf("DeleteUnused() = `%s`, want `%s`", gotErr, tc.wantErr)
			}
		})
	}
}

func TestDeleteUnusedVersions(t *testing.T) {
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
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockDatabase := mocks.NewDatabase(ctrl)

			instance := App{db: mockDatabase}

			switch tc.intention {
			case "simple":
				mockDatabase.EXPECT().Exec(gomock.Any(), gomock.Any()).Return(nil)
			}

			gotErr := instance.DeleteUnusedVersions(context.Background())

			failed := false

			if !errors.Is(gotErr, tc.wantErr) {
				failed = true
			}

			if failed {
				t.Errorf("DeleteUnusedVersions() = `%s`, want `%s`", gotErr, tc.wantErr)
			}
		})
	}
}
