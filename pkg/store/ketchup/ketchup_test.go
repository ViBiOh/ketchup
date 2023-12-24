package ketchup

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	authModel "github.com/ViBiOh/auth/v2/pkg/model"
	"github.com/ViBiOh/ketchup/pkg/mocks"
	"github.com/ViBiOh/ketchup/pkg/model"
	"github.com/jackc/pgx/v5"
	"go.uber.org/mock/gomock"
)

var (
	testCtx = model.StoreUser(context.TODO(), model.NewUser(3, testEmail, authModel.NewUser(0, "")))

	testEmail       = "nobody@localhost"
	repositoryName  = "vibioh/ketchup"
	chartRepository = "https://charts.vibioh.fr"

	repositoryVersion = "1.0.0"
)

func TestList(t *testing.T) {
	t.Parallel()

	type args struct {
		pageSize uint
	}

	cases := map[string]struct {
		args    args
		want    []model.Ketchup
		wantErr error
	}{
		"simple": {
			args{
				pageSize: 20,
			},
			[]model.Ketchup{
				{
					ID:         "cad120aa",
					Pattern:    model.DefaultPattern,
					Version:    "0.9.0",
					Frequency:  model.Daily,
					Repository: model.NewGithubRepository(model.Identifier(1), repositoryName).AddVersion(model.DefaultPattern, repositoryVersion),
					User:       model.NewUser(3, testEmail, authModel.NewUser(0, "")),
				},
				{
					ID:         "5ec6147c",
					Pattern:    model.DefaultPattern,
					Version:    repositoryVersion,
					Frequency:  model.Daily,
					Repository: model.NewHelmRepository(model.Identifier(2), chartRepository, "app").AddVersion(model.DefaultPattern, repositoryVersion),
					User:       model.NewUser(3, testEmail, authModel.NewUser(0, "")),
				},
			},
			nil,
		},
		"error": {
			args{
				pageSize: 20,
			},
			nil,
			errors.New("failed"),
		},
		"invalid kind": {
			args{
				pageSize: 20,
			},
			nil,
			errors.New("invalid value `wrong` for repository kind"),
		},
	}

	for intention, testCase := range cases {
		intention, testCase := intention, testCase

		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockDatabase := mocks.NewDatabase(ctrl)

			instance := Service{db: mockDatabase}

			switch intention {
			case "simple":
				mockRows := mocks.NewRows(ctrl)
				mockRows.EXPECT().Scan(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(pointers ...any) error {
					*pointers[0].(*string) = model.DefaultPattern
					*pointers[1].(*string) = "0.9.0"
					*pointers[2].(*string) = "daily"
					*pointers[3].(*bool) = false
					*pointers[4].(*model.Identifier) = model.Identifier(1)
					*pointers[5].(*string) = repositoryName
					*pointers[6].(*string) = ""
					*pointers[7].(*string) = "github"
					*pointers[8].(*string) = repositoryVersion

					return nil
				})
				mockRows.EXPECT().Scan(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(pointers ...any) error {
					*pointers[0].(*string) = model.DefaultPattern
					*pointers[1].(*string) = repositoryVersion
					*pointers[2].(*string) = "daily"
					*pointers[3].(*bool) = false
					*pointers[4].(*model.Identifier) = model.Identifier(2)
					*pointers[5].(*string) = chartRepository
					*pointers[6].(*string) = "app"
					*pointers[7].(*string) = "helm"
					*pointers[8].(*string) = repositoryVersion

					return nil
				})
				dummyFn := func(_ context.Context, scanner func(pgx.Rows) error, _ string, _ ...any) error {
					if err := scanner(mockRows); err != nil {
						return err
					}
					return scanner(mockRows)
				}
				mockDatabase.EXPECT().List(gomock.Any(), gomock.Any(), gomock.Any(), model.Identifier(3), uint(20)).DoAndReturn(dummyFn)
			case "error":
				mockDatabase.EXPECT().List(gomock.Any(), gomock.Any(), gomock.Any(), model.Identifier(3), uint(20)).Return(errors.New("failed"))
			case "invalid kind":
				mockRows := mocks.NewRows(ctrl)
				mockRows.EXPECT().Scan(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(pointers ...any) error {
					*pointers[0].(*string) = model.DefaultPattern
					*pointers[1].(*string) = "0.9.0"
					*pointers[2].(*string) = "daily"
					*pointers[3].(*bool) = false
					*pointers[4].(*model.Identifier) = model.Identifier(1)
					*pointers[5].(*string) = repositoryName
					*pointers[6].(*string) = ""
					*pointers[7].(*string) = "wrong"
					*pointers[8].(*string) = repositoryVersion

					return nil
				})
				dummyFn := func(_ context.Context, scanner func(pgx.Rows) error, _ string, _ ...any) error {
					if err := scanner(mockRows); err != nil {
						return err
					}
					return scanner(mockRows)
				}
				mockDatabase.EXPECT().List(gomock.Any(), gomock.Any(), gomock.Any(), model.Identifier(3), uint(20)).DoAndReturn(dummyFn)
			}

			got, gotErr := instance.List(testCtx, testCase.args.pageSize, "")
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
				t.Errorf("List() = (%+v, `%s`), want (%+v, `%s`)", got, gotErr, testCase.want, testCase.wantErr)
			}
		})
	}
}

