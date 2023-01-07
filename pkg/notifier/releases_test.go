package notifier

import (
	"context"
	"errors"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/ViBiOh/ketchup/pkg/mocks"
	"github.com/ViBiOh/ketchup/pkg/model"
	"github.com/ViBiOh/ketchup/pkg/semver"
	"github.com/golang/mock/gomock"
)

func TestGetNewStandardReleases(t *testing.T) {
	t.Parallel()

	type args struct {
		ctx context.Context
	}

	cases := map[string]struct {
		instance App
		args     args
		want     []model.Release
		wantErr  error
	}{
		"list error": {
			App{},
			args{
				ctx: context.TODO(),
			},
			nil,
			errors.New("failed"),
		},
		"github error": {
			App{},
			args{
				ctx: context.TODO(),
			},
			nil,
			nil,
		},
		"same version": {
			App{},
			args{
				ctx: context.TODO(),
			},
			nil,
			nil,
		},
		"success": {
			App{},
			args{
				ctx: context.TODO(),
			},
			[]model.Release{model.NewRelease(
				model.NewGithubRepository(model.Identifier(1), repositoryName).AddVersion(model.DefaultPattern, repositoryVersion),
				model.DefaultPattern,
				safeParse("1.1.0"),
			)},
			nil,
		},
	}

	for intention, testCase := range cases {
		intention, testCase := intention, testCase

		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepositoryService := mocks.NewRepositoryService(ctrl)

			testCase.instance.repositoryService = mockRepositoryService

			switch intention {
			case "list error":
				mockRepositoryService.EXPECT().ListByKinds(gomock.Any(), gomock.Any(), gomock.Any(), model.Github, model.Docker, model.NPM, model.Pypi).Return(nil, uint64(0), errors.New("failed"))
			case "github error":
				mockRepositoryService.EXPECT().ListByKinds(gomock.Any(), gomock.Any(), gomock.Any(), model.Github, model.Docker, model.NPM, model.Pypi).Return([]model.Repository{
					model.NewGithubRepository(model.Identifier(1), repositoryName).AddVersion(model.DefaultPattern, repositoryVersion),
				}, uint64(1), nil)
				mockRepositoryService.EXPECT().LatestVersions(gomock.Any(), gomock.Any()).Return(nil, errors.New("failed"))
			case "same version":
				mockRepositoryService.EXPECT().ListByKinds(gomock.Any(), gomock.Any(), gomock.Any(), model.Github, model.Docker, model.NPM, model.Pypi).Return([]model.Repository{
					model.NewGithubRepository(model.Identifier(1), repositoryName).AddVersion(model.DefaultPattern, repositoryVersion),
				}, uint64(1), nil)
				mockRepositoryService.EXPECT().LatestVersions(gomock.Any(), gomock.Any()).Return(map[string]semver.Version{
					model.DefaultPattern: {
						Name: "1.1.0",
					},
				}, nil)
			case "success":
				mockRepositoryService.EXPECT().ListByKinds(gomock.Any(), gomock.Any(), gomock.Any(), model.Github, model.Docker, model.NPM, model.Pypi).Return([]model.Repository{
					model.NewGithubRepository(model.Identifier(1), repositoryName).AddVersion(model.DefaultPattern, repositoryVersion),
				}, uint64(1), nil)
				mockRepositoryService.EXPECT().LatestVersions(gomock.Any(), gomock.Any()).Return(map[string]semver.Version{
					model.DefaultPattern: safeParse("1.1.0"),
					"1.0":                safeParse("1.0"),
				}, nil)
			}

			got, _, gotErr := testCase.instance.getNewStandardReleases(testCase.args.ctx)

			failed := false

			if testCase.wantErr == nil && gotErr != nil {
				failed = true
			} else if testCase.wantErr != nil && gotErr == nil {
				failed = true
			} else if testCase.wantErr != nil && !strings.Contains(gotErr.Error(), testCase.wantErr.Error()) {
				failed = true
			} else if !reflect.DeepEqual(got, testCase.want) {
				failed = true
			}

			if failed {
				t.Errorf("getNewStandardReleases() = (%+v, `%s`), want (%+v, `%s`)", got, gotErr, testCase.want, testCase.wantErr)
			}
		})
	}
}

