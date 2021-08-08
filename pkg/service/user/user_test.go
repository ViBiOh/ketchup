package user

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	authModel "github.com/ViBiOh/auth/v2/pkg/model"
	httpModel "github.com/ViBiOh/httputils/v4/pkg/model"
	"github.com/ViBiOh/ketchup/pkg/mocks"
	"github.com/ViBiOh/ketchup/pkg/model"
	"github.com/ViBiOh/ketchup/pkg/store/user/usertest"
	"github.com/golang/mock/gomock"
)

var (
	testEmail = "nobody@localhost"

	errAtomicStart = errors.New("invalid context")
)

func TestStoreInContext(t *testing.T) {
	type args struct {
		ctx context.Context
	}

	var cases = []struct {
		intention string
		instance  App
		args      args
		want      model.User
	}{
		{
			"no login user",
			New(usertest.New(), nil),
			args{
				ctx: context.Background(),
			},
			model.NoneUser,
		},
		{
			"get error",
			New(usertest.New().SetGetByLoginID(model.NoneUser, errors.New("failed")), nil),
			args{
				ctx: authModel.StoreUser(context.Background(), authModel.NewUser(1, "")),
			},
			model.NoneUser,
		},
		{
			"not found login",
			New(usertest.New(), nil),
			args{
				ctx: authModel.StoreUser(context.Background(), authModel.NewUser(1, "")),
			},
			model.NoneUser,
		},
		{
			"valid",
			New(usertest.New().SetGetByLoginID(model.NewUser(1, testEmail, authModel.NoneUser), nil), nil),
			args{
				ctx: authModel.StoreUser(context.Background(), authModel.NewUser(1, "")),
			},
			model.NewUser(1, testEmail, authModel.NewUser(0, "")),
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			if got := tc.instance.StoreInContext(tc.args.ctx); !reflect.DeepEqual(model.ReadUser(got), tc.want) {
				t.Errorf("StoreInContext() = %+v, want %+v", got, tc.want)
			}
		})
	}
}

func TestCreate(t *testing.T) {
	type args struct {
		ctx  context.Context
		item model.User
	}

	var cases = []struct {
		intention string
		instance  App
		args      args
		want      model.User
		wantErr   error
	}{
		{
			"invalid user",
			App{
				userStore: usertest.New(),
			},
			args{
				ctx:  context.TODO(),
				item: model.NewUser(1, "", authModel.NewUser(1, "")),
			},
			model.NoneUser,
			httpModel.ErrInvalid,
		},
		{
			"invalid auth",
			App{
				userStore: usertest.New(),
			},
			args{
				ctx:  context.TODO(),
				item: model.NewUser(0, testEmail, authModel.NewUser(0, "")),
			},
			model.NoneUser,
			httpModel.ErrInvalid,
		},
		{
			"start atomic error",
			App{
				userStore: usertest.New().SetGetByEmail(model.NoneUser, nil).SetDoAtomic(errAtomicStart),
			},
			args{
				ctx:  context.TODO(),
				item: model.NewUser(1, testEmail, authModel.NewUser(1, "")),
			},
			model.NoneUser,
			errAtomicStart,
		},
		{
			"login create error",
			App{
				userStore: usertest.New().SetGetByEmail(model.NoneUser, nil),
			},
			args{
				ctx:  context.Background(),
				item: model.NewUser(1, testEmail, authModel.NewUser(1, "")),
			},
			model.NoneUser,
			httpModel.ErrInternalError,
		},
		{
			"user create error",
			App{
				userStore: usertest.New().SetGetByEmail(model.NoneUser, nil).SetCreate(0, errors.New("failed")),
			},
			args{
				ctx:  context.Background(),
				item: model.NewUser(2, testEmail, authModel.NewUser(2, "")),
			},
			model.NoneUser,
			httpModel.ErrInternalError,
		},
		{
			"success",
			App{
				userStore: usertest.New().SetGetByEmail(model.NoneUser, nil).SetCreate(2, nil),
			},
			args{
				ctx:  context.Background(),
				item: model.NewUser(2, testEmail, authModel.NewUser(2, "")),
			},
			model.NewUser(2, testEmail, authModel.NewUser(2, "admin")),
			nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			authApp := mocks.NewAuth(ctrl)
			tc.instance.authApp = authApp

			switch tc.intention {
			case "invalid auth":
				authApp.EXPECT().Check(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("failed"))
			case "start atomic error":
				authApp.EXPECT().Check(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			case "login create error":
				authApp.EXPECT().Check(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
				authApp.EXPECT().Create(gomock.Any(), gomock.Any()).Return(authModel.NoneUser, errors.New("failed"))
			case "user create error":
				authApp.EXPECT().Check(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
				authApp.EXPECT().Create(gomock.Any(), gomock.Any()).Return(authModel.NoneUser, errors.New("failed"))
			case "success":
				authApp.EXPECT().Check(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
				authApp.EXPECT().Create(gomock.Any(), gomock.Any()).Return(authModel.NewUser(2, "admin"), nil)
			}

			got, gotErr := tc.instance.Create(tc.args.ctx, tc.args.item)

			failed := false

			if !errors.Is(gotErr, tc.wantErr) {
				failed = true
			} else if !reflect.DeepEqual(got, tc.want) {
				failed = true
			}

			if failed {
				t.Errorf("Create() = (%+v, `%s`), want (%+v, `%s`)", got, gotErr, tc.want, tc.wantErr)
			}
		})
	}
}

func TestCheck(t *testing.T) {
	type args struct {
		ctx context.Context
		old model.User
		new model.User
	}

	var cases = []struct {
		intention string
		instance  App
		args      args
		wantErr   error
	}{
		{
			"delete",
			App{userStore: usertest.New()},
			args{
				ctx: context.Background(),
			},
			nil,
		},
		{
			"no name",
			App{userStore: usertest.New()},
			args{
				ctx: context.Background(),
				new: model.NewUser(1, "", authModel.NewUser(1, "")),
			},
			errors.New("email is required"),
		},
		{
			"get error",
			App{userStore: usertest.New().SetGetByEmail(model.NoneUser, errors.New("failed"))},
			args{
				ctx: context.Background(),
				new: model.NewUser(1, testEmail, authModel.NewUser(1, "")),
			},
			errors.New("unable to check if email already exists"),
		},
		{
			"already used",
			App{userStore: usertest.New().SetGetByEmail(model.NewUser(1, testEmail, authModel.NewUser(1, "")), nil)},
			args{
				ctx: context.Background(),
				new: model.NewUser(1, testEmail, authModel.NewUser(1, "")),
			},
			errors.New("email already used"),
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			gotErr := tc.instance.check(tc.args.ctx, tc.args.old, tc.args.new)

			failed := false

			if tc.wantErr == nil && gotErr != nil {
				failed = true
			} else if tc.wantErr != nil && gotErr == nil {
				failed = true
			} else if tc.wantErr != nil && !strings.Contains(gotErr.Error(), tc.wantErr.Error()) {
				failed = true
			}

			if failed {
				t.Errorf("check() = `%s`, want `%s`", gotErr, tc.wantErr)
			}
		})
	}
}
