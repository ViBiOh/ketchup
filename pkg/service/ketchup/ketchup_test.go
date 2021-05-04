package ketchup

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	authModel "github.com/ViBiOh/auth/v2/pkg/model"
	httpModel "github.com/ViBiOh/httputils/v4/pkg/model"
	"github.com/ViBiOh/ketchup/pkg/model"
	"github.com/ViBiOh/ketchup/pkg/service/repository/repositorytest"
	"github.com/ViBiOh/ketchup/pkg/store/ketchup/ketchuptest"
)

var (
	ketchupRepository = "vibioh/ketchup"
	viwsRepository    = "vibioh/viws"

	errAtomicStart = errors.New("invalid context")
)

func TestList(t *testing.T) {
	type args struct {
		page     uint
		pageSize uint
	}

	var cases = []struct {
		intention string
		instance  App
		args      args
		want      []model.Ketchup
		wantCount uint64
		wantErr   error
	}{
		{
			"simple",
			New(ketchuptest.New().SetList([]model.Ketchup{
				model.NewKetchup(model.DefaultPattern, "1.2.3", model.Daily, model.NewGithubRepository(2, viwsRepository).AddVersion(model.DefaultPattern, "1.2.3")),
				model.NewKetchup(model.DefaultPattern, "1.0.0", model.Daily, model.NewGithubRepository(1, ketchupRepository).AddVersion(model.DefaultPattern, "1.0.2")),
			}, 2, nil), nil),
			args{
				page: 1,
			},
			[]model.Ketchup{
				{Pattern: model.DefaultPattern, Version: "1.0.0", Frequency: model.Daily, Semver: "Patch", Repository: model.NewGithubRepository(1, ketchupRepository).AddVersion(model.DefaultPattern, "1.0.2")},
				{Pattern: model.DefaultPattern, Version: "1.2.3", Frequency: model.Daily, Repository: model.NewGithubRepository(2, viwsRepository).AddVersion(model.DefaultPattern, "1.2.3")},
			},
			2,
			nil,
		},
		{
			"error",
			New(ketchuptest.New().SetList(nil, 0, errors.New("failed")), nil),
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

func TestListForRepositories(t *testing.T) {
	type args struct {
		repositories []model.Repository
	}

	var cases = []struct {
		intention string
		instance  App
		args      args
		want      []model.Ketchup
		wantErr   error
	}{
		{
			"simple",
			New(ketchuptest.New().SetListByRepositoriesID([]model.Ketchup{
				model.NewKetchup(model.DefaultPattern, "1.0.0", model.Daily, model.NewGithubRepository(1, ketchupRepository).AddVersion(model.DefaultPattern, "1.0.2")),
				model.NewKetchup(model.DefaultPattern, "1.2.3", model.Daily, model.NewGithubRepository(2, viwsRepository).AddVersion(model.DefaultPattern, "1.2.3")),
			}, nil), nil),
			args{
				repositories: []model.Repository{
					model.NewGithubRepository(1, ""),
					model.NewGithubRepository(2, ""),
				},
			},
			[]model.Ketchup{
				{Pattern: model.DefaultPattern, Version: "1.0.0", Frequency: model.Daily, Semver: "Patch", Repository: model.NewGithubRepository(1, ketchupRepository).AddVersion(model.DefaultPattern, "1.0.2")},
				{Pattern: model.DefaultPattern, Version: "1.2.3", Frequency: model.Daily, Repository: model.NewGithubRepository(2, viwsRepository).AddVersion(model.DefaultPattern, "1.2.3")},
			},
			nil,
		},
		{
			"error",
			New(ketchuptest.New().SetListByRepositoriesID(nil, errors.New("failed")), nil),
			args{},
			nil,
			httpModel.ErrInternalError,
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			got, gotErr := tc.instance.ListForRepositories(context.Background(), tc.args.repositories)

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
			New(ketchuptest.New().SetDoAtomic(errAtomicStart), repositorytest.New()),
			args{
				ctx:  context.TODO(),
				item: model.NoneKetchup,
			},
			model.NoneKetchup,
			errAtomicStart,
		},
		{
			"repository error",
			New(ketchuptest.New(), repositorytest.New()),
			args{
				ctx:  context.Background(),
				item: model.NoneKetchup,
			},
			model.NoneKetchup,
			httpModel.ErrInvalid,
		},
		{
			"check error",
			New(ketchuptest.New(), repositorytest.New()),
			args{
				ctx:  context.Background(),
				item: model.NewKetchup(model.DefaultPattern, "", model.Daily, model.NewGithubRepository(1, ketchupRepository)),
			},
			model.NoneKetchup,
			httpModel.ErrInvalid,
		},
		{
			"create error",
			New(ketchuptest.New().SetCreate(0, errors.New("failed")), repositorytest.New()),
			args{
				ctx:  model.StoreUser(context.Background(), model.NewUser(1, "", authModel.NewUser(0, ""))),
				item: model.NewKetchup(model.DefaultPattern, "0.0.0", model.Daily, model.NewGithubRepository(1, ketchupRepository)),
			},
			model.NoneKetchup,
			httpModel.ErrInternalError,
		},
		{
			"success",
			New(ketchuptest.New(), repositorytest.New().SetGetOrCreate(model.NewGithubRepository(1, ketchupRepository).AddVersion(model.DefaultPattern, "1.0.0"), nil)),
			args{
				ctx:  model.StoreUser(context.Background(), model.NewUser(1, "", authModel.NewUser(0, ""))),
				item: model.NewKetchup(model.DefaultPattern, "1.0.0", model.Daily, model.NewGithubRepository(1, ketchupRepository)),
			},
			model.NewKetchup(model.DefaultPattern, "1.0.0", model.Daily, model.NewGithubRepository(1, ketchupRepository).AddVersion(model.DefaultPattern, "1.0.0")),
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
		instance  App
		args      args
		want      model.Ketchup
		wantErr   error
	}{
		{
			"start atomic error",
			New(ketchuptest.New().SetDoAtomic(errAtomicStart), repositorytest.New()),
			args{
				ctx:  context.TODO(),
				item: model.NoneKetchup,
			},
			model.NoneKetchup,
			errAtomicStart,
		},
		{
			"fetch error",
			New(ketchuptest.New().SetGetByRepositoryID(model.NoneKetchup, errors.New("failed")), repositorytest.New()),
			args{
				ctx:  context.Background(),
				item: model.NoneKetchup,
			},
			model.NoneKetchup,
			httpModel.ErrInternalError,
		},
		{
			"check error",
			New(ketchuptest.New().SetGetByRepositoryID(model.NewKetchup(model.DefaultPattern, "0.9.0", model.Daily, model.NewGithubRepository(1, ketchupRepository)), nil), repositorytest.New()),
			args{
				ctx:  context.Background(),
				item: model.NewKetchup(model.DefaultPattern, "", model.Daily, model.NewGithubRepository(1, ketchupRepository)),
			},
			model.NoneKetchup,
			httpModel.ErrInvalid,
		},
		{
			"pattern change error",
			New(ketchuptest.New().SetGetByRepositoryID(model.NewKetchup(model.DefaultPattern, "0.9.0", model.Daily, model.NewGithubRepository(1, ketchupRepository)), nil), repositorytest.New().SetGetOrCreate(model.NoneRepository, errors.New("failed"))),
			args{
				ctx:  model.StoreUser(context.Background(), model.NewUser(1, "", authModel.NewUser(0, ""))),
				item: model.NewKetchup("latest", "1.0.0", model.Daily, model.NewGithubRepository(1, ketchupRepository)),
			},
			model.NoneKetchup,
			httpModel.ErrInternalError,
		},
		{
			"pattern change success",
			New(ketchuptest.New().SetGetByRepositoryID(model.Ketchup{
				Pattern:    model.DefaultPattern,
				Version:    "0.9.0",
				Frequency:  model.Daily,
				Repository: model.NewGithubRepository(1, ketchupRepository),
				User:       model.NewUser(1, "", authModel.NewUser(0, "")),
			}, nil), repositorytest.New().SetGetOrCreate(model.NewGithubRepository(1, ketchupRepository).AddVersion("latest", "1.0.1"), nil)),
			args{
				ctx:  model.StoreUser(context.Background(), model.NewUser(1, "", authModel.NewUser(0, ""))),
				item: model.NewKetchup("latest", "1.0.0", model.Daily, model.NewGithubRepository(1, ketchupRepository)),
			},
			model.Ketchup{
				Pattern:    "latest",
				Version:    "1.0.0",
				Frequency:  model.Daily,
				Repository: model.NewGithubRepository(1, ketchupRepository).AddVersion("latest", "1.0.1"),
				User:       model.NewUser(1, "", authModel.NewUser(0, "")),
			},
			nil,
		},
		{
			"update error",
			New(ketchuptest.New().SetGetByRepositoryID(model.NewKetchup(model.DefaultPattern, "0.9.0", model.Daily, model.NewGithubRepository(1, ketchupRepository)), nil).SetUpdate(errors.New("failed")), repositorytest.New()),
			args{
				ctx:  model.StoreUser(context.Background(), model.NewUser(1, "", authModel.NewUser(0, ""))),
				item: model.NewKetchup(model.DefaultPattern, "0.0.0", model.Daily, model.NewGithubRepository(2, "")),
			},
			model.NoneKetchup,
			httpModel.ErrInternalError,
		},
		{
			"success",
			New(ketchuptest.New().SetGetByRepositoryID(model.Ketchup{
				Pattern:    model.DefaultPattern,
				Version:    "0.9.0",
				Frequency:  model.Daily,
				Repository: model.NewGithubRepository(1, ketchupRepository).AddVersion(model.DefaultPattern, "1.2.3"),
				User:       model.NewUser(1, "", authModel.NewUser(0, "")),
			}, nil), repositorytest.New()),
			args{
				ctx:  model.StoreUser(context.Background(), model.NewUser(1, "", authModel.NewUser(0, ""))),
				item: model.NewKetchup(model.DefaultPattern, "1.0.0", model.Daily, model.NewGithubRepository(1, "")),
			},
			model.Ketchup{
				Pattern:    model.DefaultPattern,
				Version:    "1.0.0",
				Frequency:  model.Daily,
				Repository: model.NewGithubRepository(1, ketchupRepository).AddVersion(model.DefaultPattern, "1.2.3"),
				User:       model.NewUser(1, "", authModel.NewUser(0, "")),
			},
			nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			got, gotErr := tc.instance.Update(tc.args.ctx, tc.args.item)

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
		instance  App
		args      args
		wantErr   error
	}{
		{
			"start atomic error",
			New(ketchuptest.New().SetDoAtomic(errAtomicStart), repositorytest.New()),
			args{
				ctx:  context.TODO(),
				item: model.NoneKetchup,
			},
			errAtomicStart,
		},
		{
			"fetch error",
			New(ketchuptest.New().SetGetByRepositoryID(model.NoneKetchup, errors.New("failed")), repositorytest.New()),
			args{
				ctx:  context.Background(),
				item: model.NewKetchup(model.DefaultPattern, "", model.Daily, model.NewGithubRepository(0, "")),
			},
			httpModel.ErrInternalError,
		},
		{
			"check error",
			New(ketchuptest.New(), repositorytest.New()),
			args{
				ctx:  context.Background(),
				item: model.NewKetchup(model.DefaultPattern, "", model.Daily, model.NewGithubRepository(1, "")),
			},
			httpModel.ErrInvalid,
		},
		{
			"delete error",
			New(ketchuptest.New().SetDelete(errors.New("failed")), repositorytest.New()),
			args{
				ctx:  model.StoreUser(context.Background(), model.NewUser(1, "", authModel.NewUser(0, ""))),
				item: model.NewKetchup(model.DefaultPattern, "", model.Daily, model.NewGithubRepository(3, "")),
			},
			httpModel.ErrInternalError,
		},
		{
			"success",
			New(ketchuptest.New(), repositorytest.New()),
			args{
				ctx:  model.StoreUser(context.Background(), model.NewUser(1, "", authModel.NewUser(0, ""))),
				item: model.NewKetchup(model.DefaultPattern, "", model.Daily, model.NewGithubRepository(1, "")),
			},
			nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			gotErr := tc.instance.Delete(tc.args.ctx, tc.args.item)

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
		instance  app
		args      args
		wantErr   error
	}{
		{
			"no user",
			app{ketchupStore: ketchuptest.New()},
			args{
				ctx: context.Background(),
			},
			errors.New("you must be logged in for interacting"),
		},
		{
			"delete",
			app{ketchupStore: ketchuptest.New()},
			args{
				ctx: model.StoreUser(context.Background(), model.NewUser(1, "", authModel.NewUser(0, ""))),
				old: model.NewKetchup(model.DefaultPattern, "1.0.0", model.Daily, model.NoneRepository),
				new: model.NoneKetchup,
			},
			nil,
		},
		{
			"no pattern",
			app{ketchupStore: ketchuptest.New()},
			args{
				ctx: model.StoreUser(context.Background(), model.NewUser(1, "", authModel.NewUser(0, ""))),
				old: model.NoneKetchup,
				new: model.NewKetchup("", "", model.Daily, model.NewGithubRepository(1, "")),
			},
			errors.New("pattern is required"),
		},
		{
			"invalid pattern",
			app{ketchupStore: ketchuptest.New()},
			args{
				ctx: model.StoreUser(context.Background(), model.NewUser(1, "", authModel.NewUser(0, ""))),
				old: model.NoneKetchup,
				new: model.NewKetchup("test", "", model.Daily, model.NewGithubRepository(1, "")),
			},
			errors.New("pattern is invalid"),
		},
		{
			"no version",
			app{ketchupStore: ketchuptest.New()},
			args{
				ctx: model.StoreUser(context.Background(), model.NewUser(1, "", authModel.NewUser(0, ""))),
				old: model.NewKetchup(model.DefaultPattern, "1.0.0", model.Daily, model.NoneRepository),
				new: model.NewKetchup(model.DefaultPattern, "", model.Daily, model.NewGithubRepository(1, "")),
			},
			errors.New("version is required"),
		},
		{
			"create error",
			app{ketchupStore: ketchuptest.New().SetGetByRepositoryID(model.NoneKetchup, errors.New("failed"))},
			args{
				ctx: model.StoreUser(context.Background(), model.NewUser(1, "", authModel.NewUser(0, ""))),
				new: model.Ketchup{Version: "1.0.0", Repository: model.NewGithubRepository(0, ""), User: model.NewUser(1, "", authModel.NewUser(0, ""))},
			},
			errors.New("unable to check if ketchup already exists"),
		},
		{
			"create already exists",
			app{ketchupStore: ketchuptest.New().SetGetByRepositoryID(model.NewKetchup(model.DefaultPattern, "1.0.0", model.Daily, model.NewGithubRepository(1, ketchupRepository)), nil)},
			args{
				ctx: model.StoreUser(context.Background(), model.NewUser(1, "", authModel.NewUser(0, ""))),
				new: model.Ketchup{Pattern: model.DefaultPattern, Version: "1.0.0", Repository: model.NewGithubRepository(2, ketchupRepository), User: model.NewUser(1, "", authModel.NewUser(0, ""))},
			},
			errors.New("ketchup for vibioh/ketchup already exists"),
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