func TestGetNewHelmReleases(t *testing.T) {
	t.Parallel()

	postgresRepo := model.NewHelmRepository(model.Identifier(4), "https://charts.helm.sh/stable", "postgreql").AddVersion(model.DefaultPattern, "3.0.0")
	appRepo := model.NewHelmRepository(model.Identifier(1), "https://charts.vibioh.fr", "app").AddVersion(model.DefaultPattern, "1.0.0").AddVersion("1.0", "1.0.0").AddVersion("~1.0", "a.b.c")
	cronRepo := model.NewHelmRepository(model.Identifier(2), "https://charts.vibioh.fr", "cron").AddVersion(model.DefaultPattern, "1.0.0")
	fluxRepo := model.NewHelmRepository(model.Identifier(3), "https://charts.vibioh.fr", "flux").AddVersion(model.DefaultPattern, "1.0.0")

	type args struct {
		content string
	}

	cases := map[string]struct {
		instance  App
		args      args
		want      []model.Release
		wantCount uint64
		wantErr   error
	}{
		"fetch error": {
			App{},
			args{
				content: "test",
			},
			nil,
			0,
			errors.New("db error"),
		},
		"no repository": {
			App{},
			args{
				content: "test",
			},
			nil,
			0,
			nil,
		},
		"helm error": {
			App{},
			args{
				content: "test",
			},
			nil,
			4,
			nil,
		},
		"helm": {
			App{},
			args{
				content: "test",
			},
			[]model.Release{
				model.NewRelease(appRepo, model.DefaultPattern, safeParse("1.1.0")),
				model.NewRelease(cronRepo, model.DefaultPattern, safeParse("2.0.0")),
				model.NewRelease(postgresRepo, model.DefaultPattern, safeParse("3.1.0")),
			},
			4,
			nil,
		},
	}

	for intention, testCase := range cases {
		intention, testCase := intention, testCase

		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepositoryService := mocks.NewRepositoryService(ctrl)
			mockHelmProvider := mocks.NewHelmProvider(ctrl)

			testCase.instance.repositoryService = mockRepositoryService
			testCase.instance.helmApp = mockHelmProvider

			switch intention {
			case "fetch error":
				mockRepositoryService.EXPECT().ListByKinds(gomock.Any(), gomock.Any(), gomock.Any(), model.Helm).Return(nil, uint64(0), errors.New("db error"))
			case "no repository":
				mockRepositoryService.EXPECT().ListByKinds(gomock.Any(), gomock.Any(), gomock.Any(), model.Helm).Return(nil, uint64(0), nil)
			case "helm error":
				mockRepositoryService.EXPECT().ListByKinds(gomock.Any(), gomock.Any(), gomock.Any(), model.Helm).Return([]model.Repository{
					model.NewHelmRepository(model.Identifier(1), "https://charts.vibioh.fr", "app"),
					model.NewHelmRepository(model.Identifier(1), "https://charts.vibioh.fr", "cron"),
					model.NewHelmRepository(model.Identifier(1), "https://charts.vibioh.fr", "flux"),
					model.NewHelmRepository(model.Identifier(1), "https://charts.helm.sh/stable", "postgreql"),
				}, uint64(4), nil)
				mockHelmProvider.EXPECT().FetchIndex(gomock.Any(), gomock.Any(), gomock.Any()).Times(2).Return(nil, errors.New("helm error"))
			case "helm":
				mockRepositoryService.EXPECT().ListByKinds(gomock.Any(), gomock.Any(), gomock.Any(), model.Helm).Return([]model.Repository{
					postgresRepo,
					appRepo,
					cronRepo,
					fluxRepo,
				}, uint64(0), nil)
				mockHelmProvider.EXPECT().FetchIndex(gomock.Any(), gomock.Any(), gomock.Any()).Times(2).Return(map[string]map[string]semver.Version{
					"app": {
						model.DefaultPattern: safeParse("1.1.0"),
					},
					"cron": {
						model.DefaultPattern: safeParse("2.0.0"),
					},
					"flux": {
						model.DefaultPattern: safeParse("1.0.0"),
					},
					"postgreql": {
						model.DefaultPattern: safeParse("3.1.0"),
					},
				}, nil)
			}

			got, gotCount, gotErr := testCase.instance.getNewHelmReleases(context.TODO())

			failed := false

			sort.Sort(model.ReleaseByRepositoryIDAndPattern(got))

			if testCase.wantErr == nil && gotErr != nil {
				failed = true
			} else if testCase.wantErr != nil && gotErr == nil {
				failed = true
			} else if testCase.wantErr != nil && !strings.Contains(gotErr.Error(), testCase.wantErr.Error()) {
				failed = true
			} else if gotCount != testCase.wantCount {
				failed = true
			} else if !reflect.DeepEqual(got, testCase.want) {
				failed = true
			}

			if failed {
				t.Errorf("getNewHelmReleases() = (%+v, %d, `%s`), want (%+v, %d, `%s`)", got, gotCount, gotErr, testCase.want, testCase.wantCount, testCase.wantErr)
			}
		})
	}
}
