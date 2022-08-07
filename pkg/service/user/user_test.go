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

	cases := map[string]struct {
		instance App
		args     args
		want     model.User
	}{
		"no login user": {
			App{},
			args{
				ctx: context.Background(),
			},
			model.User{},
		},
		"get error": {
			App{},
			args{
				ctx: authModel.StoreUser(context.Background(), authModel.NewUser(1, "")),
			},
			model.User{},
		},
		"not found login": {
			App{},
			args{
				ctx: authModel.StoreUser(context.Background(), authModel.NewUser(1, "")),
			},
			model.User{},
		},
		"valid": {
			App{},
			args{
				ctx: authModel.StoreUser(context.Background(), authModel.NewUser(1, "")),
			},
			model.NewUser(1, testEmail, authModel.NewUser(0, "")),
		},
	}

	for intention, tc := range cases {
		t.Run(intention, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUserStore := mocks.NewUserStore(ctrl)
			tc.instance.userStore = mockUserStore

			switch intention {
			case "get error":
				mockUserStore.EXPECT().GetByLoginID(gomock.Any(), gomock.Any()).Return(model.User{}, errors.New("failed"))
			case "not found login":
				mockUserStore.EXPECT().GetByLoginID(gomock.Any(), gomock.Any()).Return(model.User{}, nil)
			case "valid":
				mockUserStore.EXPECT().GetByLoginID(gomock.Any(), gomock.Any()).Return(model.NewUser(1, testEmail, authModel.User{}), nil)
			}

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

	cases := map[string]struct {
		instance App
		args     args
		want     model.User
		wantErr  error
	}{
		"invalid user": {
			App{},
			args{
				ctx:  context.Background(),
				item: model.NewUser(1, "", authModel.NewUser(1, "")),
			},
			model.User{},
			httpModel.ErrInvalid,
		},
		"invalid auth": {
			App{},
			args{
				ctx:  context.Background(),
				item: model.NewUser(0, testEmail, authModel.NewUser(0, "")),
			},
			model.User{},
			httpModel.ErrInvalid,
		},
		"start atomic error": {
			App{},
			args{
				ctx:  context.Background(),
				item: model.NewUser(1, testEmail, authModel.NewUser(1, "")),
			},
			model.User{},
			errAtomicStart,
		},
		"login create error": {
			App{},
			args{
				ctx:  context.Background(),
				item: model.NewUser(1, testEmail, authModel.NewUser(1, "")),
			},
			model.User{},
			httpModel.ErrInternalError,
		},
		"user create error": {
			App{},
			args{
				ctx:  context.Background(),
				item: model.NewUser(2, testEmail, authModel.NewUser(2, "")),
			},
			model.User{},
			httpModel.ErrInternalError,
		},
		"success": {
			App{},
			args{
				ctx:  context.Background(),
				item: model.NewUser(2, testEmail, authModel.NewUser(2, "")),
			},
			model.NewUser(2, testEmail, authModel.NewUser(2, "admin")),
			nil,
		},
	}

	for intention, tc := range cases {
		t.Run(intention, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			authApp := mocks.NewAuthService(ctrl)
			mockUserStore := mocks.NewUserStore(ctrl)
			tc.instance.authApp = authApp
			tc.instance.userStore = mockUserStore

			realDoAtomic := func(ctx context.Context, action func(context.Context) error) error {
				return action(ctx)
			}

			switch intention {
			case "invalid user":
				mockUserStore.EXPECT().GetByEmail(gomock.Any(), gomock.Any()).Return(model.User{}, nil)
			case "invalid auth":
				authApp.EXPECT().Check(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("failed"))
			case "start atomic error":
				authApp.EXPECT().Check(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
				mockUserStore.EXPECT().GetByEmail(gomock.Any(), gomock.Any()).Return(model.User{}, nil)
				mockUserStore.EXPECT().DoAtomic(gomock.Any(), gomock.Any()).Return(errAtomicStart)
			case "login create error":
				mockUserStore.EXPECT().DoAtomic(gomock.Any(), gomock.Any()).DoAndReturn(realDoAtomic)
				mockUserStore.EXPECT().GetByEmail(gomock.Any(), gomock.Any()).Return(model.User{}, nil)
				authApp.EXPECT().Check(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
				authApp.EXPECT().Create(gomock.Any(), gomock.Any()).Return(authModel.User{}, errors.New("failed"))
			case "user create error":
				mockUserStore.EXPECT().DoAtomic(gomock.Any(), gomock.Any()).DoAndReturn(realDoAtomic)
				mockUserStore.EXPECT().GetByEmail(gomock.Any(), gomock.Any()).Return(model.User{}, nil)
				authApp.EXPECT().Check(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
				authApp.EXPECT().Create(gomock.Any(), gomock.Any()).Return(authModel.User{}, errors.New("failed"))
			case "success":
				mockUserStore.EXPECT().DoAtomic(gomock.Any(), gomock.Any()).DoAndReturn(realDoAtomic)
				mockUserStore.EXPECT().GetByEmail(gomock.Any(), gomock.Any()).Return(model.User{}, nil)
				mockUserStore.EXPECT().Create(gomock.Any(), gomock.Any()).Return(model.Identifier(2), nil)
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

	cases := map[string]struct {
		instance App
		args     args
		wantErr  error
	}{
		"delete": {
			App{},
			args{
				ctx: context.Background(),
			},
			nil,
		},
		"no name": {
			App{},
			args{
				ctx: context.Background(),
				new: model.NewUser(1, "", authModel.NewUser(1, "")),
			},
			errors.New("email is required"),
		},
		"get error": {
			App{},
			args{
				ctx: context.Background(),
				new: model.NewUser(1, testEmail, authModel.NewUser(1, "")),
			},
			errors.New("check if email already exists"),
		},
		"already used": {
			App{},
			args{
				ctx: context.Background(),
				new: model.NewUser(1, testEmail, authModel.NewUser(1, "")),
			},
			errors.New("email already used"),
		},
	}

	for intention, tc := range cases {
		t.Run(intention, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUserStore := mocks.NewUserStore(ctrl)

			tc.instance.userStore = mockUserStore

			switch intention {
			case "no name":
				mockUserStore.EXPECT().GetByEmail(gomock.Any(), gomock.Any()).Return(model.User{}, nil)
			case "get error":
				mockUserStore.EXPECT().GetByEmail(gomock.Any(), gomock.Any()).Return(model.User{}, errors.New("failed"))
			case "already used":
				mockUserStore.EXPECT().GetByEmail(gomock.Any(), gomock.Any()).Return(model.NewUser(1, testEmail, authModel.NewUser(1, "")), nil)
			}

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
