package ketchup

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
	ketchupRepository = "vibioh/ketchup"
	viwsRepository    = "vibioh/viws"

	errAtomicStart = errors.New("invalid context")
)

func TestList(t *testing.T) {
	t.Parallel()

	type args struct {
		pageSize uint
		last     string
	}

	cases := map[string]struct {
		args    args
		want    []model.Ketchup
		wantErr error
	}{
		"simple": {
			args{},
			[]model.Ketchup{
				{Pattern: model.DefaultPattern, Version: "1.0.0", Frequency: model.Daily, Semver: "Patch", Repository: model.NewGithubRepository(model.Identifier(1), ketchupRepository).AddVersion(model.DefaultPattern, "1.0.2")},
				{Pattern: model.DefaultPattern, Version: "1.2.3", Frequency: model.Daily, Repository: model.NewGithubRepository(model.Identifier(2), viwsRepository).AddVersion(model.DefaultPattern, "1.2.3")},
			},
			nil,
		},
		"error": {
			args{},
			nil,
			httpModel.ErrInternalError,
		},
	}

	for intention, testCase := range cases {
		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			mockKetchupStore := mocks.NewKetchupStore(ctrl)

			instance := Service{
				ketchupStore: mockKetchupStore,
			}

			switch intention {
			case "simple":
				mockKetchupStore.EXPECT().List(gomock.Any(), gomock.Any(), gomock.Any()).Return([]model.Ketchup{
					model.NewKetchup(model.DefaultPattern, "1.2.3", model.Daily, false, model.NewGithubRepository(model.Identifier(2), viwsRepository).AddVersion(model.DefaultPattern, "1.2.3")),
					model.NewKetchup(model.DefaultPattern, "1.0.0", model.Daily, false, model.NewGithubRepository(model.Identifier(1), ketchupRepository).AddVersion(model.DefaultPattern, "1.0.2")),
				}, nil)
			case "error":
				mockKetchupStore.EXPECT().List(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("failed"))
			}

			got, gotErr := instance.List(context.TODO(), testCase.args.pageSize, testCase.args.last)

			failed := false

			if !errors.Is(gotErr, testCase.wantErr) {
				failed = true
			} else if !reflect.DeepEqual(got, testCase.want) {
				failed = true
			}

			if failed {
				t.Errorf("List() = (%+v, `%s`), want (%+v, `%s`)", got, gotErr, testCase.want, testCase.wantErr)
			}
		})
	}
}

func TestListForRepositories(t *testing.T) {
	t.Parallel()

	type args struct {
		repositories []model.Repository
	}

	cases := map[string]struct {
		args    args
		want    []model.Ketchup
		wantErr error
	}{
		"simple": {
			args{
				repositories: []model.Repository{
					model.NewGithubRepository(model.Identifier(1), ""),
					model.NewGithubRepository(model.Identifier(2), ""),
				},
			},
			[]model.Ketchup{
				{Pattern: model.DefaultPattern, Version: "1.0.0", Frequency: model.Daily, Semver: "Patch", Repository: model.NewGithubRepository(model.Identifier(1), ketchupRepository).AddVersion(model.DefaultPattern, "1.0.2")},
				{Pattern: model.DefaultPattern, Version: "1.2.3", Frequency: model.Daily, Repository: model.NewGithubRepository(model.Identifier(2), viwsRepository).AddVersion(model.DefaultPattern, "1.2.3")},
			},
			nil,
		},
		"error": {
			args{},
			nil,
			httpModel.ErrInternalError,
		},
	}

	for intention, testCase := range cases {
		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			mockKetchupStore := mocks.NewKetchupStore(ctrl)

			instance := Service{
				ketchupStore: mockKetchupStore,
			}

			switch intention {
			case "simple":
				mockKetchupStore.EXPECT().ListByRepositoriesIDAndFrequencies(gomock.Any(), gomock.Any(), gomock.Any()).Return([]model.Ketchup{
					model.NewKetchup(model.DefaultPattern, "1.0.0", model.Daily, false, model.NewGithubRepository(model.Identifier(1), ketchupRepository).AddVersion(model.DefaultPattern, "1.0.2")),
					model.NewKetchup(model.DefaultPattern, "1.2.3", model.Daily, false, model.NewGithubRepository(model.Identifier(2), viwsRepository).AddVersion(model.DefaultPattern, "1.2.3")),
				}, nil)
			case "error":
				mockKetchupStore.EXPECT().ListByRepositoriesIDAndFrequencies(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("failed"))
			}

			got, gotErr := instance.ListForRepositories(context.TODO(), testCase.args.repositories, model.Daily)

			failed := false

			if !errors.Is(gotErr, testCase.wantErr) {
				failed = true
			} else if !reflect.DeepEqual(got, testCase.want) {
				failed = true
			}

			if failed {
				t.Errorf("ListForRepositories() = (%+v, `%s`), want (%+v, `%s`)", got, gotErr, testCase.want, testCase.wantErr)
			}
		})
	}
}