func TestListByRepositoriesIDAndFrequencies(t *testing.T) {
	t.Parallel()

	type args struct {
		ids       []model.Identifier
		frequency model.KetchupFrequency
	}

	cases := map[string]struct {
		args    args
		want    []model.Ketchup
		wantErr error
	}{
		"simple": {
			args{
				ids:       []model.Identifier{1, 2},
				frequency: model.Daily,
			},
			[]model.Ketchup{
				{
					Pattern:    model.DefaultPattern,
					Version:    "0.9.0",
					Frequency:  model.Daily,
					Repository: model.NewGithubRepository(model.Identifier(1), ""),
					User:       model.NewUser(1, testEmail, authModel.NewUser(0, "")),
				},
				{
					Pattern:    model.DefaultPattern,
					Version:    repositoryVersion,
					Frequency:  model.Daily,
					Repository: model.NewGithubRepository(model.Identifier(2), ""),
					User:       model.NewUser(2, "guest@domain", authModel.NewUser(0, "")),
				},
			},
			nil,
		},
		"error": {
			args{
				ids:       []model.Identifier{1, 2},
				frequency: model.Daily,
			},
			nil,
			errors.New("failed"),
		},
	}

	for intention, testCase := range cases {
		intention, testCase := intention, testCase

		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockDatabase := mocks.NewDatabase(ctrl)

			instance := Service{db: mockDatabase}

			switch intention {
			case "simple":
				mockRows := mocks.NewRows(ctrl)
				mockRows.EXPECT().Scan(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(pointers ...any) error {
					*pointers[0].(*string) = model.DefaultPattern
					*pointers[1].(*string) = "0.9.0"
					*pointers[2].(*string) = "daily"
					*pointers[3].(*bool) = false
					*pointers[4].(*model.Identifier) = model.Identifier(1)
					*pointers[5].(*model.Identifier) = model.Identifier(1)
					*pointers[6].(*string) = testEmail

					return nil
				})
				mockRows.EXPECT().Scan(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(pointers ...any) error {
					*pointers[0].(*string) = model.DefaultPattern
					*pointers[1].(*string) = repositoryVersion
					*pointers[2].(*string) = "daily"
					*pointers[3].(*bool) = false
					*pointers[4].(*model.Identifier) = model.Identifier(2)
					*pointers[5].(*model.Identifier) = model.Identifier(2)
					*pointers[6].(*string) = "guest@domain"

					return nil
				})
				dummyFn := func(_ context.Context, scanner func(pgx.Rows) error, _ string, _ ...any) error {
					if err := scanner(mockRows); err != nil {
						return err
					}
					return scanner(mockRows)
				}
				mockDatabase.EXPECT().List(gomock.Any(), gomock.Any(), gomock.Any(), testCase.args.ids, []string{"daily"}).DoAndReturn(dummyFn)
			case "error":
				mockDatabase.EXPECT().List(gomock.Any(), gomock.Any(), gomock.Any(), testCase.args.ids, []string{"daily"}).Return(errors.New("failed"))
			}

			got, gotErr := instance.ListByRepositoriesIDAndFrequencies(testCtx, testCase.args.ids, testCase.args.frequency)
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
				t.Errorf("ListByRepositoriesIDAndFrequencies() = (%+v, `%s`), want (%+v, `%s`)", got, gotErr, testCase.want, testCase.wantErr)
			}
		})
	}
}

