package repository

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	httpModel "github.com/ViBiOh/httputils/v3/pkg/model"
	"github.com/ViBiOh/ketchup/pkg/github/githubtest"
	"github.com/ViBiOh/ketchup/pkg/model"
	"github.com/ViBiOh/ketchup/pkg/semver"
)

var (
	errAtomicStart = errors.New("invalid context")
	errAtomicEnd   = errors.New("invalid context")
)

type testRepositoryStore struct{}

func (trs testRepositoryStore) DoAtomic(ctx context.Context, action func(context.Context) error) error {
	if ctx == context.TODO() {
		return errAtomicStart
	}

	err := action(ctx)
	if err != nil && strings.Contains(err.Error(), "duplicate pk") {
		return errAtomicEnd
	}

	return err
}

func (trs testRepositoryStore) List(_ context.Context, page, _ uint) ([]model.Repository, uint64, error) {
	if page == 0 {
		return nil, 0, errors.New("invalid page")
	}

	return []model.Repository{
		model.NewRepository(1, model.Github, "vibioh/ketchup").AddVersion(model.DefaultPattern, "1.0.0"),
		model.NewRepository(2, model.Github, "vibioh/viws").AddVersion(model.DefaultPattern, "1.2.3"),
	}, 2, nil
}

func (trs testRepositoryStore) Suggest(ctx context.Context, _ []uint64, _ uint64) ([]model.Repository, error) {
	if ctx == context.TODO() {
		return nil, errors.New("invalid context")
	}

	return []model.Repository{
		model.NewRepository(2, model.Github, "vibioh/viws").AddVersion(model.DefaultPattern, "1.2.3"),
	}, nil
}

func (trs testRepositoryStore) Get(_ context.Context, id uint64, _ bool) (model.Repository, error) {
	if id == 0 {
		return model.NoneRepository, errors.New("invalid id")
	}

	return model.NewRepository(id, model.Github, "vibioh/ketchup").AddVersion(model.DefaultPattern, "1.0.0"), nil
}

func (trs testRepositoryStore) GetByName(_ context.Context, name string, repositoryKind model.RepositoryKind) (model.Repository, error) {
	if name == "error" {
		return model.NoneRepository, errors.New("invalid name")
	}

	if name == "exist" {
		return model.NewRepository(3, model.Github, "vibioh/ketchup"), nil
	}

	return model.NoneRepository, nil
}

func (trs testRepositoryStore) Create(_ context.Context, o model.Repository) (uint64, error) {
	if o.Name == "vibioh" {
		return 0, errors.New("invalid name")
	}

	return 1, nil
}

func (trs testRepositoryStore) UpdateVersions(_ context.Context, o model.Repository) error {
	if o.ID == 1 {
		return errors.New("invalid id")
	}

	if o.ID == 2 {
		return errors.New("duplicate pk")
	}

	return nil
}

