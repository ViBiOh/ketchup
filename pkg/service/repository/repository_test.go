package repository

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"

	httpModel "github.com/ViBiOh/httputils/v4/pkg/model"
	"github.com/ViBiOh/ketchup/pkg/mocks"
	"github.com/ViBiOh/ketchup/pkg/model"
	"github.com/ViBiOh/ketchup/pkg/semver"
	"go.uber.org/mock/gomock"
)

var (
	ketchupRepository = "vibioh/ketchup"
	viwsRepository    = "vibioh/viws"
	chartRepository   = "https://charts.vibioh.fr"

	errAtomicStart = errors.New("invalid context")
)

func safeParse(version string) semver.Version {
	output, err := semver.Parse(version)
	if err != nil {
		fmt.Println(err)
	}
	return output
}

func TestList(t *testing.T) {
	t.Parallel()

	type args struct {
		pageSize uint
		last     string
	}

	cases := map[string]struct {
		args      args
		want      []model.Repository
		wantCount uint64
		wantErr   error
	}{
		"simple": {
			args{},
			[]model.Repository{
				model.NewGithubRepository(model.Identifier(1), ketchupRepository).AddVersion(model.DefaultPattern, "1.0.0"),
				model.NewGithubRepository(model.Identifier(2), viwsRepository).AddVersion(model.DefaultPattern, "1.2.3"),
			},
			2,
			nil,
		},
		"error": {
			args{},
			nil,
			0,
			httpModel.ErrInternalError,
		},
	}

	for intention, testCase := range cases {
		intention, testCase := intention, testCase

		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepositoryStore := mocks.NewRepositoryStore(ctrl)

			instance := App{
				repositoryStore: mockRepositoryStore,
			}

			switch intention {
			case "simple":
				mockRepositoryStore.EXPECT().List(gomock.Any(), gomock.Any(), gomock.Any()).Return([]model.Repository{
					model.NewGithubRepository(model.Identifier(1), ketchupRepository).AddVersion(model.DefaultPattern, "1.0.0"),
					model.NewGithubRepository(model.Identifier(2), viwsRepository).AddVersion(model.DefaultPattern, "1.2.3"),
				}, uint64(2), nil)
			case "error":
				mockRepositoryStore.EXPECT().List(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, uint64(0), errors.New("failed"))
			}

			got, gotCount, gotErr := instance.List(context.TODO(), testCase.args.pageSize, testCase.args.last)

			failed := false

			if !errors.Is(gotErr, testCase.wantErr) {
				failed = true
			} else if !reflect.DeepEqual(got, testCase.want) {
				failed = true
			} else if gotCount != testCase.wantCount {
				failed = true
			}

			if failed {
				t.Errorf("List() = (%+v, %d, `%s`), want (%+v, %d, `%s`)", got, gotCount, gotErr, testCase.want, testCase.wantCount, testCase.wantErr)
			}
		})
	}
}

func TestSuggest(t *testing.T) {
	t.Parallel()

	type args struct {
		ctx       context.Context
		ignoreIds []model.Identifier
		count     uint64
	}

	cases := map[string]struct {
		args    args
		want    []model.Repository
		wantErr error
	}{
		"simple": {
			args{
				ctx: context.TODO(),
			},
			[]model.Repository{
				model.NewGithubRepository(model.Identifier(1), ketchupRepository).AddVersion(model.DefaultPattern, "1.2.3"),
			},
			nil,
		},
		"error": {
			args{
				ctx: context.TODO(),
			},
			nil,
			httpModel.ErrInternalError,
		},
	}

	for intention, testCase := range cases {
		intention, testCase := intention, testCase

		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepositoryStore := mocks.NewRepositoryStore(ctrl)

			instance := App{
				repositoryStore: mockRepositoryStore,
			}

			switch intention {
			case "simple":
				mockRepositoryStore.EXPECT().Suggest(gomock.Any(), gomock.Any(), gomock.Any()).Return([]model.Repository{
					model.NewGithubRepository(model.Identifier(1), ketchupRepository).AddVersion(model.DefaultPattern, "1.2.3"),
				}, nil)
			case "error":
				mockRepositoryStore.EXPECT().Suggest(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("failed"))
			}

			got, gotErr := instance.Suggest(testCase.args.ctx, testCase.args.ignoreIds, testCase.args.count)

			failed := false

			if !errors.Is(gotErr, testCase.wantErr) {
				failed = true
			} else if !reflect.DeepEqual(got, testCase.want) {
				failed = true
			}

			if failed {
				t.Errorf("Suggest() = (%+v,`%s`), want (%+v,`%s`)", got, gotErr, testCase.want, testCase.wantErr)
			}
		})
	}
}

