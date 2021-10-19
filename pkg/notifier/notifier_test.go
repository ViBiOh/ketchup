package notifier

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	authModel "github.com/ViBiOh/auth/v2/pkg/model"
	"github.com/ViBiOh/httputils/v4/pkg/clock"
	"github.com/ViBiOh/ketchup/pkg/mocks"
	"github.com/ViBiOh/ketchup/pkg/model"
	"github.com/ViBiOh/ketchup/pkg/semver"
	"github.com/golang/mock/gomock"
)

var (
	testEmail         = "nobody@localhost"
	repositoryName    = "vibioh/ketchup"
	repositoryVersion = "1.0.0"
)

func safeParse(version string) semver.Version {
	output, err := semver.Parse(version)
	if err != nil {
		fmt.Println(err)
	}
	return output
}

func TestFlags(t *testing.T) {
	var cases = []struct {
		intention string
		want      string
	}{
		{
			"simple",
			"Usage of simple:\n  -loginID uint\n    \t[notifier] Scheduler user ID {SIMPLE_LOGIN_ID} (default 1)\n  -pushUrl string\n    \t[notifier] Pushgateway URL {SIMPLE_PUSH_URL}\n",
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			fs := flag.NewFlagSet(tc.intention, flag.ContinueOnError)
			Flags(fs, "")

			var writer strings.Builder
			fs.SetOutput(&writer)
			fs.Usage()

			result := writer.String()

			if result != tc.want {
				t.Errorf("Flags() = `%s`, want `%s`", result, tc.want)
			}
		})
	}
}

func TestGetNewRepositoryReleases(t *testing.T) {
	type args struct {
		repo model.Repository
	}

	var cases = []struct {
		intention string
		instance  App
		args      args
		want      []model.Release
	}{
		{
			"empty",
			App{},
			args{
				repo: model.NewGithubRepository(0, ""),
			},
			nil,
		},
		{
			"no new",
			App{},
			args{
				repo: model.NewGithubRepository(1, repositoryName).AddVersion(model.DefaultPattern, repositoryVersion),
			},
			nil,
		},
		{
			"invalid version",
			App{},
			args{
				repo: model.NewGithubRepository(1, repositoryName).AddVersion(model.DefaultPattern, "abcde"),
			},
			nil,
		},
		{
			"not greater",
			App{},
			args{
				repo: model.NewGithubRepository(1, repositoryName).AddVersion(model.DefaultPattern, "1.1.0"),
			},
			nil,
		},
		{
			"greater",
			App{},
			args{
				repo: model.NewGithubRepository(1, repositoryName).AddVersion(model.DefaultPattern, repositoryVersion),
			},
			[]model.Release{
				model.NewRelease(model.NewGithubRepository(1, repositoryName).AddVersion(model.DefaultPattern, repositoryVersion), model.DefaultPattern, safeParse("1.1.0")),
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepositoryService := mocks.NewRepositoryService(ctrl)

			tc.instance.repositoryService = mockRepositoryService

			switch tc.intention {
			case "empty":
				mockRepositoryService.EXPECT().LatestVersions(gomock.Any()).Return(nil, nil)
			case "no new":
				mockRepositoryService.EXPECT().LatestVersions(gomock.Any()).Return(map[string]semver.Version{
					model.DefaultPattern: safeParse(repositoryVersion),
				}, nil)
			case "invalid version":
				mockRepositoryService.EXPECT().LatestVersions(gomock.Any()).Return(map[string]semver.Version{
					model.DefaultPattern: safeParse(repositoryVersion),
				}, nil)
			case "not greater":
				mockRepositoryService.EXPECT().LatestVersions(gomock.Any()).Return(map[string]semver.Version{
					model.DefaultPattern: safeParse(repositoryVersion),
				}, nil)
			case "greater":
				mockRepositoryService.EXPECT().LatestVersions(gomock.Any()).Return(map[string]semver.Version{
					model.DefaultPattern: safeParse("1.1.0"),
				}, nil)
			}

			if got := tc.instance.getNewRepositoryReleases(tc.args.repo); !reflect.DeepEqual(got, tc.want) {
				t.Errorf("getNewRepositoryReleases() = %+v, want %+v", got, tc.want)
			}
		})
	}
}

