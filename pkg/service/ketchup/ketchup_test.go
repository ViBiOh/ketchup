package ketchup

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	httpModel "github.com/ViBiOh/httputils/v3/pkg/model"
	"github.com/ViBiOh/ketchup/pkg/model"
	"github.com/ViBiOh/ketchup/pkg/service/repository/repositorytest"
)

var (
	errAtomicStart = errors.New("invalid context")
	errAtomicEnd   = errors.New("invalid context")
)

type testKetchupStore struct{}

func (tks testKetchupStore) DoAtomic(ctx context.Context, action func(context.Context) error) error {
	if ctx == context.TODO() {
		return errAtomicStart
	}

	err := action(ctx)
	if err != nil && strings.Contains(err.Error(), "duplicate pk") {
		return errAtomicEnd
	}

	return err
}

func (tks testKetchupStore) List(_ context.Context, page, _ uint) ([]model.Ketchup, uint64, error) {
	if page == 0 {
		return nil, 0, errors.New("invalid page size")
	}

	return []model.Ketchup{
		{Pattern: model.DefaultPattern, Version: "1.0.0", Repository: model.Repository{Versions: map[string]string{model.DefaultPattern: "1.0.2"}}},
		{Pattern: model.DefaultPattern, Version: "1.2.3", Repository: model.Repository{Versions: map[string]string{model.DefaultPattern: "1.2.3"}}},
	}, 2, nil
}

func (tks testKetchupStore) ListByRepositoriesID(_ context.Context, ids []uint64) ([]model.Ketchup, error) {
	if len(ids) == 0 {
		return nil, errors.New("empty request")
	}

	return []model.Ketchup{
		{Pattern: model.DefaultPattern, Version: "1.0.0", Repository: model.Repository{ID: 1, Name: "vibioh/ketchup", Versions: map[string]string{model.DefaultPattern: "1.0.2"}}},
		{Pattern: model.DefaultPattern, Version: "1.2.3", Repository: model.Repository{ID: 2, Name: "vibioh/viws", Versions: map[string]string{model.DefaultPattern: "1.2.3"}}},
	}, nil
}

func (tks testKetchupStore) GetByRepositoryID(_ context.Context, id uint64, _ bool) (model.Ketchup, error) {
	if id == 0 {
		return model.NoneKetchup, errors.New("invalid id")
	}

	if id == 2 {
		return model.Ketchup{Pattern: model.DefaultPattern, Version: "1.0.0", Repository: model.Repository{ID: 2, Name: "vibioh/ketchup", Versions: map[string]string{model.DefaultPattern: "1.2.3"}}, User: model.User{ID: 1}}, nil
	}

	if id == 3 {
		return model.Ketchup{Pattern: model.DefaultPattern, Version: "0.0.0"}, nil
	}

	if id == 4 {
		return model.Ketchup{Pattern: model.DefaultPattern, Version: "0"}, nil
	}

	return model.NoneKetchup, nil
}

func (tks testKetchupStore) Create(_ context.Context, o model.Ketchup) (uint64, error) {
	if o.Version == "0" {
		return 0, errors.New("duplicate pk")
	}

	if o.Version == "0.0.0" {
		return 0, errors.New("invalid version")
	}

	return 0, nil
}

func (tks testKetchupStore) Update(_ context.Context, o model.Ketchup) error {
	if o.Version == "0" {
		return errors.New("duplicate pk")
	}

	if o.Version == "0.0.0" {
		return errors.New("invalid version")
	}

	return nil
}

func (tks testKetchupStore) Delete(_ context.Context, o model.Ketchup) error {
	if o.Version == "0" {
		return errors.New("duplicate pk")
	}

	if o.Version == "0.0.0" {
		return errors.New("invalid version")
	}

	return nil
}

