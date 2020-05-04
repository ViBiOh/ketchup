package ketchup

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/ViBiOh/ketchup/pkg/model"
	"github.com/ViBiOh/ketchup/pkg/service"
)

var (
	errAtomicStart = errors.New("invalid context")
	errAtomicEnd   = errors.New("invalid context")
)

type testKetchupStore struct{}

func (tks testKetchupStore) StartAtomic(ctx context.Context) (context.Context, error) {
	if ctx == context.TODO() {
		return ctx, errAtomicStart
	}

	return ctx, nil
}

func (tks testKetchupStore) EndAtomic(ctx context.Context, err error) error {
	if err != nil && strings.Contains(err.Error(), "duplicate pk") {
		return errAtomicEnd
	}

	return err
}

func (tks testKetchupStore) List(ctx context.Context, page, pageSize uint) ([]model.Ketchup, uint64, error) {
	if page == 0 {
		return nil, 0, errors.New("invalid page size")
	}

	return []model.Ketchup{
		{Version: "1.0.0", Repository: model.Repository{Version: "1.0.2"}},
		{Version: "1.2.3", Repository: model.Repository{Version: "1.2.3"}},
	}, 2, nil
}

func (tks testKetchupStore) ListByRepositoriesID(ctx context.Context, ids []uint64) ([]model.Ketchup, error) {
	if len(ids) == 0 {
		return nil, errors.New("empty request")
	}

	return []model.Ketchup{
		{Version: "1.0.0", Repository: model.Repository{ID: 1, Name: "vibioh/ketchup", Version: "1.0.2"}},
		{Version: "1.2.3", Repository: model.Repository{ID: 2, Name: "vibioh/viws", Version: "1.2.3"}},
	}, nil
}

func (tks testKetchupStore) GetByRepositoryID(ctx context.Context, id uint64, forUpdate bool) (model.Ketchup, error) {
	if id == 0 {
		return model.NoneKetchup, errors.New("invalid id")
	}

	if id == 2 {
		return model.Ketchup{Version: "1.0.0", Repository: model.Repository{ID: 2, Name: "vibioh/ketchup", Version: "1.2.3"}}, nil
	}

	if id == 3 {
		return model.Ketchup{Version: "0.0.0"}, nil
	}

	if id == 4 {
		return model.Ketchup{Version: "0"}, nil
	}

	return model.NoneKetchup, nil
}

func (tks testKetchupStore) Create(ctx context.Context, o model.Ketchup) (uint64, error) {
	if o.Version == "0" {
		return 0, errors.New("duplicate pk")
	}

	if o.Version == "0.0.0" {
		return 0, errors.New("invalid version")
	}

	return 0, nil
}

func (tks testKetchupStore) Update(ctx context.Context, o model.Ketchup) error {
	if o.Version == "0" {
		return errors.New("duplicate pk")
	}

	if o.Version == "0.0.0" {
		return errors.New("invalid version")
	}

	return nil
}

func (tks testKetchupStore) Delete(ctx context.Context, o model.Ketchup) error {
	if o.Version == "0" {
		return errors.New("duplicate pk")
	}

	if o.Version == "0.0.0" {
		return errors.New("invalid version")
	}

	return nil
}

type testRepositoryService struct{}

func (trs testRepositoryService) List(ctx context.Context, page, pageSize uint) ([]model.Repository, uint64, error) {
	return nil, 0, nil
}

func (trs testRepositoryService) GetOrCreate(ctx context.Context, name string) (model.Repository, error) {
	if len(name) == 0 {
		return model.NoneRepository, service.WrapInvalid(errors.New("invalid name"))
	}

	return model.Repository{ID: 1, Name: "vibioh/ketchup", Version: "1.2.3"}, nil
}

func (trs testRepositoryService) Update(ctx context.Context, item model.Repository) error {
	return nil
}