func (trs testRepositoryStore) DeleteUnused(ctx context.Context) error {
	if model.ReadUser(ctx) == model.NoneUser {
		return errors.New("no user found")
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
		want      []model.Repository
		wantCount uint64
		wantErr   error
	}{
		{
			"simple",
			args{
				page: 1,
			},
			[]model.Repository{
				model.NewRepository(1, model.Github, "vibioh/ketchup").AddVersion(model.DefaultPattern, "1.0.0"),
				model.NewRepository(2, model.Github, "vibioh/viws").AddVersion(model.DefaultPattern, "1.2.3"),
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
			got, gotCount, gotErr := New(testRepositoryStore{}, nil, nil).List(context.Background(), tc.args.page, tc.args.pageSize)

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

func TestSuggest(t *testing.T) {
	type args struct {
		ctx       context.Context
		ignoreIds []uint64
		count     uint64
	}

	var cases = []struct {
		intention string
		args      args
		want      []model.Repository
		wantErr   error
	}{
		{
			"simple",
			args{
				ctx: context.Background(),
			},
			[]model.Repository{
				model.NewRepository(2, model.Github, "vibioh/viws").AddVersion(model.DefaultPattern, "1.2.3"),
			},
			nil,
		},
		{
			"error",
			args{
				ctx: context.TODO(),
			},
			nil,
			httpModel.ErrInternalError,
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			got, gotErr := New(testRepositoryStore{}, nil, nil).Suggest(tc.args.ctx, tc.args.ignoreIds, tc.args.count)

			failed := false

			if !errors.Is(gotErr, tc.wantErr) {
				failed = true
			} else if !reflect.DeepEqual(got, tc.want) {
				failed = true
			}

			if failed {
				t.Errorf("Suggest() = (%+v,`%s`), want (%+v,`%s`)", got, gotErr, tc.want, tc.wantErr)
			}
		})
	}
}

func TestGetOrCreate(t *testing.T) {
	type args struct {
		ctx            context.Context
		name           string
		repositoryKind model.RepositoryKind
		pattern        string
	}

	var cases = []struct {
		intention string
		instance  App
		args      args
		want      model.Repository
		wantErr   error
	}{
		{
			"get error",
			New(testRepositoryStore{}, githubtest.New(), nil),
			args{
				ctx:            context.Background(),
				name:           "error",
				repositoryKind: model.Github,
				pattern:        model.DefaultPattern,
			},
			model.NoneRepository,
			httpModel.ErrInternalError,
		},
		{
			"exists but no pattern",
			New(testRepositoryStore{}, githubtest.New().SetLatestVersions(map[string]semver.Version{
				model.DefaultPattern: {
					Major: 1,
					Name:  "1.0.0",
				},
			}, nil), nil),
			args{
				ctx:            context.Background(),
				name:           "exist",
				repositoryKind: model.Github,
				pattern:        model.DefaultPattern,
			},
			model.NewRepository(3, model.Github, "vibioh/ketchup").AddVersion(model.DefaultPattern, "1.0.0"),
			nil,
		},
		{
			"create",
			New(testRepositoryStore{}, githubtest.New().SetLatestVersions(map[string]semver.Version{
				model.DefaultPattern: {
					Major: 1,
					Name:  "1.0.0",
				},
			}, nil), nil),
			args{
				ctx:            context.Background(),
				name:           "not found",
				repositoryKind: model.Github,
				pattern:        model.DefaultPattern,
			},
			model.NewRepository(1, model.Github, "not found").AddVersion(model.DefaultPattern, "1.0.0"),
			nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			got, gotErr := tc.instance.GetOrCreate(tc.args.ctx, tc.args.name, tc.args.repositoryKind, tc.args.pattern)

			failed := false

			if !errors.Is(gotErr, tc.wantErr) {
				failed = true
			} else if !reflect.DeepEqual(got, tc.want) {
				failed = true
			}

			if failed {
				t.Errorf("GetOrCreate() = (%+v, `%s`), want (%+v, `%s`)", got, gotErr, tc.want, tc.wantErr)
			}
		})
	}
}

func TestCreate(t *testing.T) {
	type args struct {
		ctx  context.Context
		item model.Repository
	}

	var cases = []struct {
		intention string
		instance  app
		args      args
		want      model.Repository
		wantErr   error
	}{
		{
			"invalid",
			app{
				repositoryStore: testRepositoryStore{},
				githubApp:       githubtest.New(),
			},
			args{
				ctx:  context.Background(),
				item: model.NewRepository(1, model.Github, ""),
			},
			model.NoneRepository,
			httpModel.ErrInvalid,
		},
		{
			"release error",
			app{
				repositoryStore: testRepositoryStore{},
				githubApp:       githubtest.New().SetLatestVersions(nil, errors.New("failed")),
			},
			args{
				ctx:  context.Background(),
				item: model.NewRepository(1, model.Github, "invalid"),
			},
			model.NoneRepository,
			httpModel.ErrNotFound,
		},
		{
			"create error",
			app{
				repositoryStore: testRepositoryStore{},
				githubApp: githubtest.New().SetLatestVersions(map[string]semver.Version{
					model.DefaultPattern: {
						Major: 1,
						Name:  "1.0.0",
					},
				}, nil),
			},
			args{
				ctx:  context.Background(),
				item: model.NewRepository(1, model.Github, "vibioh").AddVersion(model.DefaultPattern, "0.0.0"),
			},
			model.NewRepository(1, model.Github, "vibioh").AddVersion(model.DefaultPattern, "1.0.0"),
			httpModel.ErrInternalError,
		},
		{
			"success",
			app{
				repositoryStore: testRepositoryStore{},
				githubApp: githubtest.New().SetLatestVersions(map[string]semver.Version{
					model.DefaultPattern: {
						Major: 1,
						Name:  "1.0.0",
					},
				}, nil),
			},
			args{
				ctx:  context.Background(),
				item: model.NewRepository(1, model.Github, "vibioh/ketchup").AddVersion(model.DefaultPattern, "0.0.0"),
			},
			model.NewRepository(1, model.Github, "vibioh/ketchup").AddVersion(model.DefaultPattern, "1.0.0"),
			nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			got, gotErr := tc.instance.create(tc.args.ctx, tc.args.item)

			failed := false

			if !errors.Is(gotErr, tc.wantErr) {
				failed = true
			} else if !reflect.DeepEqual(got, tc.want) {
				failed = true
			}

			if failed {
				t.Errorf("create() = (%+v, `%s`), want (%+v, `%s`)", got, gotErr, tc.want, tc.wantErr)
			}
		})
	}
}

func TestUpdate(t *testing.T) {
	type args struct {
		ctx  context.Context
		item model.Repository
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
				item: model.NoneRepository,
			},
			errAtomicStart,
		},
		{
			"fetch error",
			args{
				ctx:  context.Background(),
				item: model.NewRepository(0, model.Github, ""),
			},
			httpModel.ErrInternalError,
		},
		{
			"invalid check",
			args{
				ctx:  context.Background(),
				item: model.NewRepository(1, model.Github, ""),
			},
			httpModel.ErrInvalid,
		},
		{
			"update error",
			args{
				ctx:  context.Background(),
				item: model.NewRepository(1, model.Github, "").AddVersion(model.DefaultPattern, "1.2.3"),
			},
			httpModel.ErrInternalError,
		},
		{
			"end atomic error",
			args{
				ctx:  context.Background(),
				item: model.NewRepository(2, model.Github, "").AddVersion(model.DefaultPattern, "1.2.3"),
			},
			errAtomicEnd,
		},
		{
			"success",
			args{
				ctx:  context.Background(),
				item: model.NewRepository(3, model.Github, "").AddVersion(model.DefaultPattern, "1.2.3"),
			},
			nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			gotErr := New(testRepositoryStore{}, nil, nil).Update(tc.args.ctx, tc.args.item)

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

func TestClean(t *testing.T) {
	type args struct {
		ctx context.Context
	}

	var cases = []struct {
		intention string
		args      args
		wantErr   error
	}{
		{
			"error",
			args{
				ctx: context.Background(),
			},
			httpModel.ErrInternalError,
		},
		{
			"success",
			args{
				ctx: model.StoreUser(context.Background(), model.User{ID: 1}),
			},
			nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			gotErr := New(testRepositoryStore{}, nil, nil).Clean(tc.args.ctx)

			failed := false

			if !errors.Is(gotErr, tc.wantErr) {
				failed = true
			}

			if failed {
				t.Errorf("Clean() = `%s`, want `%s`", gotErr, tc.wantErr)
			}
		})
	}
}

func TestCheck(t *testing.T) {
	type args struct {
		ctx context.Context
		old model.Repository
		new model.Repository
	}

	var cases = []struct {
		intention string
		args      args
		wantErr   error
	}{
		{
			"delete",
			args{
				old: model.NewRepository(1, model.Github, ""),
			},
			nil,
		},
		{
			"name required",
			args{
				new: model.NewRepository(1, model.Github, "").AddVersion(model.DefaultPattern, "1.0.0"),
			},
			errors.New("name is required"),
		},
		{
			"version required for update",
			args{
				old: model.NewRepository(1, model.Github, "vibioh/ketchup").AddVersion(model.DefaultPattern, "1.0.0"),
				new: model.NewRepository(1, model.Github, "vibioh/ketchup"),
			},
			errors.New("version is required"),
		},
		{
			"get error",
			args{
				new: model.NewRepository(1, model.Github, "error"),
			},
			errors.New("unable to check if name already exists"),
		},
		{
			"exist",
			args{
				new: model.NewRepository(1, model.Github, "exist"),
			},
			errors.New("name already exists"),
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			gotErr := app{repositoryStore: testRepositoryStore{}}.check(tc.args.ctx, tc.args.old, tc.args.new)

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

func TestSanitizeName(t *testing.T) {
	type args struct {
		name string
	}

	var cases = []struct {
		intention string
		args      args
		want      string
	}{
		{
			"nothing to do",
			args{
				name: "test",
			},
			"test",
		},
		{
			"domain prefix",
			args{
				name: "github.com/vibioh/ketchup",
			},
			"vibioh/ketchup",
		},
		{
			"full url",
			args{
				name: "https://github.com/vibioh/ketchup",
			},
			"vibioh/ketchup",
		},
		{
			"with suffix",
			args{
				name: "https://github.com/vibioh/ketchup/releases/latest",
			},
			"vibioh/ketchup",
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			if got := sanitizeName(tc.args.name); got != tc.want {
				t.Errorf("sanitizeName() = `%s`, want `%s`", got, tc.want)
			}
		})
	}
}