func TestList(t *testing.T) {
	type args struct {
		page     uint
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
				page: 1,
			},
			[]model.Ketchup{
				{Pattern: model.DefaultPattern, Version: "1.0.0", Semver: "Patch", Repository: model.Repository{Versions: map[string]string{model.DefaultPattern: "1.0.2"}}},
				{Pattern: model.DefaultPattern, Version: "1.2.3", Repository: model.Repository{Versions: map[string]string{model.DefaultPattern: "1.2.3"}}},
			},
			2,
			nil,
		},
		{
			"error",
			args{
				page: 0,
			},
			nil,
			0,
			httpModel.ErrInternalError,
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			got, gotCount, gotErr := New(testKetchupStore{}, nil).List(context.Background(), tc.args.page, tc.args.pageSize)

			failed := false

			if !errors.Is(gotErr, tc.wantErr) {
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

func TestListForRepositories(t *testing.T) {
	type args struct {
		repositories []model.Repository
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
				repositories: []model.Repository{
					{ID: 1},
					{ID: 2},
				},
			},
			[]model.Ketchup{
				{Pattern: model.DefaultPattern, Version: "1.0.0", Semver: "Patch", Repository: model.Repository{ID: 1, Name: "vibioh/ketchup", Versions: map[string]string{model.DefaultPattern: "1.0.2"}}},
				{Pattern: model.DefaultPattern, Version: "1.2.3", Repository: model.Repository{ID: 2, Name: "vibioh/viws", Versions: map[string]string{model.DefaultPattern: "1.2.3"}}},
			},
			nil,
		},
		{
			"error",
			args{},
			nil,
			httpModel.ErrInternalError,
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			got, gotErr := New(testKetchupStore{}, nil).ListForRepositories(context.Background(), tc.args.repositories)

			failed := false

			if !errors.Is(gotErr, tc.wantErr) {
				failed = true
			} else if !reflect.DeepEqual(got, tc.want) {
				failed = true
			}

			if failed {
				t.Errorf("ListForRepositories() = (%+v, `%s`), want (%+v, `%s`)", got, gotErr, tc.want, tc.wantErr)
			}
		})
	}
}

