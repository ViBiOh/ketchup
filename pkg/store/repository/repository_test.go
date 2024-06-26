package repository

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/ViBiOh/ketchup/pkg/mocks"
	"github.com/ViBiOh/ketchup/pkg/model"
	"github.com/jackc/pgx/v5"
	"go.uber.org/mock/gomock"
)

var (
	ketchupRepository = "vibioh/ketchup"
	viwsRepository    = "vibioh/viws"
	chartRepository   = "https://charts.vibioh.fr"
)

func TestList(t *testing.T) {
	t.Parallel()

	type args struct {
		pageSize uint
		last     string
	}

	cases := map[string]struct {
		args      args
		want      []model.Repository
		wantCount uint64
		wantErr   error
	}{
		"success": {
			args{
				pageSize: 20,
			},
			[]model.Repository{
				model.NewGithubRepository(model.Identifier(1), ketchupRepository).AddVersion(model.DefaultPattern, "1.0.0"),
				model.NewGithubRepository(model.Identifier(2), viwsRepository).AddVersion(model.DefaultPattern, "1.2.3"),
			},
			2,
			nil,
		},
		"invalid last": {
			args{
				pageSize: 20,
				last:     "abc",
			},
			nil,
			0,
			errors.New("invalid last key"),
		},
		"last": {
			args{
				pageSize: 20,
				last:     "2",
			},
			nil,
			0,
			nil,
		},
		"error": {
			args{
				pageSize: 20,
			},
			nil,
			0,
			errors.New("timeout"),
		},
		"scan error": {
			args{
				pageSize: 20,
			},
			nil,
			0,
			errors.New("read int"),
		},
		"invalid kind": {
			args{
				pageSize: 20,
			},
			nil,
			0,
			model.ErrUnknownRepositoryKind,
		},
	}

	for intention, testCase := range cases {
		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			mockDatabase := mocks.NewDatabase(ctrl)

			instance := Service{db: mockDatabase}

			switch intention {
			case "success":
				mockRows := mocks.NewRows(ctrl)
				mockRows.EXPECT().Scan(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(pointers ...any) error {
					*pointers[0].(*model.Identifier) = model.Identifier(1)
					*pointers[1].(*string) = "github"
					*pointers[2].(*string) = ketchupRepository
					*pointers[3].(*string) = ""

					return nil
				})
				mockRows.EXPECT().Scan(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(pointers ...any) error {
					*pointers[0].(*model.Identifier) = model.Identifier(2)
					*pointers[1].(*string) = "github"
					*pointers[2].(*string) = viwsRepository
					*pointers[3].(*string) = ""

					return nil
				})
				dummyFn := func(_ context.Context, scanner func(pgx.Rows) error, _ string, _ ...any) error {
					if err := scanner(mockRows); err != nil {
						return err
					}
					return scanner(mockRows)
				}
				mockDatabase.EXPECT().List(gomock.Any(), gomock.Any(), gomock.Any(), uint(20)).DoAndReturn(dummyFn)

				enrichRows := mocks.NewRows(ctrl)
				enrichRows.EXPECT().Scan(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(pointers ...any) error {
					*pointers[0].(*model.Identifier) = model.Identifier(1)
					*pointers[1].(*string) = model.DefaultPattern
					*pointers[2].(*string) = "1.0.0"

					return nil
				})
				enrichRows.EXPECT().Scan(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(pointers ...any) error {
					*pointers[0].(*model.Identifier) = model.Identifier(2)
					*pointers[1].(*string) = model.DefaultPattern
					*pointers[2].(*string) = "1.2.3"

					return nil
				})
				enrichFn := func(_ context.Context, scanner func(pgx.Rows) error, _ string, _ ...any) error {
					if err := scanner(enrichRows); err != nil {
						return err
					}
					return scanner(enrichRows)
				}
				mockDatabase.EXPECT().List(gomock.Any(), gomock.Any(), gomock.Any(), []model.Identifier{1, 2}).DoAndReturn(enrichFn)

			case "last":
				mockDatabase.EXPECT().List(gomock.Any(), gomock.Any(), gomock.Any(), uint64(2), uint(20)).Return(nil)

			case "error":
				mockDatabase.EXPECT().List(gomock.Any(), gomock.Any(), gomock.Any(), uint(20)).Return(errors.New("timeout"))

			case "scan error":
				mockRows := mocks.NewRows(ctrl)
				mockRows.EXPECT().Scan(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(pointers ...any) error {
					return errors.New("read int")
				})
				dummyFn := func(_ context.Context, scanner func(pgx.Rows) error, _ string, _ ...any) error {
					if err := scanner(mockRows); err != nil {
						return err
					}
					return scanner(mockRows)
				}
				mockDatabase.EXPECT().List(gomock.Any(), gomock.Any(), gomock.Any(), uint(20)).DoAndReturn(dummyFn)

			case "invalid kind":
				mockRows := mocks.NewRows(ctrl)
				mockRows.EXPECT().Scan(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(pointers ...any) error {
					*pointers[0].(*model.Identifier) = model.Identifier(1)
					*pointers[1].(*string) = "wrong"
					*pointers[2].(*string) = ketchupRepository
					*pointers[3].(*string) = ""

					return nil
				})
				dummyFn := func(_ context.Context, scanner func(pgx.Rows) error, _ string, _ ...any) error {
					if err := scanner(mockRows); err != nil {
						return err
					}
					return scanner(mockRows)
				}
				mockDatabase.EXPECT().List(gomock.Any(), gomock.Any(), gomock.Any(), uint(20)).DoAndReturn(dummyFn)
			}

			got, gotErr := instance.List(context.TODO(), testCase.args.pageSize, testCase.args.last)
			failed := false

			if testCase.wantErr == nil && gotErr != nil {
				failed = true
			} else if testCase.wantErr != nil && gotErr == nil {
				failed = true
			} else if testCase.wantErr != nil && !strings.Contains(gotErr.Error(), testCase.wantErr.Error()) {
				failed = true
			} else if !reflect.DeepEqual(got, testCase.want) {
				failed = true
			}

			if failed {
				t.Errorf("List() = (%#v, `%s`), want (%#v, `%s`)", got, gotErr, testCase.want, testCase.wantErr)
			}
		})
	}
}