func (trs testRepositoryService) Clean(ctx context.Context) error {
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
				{Version: "1.0.0", Semver: "Patch", Repository: model.Repository{Version: "1.0.2"}},
				{Version: "1.2.3", Repository: model.Repository{Version: "1.2.3"}},
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
			service.ErrInternalError,
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
				{Version: "1.0.0", Semver: "Patch", Repository: model.Repository{ID: 1, Name: "vibioh/ketchup", Version: "1.0.2"}},
				{Version: "1.2.3", Repository: model.Repository{ID: 2, Name: "vibioh/viws", Version: "1.2.3"}},
			},
			nil,
		},
		{
			"error",
			args{},
			nil,
			service.ErrInternalError,
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
			"repository error",
			args{
				ctx:  context.Background(),
				item: model.NoneKetchup,
			},
			model.NoneKetchup,
			service.ErrInvalid,
		},
		{
			"check error",
			args{
				ctx: context.Background(),
				item: model.Ketchup{
					Repository: model.Repository{ID: 1, Name: "vibioh/ketchup"},
				},
			},
			model.NoneKetchup,
			service.ErrInvalid,
		},
		{
			"create error",
			args{
				ctx:  model.StoreUser(context.Background(), model.User{ID: 1}),
				item: model.Ketchup{Version: "0.0.0", Repository: model.Repository{ID: 1, Name: "vibioh/ketchup"}},
			},
			model.NoneKetchup,
			service.ErrInternalError,
		},
		{
			"end atomic error",
			args{
				ctx:  model.StoreUser(context.Background(), model.User{ID: 1}),
				item: model.Ketchup{Version: "0", Repository: model.Repository{Name: "vibioh/ketchup"}},
			},
			model.NoneKetchup,
			errAtomicEnd,
		},
		{
			"success",
			args{
				ctx:  model.StoreUser(context.Background(), model.User{ID: 1}),
				item: model.Ketchup{Version: "1.0.0", Repository: model.Repository{Name: "vibioh/ketchup"}},
			},
			model.Ketchup{Version: "1.0.0", Repository: model.Repository{ID: 1, Name: "vibioh/ketchup", Version: "1.2.3"}},
			nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			got, gotErr := New(testKetchupStore{}, testRepositoryService{}).Create(tc.args.ctx, tc.args.item)

			failed := false

			if !errors.Is(gotErr, tc.wantErr) {
				failed = true
			} else if got != tc.want {
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
			service.ErrInternalError,
		},
		{
			"check error",
			args{
				ctx:  context.Background(),
				item: model.Ketchup{Repository: model.Repository{ID: 1}},
			},
			model.NoneKetchup,
			service.ErrInvalid,
		},
		{
			"update error",
			args{
				ctx:  model.StoreUser(context.Background(), model.User{ID: 1}),
				item: model.Ketchup{Version: "0.0.0", Repository: model.Repository{ID: 2}},
			},
			model.NoneKetchup,
			service.ErrInternalError,
		},
		{
			"end atomic error",
			args{
				ctx:  model.StoreUser(context.Background(), model.User{ID: 1}),
				item: model.Ketchup{Version: "0", Repository: model.Repository{ID: 2}},
			},
			model.NoneKetchup,
			errAtomicEnd,
		},
		{
			"success",
			args{
				ctx:  model.StoreUser(context.Background(), model.User{ID: 1}),
				item: model.Ketchup{Version: "1.0.0", Repository: model.Repository{ID: 2}},
			},
			model.Ketchup{Version: "1.0.0", Repository: model.Repository{ID: 2, Name: "vibioh/ketchup", Version: "1.2.3"}},
			nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			got, gotErr := New(testKetchupStore{}, testRepositoryService{}).Update(tc.args.ctx, tc.args.item)

			failed := false

			if !errors.Is(gotErr, tc.wantErr) {
				failed = true
			} else if got != tc.want {
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
			service.ErrInternalError,
		},
		{
			"check error",
			args{
				ctx:  context.Background(),
				item: model.Ketchup{Repository: model.Repository{ID: 1}},
			},
			service.ErrInvalid,
		},
		{
			"delete error",
			args{
				ctx:  model.StoreUser(context.Background(), model.User{ID: 1}),
				item: model.Ketchup{Repository: model.Repository{ID: 3}},
			},
			service.ErrInternalError,
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
			gotErr := New(testKetchupStore{}, testRepositoryService{}).Delete(tc.args.ctx, tc.args.item)

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
				new: model.Ketchup{Version: "1.0.0", Repository: model.Repository{ID: 0}},
			},
			errors.New("unable to check if ketchup already exists"),
		},
		{
			"create already exists",
			args{
				ctx: model.StoreUser(context.Background(), model.User{ID: 1}),
				new: model.Ketchup{Version: "1.0.0", Repository: model.Repository{ID: 2, Name: "vibioh/ketchup"}},
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
