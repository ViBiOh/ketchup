package user

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	authModel "github.com/ViBiOh/auth/v2/pkg/model"
	"github.com/ViBiOh/ketchup/pkg/model"
	"github.com/ViBiOh/ketchup/pkg/service"
)

var (
	errAtomicStart = errors.New("invalid context")
	errAtomicEnd   = errors.New("invalid context")
)

type testUserStore struct{}

func (tus testUserStore) DoAtomic(ctx context.Context, action func(context.Context) error) error {
	if ctx == context.TODO() {
		return errAtomicStart
	}

	err := action(ctx)
	if err != nil && strings.Contains(err.Error(), "duplicate pk") {
		return errAtomicEnd
	}

	return err
}

func (tus testUserStore) GetByEmail(_ context.Context, email string) (model.User, error) {
	if email == "guest@nowhere" {
		return model.NoneUser, errors.New("invalid email")
	}

	if email == "guest@localhost" {
		return model.User{
			ID: 1,
		}, nil
	}

	return model.NoneUser, nil
}

func (tus testUserStore) GetByLoginID(_ context.Context, loginID uint64) (model.User, error) {
	if loginID == 0 {
		return model.NoneUser, errors.New("invalid login id")
	}

	if loginID == 1 {
		return model.NoneUser, nil
	}

	return model.User{Email: "nobody@localhost"}, nil
}

func (tus testUserStore) Create(_ context.Context, o model.User) (uint64, error) {
	if o.Login.ID == 2 {
		return 0, errors.New("invalid id")
	}

	if o.Login.ID == 3 {
		return 0, errors.New("duplicate pk")
	}

	return 1, nil
}

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
		args      args
		want      model.User
	}{
		{
			"no login user",
			args{
				ctx: context.Background(),
			},
			model.NoneUser,
		},
		{
			"get error",
			args{
				ctx: authModel.StoreUser(context.Background(), authModel.NewUser(0, "")),
			},
			model.NoneUser,
		},
		{
			"not found login",
			args{
				ctx: authModel.StoreUser(context.Background(), authModel.NewUser(1, "")),
			},
			model.NoneUser,
		},
		{
			"valid",
			args{
				ctx: authModel.StoreUser(context.Background(), authModel.NewUser(2, "")),
			},
			model.User{Email: "nobody@localhost"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			if got := New(testUserStore{}, nil).StoreInContext(tc.args.ctx); !reflect.DeepEqual(model.ReadUser(got), tc.want) {
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
		args      args
		want      model.User
		wantErr   error
	}{
		{
			"invalid user",
			args{
				ctx: context.TODO(),
				item: model.User{
					Login: authModel.User{
						ID: 1,
					},
				},
			},
			model.NoneUser,
			service.ErrInvalid,
		},
		{
			"invalid auth",
			args{
				ctx: context.TODO(),
				item: model.User{
					Email: "nobody@localhost",
					Login: authModel.User{
						ID: 0,
					},
				},
			},
			model.NoneUser,
			service.ErrInvalid,
		},
		{
			"start atomic error",
			args{
				ctx: context.TODO(),
				item: model.User{
					Email: "nobody@localhost",
					Login: authModel.User{
						ID: 1,
					},
				},
			},
			model.NoneUser,
			errAtomicStart,
		},
		{
			"login create error",
			args{
				ctx: context.Background(),
				item: model.User{
					Email: "nobody@localhost",
					Login: authModel.User{
						ID: 1,
					},
				},
			},
			model.NoneUser,
			service.ErrInternalError,
		},
		{
			"user create error",
			args{
				ctx: context.Background(),
				item: model.User{
					Email: "nobody@localhost",
					Login: authModel.User{
						ID: 2,
					},
				},
			},
			model.NoneUser,
			service.ErrInternalError,
		},
		{
			"end atomic error",
			args{
				ctx: context.Background(),
				item: model.User{
					Email: "nobody@localhost",
					Login: authModel.User{
						ID: 3,
					},
				},
			},
			model.NoneUser,
			errAtomicEnd,
		},
		{
			"success",
			args{
				ctx: context.Background(),
				item: model.User{
					Email: "nobody@localhost",
					Login: authModel.User{
						ID: 4,
					},
				},
			},
			model.User{
				ID:    1,
				Email: "nobody@localhost",
				Login: authModel.User{
					ID:    4,
					Login: "admin",
				},
			},
			nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			got, gotErr := New(testUserStore{}, testAuthService{}).Create(tc.args.ctx, tc.args.item)

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
		args      args
		wantErr   error
	}{
		{
			"delete",
			args{
				ctx: context.Background(),
			},
			nil,
		},
		{
			"no name",
			args{
				ctx: context.Background(),
				new: model.User{
					Login: authModel.User{
						ID: 1,
					},
				},
			},
			errors.New("email is required"),
		},
		{
			"get error",
			args{
				ctx: context.Background(),
				new: model.User{
					Email: "guest@nowhere",
				},
			},
			errors.New("unable to check if email already exists"),
		},
		{
			"already used",
			args{
				ctx: context.Background(),
				new: model.User{
					Email: "guest@localhost",
				},
			},
			errors.New("email already used"),
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			gotErr := app{userStore: testUserStore{}}.check(tc.args.ctx, tc.args.old, tc.args.new)

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