func TestSuggest(t *testing.T) {
	t.Parallel()

	type args struct {
		ignoreIds []model.Identifier
		count     uint64
	}

	cases := map[string]struct {
		args    args
		want    []model.Repository
		wantErr error
	}{
		"simple": {
			args{
				ignoreIds: []model.Identifier{8000},
				count:     2,
			},
			[]model.Repository{
				model.NewGithubRepository(model.Identifier(1), ketchupRepository).AddVersion(model.DefaultPattern, "1.0.0"),
				model.NewHelmRepository(model.Identifier(2), chartRepository, "app").AddVersion(model.DefaultPattern, "1.2.3"),
			},
			nil,
		},
	}

	for intention, testCase := range cases {
		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			mockDatabase := mocks.NewDatabase(ctrl)

			instance := Service{db: mockDatabase}

			switch intention {
			case "simple":
				mockRows := mocks.NewRows(ctrl)
				mockRows.EXPECT().Scan(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(pointers ...any) error {
					*pointers[0].(*model.Identifier) = model.Identifier(1)
					*pointers[1].(*string) = "github"
					*pointers[2].(*string) = ketchupRepository
					*pointers[3].(*string) = ""

					return nil
				})
				mockRows.EXPECT().Scan(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(pointers ...any) error {
					*pointers[0].(*model.Identifier) = model.Identifier(2)
					*pointers[1].(*string) = "helm"
					*pointers[2].(*string) = chartRepository
					*pointers[3].(*string) = "app"

					return nil
				})
				dummyFn := func(_ context.Context, scanner func(pgx.Rows) error, _ string, _ ...any) error {
					if err := scanner(mockRows); err != nil {
						return err
					}
					return scanner(mockRows)
				}
				mockDatabase.EXPECT().List(gomock.Any(), gomock.Any(), gomock.Any(), uint64(2), []model.Identifier{8000}).DoAndReturn(dummyFn)

				enrichRows := mocks.NewRows(ctrl)
				enrichRows.EXPECT().Scan(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(pointers ...any) error {
					*pointers[0].(*model.Identifier) = model.Identifier(1)
					*pointers[1].(*string) = model.DefaultPattern
					*pointers[2].(*string) = "1.0.0"

					return nil
				})
				enrichRows.EXPECT().Scan(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(pointers ...any) error {
					*pointers[0].(*model.Identifier) = model.Identifier(2)
					*pointers[1].(*string) = model.DefaultPattern
					*pointers[2].(*string) = "1.2.3"

					return nil
				})
				enrichFn := func(_ context.Context, scanner func(pgx.Rows) error, _ string, _ ...any) error {
					if err := scanner(enrichRows); err != nil {
						return err
					}
					return scanner(enrichRows)
				}
				mockDatabase.EXPECT().List(gomock.Any(), gomock.Any(), gomock.Any(), []model.Identifier{1, 2}).DoAndReturn(enrichFn)
			}

			got, gotErr := instance.Suggest(context.TODO(), testCase.args.ignoreIds, testCase.args.count)
			failed := false

			if testCase.wantErr == nil && gotErr != nil {
				failed = true
			} else if testCase.wantErr != nil && gotErr == nil {
				failed = true
			} else if testCase.wantErr != nil && !strings.Contains(gotErr.Error(), testCase.wantErr.Error()) {
				failed = true
			} else if !reflect.DeepEqual(got, testCase.want) {
				failed = true
			}

			if failed {
				t.Errorf("Suggest() = (%#v, `%s`), want (%#v, `%s`)", got, gotErr, testCase.want, testCase.wantErr)
			}
		})
	}
}