func TestCreate(t *testing.T) {
	type args struct {
		ctx  context.Context
		item model.Ketchup
	}

	var cases = []struct {
		intention string
		instance  App
		args      args
		want      model.Ketchup
		wantErr   error
	}{
		{
			"start atomic error",
			New(testKetchupStore{}, repositorytest.New()),
			args{
				ctx:  context.TODO(),
				item: model.NoneKetchup,
			},
			model.NoneKetchup,
			errAtomicStart,
		},
		{
			"repository error",
			New(testKetchupStore{}, repositorytest.New()),
			args{
				ctx:  context.Background(),
				item: model.NoneKetchup,
			},
			model.NoneKetchup,
			httpModel.ErrInvalid,
		},
		{
			"check error",
			New(testKetchupStore{}, repositorytest.New()),
			args{
				ctx: context.Background(),
				item: model.Ketchup{
					Repository: model.Repository{ID: 1, Name: "vibioh/ketchup"},
				},
			},
			model.NoneKetchup,
			httpModel.ErrInvalid,
		},
		{
			"create error",
			New(testKetchupStore{}, repositorytest.New()),
			args{
				ctx:  model.StoreUser(context.Background(), model.User{ID: 1}),
				item: model.Ketchup{Pattern: model.DefaultPattern, Version: "0.0.0", Repository: model.Repository{ID: 1, Name: "vibioh/ketchup"}},
			},
			model.NoneKetchup,
			httpModel.ErrInternalError,
		},
		{
			"end atomic error",
			New(testKetchupStore{}, repositorytest.New()),
			args{
				ctx:  model.StoreUser(context.Background(), model.User{ID: 1}),
				item: model.Ketchup{Pattern: model.DefaultPattern, Version: "0", Repository: model.Repository{Name: "vibioh/ketchup"}},
			},
			model.NoneKetchup,
			errAtomicEnd,
		},
		{
			"success",
			New(testKetchupStore{}, repositorytest.New().SetGetOrCreate(model.Repository{
				ID:   1,
				Kind: model.Github,
				Name: "vibioh/ketchup",
				Versions: map[string]string{
					model.DefaultPattern: "1.0.0",
				},
			}, nil),
			),
			args{
				ctx: model.StoreUser(context.Background(), model.User{ID: 1}),
				item: model.Ketchup{
					Pattern: model.DefaultPattern,
					Version: "1.0.0",
					Repository: model.Repository{
						Kind: model.Github,
						Name: "vibioh/ketchup",
					},
				},
			},
			model.Ketchup{
				Pattern: model.DefaultPattern,
				Version: "1.0.0",
				Repository: model.Repository{
					ID:   1,
					Kind: model.Github,
					Name: "vibioh/ketchup",
					Versions: map[string]string{
						model.DefaultPattern: "1.0.0",
					},
				},
			},
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

func TestUpdate(t *testing.T) {
	type args struct {
		ctx  context.Context
		item model.Ketchup
	}

	var cases = []struct {
		intention string
		args      args
		want      model.Ketchup
		wantErr   error
	}{
		{
			"start atomic error",
			args{
				ctx:  context.TODO(),
				item: model.NoneKetchup,
			},
			model.NoneKetchup,
			errAtomicStart,
		},
		{
			"fetch error",
			args{
				ctx:  context.Background(),
				item: model.NoneKetchup,
			},
			model.NoneKetchup,
			httpModel.ErrInternalError,
		},
		{
			"check error",
			args{
				ctx:  context.Background(),
				item: model.Ketchup{Repository: model.Repository{ID: 1}},
			},
			model.NoneKetchup,
			httpModel.ErrInvalid,
		},
		{
			"update error",
			args{
				ctx:  model.StoreUser(context.Background(), model.User{ID: 1}),
				item: model.Ketchup{Pattern: model.DefaultPattern, Version: "0.0.0", Repository: model.Repository{ID: 2}},
			},
			model.NoneKetchup,
			httpModel.ErrInternalError,
		},
		{
			"end atomic error",
			args{
				ctx:  model.StoreUser(context.Background(), model.User{ID: 1}),
				item: model.Ketchup{Pattern: model.DefaultPattern, Version: "0", Repository: model.Repository{ID: 2}},
			},
			model.NoneKetchup,
			errAtomicEnd,
		},
		{
			"success",
			args{
				ctx:  model.StoreUser(context.Background(), model.User{ID: 1}),
				item: model.Ketchup{Pattern: model.DefaultPattern, Version: "1.0.0", Repository: model.Repository{ID: 2}},
			},
			model.Ketchup{Pattern: model.DefaultPattern, Version: "1.0.0", Repository: model.Repository{ID: 2, Name: "vibioh/ketchup", Versions: map[string]string{model.DefaultPattern: "1.2.3"}}, User: model.User{ID: 1}},
			nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			got, gotErr := New(testKetchupStore{}, repositorytest.New()).Update(tc.args.ctx, tc.args.item)

			failed := false

			if !errors.Is(gotErr, tc.wantErr) {
				failed = true
			} else if !reflect.DeepEqual(got, tc.want) {
				failed = true
			}

			if failed {
				t.Errorf("Update() = (%+v, `%s`), want (%+v, `%s`)", got, gotErr, tc.want, tc.wantErr)
			}
		})
	}
}

func TestDelete(t *testing.T) {
	type args struct {
		ctx  context.Context
		item model.Ketchup
	}

	var cases = []struct {
		intention string
		args      args
		wantErr   error
	}{
		{
			"start atomic error",
			args{
				ctx:  context.TODO(),
				item: model.NoneKetchup,
			},
			errAtomicStart,
		},
		{
			"fetch error",
			args{
				ctx:  context.Background(),
				item: model.Ketchup{Repository: model.Repository{ID: 0}},
			},
			httpModel.ErrInternalError,
		},
		{
			"check error",
			args{
				ctx:  context.Background(),
				item: model.Ketchup{Repository: model.Repository{ID: 1}},
			},
			httpModel.ErrInvalid,
		},
		{
			"delete error",
			args{
				ctx:  model.StoreUser(context.Background(), model.User{ID: 1}),
				item: model.Ketchup{Repository: model.Repository{ID: 3}},
			},
			httpModel.ErrInternalError,
		},
		{
			"end atomic error",
			args{
				ctx:  model.StoreUser(context.Background(), model.User{ID: 1}),
				item: model.Ketchup{Repository: model.Repository{ID: 4}},
			},
			errAtomicEnd,
		},
		{
			"success",
			args{
				ctx:  model.StoreUser(context.Background(), model.User{ID: 1}),
				item: model.Ketchup{Repository: model.Repository{ID: 1}},
			},
			nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			gotErr := New(testKetchupStore{}, repositorytest.New()).Delete(tc.args.ctx, tc.args.item)

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

func TestCheck(t *testing.T) {
	type args struct {
		ctx context.Context
		old model.Ketchup
		new model.Ketchup
	}

	var cases = []struct {
		intention string
		args      args
		wantErr   error
	}{
		{
			"no user",
			args{
				ctx: context.Background(),
			},
			errors.New("you must be logged in for interacting"),
		},
		{
			"delete",
			args{
				ctx: model.StoreUser(context.Background(), model.User{ID: 1}),
				old: model.Ketchup{Version: "1.0.0"},
				new: model.NoneKetchup,
			},
			nil,
		},
		{
			"no version",
			args{
				ctx: model.StoreUser(context.Background(), model.User{ID: 1}),
				old: model.Ketchup{Version: "1.0.0"},
				new: model.Ketchup{Version: "", Repository: model.Repository{ID: 1}},
			},
			errors.New("version is required"),
		},
		{
			"create error",
			args{
				ctx: model.StoreUser(context.Background(), model.User{ID: 1}),
				new: model.Ketchup{Version: "1.0.0", Repository: model.Repository{ID: 0}, User: model.User{ID: 1}},
			},
			errors.New("unable to check if ketchup already exists"),
		},
		{
			"create already exists",
			args{
				ctx: model.StoreUser(context.Background(), model.User{ID: 1}),
				new: model.Ketchup{Version: "1.0.0", Repository: model.Repository{ID: 2, Name: "vibioh/ketchup"}, User: model.User{ID: 1}},
			},
			errors.New("ketchup for vibioh/ketchup already exists"),
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			gotErr := app{ketchupStore: testKetchupStore{}}.check(tc.args.ctx, tc.args.old, tc.args.new)

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
