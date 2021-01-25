package scheduler

import (
	"context"
	"errors"
	"flag"
	"reflect"
	"strings"
	"testing"

	"github.com/ViBiOh/ketchup/pkg/model"
	"github.com/ViBiOh/ketchup/pkg/semver"
	"github.com/ViBiOh/ketchup/pkg/service/ketchup/ketchuptest"
	"github.com/ViBiOh/ketchup/pkg/service/repository/repositorytest"
	"github.com/ViBiOh/mailer/pkg/client/clienttest"
)

var (
	repositoryName    = "vibioh/ketchup"
	repositoryVersion = "1.0.0"
)

func ketchupRepository() model.Repository {
	return model.Repository{
		ID:   1,
		Kind: model.Github,
		Name: repositoryName,
		Versions: map[string]string{
			model.DefaultPattern: repositoryVersion,
		},
	}
}

func TestFlags(t *testing.T) {
	var cases = []struct {
		intention string
		want      string
	}{
		{
			"simple",
			"Usage of simple:\n  -hour string\n    \t[scheduler] Hour of cron, 24-hour format {SIMPLE_HOUR} (default \"08:00\")\n  -loginID uint\n    \t[scheduler] Scheduler user ID {SIMPLE_LOGIN_ID} (default 1)\n  -timezone string\n    \t[scheduler] Timezone {SIMPLE_TIMEZONE} (default \"Europe/Paris\")\n",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.intention, func(t *testing.T) {
			fs := flag.NewFlagSet(testCase.intention, flag.ContinueOnError)
			Flags(fs, "")

			var writer strings.Builder
			fs.SetOutput(&writer)
			fs.Usage()

			result := writer.String()

			if result != testCase.want {
				t.Errorf("Flags() = `%s`, want `%s`", result, testCase.want)
			}
		})
	}
}

func TestGetNewReleases(t *testing.T) {
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
				repositoryService: repositorytest.New().SetList(nil, 0, errors.New("failed")),
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
				repositoryService: repositorytest.New().SetList([]model.Repository{ketchupRepository()}, 0, nil).SetLatestVersions(nil, errors.New("failed")),
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
				repositoryService: repositorytest.New().SetList([]model.Repository{
					ketchupRepository(),
				}, 0, nil).SetLatestVersions(map[string]semver.Version{
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
			"update error",
			app{
				repositoryService: repositorytest.New().SetList([]model.Repository{ketchupRepository()}, 0, nil).SetLatestVersions(map[string]semver.Version{
					model.DefaultPattern: {
						Name:  "1.1.0",
						Major: 1,
						Minor: 1,
					},
				}, nil).SetUpdate(errors.New("failed")),
			},
			args{
				ctx: context.Background(),
			},
			nil,
			errors.New("unable to update repo `vibioh/ketchup`"),
		},
		{
			"success",
			app{
				repositoryService: repositorytest.New().SetList([]model.Repository{ketchupRepository()}, 0, nil).SetLatestVersions(map[string]semver.Version{
					model.DefaultPattern: {
						Name:  "1.1.0",
						Major: 1,
						Minor: 1,
					},
				}, nil).SetUpdate(nil),
			},
			args{
				ctx: context.Background(),
			},
			[]model.Release{model.NewRelease(
				model.Repository{
					ID:   1,
					Kind: model.Github,
					Name: repositoryName,
					Versions: map[string]string{
						model.DefaultPattern: "1.1.0",
					},
				},
				model.DefaultPattern,
				semver.Version{
					Name:  "1.1.0",
					Major: 1,
					Minor: 1,
					Patch: 0,
				},
			)},
			nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			pageSize = 1
			got, gotErr := tc.instance.getNewReleases(tc.args.ctx)
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
				t.Errorf("getNewReleases() = (%+v, `%s`), want (%+v, `%s`)", got, gotErr, tc.want, tc.wantErr)
			}
		})
	}
}

