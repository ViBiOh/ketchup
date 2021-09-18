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
	"github.com/golang/mock/gomock"
	"github.com/jackc/pgx/v4"
)

var (
	testCtx = model.StoreUser(context.Background(), model.NewUser(3, testEmail, authModel.NewUser(0, "")))

	testEmail       = "nobody@localhost"
	repositoryName  = "vibioh/ketchup"
	viwsRepository  = "vibioh/viws"
	chartRepository = "https://charts.vibioh.fr"

	repositoryVersion = "1.0.0"
)

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
					ID:         "4bdfbeb6",
					Pattern:    model.DefaultPattern,
					Version:    "0.9.0",
					Frequency:  model.Daily,
					Repository: model.NewGithubRepository(1, repositoryName).AddVersion(model.DefaultPattern, repositoryVersion),
					User:       model.NewUser(3, testEmail, authModel.NewUser(0, "")),
				},
				{
					ID:         "844123b7",
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
			"error",
			args{
				pageSize: 20,
			},
			[]model.Ketchup{},
			0,
			errors.New("failed"),
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
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockDatabase := mocks.NewDatabase(ctrl)

			instance := App{db: mockDatabase}

			switch tc.intention {
			case "simple":
				mockRows := mocks.NewRows(ctrl)
				mockRows.EXPECT().Scan(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(pointers ...interface{}) error {
					*pointers[0].(*string) = model.DefaultPattern
					*pointers[1].(*string) = "0.9.0"
					*pointers[2].(*string) = "daily"
					*pointers[3].(*uint64) = 1
					*pointers[4].(*string) = repositoryName
					*pointers[5].(*string) = ""
					*pointers[6].(*string) = "github"
					*pointers[7].(*string) = repositoryVersion
					*pointers[8].(*uint64) = 2

					return nil
				})
				mockRows.EXPECT().Scan(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(pointers ...interface{}) error {
					*pointers[0].(*string) = model.DefaultPattern
					*pointers[1].(*string) = repositoryVersion
					*pointers[2].(*string) = "daily"
					*pointers[3].(*uint64) = 2
					*pointers[4].(*string) = chartRepository
					*pointers[5].(*string) = "app"
					*pointers[6].(*string) = "helm"
					*pointers[7].(*string) = repositoryVersion
					*pointers[8].(*uint64) = 2

					return nil
				})
				dummyFn := func(_ context.Context, scanner func(pgx.Rows) error, _ string, _ ...interface{}) error {
					if err := scanner(mockRows); err != nil {
						return err
					}
					return scanner(mockRows)
				}
				mockDatabase.EXPECT().List(gomock.Any(), gomock.Any(), gomock.Any(), uint64(3), uint(20)).DoAndReturn(dummyFn)
			case "error":
				mockDatabase.EXPECT().List(gomock.Any(), gomock.Any(), gomock.Any(), uint64(3), uint(20)).Return(errors.New("failed"))
			case "invalid kind":
				mockRows := mocks.NewRows(ctrl)
				mockRows.EXPECT().Scan(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(pointers ...interface{}) error {
					*pointers[0].(*string) = model.DefaultPattern
					*pointers[1].(*string) = "0.9.0"
					*pointers[2].(*string) = "daily"
					*pointers[3].(*uint64) = 1
					*pointers[4].(*string) = repositoryName
					*pointers[5].(*string) = ""
					*pointers[6].(*string) = "wrong"
					*pointers[7].(*string) = repositoryVersion
					*pointers[8].(*uint64) = 1

					return nil
				})
				dummyFn := func(_ context.Context, scanner func(pgx.Rows) error, _ string, _ ...interface{}) error {
					if err := scanner(mockRows); err != nil {
						return err
					}
					return scanner(mockRows)
				}
				mockDatabase.EXPECT().List(gomock.Any(), gomock.Any(), gomock.Any(), uint64(3), uint(20)).DoAndReturn(dummyFn)
			}

			got, gotCount, gotErr := instance.List(testCtx, tc.args.pageSize, "")
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
				ids:       []uint64{1, 2},
				frequency: model.Daily,
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
			"error",
			args{
				ids:       []uint64{1, 2},
				frequency: model.Daily,
			},
			make([]model.Ketchup, 0),
			errors.New("failed"),
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
				mockRows.EXPECT().Scan(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(pointers ...interface{}) error {
					*pointers[0].(*string) = model.DefaultPattern
					*pointers[1].(*string) = "0.9.0"
					*pointers[2].(*string) = "daily"
					*pointers[3].(*uint64) = 1
					*pointers[4].(*uint64) = 1
					*pointers[5].(*string) = testEmail

					return nil
				})
				mockRows.EXPECT().Scan(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(pointers ...interface{}) error {
					*pointers[0].(*string) = model.DefaultPattern
					*pointers[1].(*string) = repositoryVersion
					*pointers[2].(*string) = "daily"
					*pointers[3].(*uint64) = 2
					*pointers[4].(*uint64) = 2
					*pointers[5].(*string) = "guest@domain"

					return nil
				})
				dummyFn := func(_ context.Context, scanner func(pgx.Rows) error, _ string, _ ...interface{}) error {
					if err := scanner(mockRows); err != nil {
						return err
					}
					return scanner(mockRows)
				}
				mockDatabase.EXPECT().List(gomock.Any(), gomock.Any(), gomock.Any(), tc.args.ids, "daily").DoAndReturn(dummyFn)
			case "error":
				mockDatabase.EXPECT().List(gomock.Any(), gomock.Any(), gomock.Any(), tc.args.ids, "daily").Return(errors.New("failed"))
			}

			got, gotErr := instance.ListByRepositoriesID(testCtx, tc.args.ids, tc.args.frequency)
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
	}
}