func TestGet(t *testing.T) {
	t.Parallel()

	type args struct {
		id        model.Identifier
		forUpdate bool
	}

	cases := map[string]struct {
		args      args
		expectSQL string
		want      model.Repository
		wantErr   error
	}{
		"simple": {
			args{
				id: 1,
			},
			"SELECT id, kind, name, part FROM ketchup.repository WHERE id =",
			model.NewHelmRepository(model.Identifier(1), chartRepository, "app").AddVersion(model.DefaultPattern, "1.0.0"),
			nil,
		},
		"no rows": {
			args{
				id:        1,
				forUpdate: true,
			},
			"SELECT id, kind, name, part FROM ketchup.repository WHERE id =",
			model.NewEmptyRepository(),
			nil,
		},
		"scan error": {
			args{
				id:        1,
				forUpdate: true,
			},
			"SELECT id, kind, name, part FROM ketchup.repository WHERE id =",
			model.NewEmptyRepository(),
			errors.New("read int"),
		},
	}

	for intention, testCase := range cases {
		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			mockDatabase := mocks.NewDatabase(ctrl)

			instance := Service{db: mockDatabase}

			switch intention {
			case "simple":
				mockRow := mocks.NewRow(ctrl)
				mockRow.EXPECT().Scan(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(pointers ...any) error {
					*pointers[0].(*model.Identifier) = model.Identifier(1)
					*pointers[1].(*string) = "helm"
					*pointers[2].(*string) = chartRepository
					*pointers[3].(*string) = "app"

					return nil
				})
				dummyFn := func(_ context.Context, scanner func(pgx.Row) error, _ string, _ ...any) error {
					return scanner(mockRow)
				}
				mockDatabase.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any(), model.Identifier(1)).DoAndReturn(dummyFn)

				enrichRows := mocks.NewRows(ctrl)
				enrichRows.EXPECT().Scan(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(pointers ...any) error {
					*pointers[0].(*model.Identifier) = model.Identifier(1)
					*pointers[1].(*string) = model.DefaultPattern
					*pointers[2].(*string) = "1.0.0"

					return nil
				})
				enrichFn := func(_ context.Context, scanner func(pgx.Rows) error, _ string, _ ...any) error {
					return scanner(enrichRows)
				}
				mockDatabase.EXPECT().List(gomock.Any(), gomock.Any(), gomock.Any(), []model.Identifier{1}).DoAndReturn(enrichFn)
			case "no rows":
				mockRow := mocks.NewRow(ctrl)
				mockRow.EXPECT().Scan(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(pointers ...any) error {
					return pgx.ErrNoRows
				})
				dummyFn := func(_ context.Context, scanner func(pgx.Row) error, _ string, _ ...any) error {
					return scanner(mockRow)
				}
				mockDatabase.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any(), model.Identifier(1)).DoAndReturn(dummyFn)
			case "scan error":
				mockRow := mocks.NewRow(ctrl)
				mockRow.EXPECT().Scan(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(pointers ...any) error {
					return errors.New("read int")
				})
				dummyFn := func(_ context.Context, scanner func(pgx.Row) error, _ string, _ ...any) error {
					return scanner(mockRow)
				}
				mockDatabase.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any(), model.Identifier(1)).DoAndReturn(dummyFn)
			}

			got, gotErr := instance.Get(context.TODO(), testCase.args.id, testCase.args.forUpdate)

			failed := false

			if testCase.wantErr == nil && gotErr != nil {
				failed = true
			} else if testCase.wantErr != nil && gotErr == nil {
				failed = true
			} else if testCase.wantErr != nil && !strings.Contains(gotErr.Error(), testCase.wantErr.Error()) {
				failed = true
			} else if !reflect.DeepEqual(got, testCase.want) {
				failed = true
			}

			if failed {
				t.Errorf("Get() = (%#v, `%s`), want (%#v, `%s`)", got, gotErr, testCase.want, testCase.wantErr)
			}
		})
	}
}