func TestCheckRepositoryVersion(t *testing.T) {
	type args struct {
		repo model.Repository
	}

	var cases = []struct {
		intention string
		instance  app
		args      args
		want      []model.Release
	}{
		{
			"empty",
			app{
				repositoryService: repositorytest.New().SetLatestVersions(nil, nil),
			},
			args{
				repo: model.Repository{},
			},
			make([]model.Release, 0),
		},
		{
			"no new",
			app{
				repositoryService: repositorytest.New().SetLatestVersions(map[string]semver.Version{
					model.DefaultPattern: {
						Name:  repositoryVersion,
						Major: 1,
						Minor: 0,
						Patch: 0,
					},
				}, nil),
			},
			args{
				repo: ketchupRepository(),
			},
			make([]model.Release, 0),
		},
		{
			"invalid version",
			app{
				repositoryService: repositorytest.New().SetLatestVersions(map[string]semver.Version{
					model.DefaultPattern: {
						Name:  repositoryVersion,
						Major: 1,
						Minor: 0,
						Patch: 0,
					},
				}, nil),
			},
			args{
				repo: model.Repository{
					ID:   1,
					Name: repositoryName,
					Versions: map[string]string{
						model.DefaultPattern: "abcde",
					},
				},
			},
			make([]model.Release, 0),
		},
		{
			"not greater",
			app{
				repositoryService: repositorytest.New().SetLatestVersions(map[string]semver.Version{
					model.DefaultPattern: {
						Name:  repositoryVersion,
						Major: 1,
						Minor: 0,
						Patch: 0,
					},
				}, nil),
			},
			args{
				repo: model.Repository{
					ID:   1,
					Name: repositoryName,
					Versions: map[string]string{
						model.DefaultPattern: "1.1.0",
					},
				},
			},
			make([]model.Release, 0),
		},
		{
			"greater",
			app{
				repositoryService: repositorytest.New().SetLatestVersions(map[string]semver.Version{
					model.DefaultPattern: {
						Name:  "1.1.0",
						Major: 1,
						Minor: 1,
						Patch: 0,
					},
				}, nil),
			},
			args{
				repo: ketchupRepository(),
			},
			[]model.Release{
				model.NewRelease(ketchupRepository(), model.DefaultPattern, semver.Version{
					Name:  "1.1.0",
					Major: 1,
					Minor: 1,
					Patch: 0,
				}),
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			if got := tc.instance.checkRepositoryVersion(tc.args.repo); !reflect.DeepEqual(got, tc.want) {
				t.Errorf("checkRepositoryVersion() = %+v, want %+v", got, tc.want)
			}
		})
	}
}

func TestGetKetchupToNotify(t *testing.T) {
	firstRelease := []model.Release{
		{
			Pattern: model.DefaultPattern,
			Version: semver.Version{
				Name: "1.1.0",
			},
			Repository: model.Repository{
				ID:   1,
				Name: repositoryName,
				Versions: map[string]string{
					model.DefaultPattern: "1.1.0",
					"latest":             "1.1.0-beta",
				},
			},
		},
		{
			Pattern: "latest",
			Version: semver.Version{
				Name: "1.1.0-beta",
			},
			Repository: model.Repository{
				ID:   1,
				Name: repositoryName,
				Versions: map[string]string{
					model.DefaultPattern: "1.1.0",
					"latest":             "1.1.0-beta",
				},
			},
		},
	}

	secondRelease := []model.Release{
		{
			Pattern: model.DefaultPattern,
			Version: semver.Version{
				Name: "1.0.1",
			},
			Repository: model.Repository{
				ID:       1,
				Name:     "vibioh/viws",
				Versions: map[string]string{model.DefaultPattern: "1.1.0"},
			},
		},
		{
			Pattern: model.DefaultPattern,
			Version: semver.Version{
				Name: "1.0.1",
			},
			Repository: model.Repository{
				ID:       2,
				Name:     "vibioh/dotfiles",
				Versions: map[string]string{model.DefaultPattern: "1.1.0"},
			},
		},
	}

	type args struct {
		ctx      context.Context
		releases []model.Release
	}

	var cases = []struct {
		intention string
		args      args
		want      map[model.User][]model.Release
		wantErr   error
	}{
		{
			"list error",
			args{
				ctx: context.TODO(),
			},
			nil,
			errors.New("invalid context"),
		},
		{
			"empty",
			args{
				ctx: context.Background(),
			},
			make(map[model.User][]model.Release),
			nil,
		},
		{
			"one release, n ketchups",
			args{
				ctx:      context.Background(),
				releases: firstRelease,
			},
			map[model.User][]model.Release{
				{ID: 2, Email: "guest@nowhere"}:    {firstRelease[0]},
				{ID: 1, Email: "nobody@localhost"}: {firstRelease[1], firstRelease[0]},
			},
			nil,
		},
		{
			"one release, n ketchups",
			args{
				ctx:      context.Background(),
				releases: secondRelease,
			},
			map[model.User][]model.Release{
				{ID: 1, Email: "nobody@localhost"}: {secondRelease[0]},
				{ID: 2, Email: "guest@nowhere"}:    secondRelease,
			},
			nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			got, gotErr := app{ketchupService: ketchuptest.NewApp()}.getKetchupToNotify(tc.args.ctx, tc.args.releases)

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
				t.Errorf("getKetchupToNotify() = (%+v, `%s`), want (%+v, `%s`)", got, gotErr, tc.want, tc.wantErr)
			}
		})
	}
}