func TestGetOrCreate(t *testing.T) {
	t.Parallel()

	type args struct {
		ctx            context.Context
		repositoryKind model.RepositoryKind
		name           string
		part           string
		pattern        string
	}

	cases := map[string]struct {
		args    args
		want    model.Repository
		wantErr error
	}{
		"get error": {
			args{
				ctx:            context.TODO(),
				name:           "error",
				repositoryKind: model.Github,
				pattern:        model.DefaultPattern,
			},
			model.NewEmptyRepository(),
			httpModel.ErrInternalError,
		},
		"exists with pattern": {
			args{
				ctx:            context.TODO(),
				name:           "exist",
				repositoryKind: model.Github,
				pattern:        model.DefaultPattern,
			},
			model.NewGithubRepository(model.Identifier(1), ketchupRepository).AddVersion(model.DefaultPattern, "1.0.0"),
			nil,
		},
		"exists no pattern error": {
			args{
				ctx:            context.TODO(),
				name:           "exist",
				repositoryKind: model.Github,
				pattern:        model.DefaultPattern,
			},
			model.NewEmptyRepository(),
			httpModel.ErrInternalError,
		},
		"exists pattern not found": {
			args{
				ctx:            context.TODO(),
				name:           "exist",
				repositoryKind: model.Github,
				pattern:        model.DefaultPattern,
			},
			model.NewEmptyRepository(),
			httpModel.ErrNotFound,
		},
		"exists but no pattern": {
			args{
				ctx:            context.TODO(),
				name:           "exist",
				repositoryKind: model.Github,
				pattern:        model.DefaultPattern,
			},
			model.NewGithubRepository(model.Identifier(1), ketchupRepository).AddVersion(model.DefaultPattern, "1.0.0"),
			nil,
		},
		"update error": {
			args{
				ctx:            context.TODO(),
				name:           "exist",
				repositoryKind: model.Helm,
				pattern:        model.DefaultPattern,
			},
			model.NewEmptyRepository(),
			httpModel.ErrInternalError,
		},
		"create": {
			args{
				ctx:            context.TODO(),
				name:           "not found",
				repositoryKind: model.Github,
				pattern:        model.DefaultPattern,
			},
			model.NewGithubRepository(model.Identifier(1), "not found").AddVersion(model.DefaultPattern, "1.0.0"),
			nil,
		},
	}

	for intention, testCase := range cases {
		intention, testCase := intention, testCase

		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepositoryStore := mocks.NewRepositoryStore(ctrl)
			mockGithub := mocks.NewGenericProvider(ctrl)

			instance := App{
				repositoryStore: mockRepositoryStore,
				githubApp:       mockGithub,
			}

			switch intention {
			case "get error":
				mockRepositoryStore.EXPECT().GetByName(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(model.NewEmptyRepository(), errors.New("failed"))
			case "exists with pattern":
				mockRepositoryStore.EXPECT().GetByName(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(model.NewGithubRepository(model.Identifier(1), ketchupRepository).AddVersion(model.DefaultPattern, "1.0.0"), nil)
			case "exists no pattern error":
				mockRepositoryStore.EXPECT().GetByName(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(model.NewGithubRepository(model.Identifier(1), ketchupRepository), nil)
				mockGithub.EXPECT().LatestVersions(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("failed"))
			case "exists pattern not found":
				mockRepositoryStore.EXPECT().GetByName(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(model.NewGithubRepository(model.Identifier(1), ketchupRepository), nil)
				mockGithub.EXPECT().LatestVersions(gomock.Any(), gomock.Any(), gomock.Any()).Return(map[string]semver.Version{
					"latest": safeParse("1.0.0"),
				}, nil)
			case "exists but no pattern":
				mockRepositoryStore.EXPECT().GetByName(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(model.NewGithubRepository(model.Identifier(1), ketchupRepository), nil)
				mockRepositoryStore.EXPECT().UpdateVersions(gomock.Any(), gomock.Any()).Return(nil)
				mockGithub.EXPECT().LatestVersions(gomock.Any(), gomock.Any(), gomock.Any()).Return(map[string]semver.Version{
					model.DefaultPattern: safeParse("1.0.0"),
				}, nil)
			case "update error":
				mockRepositoryStore.EXPECT().GetByName(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(model.NewGithubRepository(model.Identifier(1), ketchupRepository), nil)
				mockRepositoryStore.EXPECT().UpdateVersions(gomock.Any(), gomock.Any()).Return(errors.New("failed"))
				mockGithub.EXPECT().LatestVersions(gomock.Any(), gomock.Any(), gomock.Any()).Return(map[string]semver.Version{
					model.DefaultPattern: safeParse("1.0.0"),
				}, nil)
			case "create":
				mockRepositoryStore.EXPECT().GetByName(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(model.NewEmptyRepository(), nil)
				mockRepositoryStore.EXPECT().Create(gomock.Any(), gomock.Any()).Return(model.Identifier(1), nil)
				mockRepositoryStore.EXPECT().GetByName(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(model.NewEmptyRepository(), nil)
				dummyFn := func(ctx context.Context, do func(ctx context.Context) error) error {
					return do(ctx)
				}
				mockRepositoryStore.EXPECT().DoAtomic(gomock.Any(), gomock.Any()).DoAndReturn(dummyFn)
				mockGithub.EXPECT().LatestVersions(gomock.Any(), gomock.Any(), gomock.Any()).Return(map[string]semver.Version{
					model.DefaultPattern: safeParse("1.0.0"),
				}, nil)
			}

			got, gotErr := instance.GetOrCreate(testCase.args.ctx, testCase.args.repositoryKind, testCase.args.name, testCase.args.part, testCase.args.pattern)

			failed := false

			if !errors.Is(gotErr, testCase.wantErr) {
				failed = true
			} else if !reflect.DeepEqual(got, testCase.want) {
				failed = true
			}

			if failed {
				t.Errorf("GetOrCreate() = (%+v, `%s`), want (%+v, `%s`)", got, gotErr, testCase.want, testCase.wantErr)
			}
		})
	}
}

func TestCreate(t *testing.T) {
	t.Parallel()

	type args struct {
		ctx  context.Context
		item model.Repository
	}

	cases := map[string]struct {
		args    args
		want    model.Repository
		wantErr error
	}{
		"invalid": {
			args{
				ctx:  context.TODO(),
				item: model.NewGithubRepository(model.Identifier(1), ""),
			},
			model.NewEmptyRepository(),
			httpModel.ErrInvalid,
		},
		"release error": {
			args{
				ctx:  context.TODO(),
				item: model.NewGithubRepository(model.Identifier(1), "invalid"),
			},
			model.NewEmptyRepository(),
			httpModel.ErrNotFound,
		},
		"create error": {
			args{
				ctx:  context.TODO(),
				item: model.NewGithubRepository(model.Identifier(1), "vibioh").AddVersion(model.DefaultPattern, "0.0.0"),
			},
			model.NewGithubRepository(model.Identifier(1), "vibioh").AddVersion(model.DefaultPattern, "1.0.0"),
			httpModel.ErrInternalError,
		},
		"success": {
			args{
				ctx:  context.TODO(),
				item: model.NewGithubRepository(model.Identifier(1), ketchupRepository).AddVersion(model.DefaultPattern, "0.0.0"),
			},
			model.NewGithubRepository(model.Identifier(1), ketchupRepository).AddVersion(model.DefaultPattern, "1.0.0"),
			nil,
		},
	}

	for intention, testCase := range cases {
		intention, testCase := intention, testCase

		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepositoryStore := mocks.NewRepositoryStore(ctrl)
			mockGithub := mocks.NewGenericProvider(ctrl)

			instance := App{
				repositoryStore: mockRepositoryStore,
				githubApp:       mockGithub,
			}

			switch intention {
			case "invalid":
				mockRepositoryStore.EXPECT().GetByName(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(model.NewEmptyRepository(), nil)
			case "release error":
				mockRepositoryStore.EXPECT().GetByName(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(model.NewEmptyRepository(), nil)
			case "create error":
				mockRepositoryStore.EXPECT().GetByName(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(model.NewEmptyRepository(), nil)
				dummyFn := func(ctx context.Context, do func(ctx context.Context) error) error {
					return do(ctx)
				}
				mockRepositoryStore.EXPECT().DoAtomic(gomock.Any(), gomock.Any()).DoAndReturn(dummyFn)
				mockRepositoryStore.EXPECT().Create(gomock.Any(), gomock.Any()).Return(model.Identifier(0), errors.New("failed"))
				mockGithub.EXPECT().LatestVersions(gomock.Any(), gomock.Any(), gomock.Any()).Return(map[string]semver.Version{
					model.DefaultPattern: safeParse("1.0.0"),
				}, nil)
			case "success":
				mockRepositoryStore.EXPECT().GetByName(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(model.NewEmptyRepository(), nil)
				dummyFn := func(ctx context.Context, do func(ctx context.Context) error) error {
					return do(ctx)
				}
				mockRepositoryStore.EXPECT().DoAtomic(gomock.Any(), gomock.Any()).DoAndReturn(dummyFn)
				mockRepositoryStore.EXPECT().Create(gomock.Any(), gomock.Any()).Return(model.Identifier(1), nil)
				mockGithub.EXPECT().LatestVersions(gomock.Any(), gomock.Any(), gomock.Any()).Return(map[string]semver.Version{
					model.DefaultPattern: safeParse("1.0.0"),
				}, nil)
			}

			got, gotErr := instance.create(testCase.args.ctx, testCase.args.item)

			failed := false

			if !errors.Is(gotErr, testCase.wantErr) {
				failed = true
			} else if !reflect.DeepEqual(got, testCase.want) {
				failed = true
			}

			if failed {
				t.Errorf("create() = (%+v, `%s`), want (%+v, `%s`)", got, gotErr, testCase.want, testCase.wantErr)
			}
		})
	}
}

func TestUpdate(t *testing.T) {
	t.Parallel()

	type args struct {
		ctx  context.Context
		item model.Repository
	}

	cases := map[string]struct {
		args    args
		wantErr error
	}{
		"start atomic error": {
			args{
				ctx:  context.TODO(),
				item: model.NewEmptyRepository(),
			},
			errAtomicStart,
		},
		"fetch error": {
			args{
				ctx:  context.TODO(),
				item: model.NewGithubRepository(model.Identifier(0), ""),
			},
			httpModel.ErrInternalError,
		},
		"invalid check": {
			args{
				ctx:  context.TODO(),
				item: model.NewGithubRepository(model.Identifier(1), ""),
			},
			httpModel.ErrInvalid,
		},
		"update error": {
			args{
				ctx:  context.TODO(),
				item: model.NewGithubRepository(model.Identifier(1), "").AddVersion(model.DefaultPattern, "1.2.3"),
			},
			httpModel.ErrInternalError,
		},
		"success": {
			args{
				ctx:  context.TODO(),
				item: model.NewGithubRepository(model.Identifier(3), "").AddVersion(model.DefaultPattern, "1.2.3"),
			},
			nil,
		},
	}

	for intention, testCase := range cases {
		intention, testCase := intention, testCase

		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepositoryStore := mocks.NewRepositoryStore(ctrl)

			instance := App{
				repositoryStore: mockRepositoryStore,
			}

			switch intention {
			case "start atomic error":
				mockRepositoryStore.EXPECT().DoAtomic(gomock.Any(), gomock.Any()).Return(errAtomicStart)
			case "fetch error":
				mockRepositoryStore.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any()).Return(model.NewEmptyRepository(), errors.New("failed"))
				dummyFn := func(ctx context.Context, do func(ctx context.Context) error) error {
					return do(ctx)
				}
				mockRepositoryStore.EXPECT().DoAtomic(gomock.Any(), gomock.Any()).DoAndReturn(dummyFn)
			case "invalid check":
				mockRepositoryStore.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any()).Return(model.NewGithubRepository(model.Identifier(1), ketchupRepository), nil)
				mockRepositoryStore.EXPECT().GetByName(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(model.NewEmptyRepository(), nil)
				dummyFn := func(ctx context.Context, do func(ctx context.Context) error) error {
					return do(ctx)
				}
				mockRepositoryStore.EXPECT().DoAtomic(gomock.Any(), gomock.Any()).DoAndReturn(dummyFn)
			case "update error":
				mockRepositoryStore.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any()).Return(model.NewGithubRepository(model.Identifier(1), ketchupRepository), nil)
				mockRepositoryStore.EXPECT().GetByName(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(model.NewEmptyRepository(), nil)
				mockRepositoryStore.EXPECT().UpdateVersions(gomock.Any(), gomock.Any()).Return(errors.New("failed"))
				dummyFn := func(ctx context.Context, do func(ctx context.Context) error) error {
					return do(ctx)
				}
				mockRepositoryStore.EXPECT().DoAtomic(gomock.Any(), gomock.Any()).DoAndReturn(dummyFn)
			case "success":
				mockRepositoryStore.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any()).Return(model.NewGithubRepository(model.Identifier(1), ketchupRepository), nil)
				mockRepositoryStore.EXPECT().GetByName(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(model.NewEmptyRepository(), nil)
				mockRepositoryStore.EXPECT().UpdateVersions(gomock.Any(), gomock.Any()).Return(nil)
				dummyFn := func(ctx context.Context, do func(ctx context.Context) error) error {
					return do(ctx)
				}
				mockRepositoryStore.EXPECT().DoAtomic(gomock.Any(), gomock.Any()).DoAndReturn(dummyFn)
			}

			gotErr := instance.Update(testCase.args.ctx, testCase.args.item)

			failed := false

			if !errors.Is(gotErr, testCase.wantErr) {
				failed = true
			}

			if failed {
				t.Errorf("Update() = `%s`, want `%s`", gotErr, testCase.wantErr)
			}
		})
	}
}

func TestClean(t *testing.T) {
	t.Parallel()

	type args struct {
		ctx context.Context
	}

	cases := map[string]struct {
		args    args
		wantErr error
	}{
		"error": {
			args{
				ctx: context.TODO(),
			},
			httpModel.ErrInternalError,
		},
		"error versions": {
			args{
				ctx: context.TODO(),
			},
			httpModel.ErrInternalError,
		},
		"success": {
			args{
				ctx: context.TODO(),
			},
			nil,
		},
	}

	for intention, testCase := range cases {
		intention, testCase := intention, testCase

		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepositoryStore := mocks.NewRepositoryStore(ctrl)

			instance := App{
				repositoryStore: mockRepositoryStore,
			}

			switch intention {
			case "error":
				dummyFn := func(ctx context.Context, do func(ctx context.Context) error) error {
					return do(ctx)
				}
				mockRepositoryStore.EXPECT().DoAtomic(gomock.Any(), gomock.Any()).DoAndReturn(dummyFn)
				mockRepositoryStore.EXPECT().DeleteUnused(gomock.Any()).Return(errors.New("failed"))
			case "error versions":
				dummyFn := func(ctx context.Context, do func(ctx context.Context) error) error {
					return do(ctx)
				}
				mockRepositoryStore.EXPECT().DoAtomic(gomock.Any(), gomock.Any()).DoAndReturn(dummyFn)
				mockRepositoryStore.EXPECT().DeleteUnused(gomock.Any()).Return(nil)
				mockRepositoryStore.EXPECT().DeleteUnusedVersions(gomock.Any()).Return(errors.New("failed"))
			case "success":
				dummyFn := func(ctx context.Context, do func(ctx context.Context) error) error {
					return do(ctx)
				}
				mockRepositoryStore.EXPECT().DoAtomic(gomock.Any(), gomock.Any()).DoAndReturn(dummyFn)
				mockRepositoryStore.EXPECT().DeleteUnused(gomock.Any()).Return(nil)
				mockRepositoryStore.EXPECT().DeleteUnusedVersions(gomock.Any()).Return(nil)
			}

			gotErr := instance.Clean(testCase.args.ctx)

			failed := false

			if !errors.Is(gotErr, testCase.wantErr) {
				failed = true
			}

			if failed {
				t.Errorf("Clean() = `%s`, want `%s`", gotErr, testCase.wantErr)
			}
		})
	}
}

func TestCheck(t *testing.T) {
	t.Parallel()

	type args struct {
		ctx context.Context
		old model.Repository
		new model.Repository
	}

	cases := map[string]struct {
		args    args
		wantErr error
	}{
		"delete": {
			args{
				old: model.NewGithubRepository(model.Identifier(1), ""),
			},
			nil,
		},
		"name required": {
			args{
				new: model.NewGithubRepository(model.Identifier(1), "").AddVersion(model.DefaultPattern, "1.0.0"),
			},
			errors.New("name is required"),
		},
		"no kind change": {
			args{
				old: model.NewGithubRepository(model.Identifier(1), ketchupRepository).AddVersion(model.DefaultPattern, "1.0.0"),
				new: model.NewHelmRepository(model.Identifier(1), chartRepository, "app"),
			},
			errors.New("kind cannot be changed"),
		},
		"version required for update": {
			args{
				old: model.NewGithubRepository(model.Identifier(1), ketchupRepository).AddVersion(model.DefaultPattern, "1.0.0"),
				new: model.NewGithubRepository(model.Identifier(1), ketchupRepository),
			},
			errors.New("version is required"),
		},
		"get error": {
			args{
				new: model.NewGithubRepository(model.Identifier(1), "error"),
			},
			errors.New("check if name already exists"),
		},
		"exist": {
			args{
				new: model.NewGithubRepository(model.Identifier(1), "exist"),
			},
			errors.New("name already exists"),
		},
	}

	for intention, testCase := range cases {
		intention, testCase := intention, testCase

		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepositoryStore := mocks.NewRepositoryStore(ctrl)

			instance := App{
				repositoryStore: mockRepositoryStore,
			}

			switch intention {
			case "name required":
				mockRepositoryStore.EXPECT().GetByName(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(model.NewEmptyRepository(), nil)
			case "no kind change":
				mockRepositoryStore.EXPECT().GetByName(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(model.NewEmptyRepository(), nil)
			case "version required for update":
				mockRepositoryStore.EXPECT().GetByName(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(model.NewEmptyRepository(), nil)
			case "get error":
				mockRepositoryStore.EXPECT().GetByName(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(model.NewEmptyRepository(), errors.New("failed"))
			case "exist":
				mockRepositoryStore.EXPECT().GetByName(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(model.NewGithubRepository(model.Identifier(2), ketchupRepository), nil)
			}

			gotErr := instance.check(testCase.args.ctx, testCase.args.old, testCase.args.new)

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

func TestSanitizeName(t *testing.T) {
	t.Parallel()

	type args struct {
		name string
	}

	cases := map[string]struct {
		args args
		want string
	}{
		"nothing to do": {
			args{
				name: "test",
			},
			"test",
		},
		"domain prefix": {
			args{
				name: "github.com/vibioh/ketchup",
			},
			ketchupRepository,
		},
		"full url": {
			args{
				name: "https://github.com/vibioh/ketchup",
			},
			ketchupRepository,
		},
		"with suffix": {
			args{
				name: "https://github.com/vibioh/ketchup/releases/latest",
			},
			ketchupRepository,
		},
	}

	for intention, testCase := range cases {
		intention, testCase := intention, testCase

		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			if got := sanitizeName(testCase.args.name); got != testCase.want {
				t.Errorf("sanitizeName() = `%s`, want `%s`", got, testCase.want)
			}
		})
	}
}