func TestGetByRepository(t *testing.T) {
	type args struct {
		id        uint64
		pattern   string
		forUpdate bool
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
				id:      1,
				pattern: "stable",
			},
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
				id:        1,
				pattern:   "stable",
				forUpdate: true,
			},
			model.NoneKetchup,
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
				mockRow.EXPECT().Scan(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(pointers ...interface{}) error {
					*pointers[0].(*string) = model.DefaultPattern
					*pointers[1].(*string) = "0.9.0"
					*pointers[2].(*string) = "daily"
					*pointers[3].(*uint64) = 1
					*pointers[4].(*uint64) = 3
					*pointers[5].(*string) = repositoryName
					*pointers[6].(*string) = ""
					*pointers[7].(*string) = "github"

					return nil
				})
				dummyFn := func(_ context.Context, scanner func(pgx.Row) error, _ string, _ ...interface{}) error {
					return scanner(mockRow)
				}
				mockDatabase.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any(), tc.args.id, uint64(3), tc.args.pattern).DoAndReturn(dummyFn)
			case "no rows":
				mockRow := mocks.NewRow(ctrl)
				mockRow.EXPECT().Scan(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(pointers ...interface{}) error {
					return pgx.ErrNoRows
				})
				dummyFn := func(_ context.Context, scanner func(pgx.Row) error, _ string, _ ...interface{}) error {
					return scanner(mockRow)
				}
				mockDatabase.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any(), tc.args.id, uint64(3), tc.args.pattern).DoAndReturn(dummyFn)
			case "error":
				mockDatabase.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any(), tc.args.id, uint64(3), tc.args.pattern).Return(errors.New("failed"))
			}

			got, gotErr := instance.GetByRepository(testCtx, tc.args.id, tc.args.pattern, tc.args.forUpdate)

			failed := false

			if !errors.Is(gotErr, tc.wantErr) {
				failed = true
			} else if !reflect.DeepEqual(got, tc.want) {
				failed = true
			}

			if failed {
				t.Errorf("GetByRepository() = (%+v, `%s`), want (%+v, `%s`)", got, gotErr, tc.want, tc.wantErr)
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
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockDatabase := mocks.NewDatabase(ctrl)

			instance := App{db: mockDatabase}

			switch tc.intention {
			case "simple":
				mockDatabase.EXPECT().Create(gomock.Any(), gomock.Any(), model.DefaultPattern, "0.9.0", "daily", uint64(1), uint64(3)).Return(uint64(1), nil)
			}

			got, gotErr := instance.Create(testCtx, tc.args.o)

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
	}
}

func TestUpdate(t *testing.T) {
	type args struct {
		o          model.Ketchup
		oldPattern string
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
				oldPattern: "stable",
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
				mockDatabase.EXPECT().Exec(gomock.Any(), gomock.Any(), uint64(1), uint64(3), model.DefaultPattern, model.DefaultPattern, "0.9.0", "daily").Return(nil)
			}

			gotErr := instance.Update(testCtx, tc.args.o, tc.args.oldPattern)

			failed := false

			if !errors.Is(gotErr, tc.wantErr) {
				failed = true
			}

			if failed {
				t.Errorf("Update() = `%s`, want `%s`", gotErr, tc.wantErr)
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
					Pattern:    "stable",
					Repository: model.NewGithubRepository(1, ""),
				},
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
				mockDatabase.EXPECT().Exec(gomock.Any(), gomock.Any(), uint64(1), uint64(3), model.DefaultPattern).Return(nil)
			}

			gotErr := instance.Delete(testCtx, tc.args.o)

			failed := false

			if !errors.Is(gotErr, tc.wantErr) {
				failed = true
			}

			if failed {
				t.Errorf("Delete() = `%s`, want `%s`", gotErr, tc.wantErr)
			}
		})
	}
}
