package user

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	authModel "github.com/ViBiOh/auth/v3/pkg/model"
	httpModel "github.com/ViBiOh/httputils/v4/pkg/model"
	"github.com/ViBiOh/ketchup/pkg/mocks"
	"github.com/ViBiOh/ketchup/pkg/model"
	"go.uber.org/mock/gomock"
)

var (
	testEmail = "nobody@localhost"

	errAtomicStart = errors.New("invalid context")
)

func TestStoreInContext(t *testing.T) {
	t.Parallel()

	type args struct {
		ctx context.Context
	}

	cases := map[string]struct {
		instance Service
		args     args
		want     model.User
	}{
		"no login user": {
			Service{},
			args{
				ctx: context.TODO(),
			},
			model.User{},
		},
		"get error": {
			Service{},
			args{
				ctx: authModel.StoreUser(context.TODO(), authModel.NewUser(1, "")),
			},
			model.User{},
		},
		"not found login": {
			Service{},
			args{
				ctx: authModel.StoreUser(context.TODO(), authModel.NewUser(1, "")),
			},
			model.User{},
		},
		"valid": {
			Service{},
			args{
				ctx: authModel.StoreUser(context.TODO(), authModel.NewUser(1, "")),
			},
			model.NewUser(1, testEmail, authModel.NewUser(0, "")),
		},
	}

	for intention, testCase := range cases {
		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			mockUserStore := mocks.NewUserStore(ctrl)
			testCase.instance.store = mockUserStore

			switch intention {
			case "get error":
				mockUserStore.EXPECT().GetByLoginID(gomock.Any(), gomock.Any()).Return(model.User{}, errors.New("failed"))
			case "not found login":
				mockUserStore.EXPECT().GetByLoginID(gomock.Any(), gomock.Any()).Return(model.User{}, nil)
			case "valid":
				mockUserStore.EXPECT().GetByLoginID(gomock.Any(), gomock.Any()).Return(model.NewUser(1, testEmail, authModel.User{}), nil)
			}

			if got := testCase.instance.StoreInContext(testCase.args.ctx); !reflect.DeepEqual(model.ReadUser(got), testCase.want) {
				t.Errorf("StoreInContext() = %+v, want %+v", got, testCase.want)
			}
		})
	}
}

