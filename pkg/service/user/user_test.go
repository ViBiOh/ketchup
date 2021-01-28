package user

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	authModel "github.com/ViBiOh/auth/v2/pkg/model"
	httpModel "github.com/ViBiOh/httputils/v3/pkg/model"
	"github.com/ViBiOh/ketchup/pkg/model"
	"github.com/ViBiOh/ketchup/pkg/store/user/usertest"
)

var (
	errAtomicStart = errors.New("invalid context")
	errAtomicEnd   = errors.New("invalid context")
)

type testAuthService struct{}

func (tas testAuthService) Unmarshal(_ []byte, _ string) (authModel.User, error) {
	return authModel.NoneUser, nil
}

func (tas testAuthService) Check(_ context.Context, _, new authModel.User) error {
	if new.ID == 0 {
		return errors.New("id is invalid")
	}

	return nil
}

func (tas testAuthService) List(_ context.Context, _, _ uint, _ string, _ bool, _ map[string][]string) ([]authModel.User, uint, error) {
	return nil, 0, nil
}

func (tas testAuthService) Get(_ context.Context, _ uint64) (authModel.User, error) {
	return authModel.NoneUser, nil
}

func (tas testAuthService) Create(_ context.Context, o authModel.User) (authModel.User, error) {
	if o.ID == 1 {
		return authModel.NoneUser, errors.New("invalid id")
	}

	return authModel.NewUser(o.ID, "admin"), nil
}

func (tas testAuthService) Update(_ context.Context, _ authModel.User) (authModel.User, error) {
	return authModel.NoneUser, nil
}

func (tas testAuthService) Delete(_ context.Context, _ authModel.User) error {
	return nil
}

func (tas testAuthService) CheckRights(_ context.Context, _ uint64) error {
	return nil
}

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
			New(usertest.New().SetGetByLoginID(model.NewUser(1, "nobody@localhost", authModel.NoneUser), nil), nil),
			args{
				ctx: authModel.StoreUser(context.Background(), authModel.NewUser(1, "")),
			},
			model.User{ID: 1, Email: "nobody@localhost"},
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
			New(usertest.New(), testAuthService{}),
			args{
				ctx:  context.TODO(),
				item: model.NewUser(1, "", authModel.NewUser(1, "")),
			},
			model.NoneUser,
			httpModel.ErrInvalid,
		},
		{
			"invalid auth",
			New(usertest.New(), testAuthService{}),
			args{
				ctx:  context.TODO(),
				item: model.NewUser(0, "nobody@localhost", authModel.NewUser(0, "")),
			},
			model.NoneUser,
			httpModel.ErrInvalid,
		},
		{
			"start atomic error",
			New(usertest.New().SetGetByEmail(model.NoneUser, nil).SetDoAtomic(errAtomicStart), testAuthService{}),
			args{
				ctx:  context.TODO(),
				item: model.NewUser(1, "nobody@localhost", authModel.NewUser(1, "")),
			},
			model.NoneUser,
			errAtomicStart,
		},
		{
			"login create error",
			New(usertest.New().SetGetByEmail(model.NoneUser, nil), testAuthService{}),
			args{
				ctx:  context.Background(),
				item: model.NewUser(1, "nobody@localhost", authModel.NewUser(1, "")),
			},
			model.NoneUser,
			httpModel.ErrInternalError,
		},
		{
			"user create error",
			New(usertest.New().SetGetByEmail(model.NoneUser, nil).SetCreate(0, errors.New("failed")), testAuthService{}),
			args{
				ctx:  context.Background(),
				item: model.NewUser(2, "nobody@localhost", authModel.NewUser(2, "")),
			},
			model.NoneUser,
			httpModel.ErrInternalError,
		},
		{
			"success",
			New(usertest.New().SetGetByEmail(model.NoneUser, nil).SetCreate(2, nil), testAuthService{}),
			args{
				ctx:  context.Background(),
				item: model.NewUser(2, "nobody@localhost", authModel.NewUser(2, "")),
			},
			model.NewUser(2, "nobody@localhost", authModel.NewUser(2, "admin")),
			nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
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
		instance  app
		args      args
		wantErr   error
	}{
		{
			"delete",
			app{userStore: usertest.New()},
			args{
				ctx: context.Background(),
			},
			nil,
		},
		{
			"no name",
			app{userStore: usertest.New()},
			args{
				ctx: context.Background(),
				new: model.NewUser(1, "", authModel.NewUser(1, "")),
			},
			errors.New("email is required"),
		},
		{
			"get error",
			app{userStore: usertest.New().SetGetByEmail(model.NoneUser, errors.New("failed"))},
			args{
				ctx: context.Background(),
				new: model.NewUser(1, "nobody@localhost", authModel.NewUser(1, "")),
			},
			errors.New("unable to check if email already exists"),
		},
		{
			"already used",
			app{userStore: usertest.New().SetGetByEmail(model.NewUser(1, "nobody@localhost", authModel.NewUser(1, "")), nil)},
			args{
				ctx: context.Background(),
				new: model.NewUser(1, "nobody@localhost", authModel.NewUser(1, "")),
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