func TestCreate(t *testing.T) {
	t.Parallel()

	type args struct {
		ctx  context.Context
		item model.Ketchup
	}

	cases := map[string]struct {
		args    args
		want    model.Ketchup
		wantErr error
	}{
		"start atomic error": {
			args{
				ctx:  context.TODO(),
				item: model.Ketchup{},
			},
			model.Ketchup{},
			errAtomicStart,
		},
		"repository error": {
			args{
				ctx:  context.TODO(),
				item: model.Ketchup{},
			},
			model.Ketchup{},
			httpModel.ErrInvalid,
		},
		"check error": {
			args{
				ctx:  context.TODO(),
				item: model.NewKetchup(model.DefaultPattern, "", model.Daily, false, model.NewGithubRepository(model.Identifier(1), ketchupRepository)),
			},
			model.Ketchup{},
			httpModel.ErrInvalid,
		},
		"create error": {
			args{
				ctx:  model.StoreUser(context.TODO(), model.NewUser(1, "", authModel.NewUser(0, ""))),
				item: model.NewKetchup(model.DefaultPattern, "0.0.0", model.Daily, false, model.NewGithubRepository(model.Identifier(1), ketchupRepository)),
			},
			model.Ketchup{},
			httpModel.ErrInternalError,
		},
		"success": {
			args{
				ctx:  model.StoreUser(context.TODO(), model.NewUser(1, "", authModel.NewUser(0, ""))),
				item: model.NewKetchup(model.DefaultPattern, "1.0.0", model.Daily, false, model.NewGithubRepository(model.Identifier(1), ketchupRepository)),
			},
			model.NewKetchup(model.DefaultPattern, "1.0.0", model.Daily, false, model.NewGithubRepository(model.Identifier(1), ketchupRepository).AddVersion(model.DefaultPattern, "1.0.0")),
			nil,
		},
	}

	for intention, testCase := range cases {
		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			mockKetchupStore := mocks.NewKetchupStore(ctrl)
			mockRepositoryService := mocks.NewRepositoryService(ctrl)

			instance := Service{
				ketchupStore: mockKetchupStore,
				repository:   mockRepositoryService,
			}

			switch intention {
			case "start atomic error":
				mockKetchupStore.EXPECT().DoAtomic(gomock.Any(), gomock.Any()).Return(errAtomicStart)
			case "repository error", "check error":
				dummyFn := func(ctx context.Context, do func(ctx context.Context) error) error {
					return do(ctx)
				}
				mockKetchupStore.EXPECT().DoAtomic(gomock.Any(), gomock.Any()).DoAndReturn(dummyFn)
				mockRepositoryService.EXPECT().GetOrCreate(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(model.NewEmptyRepository(), nil)
			case "create error":
				dummyFn := func(ctx context.Context, do func(ctx context.Context) error) error {
					return do(ctx)
				}
				mockKetchupStore.EXPECT().DoAtomic(gomock.Any(), gomock.Any()).DoAndReturn(dummyFn)
				mockRepositoryService.EXPECT().GetOrCreate(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(model.NewEmptyRepository(), nil)
				mockKetchupStore.EXPECT().Create(gomock.Any(), gomock.Any()).Return(model.Identifier(0), errors.New("failed"))
			case "success":
				dummyFn := func(ctx context.Context, do func(ctx context.Context) error) error {
					return do(ctx)
				}
				mockKetchupStore.EXPECT().DoAtomic(gomock.Any(), gomock.Any()).DoAndReturn(dummyFn)
				mockRepositoryService.EXPECT().GetOrCreate(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(model.NewGithubRepository(model.Identifier(1), ketchupRepository).AddVersion(model.DefaultPattern, "1.0.0"), nil)
				mockKetchupStore.EXPECT().GetByRepository(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(model.Ketchup{}, nil)
				mockKetchupStore.EXPECT().Create(gomock.Any(), gomock.Any()).Return(model.Identifier(1), nil)
			}

			got, gotErr := instance.Create(testCase.args.ctx, testCase.args.item)

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

func TestUpdate(t *testing.T) {
	t.Parallel()

	type args struct {
		ctx        context.Context
		oldPattern string
		item       model.Ketchup
	}

	cases := map[string]struct {
		args    args
		want    model.Ketchup
		wantErr error
	}{
		"start atomic error": {
			args{
				ctx:  context.TODO(),
				item: model.Ketchup{},
			},
			model.Ketchup{},
			errAtomicStart,
		},
		"fetch error": {
			args{
				ctx:  context.TODO(),
				item: model.Ketchup{},
			},
			model.Ketchup{},
			httpModel.ErrInternalError,
		},
		"check error": {
			args{
				ctx:  context.TODO(),
				item: model.NewKetchup(model.DefaultPattern, "", model.Daily, false, model.NewGithubRepository(model.Identifier(1), ketchupRepository)),
			},
			model.Ketchup{},
			httpModel.ErrInvalid,
		},
		"pattern change error": {
			args{
				ctx:        model.StoreUser(context.TODO(), model.NewUser(1, "", authModel.NewUser(0, ""))),
				oldPattern: "0.9.0",
				item:       model.NewKetchup("latest", "1.0.0", model.Daily, false, model.NewGithubRepository(model.Identifier(1), ketchupRepository)),
			},
			model.Ketchup{},
			httpModel.ErrInternalError,
		},
		"pattern change success": {
			args{
				ctx:        model.StoreUser(context.TODO(), model.NewUser(1, "", authModel.NewUser(0, ""))),
				oldPattern: "0.9.0",
				item:       model.NewKetchup("latest", "1.0.0", model.Daily, false, model.NewGithubRepository(model.Identifier(1), ketchupRepository)),
			},
			model.Ketchup{
				Pattern:    "latest",
				Version:    "1.0.0",
				Frequency:  model.Daily,
				Repository: model.NewGithubRepository(model.Identifier(1), ketchupRepository).AddVersion("latest", "1.0.1"),
				User:       model.NewUser(1, "", authModel.NewUser(0, "")),
			},
			nil,
		},
		"update error": {
			args{
				ctx:  model.StoreUser(context.TODO(), model.NewUser(1, "", authModel.NewUser(0, ""))),
				item: model.NewKetchup(model.DefaultPattern, "0.0.0", model.Daily, false, model.NewGithubRepository(model.Identifier(2), "")),
			},
			model.Ketchup{},
			httpModel.ErrInternalError,
		},
		"success": {
			args{
				ctx:        model.StoreUser(context.TODO(), model.NewUser(1, "", authModel.NewUser(0, ""))),
				oldPattern: "0.9.0",
				item:       model.NewKetchup(model.DefaultPattern, "1.0.0", model.Daily, false, model.NewGithubRepository(model.Identifier(1), "")),
			},
			model.Ketchup{
				Pattern:    model.DefaultPattern,
				Version:    "1.0.0",
				Frequency:  model.Daily,
				Repository: model.NewGithubRepository(model.Identifier(1), ketchupRepository).AddVersion(model.DefaultPattern, "1.2.3"),
				User:       model.NewUser(1, "", authModel.NewUser(0, "")),
			},
			nil,
		},
	}

	for intention, testCase := range cases {
		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			mockKetchupStore := mocks.NewKetchupStore(ctrl)
			mockRepositoryService := mocks.NewRepositoryService(ctrl)

			instance := Service{
				ketchupStore: mockKetchupStore,
				repository:   mockRepositoryService,
			}

			switch intention {
			case "start atomic error":
				mockKetchupStore.EXPECT().DoAtomic(gomock.Any(), gomock.Any()).Return(errAtomicStart)
			case "fetch error":
				dummyFn := func(ctx context.Context, do func(ctx context.Context) error) error {
					return do(ctx)
				}
				mockKetchupStore.EXPECT().DoAtomic(gomock.Any(), gomock.Any()).DoAndReturn(dummyFn)
				mockKetchupStore.EXPECT().GetByRepository(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(model.Ketchup{}, errors.New("failed"))
			case "check error":
				dummyFn := func(ctx context.Context, do func(ctx context.Context) error) error {
					return do(ctx)
				}
				mockKetchupStore.EXPECT().DoAtomic(gomock.Any(), gomock.Any()).DoAndReturn(dummyFn)
				mockKetchupStore.EXPECT().GetByRepository(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(model.NewKetchup(model.DefaultPattern, "0.9.0", model.Daily, false, model.NewGithubRepository(model.Identifier(1), ketchupRepository)), nil)
			case "pattern change error":
				dummyFn := func(ctx context.Context, do func(ctx context.Context) error) error {
					return do(ctx)
				}
				mockKetchupStore.EXPECT().DoAtomic(gomock.Any(), gomock.Any()).DoAndReturn(dummyFn)
				mockKetchupStore.EXPECT().GetByRepository(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(model.NewKetchup(model.DefaultPattern, "0.9.0", model.Daily, false, model.NewGithubRepository(model.Identifier(1), ketchupRepository)), nil)
				mockRepositoryService.EXPECT().GetOrCreate(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(model.NewEmptyRepository(), errors.New("failed"))
			case "pattern change success":
				dummyFn := func(ctx context.Context, do func(ctx context.Context) error) error {
					return do(ctx)
				}
				mockKetchupStore.EXPECT().DoAtomic(gomock.Any(), gomock.Any()).DoAndReturn(dummyFn)
				mockKetchupStore.EXPECT().GetByRepository(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(model.Ketchup{
					Pattern:    model.DefaultPattern,
					Version:    "0.9.0",
					Frequency:  model.Daily,
					Repository: model.NewGithubRepository(model.Identifier(1), ketchupRepository),
					User:       model.NewUser(1, "", authModel.NewUser(0, "")),
				}, nil)
				mockRepositoryService.EXPECT().GetOrCreate(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(model.NewGithubRepository(model.Identifier(1), ketchupRepository).AddVersion("latest", "1.0.1"), nil)
				mockKetchupStore.EXPECT().Update(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			case "update error":
				dummyFn := func(ctx context.Context, do func(ctx context.Context) error) error {
					return do(ctx)
				}
				mockKetchupStore.EXPECT().DoAtomic(gomock.Any(), gomock.Any()).DoAndReturn(dummyFn)
				mockKetchupStore.EXPECT().GetByRepository(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(model.NewKetchup(model.DefaultPattern, "0.9.0", model.Daily, false, model.NewGithubRepository(model.Identifier(1), ketchupRepository)), nil)
				mockKetchupStore.EXPECT().Update(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("failed"))
			case "success":
				dummyFn := func(ctx context.Context, do func(ctx context.Context) error) error {
					return do(ctx)
				}
				mockKetchupStore.EXPECT().DoAtomic(gomock.Any(), gomock.Any()).DoAndReturn(dummyFn)
				mockKetchupStore.EXPECT().GetByRepository(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(model.Ketchup{
					Pattern:    model.DefaultPattern,
					Version:    "0.9.0",
					Frequency:  model.Daily,
					Repository: model.NewGithubRepository(model.Identifier(1), ketchupRepository).AddVersion(model.DefaultPattern, "1.2.3"),
					User:       model.NewUser(1, "", authModel.NewUser(0, "")),
				}, nil)
				mockKetchupStore.EXPECT().Update(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			}

			got, gotErr := instance.Update(testCase.args.ctx, testCase.args.oldPattern, testCase.args.item)

			failed := false

			if !errors.Is(gotErr, testCase.wantErr) {
				failed = true
			} else if !reflect.DeepEqual(got, testCase.want) {
				failed = true
			}

			if failed {
				t.Errorf("Update() = (%+v, `%s`), want (%+v, `%s`)", got, gotErr, testCase.want, testCase.wantErr)
			}
		})
	}
}

func TestDelete(t *testing.T) {
	t.Parallel()

	type args struct {
		ctx  context.Context
		item model.Ketchup
	}

	cases := map[string]struct {
		args    args
		wantErr error
	}{
		"start atomic error": {
			args{
				ctx:  context.TODO(),
				item: model.Ketchup{},
			},
			errAtomicStart,
		},
		"fetch error": {
			args{
				ctx:  context.TODO(),
				item: model.NewKetchup(model.DefaultPattern, "", model.Daily, false, model.NewGithubRepository(model.Identifier(0), "")),
			},
			httpModel.ErrInternalError,
		},
		"check error": {
			args{
				ctx:  context.TODO(),
				item: model.NewKetchup(model.DefaultPattern, "", model.Daily, false, model.NewGithubRepository(model.Identifier(1), "")),
			},
			httpModel.ErrInvalid,
		},
		"delete error": {
			args{
				ctx:  model.StoreUser(context.TODO(), model.NewUser(1, "", authModel.NewUser(0, ""))),
				item: model.NewKetchup(model.DefaultPattern, "", model.Daily, false, model.NewGithubRepository(model.Identifier(3), "")),
			},
			httpModel.ErrInternalError,
		},
		"success": {
			args{
				ctx:  model.StoreUser(context.TODO(), model.NewUser(1, "", authModel.NewUser(0, ""))),
				item: model.NewKetchup(model.DefaultPattern, "", model.Daily, false, model.NewGithubRepository(model.Identifier(1), "")),
			},
			nil,
		},
	}

	for intention, testCase := range cases {
		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			mockKetchupStore := mocks.NewKetchupStore(ctrl)

			instance := Service{
				ketchupStore: mockKetchupStore,
			}

			switch intention {
			case "start atomic error":
				mockKetchupStore.EXPECT().DoAtomic(gomock.Any(), gomock.Any()).Return(errAtomicStart)
			case "fetch error":
				dummyFn := func(ctx context.Context, do func(ctx context.Context) error) error {
					return do(ctx)
				}
				mockKetchupStore.EXPECT().DoAtomic(gomock.Any(), gomock.Any()).DoAndReturn(dummyFn)
				mockKetchupStore.EXPECT().GetByRepository(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(model.Ketchup{}, errors.New("failed"))
			case "check error":
				dummyFn := func(ctx context.Context, do func(ctx context.Context) error) error {
					return do(ctx)
				}
				mockKetchupStore.EXPECT().DoAtomic(gomock.Any(), gomock.Any()).DoAndReturn(dummyFn)
				mockKetchupStore.EXPECT().GetByRepository(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(model.NewKetchup(model.DefaultPattern, "", model.Daily, false, model.NewGithubRepository(model.Identifier(1), "")), nil)
			case "delete error":
				dummyFn := func(ctx context.Context, do func(ctx context.Context) error) error {
					return do(ctx)
				}
				mockKetchupStore.EXPECT().DoAtomic(gomock.Any(), gomock.Any()).DoAndReturn(dummyFn)
				mockKetchupStore.EXPECT().GetByRepository(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(model.NewKetchup(model.DefaultPattern, "", model.Daily, false, model.NewGithubRepository(model.Identifier(1), "")), nil)
				mockKetchupStore.EXPECT().Delete(gomock.Any(), gomock.Any()).Return(errors.New("failed"))
			case "success":
				dummyFn := func(ctx context.Context, do func(ctx context.Context) error) error {
					return do(ctx)
				}
				mockKetchupStore.EXPECT().DoAtomic(gomock.Any(), gomock.Any()).DoAndReturn(dummyFn)
				mockKetchupStore.EXPECT().GetByRepository(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(model.NewKetchup(model.DefaultPattern, "", model.Daily, false, model.NewGithubRepository(model.Identifier(1), "")), nil)
				mockKetchupStore.EXPECT().Delete(gomock.Any(), gomock.Any()).Return(nil)
			}

			gotErr := instance.Delete(testCase.args.ctx, testCase.args.item)

			failed := false

			if !errors.Is(gotErr, testCase.wantErr) {
				failed = true
			}

			if failed {
				t.Errorf("Delete() = `%s`, want `%s`", gotErr, testCase.wantErr)
			}
		})
	}
}

func TestCheck(t *testing.T) {
	t.Parallel()

	type args struct {
		ctx context.Context
		old model.Ketchup
		new model.Ketchup
	}

	cases := map[string]struct {
		args    args
		wantErr error
	}{
		"no user": {
			args{
				ctx: context.TODO(),
			},
			errors.New("you must be logged in for interacting"),
		},
		"delete": {
			args{
				ctx: model.StoreUser(context.TODO(), model.NewUser(1, "", authModel.NewUser(0, ""))),
				old: model.NewKetchup(model.DefaultPattern, "1.0.0", model.Daily, false, model.NewEmptyRepository()),
				new: model.Ketchup{},
			},
			nil,
		},
		"no pattern": {
			args{
				ctx: model.StoreUser(context.TODO(), model.NewUser(1, "", authModel.NewUser(0, ""))),
				old: model.Ketchup{},
				new: model.NewKetchup("", "", model.Daily, false, model.NewGithubRepository(model.Identifier(1), "")),
			},
			errors.New("pattern is required"),
		},
		"invalid pattern": {
			args{
				ctx: model.StoreUser(context.TODO(), model.NewUser(1, "", authModel.NewUser(0, ""))),
				old: model.Ketchup{},
				new: model.NewKetchup("test", "", model.Daily, false, model.NewGithubRepository(model.Identifier(1), "")),
			},
			errors.New("pattern is invalid"),
		},
		"no version": {
			args{
				ctx: model.StoreUser(context.TODO(), model.NewUser(1, "", authModel.NewUser(0, ""))),
				old: model.NewKetchup(model.DefaultPattern, "1.0.0", model.Daily, false, model.NewEmptyRepository()),
				new: model.NewKetchup(model.DefaultPattern, "", model.Daily, false, model.NewGithubRepository(model.Identifier(1), "")),
			},
			errors.New("version is required"),
		},
		"create error": {
			args{
				ctx: model.StoreUser(context.TODO(), model.NewUser(1, "", authModel.NewUser(0, ""))),
				new: model.Ketchup{Version: "1.0.0", Pattern: "stable", Repository: model.NewGithubRepository(model.Identifier(1), ""), User: model.NewUser(1, "", authModel.NewUser(0, ""))},
			},
			errors.New("check if ketchup already exists"),
		},
		"create already exists": {
			args{
				ctx: model.StoreUser(context.TODO(), model.NewUser(1, "", authModel.NewUser(0, ""))),
				new: model.Ketchup{Pattern: model.DefaultPattern, Version: "1.0.0", Repository: model.NewGithubRepository(model.Identifier(2), ketchupRepository), User: model.NewUser(1, "", authModel.NewUser(0, ""))},
			},
			errors.New("ketchup for `vibioh/ketchup` with pattern `stable` already exists"),
		},
	}

	for intention, testCase := range cases {
		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			mockKetchupStore := mocks.NewKetchupStore(ctrl)

			instance := Service{
				ketchupStore: mockKetchupStore,
			}

			switch intention {
			case "no pattern", "invalid pattern", "no version":
				mockKetchupStore.EXPECT().GetByRepository(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(model.Ketchup{}, nil)
			case "create error":
				mockKetchupStore.EXPECT().GetByRepository(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(model.Ketchup{}, errors.New("failed"))
			case "create already exists":
				mockKetchupStore.EXPECT().GetByRepository(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(model.NewKetchup(model.DefaultPattern, "1.0.0", model.Daily, false, model.NewGithubRepository(model.Identifier(1), ketchupRepository)), nil)
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
