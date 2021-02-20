package repository

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"

	httpModel "github.com/ViBiOh/httputils/v4/pkg/model"
	"github.com/ViBiOh/ketchup/pkg/github/githubtest"
	"github.com/ViBiOh/ketchup/pkg/helm/helmtest"
	"github.com/ViBiOh/ketchup/pkg/model"
	"github.com/ViBiOh/ketchup/pkg/semver"
	"github.com/ViBiOh/ketchup/pkg/store/repository/repositorytest"
)

var (
	ketchupRepository = "vibioh/ketchup"
	viwsRepository    = "vibioh/viws"

	errAtomicStart = errors.New("invalid context")
	errAtomicEnd   = errors.New("invalid context")
)

func safeParse(version string) semver.Version {
	output, err := semver.Parse(version)
	if err != nil {
		fmt.Println(err)
	}
	return output
}

func TestList(t *testing.T) {
	type args struct {
		page     uint
		pageSize uint
	}

	var cases = []struct {
		intention string
		instance  App
		args      args
		want      []model.Repository
		wantCount uint64
		wantErr   error
	}{
		{
			"simple",
			New(repositorytest.New().SetList([]model.Repository{
				model.NewRepository(1, model.Github, ketchupRepository).AddVersion(model.DefaultPattern, "1.0.0"),
				model.NewRepository(2, model.Github, viwsRepository).AddVersion(model.DefaultPattern, "1.2.3"),
			}, 2, nil), nil, nil),
			args{
				page: 1,
			},
			[]model.Repository{
				model.NewRepository(1, model.Github, ketchupRepository).AddVersion(model.DefaultPattern, "1.0.0"),
				model.NewRepository(2, model.Github, viwsRepository).AddVersion(model.DefaultPattern, "1.2.3"),
			},
			2,
			nil,
		},
		{
			"error",
			New(repositorytest.New().SetList(nil, 0, errors.New("failed")), nil, nil),
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
			got, gotCount, gotErr := tc.instance.List(context.Background(), tc.args.page, tc.args.pageSize)

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
		instance  App
		args      args
		want      []model.Repository
		wantErr   error
	}{
		{
			"simple",
			New(repositorytest.New().SetSuggest([]model.Repository{
				model.NewRepository(1, model.Github, ketchupRepository).AddVersion(model.DefaultPattern, "1.2.3"),
			}, nil), nil, nil),
			args{
				ctx: context.Background(),
			},
			[]model.Repository{
				model.NewRepository(1, model.Github, ketchupRepository).AddVersion(model.DefaultPattern, "1.2.3"),
			},
			nil,
		},
		{
			"error",
			New(repositorytest.New().SetSuggest(nil, errors.New("failed")), nil, nil),
			args{
				ctx: context.Background(),
			},
			nil,
			httpModel.ErrInternalError,
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			got, gotErr := tc.instance.Suggest(tc.args.ctx, tc.args.ignoreIds, tc.args.count)

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
			New(repositorytest.New().SetGetByName(model.NoneRepository, errors.New("failed")), githubtest.New(), nil),
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
			"exists with pattern",
			New(repositorytest.New().SetGetByName(model.NewRepository(1, model.Github, ketchupRepository).AddVersion(model.DefaultPattern, "1.0.0"), nil), githubtest.New(), nil),
			args{
				ctx:            context.Background(),
				name:           "exist",
				repositoryKind: model.Github,
				pattern:        model.DefaultPattern,
			},
			model.NewRepository(1, model.Github, ketchupRepository).AddVersion(model.DefaultPattern, "1.0.0"),
			nil,
		},
		{
			"exists no pattern error",
			New(repositorytest.New().SetGetByName(model.NewRepository(1, model.Github, ketchupRepository), nil), githubtest.New().SetLatestVersions(nil, errors.New("failed")), nil),
			args{
				ctx:            context.Background(),
				name:           "exist",
				repositoryKind: model.Github,
				pattern:        model.DefaultPattern,
			},
			model.NoneRepository,
			httpModel.ErrInternalError,
		},
		{
			"exists pattern not found",
			New(repositorytest.New().SetGetByName(model.NewRepository(1, model.Github, ketchupRepository), nil), githubtest.New().SetLatestVersions(map[string]semver.Version{
				"latest": safeParse("1.0.0"),
			}, nil), nil),
			args{
				ctx:            context.Background(),
				name:           "exist",
				repositoryKind: model.Github,
				pattern:        model.DefaultPattern,
			},
			model.NoneRepository,
			httpModel.ErrNotFound,
		},
		{
			"exists but no pattern",
			New(repositorytest.New().SetGetByName(model.NewRepository(1, model.Github, ketchupRepository), nil), githubtest.New().SetLatestVersions(map[string]semver.Version{
				model.DefaultPattern: safeParse("1.0.0"),
			}, nil), nil),
			args{
				ctx:            context.Background(),
				name:           "exist",
				repositoryKind: model.Github,
				pattern:        model.DefaultPattern,
			},
			model.NewRepository(1, model.Github, ketchupRepository).AddVersion(model.DefaultPattern, "1.0.0"),
			nil,
		},
		{
			"update error",
			New(repositorytest.New().SetGetByName(model.NewRepository(1, model.Helm, ketchupRepository), nil).SetUpdateVersions(errors.New("failed")), githubtest.New(), helmtest.New().SetLatestVersions(map[string]semver.Version{
				model.DefaultPattern: safeParse("1.0.0"),
			}, nil)),
			args{
				ctx:            context.Background(),
				name:           "exist",
				repositoryKind: model.Helm,
				pattern:        model.DefaultPattern,
			},
			model.NoneRepository,
			httpModel.ErrInternalError,
		},
		{
			"create",
			New(repositorytest.New().SetCreate(1, nil), githubtest.New().SetLatestVersions(map[string]semver.Version{
				model.DefaultPattern: safeParse("1.0.0"),
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
				repositoryStore: repositorytest.New(),
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
				repositoryStore: repositorytest.New(),
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
				repositoryStore: repositorytest.New().SetCreate(0, errors.New("failed")),
				githubApp: githubtest.New().SetLatestVersions(map[string]semver.Version{
					model.DefaultPattern: safeParse("1.0.0"),
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
				repositoryStore: repositorytest.New().SetCreate(1, nil),
				githubApp: githubtest.New().SetLatestVersions(map[string]semver.Version{
					model.DefaultPattern: safeParse("1.0.0"),
				}, nil),
			},
			args{
				ctx:  context.Background(),
				item: model.NewRepository(1, model.Github, ketchupRepository).AddVersion(model.DefaultPattern, "0.0.0"),
			},
			model.NewRepository(1, model.Github, ketchupRepository).AddVersion(model.DefaultPattern, "1.0.0"),
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
		instance  App
		args      args
		wantErr   error
	}{
		{
			"start atomic error",
			New(repositorytest.New().SetDoAtomic(errAtomicStart), nil, nil),
			args{
				ctx:  context.TODO(),
				item: model.NoneRepository,
			},
			errAtomicStart,
		},
		{
			"fetch error",
			New(repositorytest.New().SetGet(model.NoneRepository, errors.New("failed")), nil, nil),
			args{
				ctx:  context.Background(),
				item: model.NewRepository(0, model.Github, ""),
			},
			httpModel.ErrInternalError,
		},
		{
			"invalid check",
			New(repositorytest.New().SetGet(model.NewRepository(1, model.Github, ketchupRepository), nil), nil, nil),
			args{
				ctx:  context.Background(),
				item: model.NewRepository(1, model.Github, ""),
			},
			httpModel.ErrInvalid,
		},
		{
			"update error",
			New(repositorytest.New().SetGet(model.NewRepository(1, model.Github, ketchupRepository), nil).SetUpdateVersions(errors.New("failed")), nil, nil),
			args{
				ctx:  context.Background(),
				item: model.NewRepository(1, model.Github, "").AddVersion(model.DefaultPattern, "1.2.3"),
			},
			httpModel.ErrInternalError,
		},
		{
			"success",
			New(repositorytest.New().SetGet(model.NewRepository(1, model.Github, ketchupRepository), nil), nil, nil),
			args{
				ctx:  context.Background(),
				item: model.NewRepository(3, model.Github, "").AddVersion(model.DefaultPattern, "1.2.3"),
			},
			nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			gotErr := tc.instance.Update(tc.args.ctx, tc.args.item)

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
		instance  App
		args      args
		wantErr   error
	}{
		{
			"error",
			New(repositorytest.New().SetDeleteUnused(errors.New("failed")), nil, nil),
			args{
				ctx: context.Background(),
			},
			httpModel.ErrInternalError,
		},
		{
			"error versions",
			New(repositorytest.New().SetDeleteUnusedVersions(errors.New("failed")), nil, nil),
			args{
				ctx: context.Background(),
			},
			httpModel.ErrInternalError,
		},
		{
			"success",
			New(repositorytest.New(), nil, nil),
			args{
				ctx: context.Background(),
			},
			nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			gotErr := tc.instance.Clean(tc.args.ctx)

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
		instance  app
		args      args
		wantErr   error
	}{
		{
			"delete",
			app{repositoryStore: repositorytest.New()},
			args{
				old: model.NewRepository(1, model.Github, ""),
			},
			nil,
		},
		{
			"name required",
			app{repositoryStore: repositorytest.New()},
			args{
				new: model.NewRepository(1, model.Github, "").AddVersion(model.DefaultPattern, "1.0.0"),
			},
			errors.New("name is required"),
		},
		{
			"no kind change",
			app{repositoryStore: repositorytest.New()},
			args{
				old: model.NewRepository(1, model.Github, ketchupRepository).AddVersion(model.DefaultPattern, "1.0.0"),
				new: model.NewRepository(1, model.Helm, ketchupRepository),
			},
			errors.New("kind cannot be changed"),
		},
		{
			"version required for update",
			app{repositoryStore: repositorytest.New()},
			args{
				old: model.NewRepository(1, model.Github, ketchupRepository).AddVersion(model.DefaultPattern, "1.0.0"),
				new: model.NewRepository(1, model.Github, ketchupRepository),
			},
			errors.New("version is required"),
		},
		{
			"get error",
			app{repositoryStore: repositorytest.New().SetGetByName(model.NoneRepository, errors.New("failed"))},
			args{
				new: model.NewRepository(1, model.Github, "error"),
			},
			errors.New("unable to check if name already exists"),
		},
		{
			"exist",
			app{repositoryStore: repositorytest.New().SetGetByName(model.NewRepository(2, model.Github, ketchupRepository), nil)},
			args{
				new: model.NewRepository(1, model.Github, "exist"),
			},
			errors.New("name already exists"),
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
			ketchupRepository,
		},
		{
			"full url",
			args{
				name: "https://github.com/vibioh/ketchup",
			},
			ketchupRepository,
		},
		{
			"with suffix",
			args{
				name: "https://github.com/vibioh/ketchup/releases/latest",
			},
			ketchupRepository,
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