func TestGetKetchupToNotify(t *testing.T) {
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
				ctx: context.Background(),
			},
			nil,
			errors.New("failed"),
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
				ctx: context.Background(),
				releases: []model.Release{
					{
						Pattern: model.DefaultPattern,
						Version: semver.Version{
							Name: "1.1.0",
						},
						Repository: model.NewGithubRepository(1, repositoryName),
					},
					{
						Pattern: model.DefaultPattern,
						Version: semver.Version{
							Name: "1.1.0",
						},
						Repository: model.NewGithubRepository(2, "vibioh/dotfiles"),
					},
					{
						Pattern: model.DefaultPattern,
						Version: semver.Version{
							Name: "1.1.0",
						},
						Repository: model.NewGithubRepository(3, "vibioh/zzz"),
					},
				},
			},
			map[model.User][]model.Release{
				{ID: 2, Email: "guest@nowhere"}: {{
					Pattern: model.DefaultPattern,
					Version: semver.Version{
						Name: "1.1.0",
					},
					Repository: model.NewGithubRepository(1, repositoryName),
				}},
				{ID: 1, Email: testEmail}: {{
					Pattern: model.DefaultPattern,
					Version: semver.Version{
						Name: "1.1.0",
					},
					Repository: model.NewGithubRepository(1, repositoryName),
				}, {
					Pattern: model.DefaultPattern,
					Version: semver.Version{
						Name: "1.1.0",
					},
					Repository: model.NewGithubRepository(2, "vibioh/dotfiles"),
				}, {
					Pattern: model.DefaultPattern,
					Version: semver.Version{
						Name: "1.1.0",
					},
					Repository: model.NewGithubRepository(3, "vibioh/zzz"),
				}},
			},
			nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockKetchupService := mocks.NewKetchupService(ctrl)

			instance := App{
				ketchupService: mockKetchupService,
				clock:          clock.New(time.Unix(1609459200, 0)),
			}

			switch tc.intention {
			case "list error":
				mockKetchupService.EXPECT().ListForRepositories(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("failed"))
			case "empty":
				mockKetchupService.EXPECT().ListForRepositories(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil)
			case "one release, n ketchups":
				mockKetchupService.EXPECT().ListForRepositories(gomock.Any(), gomock.Any(), gomock.Any()).Return([]model.Ketchup{
					{
						Pattern:    model.DefaultPattern,
						Repository: model.NewGithubRepository(1, repositoryName),
						User:       model.NewUser(1, testEmail, authModel.NewUser(0, "")),
						Version:    repositoryVersion,
					},
					{
						Pattern:    model.DefaultPattern,
						Repository: model.NewGithubRepository(1, repositoryName),
						User:       model.NewUser(2, "guest@nowhere", authModel.NewUser(0, "")),
						Version:    repositoryVersion,
					},
					{
						Pattern:    model.DefaultPattern,
						Repository: model.NewGithubRepository(2, "vibioh/dotfiles"),
						User:       model.NewUser(1, testEmail, authModel.NewUser(0, "")),
						Version:    repositoryVersion,
					},
					{
						Pattern:    "^1.1-0",
						Repository: model.NewGithubRepository(2, "vibioh/dotfiles"),
						User:       model.NewUser(2, "guest@nowhere", authModel.NewUser(0, "")),
						Version:    repositoryVersion,
					},
					{
						Pattern:    model.DefaultPattern,
						Repository: model.NewGithubRepository(3, "vibioh/zzz"),
						User:       model.NewUser(1, testEmail, authModel.NewUser(0, "")),
						Version:    repositoryVersion,
					},
					{
						Pattern:    model.DefaultPattern,
						Repository: model.NewGithubRepository(3, "vibioh/zzz"),
						User:       model.NewUser(2, "guest@nowhere", authModel.NewUser(0, "")),
						Version:    "1.1.0",
					},
				}, nil)
			}

			got, gotErr := instance.getKetchupToNotify(tc.args.ctx, tc.args.releases)

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
		instance  App
		args      args
		wantErr   error
	}{
		{
			"empty",
			App{},
			args{
				ctx:             context.Background(),
				ketchupToNotify: nil,
			},
			nil,
		},
		{
			"no mailer",
			App{},
			args{
				ctx: context.Background(),
				ketchupToNotify: map[model.User][]model.Release{
					{
						ID:    1,
						Email: testEmail,
					}: {
						{
							Repository: model.NewGithubRepository(1, repositoryName),
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
			App{},
			args{
				ctx: context.Background(),
				ketchupToNotify: map[model.User][]model.Release{
					{
						ID:    1,
						Email: testEmail,
					}: {
						{
							Repository: model.NewGithubRepository(1, repositoryName),
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
			App{},
			args{
				ctx: context.TODO(),
				ketchupToNotify: map[model.User][]model.Release{
					{
						ID:    1,
						Email: testEmail,
					}: {
						{
							Repository: model.NewGithubRepository(1, repositoryName),
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
			App{},
			args{
				ctx: context.Background(),
				ketchupToNotify: map[model.User][]model.Release{
					{
						ID:    1,
						Email: testEmail,
					}: {
						{
							Repository: model.NewGithubRepository(1, repositoryName),
							Version: semver.Version{
								Name: repositoryVersion,
							},
						},
						{
							Repository: model.NewGithubRepository(2, "vibioh/viws"),
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
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mailerApp := mocks.NewMailer(ctrl)
			tc.instance.mailerApp = mailerApp

			switch tc.intention {
			case "no mailer":
				tc.instance.mailerApp = nil
			case "mailer disabled":
				mailerApp.EXPECT().Enabled().Return(false)
			case "mailer error":
				mailerApp.EXPECT().Enabled().Return(true)
				mailerApp.EXPECT().Send(gomock.Any(), gomock.Any()).Return(errors.New("invalid context"))
			case "multiple releases":
				mailerApp.EXPECT().Enabled().Return(false)
			}

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