func TestGetByRepository(t *testing.T) {
	t.Parallel()

	type args struct {
		id        model.Identifier
		pattern   string
		forUpdate bool
	}

	cases := map[string]struct {
		args    args
		want    model.Ketchup
		wantErr error
	}{
		"simple": {
			args{
				id:      1,
				pattern: "stable",
			},
			model.Ketchup{
				Pattern:    model.DefaultPattern,
				Version:    "0.9.0",
				Frequency:  model.Daily,
				Repository: model.NewGithubRepository(model.Identifier(1), repositoryName),
				User:       model.NewUser(3, testEmail, authModel.NewUser(0, "")),
			},
			nil,
		},
		"no rows": {
			args{
				id:        1,
				pattern:   "stable",
				forUpdate: true,
			},
			model.Ketchup{},
			nil,
		},
	}

	for intention, testCase := range cases {
		intention, testCase := intention, testCase

		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockDatabase := mocks.NewDatabase(ctrl)

			instance := Service{db: mockDatabase}

			switch intention {
			case "simple":
				mockRow := mocks.NewRow(ctrl)
				mockRow.EXPECT().Scan(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(pointers ...any) error {
					*pointers[0].(*string) = model.DefaultPattern
					*pointers[1].(*string) = "0.9.0"
					*pointers[2].(*string) = "daily"
					*pointers[3].(*bool) = false
					*pointers[4].(*model.Identifier) = model.Identifier(1)
					*pointers[5].(*model.Identifier) = model.Identifier(3)
					*pointers[6].(*string) = repositoryName
					*pointers[7].(*string) = ""
					*pointers[8].(*string) = "github"

					return nil
				})
				dummyFn := func(_ context.Context, scanner func(pgx.Row) error, _ string, _ ...any) error {
					return scanner(mockRow)
				}
				mockDatabase.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any(), testCase.args.id, model.Identifier(3), testCase.args.pattern).DoAndReturn(dummyFn)
			case "no rows":
				mockRow := mocks.NewRow(ctrl)
				mockRow.EXPECT().Scan(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(pointers ...any) error {
					return pgx.ErrNoRows
				})
				dummyFn := func(_ context.Context, scanner func(pgx.Row) error, _ string, _ ...any) error {
					return scanner(mockRow)
				}
				mockDatabase.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any(), testCase.args.id, model.Identifier(3), testCase.args.pattern).DoAndReturn(dummyFn)
			case "error":
				mockDatabase.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any(), testCase.args.id, model.Identifier(3), testCase.args.pattern).Return(errors.New("failed"))
			}

			got, gotErr := instance.GetByRepository(testCtx, testCase.args.id, testCase.args.pattern, testCase.args.forUpdate)

			failed := false

			if !errors.Is(gotErr, testCase.wantErr) {
				failed = true
			} else if !reflect.DeepEqual(got, testCase.want) {
				failed = true
			}

			if failed {
				t.Errorf("GetByRepository() = (%+v, `%s`), want (%+v, `%s`)", got, gotErr, testCase.want, testCase.wantErr)
			}
		})
	}
}

func TestCreate(t *testing.T) {
	t.Parallel()

	type args struct {
		o model.Ketchup
	}

	cases := map[string]struct {
		args    args
		want    model.Identifier
		wantErr error
	}{
		"simple": {
			args{
				o: model.Ketchup{
					Pattern:    model.DefaultPattern,
					Version:    "0.9.0",
					Frequency:  model.Daily,
					Repository: model.NewGithubRepository(model.Identifier(1), ""),
				},
			},
			1,
			nil,
		},
	}

	for intention, testCase := range cases {
		intention, testCase := intention, testCase

		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockDatabase := mocks.NewDatabase(ctrl)

			instance := Service{db: mockDatabase}

			switch intention {
			case "simple":
				mockDatabase.EXPECT().Create(gomock.Any(), gomock.Any(), model.DefaultPattern, "0.9.0", "daily", gomock.Any(), model.Identifier(1), model.Identifier(3)).Return(uint64(1), nil)
			}

			got, gotErr := instance.Create(testCtx, testCase.args.o)

			failed := false

			if !errors.Is(gotErr, testCase.wantErr) {
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

func TestUpdate(t *testing.T) {
	t.Parallel()

	type args struct {
		o          model.Ketchup
		oldPattern string
	}

	cases := map[string]struct {
		args    args
		wantErr error
	}{
		"simple": {
			args{
				o: model.Ketchup{
					Pattern:    model.DefaultPattern,
					Version:    "0.9.0",
					Frequency:  model.Daily,
					Repository: model.NewGithubRepository(model.Identifier(1), ""),
				},
				oldPattern: "stable",
			},
			nil,
		},
	}

	for intention, testCase := range cases {
		intention, testCase := intention, testCase

		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockDatabase := mocks.NewDatabase(ctrl)

			instance := Service{db: mockDatabase}

			switch intention {
			case "simple":
				mockDatabase.EXPECT().One(gomock.Any(), gomock.Any(), model.Identifier(1), model.Identifier(3), model.DefaultPattern, model.DefaultPattern, "0.9.0", "daily", gomock.Any()).Return(nil)
			}

			gotErr := instance.Update(testCtx, testCase.args.o, testCase.args.oldPattern)

			failed := false

			if !errors.Is(gotErr, testCase.wantErr) {
				failed = true
			}

			if failed {
				t.Errorf("Update() = `%s`, want `%s`", gotErr, testCase.wantErr)
			}
		})
	}
}

func TestDelete(t *testing.T) {
	t.Parallel()

	type args struct {
		o model.Ketchup
	}

	cases := map[string]struct {
		args    args
		wantErr error
	}{
		"simple": {
			args{
				o: model.Ketchup{
					Pattern:    "stable",
					Repository: model.NewGithubRepository(model.Identifier(1), ""),
				},
			},
			nil,
		},
	}

	for intention, testCase := range cases {
		intention, testCase := intention, testCase

		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockDatabase := mocks.NewDatabase(ctrl)

			instance := Service{db: mockDatabase}

			switch intention {
			case "simple":
				mockDatabase.EXPECT().One(gomock.Any(), gomock.Any(), model.Identifier(1), model.Identifier(3), model.DefaultPattern).Return(nil)
			}

			gotErr := instance.Delete(testCtx, testCase.args.o)

			failed := false

			if !errors.Is(gotErr, testCase.wantErr) {
				failed = true
			}

			if failed {
				t.Errorf("Delete() = `%s`, want `%s`", gotErr, testCase.wantErr)
			}
		})
	}
}