func TestCreate(t *testing.T) {
	t.Parallel()

	type args struct {
		o model.Repository
	}

	cases := map[string]struct {
		args    args
		want    model.Identifier
		wantErr error
	}{
		"error lock": {
			args{
				o: model.NewGithubRepository(model.Identifier(0), ketchupRepository).AddVersion(model.DefaultPattern, "1.0.0"),
			},
			0,
			errors.New("obtain lock"),
		},
		"error get": {
			args{
				o: model.NewGithubRepository(model.Identifier(0), ketchupRepository).AddVersion(model.DefaultPattern, "1.0.0"),
			},
			0,
			errors.New("read"),
		},
		"found get": {
			args{
				o: model.NewGithubRepository(model.Identifier(0), ketchupRepository).AddVersion(model.DefaultPattern, "1.0.0"),
			},
			0,
			errors.New("repository already exists with name"),
		},
		"error create": {
			args{
				o: model.NewGithubRepository(model.Identifier(0), ketchupRepository).AddVersion(model.DefaultPattern, "1.0.0"),
			},
			0,
			errors.New("timeout"),
		},
		"success": {
			args{
				o: model.NewGithubRepository(model.Identifier(0), ketchupRepository).AddVersion(model.DefaultPattern, "1.0.0"),
			},
			1,
			nil,
		},
	}

	for intention, testCase := range cases {
		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			mockDatabase := mocks.NewDatabase(ctrl)

			instance := Service{db: mockDatabase}

			switch intention {
			case "error lock":
				mockDatabase.EXPECT().Exec(gomock.Any(), gomock.Any()).Return(errors.New("obtain lock"))
			case "error get":
				mockDatabase.EXPECT().Exec(gomock.Any(), gomock.Any()).Return(nil)
				mockDatabase.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any(), "github", ketchupRepository, "").Return(errors.New("read"))
			case "found get":
				mockDatabase.EXPECT().Exec(gomock.Any(), gomock.Any()).Return(nil)

				mockRow := mocks.NewRow(ctrl)
				mockRow.EXPECT().Scan(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(pointers ...any) error {
					*pointers[0].(*model.Identifier) = model.Identifier(1)
					*pointers[1].(*string) = "github"
					*pointers[2].(*string) = ketchupRepository
					*pointers[3].(*string) = ""

					return nil
				})
				dummyFn := func(_ context.Context, scanner func(pgx.Row) error, _ string, _ ...any) error {
					return scanner(mockRow)
				}
				mockDatabase.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any(), "github", ketchupRepository, "").DoAndReturn(dummyFn)

				enrichRows := mocks.NewRows(ctrl)
				enrichRows.EXPECT().Scan(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(pointers ...any) error {
					*pointers[0].(*model.Identifier) = model.Identifier(1)
					*pointers[1].(*string) = model.DefaultPattern
					*pointers[2].(*string) = "1.0.0"

					return nil
				})
				enrichFn := func(_ context.Context, scanner func(pgx.Rows) error, _ string, _ ...any) error {
					return scanner(enrichRows)
				}
				mockDatabase.EXPECT().List(gomock.Any(), gomock.Any(), gomock.Any(), []model.Identifier{1}).DoAndReturn(enrichFn)
			case "error create":
				mockDatabase.EXPECT().Exec(gomock.Any(), gomock.Any()).Return(nil)
				mockDatabase.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any(), "github", ketchupRepository, "").Return(nil)
				mockDatabase.EXPECT().Create(gomock.Any(), gomock.Any(), "github", ketchupRepository, "").Return(uint64(0), errors.New("timeout"))
			case "success":
				mockDatabase.EXPECT().Exec(gomock.Any(), gomock.Any()).Return(nil)
				mockDatabase.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any(), "github", ketchupRepository, "").Return(nil)
				mockDatabase.EXPECT().Create(gomock.Any(), gomock.Any(), "github", ketchupRepository, "").Return(uint64(1), nil)
				mockDatabase.EXPECT().List(gomock.Any(), gomock.Any(), gomock.Any(), model.Identifier(1)).Return(nil)
				mockDatabase.EXPECT().One(gomock.Any(), gomock.Any(), model.Identifier(1), model.DefaultPattern, "1.0.0").Return(nil)
			}

			got, gotErr := instance.Create(context.TODO(), testCase.args.o)

			failed := false

			if testCase.wantErr == nil && gotErr != nil {
				failed = true
			} else if testCase.wantErr != nil && gotErr == nil {
				failed = true
			} else if testCase.wantErr != nil && !strings.Contains(gotErr.Error(), testCase.wantErr.Error()) {
				failed = true
			} else if got != testCase.want {
				failed = true
			}

			if failed {
				t.Errorf("Create() = (%d, `%s`), want (%d, `%s`)", got, gotErr, testCase.want, testCase.wantErr)
			}
		})
	}
}

