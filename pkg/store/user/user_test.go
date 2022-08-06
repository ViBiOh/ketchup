package user

import (
	"context"
	"errors"
	"reflect"
	"testing"

	authModel "github.com/ViBiOh/auth/v2/pkg/model"
	"github.com/ViBiOh/ketchup/pkg/mocks"
	"github.com/ViBiOh/ketchup/pkg/model"
	"github.com/golang/mock/gomock"
	"github.com/jackc/pgx/v4"
)

var testEmail = "nobody@localhost"

func TestGetByEmail(t *testing.T) {
	type args struct {
		email string
	}

	cases := map[string]struct {
		args    args
		want    model.User
		wantErr error
	}{
		"simple": {
			args{
				email: testEmail,
			},
			model.NewUser(1, testEmail, authModel.NewUser(1, "")),
			nil,
		},
		"no rows": {
			args{
				email: testEmail,
			},
			model.User{},
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
			case "simple":
				mockRow := mocks.NewRow(ctrl)
				mockRow.EXPECT().Scan(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(pointers ...any) error {
					*pointers[0].(*model.Identifier) = model.Identifier(1)
					*pointers[1].(*string) = testEmail
					*pointers[2].(*uint64) = 1

					return nil
				})
				dummyFn := func(_ context.Context, scanner func(pgx.Row) error, _ string, _ ...any) error {
					return scanner(mockRow)
				}
				mockDatabase.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any(), testEmail).DoAndReturn(dummyFn)
			case "no rows":
				mockRow := mocks.NewRow(ctrl)
				mockRow.EXPECT().Scan(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(pointers ...any) error {
					return pgx.ErrNoRows
				})
				dummyFn := func(_ context.Context, scanner func(pgx.Row) error, _ string, _ ...any) error {
					return scanner(mockRow)
				}
				mockDatabase.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any(), testEmail).DoAndReturn(dummyFn)
			}

			got, gotErr := instance.GetByEmail(context.Background(), tc.args.email)

			failed := false

			if !errors.Is(gotErr, tc.wantErr) {
				failed = true
			} else if !reflect.DeepEqual(got, tc.want) {
				failed = true
			}

			if failed {
				t.Errorf("GetByEmail() = (%+v, `%s`), want (%+v, `%s`)", got, gotErr, tc.want, tc.wantErr)
			}
		})
	}
}

func TestGetByLoginID(t *testing.T) {
	type args struct {
		loginID uint64
	}

	cases := map[string]struct {
		args    args
		want    model.User
		wantErr error
	}{
		"simple": {
			args{
				loginID: 2,
			},
			model.NewUser(1, testEmail, authModel.NewUser(2, "")),
			nil,
		},
		"no rows": {
			args{
				loginID: 2,
			},
			model.User{},
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
			case "simple":
				mockRow := mocks.NewRow(ctrl)
				mockRow.EXPECT().Scan(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(pointers ...any) error {
					*pointers[0].(*model.Identifier) = 1
					*pointers[1].(*string) = testEmail
					*pointers[2].(*uint64) = 2

					return nil
				})
				dummyFn := func(_ context.Context, scanner func(pgx.Row) error, _ string, _ ...any) error {
					return scanner(mockRow)
				}
				mockDatabase.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any(), uint64(2)).DoAndReturn(dummyFn)
			case "no rows":
				mockRow := mocks.NewRow(ctrl)
				mockRow.EXPECT().Scan(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(pointers ...any) error {
					return pgx.ErrNoRows
				})
				dummyFn := func(_ context.Context, scanner func(pgx.Row) error, _ string, _ ...any) error {
					return scanner(mockRow)
				}
				mockDatabase.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any(), uint64(2)).DoAndReturn(dummyFn)
			}

			got, gotErr := instance.GetByLoginID(context.Background(), tc.args.loginID)

			failed := false

			if !errors.Is(gotErr, tc.wantErr) {
				failed = true
			} else if !reflect.DeepEqual(got, tc.want) {
				failed = true
			}

			if failed {
				t.Errorf("GetByLoginID() = (%+v, `%s`), want (%+v, `%s`)", got, gotErr, tc.want, tc.wantErr)
			}
		})
	}
}

func TestCreate(t *testing.T) {
	type args struct {
		o model.User
	}

	cases := map[string]struct {
		args    args
		want    model.Identifier
		wantErr error
	}{
		"simple": {
			args{
				o: model.NewUser(0, testEmail, authModel.User{
					ID:       1,
					Login:    "vibioh",
					Password: "secret",
				}),
			},
			1,
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
			case "simple":
				mockDatabase.EXPECT().Create(gomock.Any(), gomock.Any(), testEmail, uint64(1)).Return(uint64(1), nil)
			}

			got, gotErr := instance.Create(context.Background(), tc.args.o)

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

func TestCount(t *testing.T) {
	cases := map[string]struct {
		want    uint64
		wantErr error
	}{
		"simple": {
			1,
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
			case "simple":
				mockRow := mocks.NewRow(ctrl)
				mockRow.EXPECT().Scan(gomock.Any()).DoAndReturn(func(pointers ...any) error {
					*pointers[0].(*uint64) = 1

					return nil
				})
				dummyFn := func(_ context.Context, scanner func(pgx.Row) error, _ string, _ ...any) error {
					return scanner(mockRow)
				}
				mockDatabase.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(dummyFn)
			}

			got, gotErr := instance.Count(context.Background())

			failed := false

			if !errors.Is(gotErr, tc.wantErr) {
				failed = true
			} else if got != tc.want {
				failed = true
			}

			if failed {
				t.Errorf("Count() = (%d, `%s`), want (%d, `%s`)", got, gotErr, tc.want, tc.wantErr)
			}
		})
	}
}