func TestSendNotification(t *testing.T) {
	type args struct {
		ctx             context.Context
		ketchupToNotify map[model.User][]model.Release
	}

	var cases = []struct {
		intention string
		instance  app
		args      args
		wantErr   error
	}{
		{
			"empty",
			app{},
			args{
				ctx:             context.Background(),
				ketchupToNotify: nil,
			},
			nil,
		},
		{
			"no mailer",
			app{},
			args{
				ctx: context.Background(),
				ketchupToNotify: map[model.User][]model.Release{
					{
						ID:    1,
						Email: "nobody@localhost",
					}: {
						{
							Repository: model.Repository{
								Name: repositoryName,
							},
							Version: semver.Version{
								Name: repositoryVersion,
							},
						},
					},
				},
			},
			nil,
		},
		{
			"mailer disabled",
			app{
				mailerApp: clienttest.New(false),
			},
			args{
				ctx: context.Background(),
				ketchupToNotify: map[model.User][]model.Release{
					{
						ID:    1,
						Email: "nobody@localhost",
					}: {
						{
							Repository: model.Repository{
								Name: repositoryName,
							},
							Version: semver.Version{
								Name: repositoryVersion,
							},
						},
					},
				},
			},
			nil,
		},
		{
			"mailer error",
			app{
				mailerApp: clienttest.New(true),
			},
			args{
				ctx: context.TODO(),
				ketchupToNotify: map[model.User][]model.Release{
					{
						ID:    1,
						Email: "nobody@localhost",
					}: {
						{
							Repository: model.Repository{
								Name: repositoryName,
							},
							Version: semver.Version{
								Name: repositoryVersion,
							},
						},
					},
				},
			},
			errors.New("unable to send email to nobody@localhost: invalid context"),
		},
		{
			"multiple releases",
			app{
				mailerApp: clienttest.New(true),
			},
			args{
				ctx: context.Background(),
				ketchupToNotify: map[model.User][]model.Release{
					{
						ID:    1,
						Email: "nobody@localhost",
					}: {
						{
							Repository: model.Repository{
								Name: repositoryName,
							},
							Version: semver.Version{
								Name: repositoryVersion,
							},
						},
						{
							Repository: model.Repository{
								Name: "vibioh/viws",
							},
							Version: semver.Version{
								Name: repositoryVersion,
							},
						},
					},
				},
			},
			nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			gotErr := tc.instance.sendNotification(tc.args.ctx, tc.args.ketchupToNotify)

			failed := false

			if tc.wantErr == nil && gotErr != nil {
				failed = true
			} else if tc.wantErr != nil && gotErr == nil {
				failed = true
			} else if tc.wantErr != nil && !strings.Contains(gotErr.Error(), tc.wantErr.Error()) {
				failed = true
			}

			if failed {
				t.Errorf("sendNotification() = `%s`, want `%s`", gotErr, tc.wantErr)
			}
		})
	}
}