func TestCreate(t *testing.T) {
	t.Parallel()

	type args struct {
		ctx      context.Context
		email    string
		login    string
		password string
	}

	cases := map[string]struct {
		instance Service
		args     args
		want     model.User
		wantErr  error
	}{
		"invalid user": {
			Service{},
			args{
				ctx: context.TODO(),
			},
			model.User{},
			httpModel.ErrInvalid,
		},
		"invalid auth": {
			Service{},
			args{
				ctx:   context.TODO(),
				email: testEmail,
			},
			model.User{},
			httpModel.ErrInvalid,
		},
		"start atomic error": {
			Service{},
			args{
				ctx:      context.TODO(),
				email:    testEmail,
				login:    "admin",
				password: "password",
			},
			model.User{},
			errAtomicStart,
		},
		"login create error": {
			Service{},
			args{
				ctx:      context.TODO(),
				email:    testEmail,
				login:    "admin",
				password: "password",
			},
			model.User{},
			httpModel.ErrInternalError,
		},
		"user create error": {
			Service{},
			args{
				ctx:      context.TODO(),
				email:    testEmail,
				login:    "admin",
				password: "password",
			},
			model.User{},
			httpModel.ErrInternalError,
		},
		"success": {
			Service{},
			args{
				ctx:      context.TODO(),
				email:    testEmail,
				login:    "admin",
				password: "password",
			},
			model.NewUser(2, testEmail, authModel.NewUser(2, "admin")),
			nil,
		},
	}

	for intention, testCase := range cases {
		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			authService := mocks.NewAuthService(ctrl)
			mockUserStore := mocks.NewUserStore(ctrl)
			testCase.instance.auth = authService
			testCase.instance.store = mockUserStore

			realDoAtomic := func(ctx context.Context, action func(context.Context) error) error {
				return action(ctx)
			}

			switch intention {
			case "invalid user":
				mockUserStore.EXPECT().GetByEmail(gomock.Any(), gomock.Any()).Return(model.User{}, nil)
			case "invalid auth":
				mockUserStore.EXPECT().GetByEmail(gomock.Any(), gomock.Any()).Return(model.User{}, nil)
			case "start atomic error":
				mockUserStore.EXPECT().GetByEmail(gomock.Any(), gomock.Any()).Return(model.User{}, nil)
				mockUserStore.EXPECT().DoAtomic(gomock.Any(), gomock.Any()).Return(errAtomicStart)
			case "login create error":
				mockUserStore.EXPECT().DoAtomic(gomock.Any(), gomock.Any()).DoAndReturn(realDoAtomic)
				mockUserStore.EXPECT().GetByEmail(gomock.Any(), gomock.Any()).Return(model.User{}, nil)
				authService.EXPECT().CreateBasic(gomock.Any(), gomock.Any(), gomock.Any()).Return(authModel.User{}, errors.New("failed"))
			case "user create error":
				mockUserStore.EXPECT().DoAtomic(gomock.Any(), gomock.Any()).DoAndReturn(realDoAtomic)
				mockUserStore.EXPECT().GetByEmail(gomock.Any(), gomock.Any()).Return(model.User{}, nil)
				authService.EXPECT().CreateBasic(gomock.Any(), gomock.Any(), gomock.Any()).Return(authModel.User{}, errors.New("failed"))
			case "success":
				mockUserStore.EXPECT().DoAtomic(gomock.Any(), gomock.Any()).DoAndReturn(realDoAtomic)
				mockUserStore.EXPECT().GetByEmail(gomock.Any(), gomock.Any()).Return(model.User{}, nil)
				mockUserStore.EXPECT().Create(gomock.Any(), gomock.Any()).Return(model.Identifier(2), nil)
				authService.EXPECT().CreateBasic(gomock.Any(), gomock.Any(), gomock.Any()).Return(authModel.NewUser(2, "admin"), nil)
			}

			got, gotErr := testCase.instance.Create(testCase.args.ctx, testCase.args.email, testCase.args.login, testCase.args.password)

			failed := false

			if !errors.Is(gotErr, testCase.wantErr) {
				failed = true
			} else if !reflect.DeepEqual(got, testCase.want) {
				failed = true
			}

			if failed {
				t.Errorf("Create() = (%+v, `%s`), want (%+v, `%s`)", got, gotErr, testCase.want, testCase.wantErr)
			}
		})
	}
}

func TestCheck(t *testing.T) {
	t.Parallel()

	type args struct {
		ctx      context.Context
		email    string
		login    string
		password string
	}

	cases := map[string]struct {
		instance Service
		args     args
		wantErr  error
	}{
		"no name": {
			Service{},
			args{
				ctx: context.TODO(),
			},
			errors.New("email is required"),
		},
		"get error": {
			Service{},
			args{
				ctx:      context.TODO(),
				email:    testEmail,
				login:    "admin",
				password: "password",
			},
			errors.New("check if email already exists"),
		},
		"already used": {
			Service{},
			args{
				ctx:      context.TODO(),
				email:    testEmail,
				login:    "admin",
				password: "password",
			},
			errors.New("email already used"),
		},
	}

	for intention, testCase := range cases {
		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			mockUserStore := mocks.NewUserStore(ctrl)

			testCase.instance.store = mockUserStore

			switch intention {
			case "no name":
				mockUserStore.EXPECT().GetByEmail(gomock.Any(), gomock.Any()).Return(model.User{}, nil)
			case "get error":
				mockUserStore.EXPECT().GetByEmail(gomock.Any(), gomock.Any()).Return(model.User{}, errors.New("failed"))
			case "already used":
				mockUserStore.EXPECT().GetByEmail(gomock.Any(), gomock.Any()).Return(model.NewUser(1, testEmail, authModel.NewUser(1, "")), nil)
			}

			gotErr := testCase.instance.check(testCase.args.ctx, testCase.args.email, testCase.args.login, testCase.args.password)

			failed := false

			if testCase.wantErr == nil && gotErr != nil {
				failed = true
			} else if testCase.wantErr != nil && gotErr == nil {
				failed = true
			} else if testCase.wantErr != nil && !strings.Contains(gotErr.Error(), testCase.wantErr.Error()) {
				failed = true
			}

			if failed {
				t.Errorf("check() = `%s`, want `%s`", gotErr, testCase.wantErr)
			}
		})
	}
}
