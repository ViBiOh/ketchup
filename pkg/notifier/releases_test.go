package notifier

import (
	"context"
	"errors"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/ViBiOh/ketchup/pkg/model"
	"github.com/ViBiOh/ketchup/pkg/provider/helm/helmtest"
	"github.com/ViBiOh/ketchup/pkg/semver"
	"github.com/ViBiOh/ketchup/pkg/service/repository/repositorytest"
)

func TestGetNewStandardReleases(t *testing.T) {
	type args struct {
		ctx context.Context
	}

	var cases = []struct {
		intention string
		instance  app
		args      args
		want      []model.Release
		wantErr   error
	}{
		{
			"list error",
			app{
				repositoryService: repositorytest.New().SetListByKinds(nil, 0, errors.New("failed")),
			},
			args{
				ctx: context.Background(),
			},
			nil,
			errors.New("failed"),
		},
		{
			"github error",
			app{
				repositoryService: repositorytest.New().SetListByKinds([]model.Repository{
					model.NewGithubRepository(1, repositoryName).AddVersion(model.DefaultPattern, repositoryVersion)}, 1, nil).SetLatestVersions(nil, errors.New("failed")),
			},
			args{
				ctx: context.Background(),
			},
			nil,
			nil,
		},
		{
			"same version",
			app{
				repositoryService: repositorytest.New().SetListByKinds([]model.Repository{
					model.NewGithubRepository(1, repositoryName).AddVersion(model.DefaultPattern, repositoryVersion),
				}, 1, nil).SetLatestVersions(map[string]semver.Version{
					model.DefaultPattern: {
						Name: "1.1.0",
					},
				}, nil).SetUpdate(errors.New("failed")),
			},
			args{
				ctx: context.Background(),
			},
			nil,
			nil,
		},
		{
			"success",
			app{
				repositoryService: repositorytest.New().SetListByKinds([]model.Repository{
					model.NewGithubRepository(1, repositoryName).AddVersion(model.DefaultPattern, repositoryVersion),
				}, 1, nil).SetLatestVersions(map[string]semver.Version{
					model.DefaultPattern: safeParse("1.1.0"),
					"1.0":                safeParse("1.0"),
				}, nil).SetUpdate(nil),
			},
			args{
				ctx: context.Background(),
			},
			[]model.Release{model.NewRelease(
				model.NewGithubRepository(1, repositoryName).AddVersion(model.DefaultPattern, repositoryVersion),
				model.DefaultPattern,
				safeParse("1.1.0"),
			)},
			nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			got, _, gotErr := tc.instance.getNewStandardReleases(tc.args.ctx)
			pageSize = 20

			failed := false

			if tc.wantErr == nil && gotErr != nil {
				failed = true
			} else if tc.wantErr != nil && gotErr == nil {
				failed = true
			} else if tc.wantErr != nil && !strings.Contains(gotErr.Error(), tc.wantErr.Error()) {
				failed = true
			} else if !reflect.DeepEqual(got, tc.want) {
				failed = true
			}

			if failed {
				t.Errorf("getNewStandardReleases() = (%+v, `%s`), want (%+v, `%s`)", got, gotErr, tc.want, tc.wantErr)
			}
		})
	}
}

func TestGetNewHelmReleases(t *testing.T) {
	postgresRepo := model.NewHelmRepository(4, "https://charts.helm.sh/stable", "postgreql").AddVersion(model.DefaultPattern, "3.0.0")
	appRepo := model.NewHelmRepository(1, "https://charts.vibioh.fr", "app").AddVersion(model.DefaultPattern, "1.0.0").AddVersion("1.0", "1.0.0").AddVersion("~1.0", "a.b.c")
	cronRepo := model.NewHelmRepository(2, "https://charts.vibioh.fr", "cron").AddVersion(model.DefaultPattern, "1.0.0")
	fluxRepo := model.NewHelmRepository(3, "https://charts.vibioh.fr", "flux").AddVersion(model.DefaultPattern, "1.0.0")

	type args struct {
		content string
	}

	var cases = []struct {
		intention string
		instance  app
		args      args
		want      []model.Release
		wantCount uint64
		wantErr   error
	}{
		{
			"fetch error",
			app{
				repositoryService: repositorytest.New().SetListByKinds(nil, 0, errors.New("db error")),
				helmApp:           helmtest.New(),
			},
			args{
				content: "test",
			},
			nil,
			0,
			errors.New("db error"),
		},
		{
			"no repository",
			app{
				repositoryService: repositorytest.New().SetListByKinds(nil, 0, nil),
				helmApp:           helmtest.New(),
			},
			args{
				content: "test",
			},
			nil,
			0,
			nil,
		},
		{
			"helm error",
			app{
				repositoryService: repositorytest.New().SetListByKinds([]model.Repository{
					model.NewHelmRepository(1, "https://charts.vibioh.fr", "app"),
					model.NewHelmRepository(1, "https://charts.vibioh.fr", "cron"),
					model.NewHelmRepository(1, "https://charts.vibioh.fr", "flux"),
					model.NewHelmRepository(1, "https://charts.helm.sh/stable", "postgreql"),
				}, 4, nil),
				helmApp: helmtest.New().SetFetchIndex(nil, errors.New("helm error")),
			},
			args{
				content: "test",
			},
			nil,
			4,
			nil,
		},
		{
			"helm",
			app{
				repositoryService: repositorytest.New().SetListByKinds([]model.Repository{
					postgresRepo,
					appRepo,
					cronRepo,
					fluxRepo,
				}, 0, nil),
				helmApp: helmtest.New().SetFetchIndex(map[string]map[string]semver.Version{
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
				}, nil),
			},
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

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			got, gotCount, gotErr := tc.instance.getNewHelmReleases(context.Background())

			failed := false

			sort.Sort(model.ReleaseByRepositoryIDAndPattern(got))

			if tc.wantErr == nil && gotErr != nil {
				failed = true
			} else if tc.wantErr != nil && gotErr == nil {
				failed = true
			} else if tc.wantErr != nil && !strings.Contains(gotErr.Error(), tc.wantErr.Error()) {
				failed = true
			} else if gotCount != tc.wantCount {
				failed = true
			} else if !reflect.DeepEqual(got, tc.want) {
				failed = true
			}

			if failed {
				t.Errorf("getNewHelmReleases() = (%+v, %d, `%s`), want (%+v, %d, `%s`)", got, gotCount, gotErr, tc.want, tc.wantCount, tc.wantErr)
			}
		})
	}
}