func TestDeleteUnused(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		wantErr error
	}{
		"simple": {
			nil,
		},
	}

	for intention, testCase := range cases {
		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			mockDatabase := mocks.NewDatabase(ctrl)

			instance := Service{db: mockDatabase}

			switch intention {
			case "simple":
				mockDatabase.EXPECT().Exec(gomock.Any(), gomock.Any()).Return(nil)
			}

			gotErr := instance.DeleteUnused(context.TODO())

			failed := false

			if !errors.Is(gotErr, testCase.wantErr) {
				failed = true
			}

			if failed {
				t.Errorf("DeleteUnused() = `%s`, want `%s`", gotErr, testCase.wantErr)
			}
		})
	}
}

func TestDeleteUnusedVersions(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		wantErr error
	}{
		"simple": {
			nil,
		},
	}

	for intention, testCase := range cases {
		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			mockDatabase := mocks.NewDatabase(ctrl)

			instance := Service{db: mockDatabase}

			switch intention {
			case "simple":
				mockDatabase.EXPECT().Exec(gomock.Any(), gomock.Any()).Return(nil)
			}

			gotErr := instance.DeleteUnusedVersions(context.TODO())

			failed := false

			if !errors.Is(gotErr, testCase.wantErr) {
				failed = true
			}

			if failed {
				t.Errorf("DeleteUnusedVersions() = `%s`, want `%s`", gotErr, testCase.wantErr)
			}
		})
	}
}
