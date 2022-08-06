package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/ViBiOh/ketchup/pkg/mocks"
	"github.com/ViBiOh/ketchup/pkg/model"
	"github.com/golang/mock/gomock"
	"github.com/jackc/pgx/v4"
)

var errFailed = errors.New("timeout")

func TestUpdateVersions(t *testing.T) {
	type args struct {
		o model.Repository
	}

	cases := map[string]struct {
		args    args
		wantErr error
	}{
		"no version": {
			args{
				o: model.NewEmptyRepository(),
			},
			nil,
		},
		"create error": {
			args{
				o: model.NewGithubRepository(model.Identifier(0), "").AddVersion(model.DefaultPattern, "1.0.0"),
			},
			errFailed,
		},
		"create": {
			args{
				o: model.NewGithubRepository(model.Identifier(0), "").AddVersion(model.DefaultPattern, "1.0.0"),
			},
			nil,
		},
		"no update": {
			args{
				o: model.NewGithubRepository(model.Identifier(0), "").AddVersion(model.DefaultPattern, "1.0.0"),
			},
			nil,
		},
		"update error": {
			args{
				o: model.NewGithubRepository(model.Identifier(0), "").AddVersion(model.DefaultPattern, "1.0.0"),
			},
			errFailed,
		},
		"update": {
			args{
				o: model.NewGithubRepository(model.Identifier(0), "").AddVersion(model.DefaultPattern, "1.0.0"),
			},
			nil,
		},
		"delete error": {
			args{
				o: model.NewGithubRepository(model.Identifier(0), ""),
			},
			errFailed,
		},
		"delete": {
			args{
				o: model.NewGithubRepository(model.Identifier(0), ""),
			},
			nil,
		},
	}

	for intention, tc := range cases {
		t.Run(intention, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockDatabase := mocks.NewDatabase(ctrl)

			instance := App{db: mockDatabase}

			switch intention {
			case "no version":
				mockDatabase.EXPECT().List(gomock.Any(), gomock.Any(), gomock.Any(), model.Identifier(0)).Return(nil)
			case "create error":
				mockDatabase.EXPECT().List(gomock.Any(), gomock.Any(), gomock.Any(), model.Identifier(0)).Return(nil)
				mockDatabase.EXPECT().One(gomock.Any(), gomock.Any(), model.Identifier(0), model.DefaultPattern, "1.0.0").Return(errFailed)
			case "create":
				mockDatabase.EXPECT().List(gomock.Any(), gomock.Any(), gomock.Any(), model.Identifier(0)).Return(nil)
				mockDatabase.EXPECT().One(gomock.Any(), gomock.Any(), model.Identifier(0), model.DefaultPattern, "1.0.0").Return(nil)
			case "no update":
				mockRows := mocks.NewRows(ctrl)
				mockRows.EXPECT().Scan(gomock.Any(), gomock.Any()).DoAndReturn(func(pointers ...any) error {
					*pointers[0].(*string) = model.DefaultPattern
					*pointers[1].(*string) = "1.0.0"

					return nil
				})
				dummyFn := func(_ context.Context, scanner func(pgx.Rows) error, _ string, _ ...any) error {
					return scanner(mockRows)
				}
				mockDatabase.EXPECT().List(gomock.Any(), gomock.Any(), gomock.Any(), model.Identifier(0)).DoAndReturn(dummyFn)
			case "update error":
				mockRows := mocks.NewRows(ctrl)
				mockRows.EXPECT().Scan(gomock.Any(), gomock.Any()).DoAndReturn(func(pointers ...any) error {
					*pointers[0].(*string) = model.DefaultPattern
					*pointers[1].(*string) = "0.9.0"

					return nil
				})
				dummyFn := func(_ context.Context, scanner func(pgx.Rows) error, _ string, _ ...any) error {
					return scanner(mockRows)
				}
				mockDatabase.EXPECT().List(gomock.Any(), gomock.Any(), gomock.Any(), model.Identifier(0)).DoAndReturn(dummyFn)
				mockDatabase.EXPECT().One(gomock.Any(), gomock.Any(), model.Identifier(0), model.DefaultPattern, "1.0.0").Return(errFailed)
			case "update":
				mockRows := mocks.NewRows(ctrl)
				mockRows.EXPECT().Scan(gomock.Any(), gomock.Any()).DoAndReturn(func(pointers ...any) error {
					*pointers[0].(*string) = model.DefaultPattern
					*pointers[1].(*string) = "0.9.0"

					return nil
				})
				dummyFn := func(_ context.Context, scanner func(pgx.Rows) error, _ string, _ ...any) error {
					return scanner(mockRows)
				}
				mockDatabase.EXPECT().List(gomock.Any(), gomock.Any(), gomock.Any(), model.Identifier(0)).DoAndReturn(dummyFn)
				mockDatabase.EXPECT().One(gomock.Any(), gomock.Any(), model.Identifier(0), model.DefaultPattern, "1.0.0").Return(nil)
			case "delete error":
				mockRows := mocks.NewRows(ctrl)
				mockRows.EXPECT().Scan(gomock.Any(), gomock.Any()).DoAndReturn(func(pointers ...any) error {
					*pointers[0].(*string) = model.DefaultPattern
					*pointers[1].(*string) = "0.9.0"

					return nil
				})
				dummyFn := func(_ context.Context, scanner func(pgx.Rows) error, _ string, _ ...any) error {
					return scanner(mockRows)
				}
				mockDatabase.EXPECT().List(gomock.Any(), gomock.Any(), gomock.Any(), model.Identifier(0)).DoAndReturn(dummyFn)
				mockDatabase.EXPECT().One(gomock.Any(), gomock.Any(), model.Identifier(0), model.DefaultPattern).Return(errFailed)
			case "delete":
				mockRows := mocks.NewRows(ctrl)
				mockRows.EXPECT().Scan(gomock.Any(), gomock.Any()).DoAndReturn(func(pointers ...any) error {
					*pointers[0].(*string) = model.DefaultPattern
					*pointers[1].(*string) = "0.9.0"

					return nil
				})
				dummyFn := func(_ context.Context, scanner func(pgx.Rows) error, _ string, _ ...any) error {
					return scanner(mockRows)
				}
				mockDatabase.EXPECT().List(gomock.Any(), gomock.Any(), gomock.Any(), model.Identifier(0)).DoAndReturn(dummyFn)
				mockDatabase.EXPECT().One(gomock.Any(), gomock.Any(), model.Identifier(0), model.DefaultPattern).Return(nil)
			}

			gotErr := instance.UpdateVersions(context.Background(), tc.args.o)

			failed := false

			if !errors.Is(gotErr, tc.wantErr) {
				failed = true
			}

			if failed {
				t.Errorf("UpdateVersions() = `%s`, want `%s`", gotErr, tc.wantErr)
			}
		})
	}
}
