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
	"github.com/ViBiOh/ketchup/pkg/mocks"
	"github.com/ViBiOh/ketchup/pkg/model"
	"github.com/ViBiOh/ketchup/pkg/semver"
	"go.uber.org/mock/gomock"
)

var (
	testEmail         = "nobody@localhost"
	repositoryName    = "vibioh/ketchup"
	repositoryVersion = "1.0.0"
)

func safeParse(version string) semver.Version {
	output, err := semver.Parse(version, "")
	if err != nil {
		fmt.Println(err)
	}
	return output
}

func TestFlags(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		want string
	}{
		"simple": {
			"Usage of simple:\n  -dryRun\n    \t[notifier] Run in dry-run ${SIMPLE_DRY_RUN}\n",
		},
	}

	for intention, testCase := range cases {
		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			fs := flag.NewFlagSet(intention, flag.ContinueOnError)
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

func TestGetNewRepositoryReleases(t *testing.T) {
	t.Parallel()

	type args struct {
		repo model.Repository
	}

	cases := map[string]struct {
		instance Service
		args     args
		want     []model.Release
	}{
		"empty": {
			Service{},
			args{
				repo: model.NewGithubRepository(model.Identifier(0), ""),
			},
			nil,
		},
		"no new": {
			Service{},
			args{
				repo: model.NewGithubRepository(model.Identifier(1), repositoryName).AddVersion(model.DefaultPattern, repositoryVersion),
			},
			nil,
		},
		"invalid version": {
			Service{},
			args{
				repo: model.NewGithubRepository(model.Identifier(1), repositoryName).AddVersion(model.DefaultPattern, "abcde"),
			},
			nil,
		},
		"not greater": {
			Service{},
			args{
				repo: model.NewGithubRepository(model.Identifier(1), repositoryName).AddVersion(model.DefaultPattern, "1.1.0"),
			},
			nil,
		},
		"greater": {
			Service{},
			args{
				repo: model.NewGithubRepository(model.Identifier(1), repositoryName).AddVersion(model.DefaultPattern, repositoryVersion),
			},
			[]model.Release{
				model.NewRelease(model.NewGithubRepository(model.Identifier(1), repositoryName).AddVersion(model.DefaultPattern, repositoryVersion), model.DefaultPattern, safeParse("1.1.0")),
			},
		},
	}

	for intention, testCase := range cases {
		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			mockRepositoryService := mocks.NewRepositoryService(ctrl)

			testCase.instance.repository = mockRepositoryService

			switch intention {
			case "empty":
				mockRepositoryService.EXPECT().LatestVersions(gomock.Any(), gomock.Any()).Return(nil, nil)
			case "no new":
				mockRepositoryService.EXPECT().LatestVersions(gomock.Any(), gomock.Any()).Return(map[string]semver.Version{
					model.DefaultPattern: safeParse(repositoryVersion),
				}, nil)
			case "invalid version":
				mockRepositoryService.EXPECT().LatestVersions(gomock.Any(), gomock.Any()).Return(map[string]semver.Version{
					model.DefaultPattern: safeParse(repositoryVersion),
				}, nil)
			case "not greater":
				mockRepositoryService.EXPECT().LatestVersions(gomock.Any(), gomock.Any()).Return(map[string]semver.Version{
					model.DefaultPattern: safeParse(repositoryVersion),
				}, nil)
			case "greater":
				mockRepositoryService.EXPECT().LatestVersions(gomock.Any(), gomock.Any()).Return(map[string]semver.Version{
					model.DefaultPattern: safeParse("1.1.0"),
				}, nil)
			}

			if got := testCase.instance.getNewRepositoryReleases(context.TODO(), testCase.args.repo); !reflect.DeepEqual(got, testCase.want) {
				t.Errorf("getNewRepositoryReleases() = %+v, want %+v", got, testCase.want)
			}
		})
	}
}

func TestGetKetchupToNotify(t *testing.T) {
	t.Parallel()

	type args struct {
		ctx      context.Context
		releases []model.Release
	}

	cases := map[string]struct {
		args     args
		want     map[model.User][]model.Release
		wantUser map[model.Identifier]uint8
		wantErr  error
	}{
		"list error": {
			args{
				ctx: context.TODO(),
			},
			nil,
			nil,
			errors.New("failed"),
		},
		"empty": {
			args{
				ctx: context.TODO(),
			},
			make(map[model.User][]model.Release),
			make(map[model.Identifier]uint8),
			nil,
		},
		"one release, n ketchups": {
			args{
				ctx: context.TODO(),
				releases: []model.Release{
					{
						Pattern: model.DefaultPattern,
						Version: semver.Version{
							Name: "1.1.0",
						},
						Repository: model.NewGithubRepository(model.Identifier(1), repositoryName),
					},
					{
						Pattern: model.DefaultPattern,
						Version: semver.Version{
							Name: "1.1.0",
						},
						Repository: model.NewGithubRepository(model.Identifier(2), "vibioh/dotfiles"),
					},
					{
						Pattern: model.DefaultPattern,
						Version: semver.Version{
							Name: "1.1.0",
						},
						Repository: model.NewGithubRepository(model.Identifier(3), "vibioh/zzz"),
					},
				},
			},
			map[model.User][]model.Release{
				{ID: 2, Email: "guest@nowhere"}: {{
					Pattern: model.DefaultPattern,
					Version: semver.Version{
						Name: "1.1.0",
					},
					Repository: model.NewGithubRepository(model.Identifier(1), repositoryName),
				}},
				{ID: 1, Email: testEmail}: {{
					Pattern: model.DefaultPattern,
					Version: semver.Version{
						Name: "1.1.0",
					},
					Repository: model.NewGithubRepository(model.Identifier(1), repositoryName),
				}, {
					Pattern: model.DefaultPattern,
					Version: semver.Version{
						Name: "1.1.0",
					},
					Repository: model.NewGithubRepository(model.Identifier(2), "vibioh/dotfiles"),
				}, {
					Pattern: model.DefaultPattern,
					Version: semver.Version{
						Name: "1.1.0",
					},
					Repository: model.NewGithubRepository(model.Identifier(3), "vibioh/zzz"),
				}},
			},
			map[model.Identifier]uint8{
				1: 3,
				2: 1,
			},
			nil,
		},
	}

	for intention, testCase := range cases {
		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			mockKetchupService := mocks.NewKetchupService(ctrl)

			instance := Service{
				ketchup: mockKetchupService,
				clock:   func() time.Time { return time.Unix(1609459200, 0) },
			}

			switch intention {
			case "list error":
				mockKetchupService.EXPECT().ListForRepositories(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("failed"))
			case "empty":
				mockKetchupService.EXPECT().ListForRepositories(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil)
			case "one release, n ketchups":
				mockKetchupService.EXPECT().ListForRepositories(gomock.Any(), gomock.Any(), gomock.Any()).Return([]model.Ketchup{
					{
						Pattern:    model.DefaultPattern,
						Repository: model.NewGithubRepository(model.Identifier(1), repositoryName),
						User:       model.NewUser(1, testEmail, authModel.NewUser(0, "")),
						Version:    repositoryVersion,
						Frequency:  model.Daily,
					},
					{
						Pattern:    model.DefaultPattern,
						Repository: model.NewGithubRepository(model.Identifier(1), repositoryName),
						User:       model.NewUser(2, "guest@nowhere", authModel.NewUser(0, "")),
						Version:    repositoryVersion,
						Frequency:  model.Daily,
					},
					{
						Pattern:    model.DefaultPattern,
						Repository: model.NewGithubRepository(model.Identifier(2), "vibioh/dotfiles"),
						User:       model.NewUser(1, testEmail, authModel.NewUser(0, "")),
						Version:    repositoryVersion,
						Frequency:  model.Daily,
					},
					{
						Pattern:    "^1.1-0",
						Repository: model.NewGithubRepository(model.Identifier(2), "vibioh/dotfiles"),
						User:       model.NewUser(2, "guest@nowhere", authModel.NewUser(0, "")),
						Version:    repositoryVersion,
						Frequency:  model.Daily,
					},
					{
						Pattern:    model.DefaultPattern,
						Repository: model.NewGithubRepository(model.Identifier(3), "vibioh/zzz"),
						User:       model.NewUser(1, testEmail, authModel.NewUser(0, "")),
						Version:    repositoryVersion,
						Frequency:  model.Weekly,
					},
					{
						Pattern:    model.DefaultPattern,
						Repository: model.NewGithubRepository(model.Identifier(3), "vibioh/zzz"),
						User:       model.NewUser(2, "guest@nowhere", authModel.NewUser(0, "")),
						Version:    "1.1.0",
						Frequency:  model.Daily,
					},
				}, nil)
			}

			got, gotUser, gotErr := instance.getKetchupToNotify(testCase.args.ctx, testCase.args.releases)

			failed := false

			if testCase.wantErr == nil && gotErr != nil {
				failed = true
			} else if testCase.wantErr != nil && gotErr == nil {
				failed = true
			} else if testCase.wantErr != nil && !strings.Contains(gotErr.Error(), testCase.wantErr.Error()) {
				failed = true
			} else if !reflect.DeepEqual(got, testCase.want) {
				failed = true
			} else if !reflect.DeepEqual(gotUser, testCase.wantUser) {
				failed = true
			}

			if failed {
				t.Errorf("getKetchupToNotify() = (%+v, %+v, `%s`), want (%+v, %+v, `%s`)", got, gotUser, gotErr, testCase.want, testCase.wantUser, testCase.wantErr)
			}
		})
	}
}

func TestSendNotification(t *testing.T) {
	t.Parallel()

	type args struct {
		ctx             context.Context
		ketchupToNotify map[model.User][]model.Release
		userStatuses    map[model.Identifier]uint8
	}

	cases := map[string]struct {
		instance Service
		args     args
		wantErr  error
	}{
		"empty": {
			Service{},
			args{
				ctx:             context.TODO(),
				ketchupToNotify: nil,
			},
			nil,
		},
		"no mailer": {
			Service{},
			args{
				ctx: context.TODO(),
				ketchupToNotify: map[model.User][]model.Release{
					{
						ID:    1,
						Email: testEmail,
					}: {
						{
							Repository: model.NewGithubRepository(model.Identifier(1), repositoryName),
							Version: semver.Version{
								Name: repositoryVersion,
							},
						},
					},
				},
			},
			nil,
		},
		"mailer disabled": {
			Service{},
			args{
				ctx: context.TODO(),
				ketchupToNotify: map[model.User][]model.Release{
					{
						ID:    1,
						Email: testEmail,
					}: {
						{
							Repository: model.NewGithubRepository(model.Identifier(1), repositoryName),
							Version: semver.Version{
								Name: repositoryVersion,
							},
						},
					},
				},
			},
			nil,
		},
		"mailer error": {
			Service{},
			args{
				ctx: context.TODO(),
				ketchupToNotify: map[model.User][]model.Release{
					{
						ID:    1,
						Email: testEmail,
					}: {
						{
							Repository: model.NewGithubRepository(model.Identifier(1), repositoryName),
							Version: semver.Version{
								Name: repositoryVersion,
							},
						},
					},
				},
			},
			errors.New("send email to nobody@localhost: invalid context"),
		},
		"multiple releases": {
			Service{},
			args{
				ctx: context.TODO(),
				ketchupToNotify: map[model.User][]model.Release{
					{
						ID:    1,
						Email: testEmail,
					}: {
						{
							Repository: model.NewGithubRepository(model.Identifier(1), repositoryName),
							Version: semver.Version{
								Name: repositoryVersion,
							},
						},
						{
							Repository: model.NewGithubRepository(model.Identifier(2), "vibioh/viws"),
							Version: semver.Version{
								Name: repositoryVersion,
							},
						},
					},
				},
				userStatuses: map[model.Identifier]uint8{
					1: 1,
					2: 3,
				},
			},
			nil,
		},
	}

	for intention, testCase := range cases {
		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			mailerService := mocks.NewMailer(ctrl)
			testCase.instance.mailer = mailerService

			switch intention {
			case "no mailer":
				testCase.instance.mailer = nil
			case "mailer disabled":
				mailerService.EXPECT().Enabled().Return(false)
			case "mailer error":
				mailerService.EXPECT().Enabled().Return(true)
				mailerService.EXPECT().Send(gomock.Any(), gomock.Any()).Return(errors.New("invalid context"))
			case "multiple releases":
				mailerService.EXPECT().Enabled().Return(false)
			}

			gotErr := testCase.instance.sendNotification(testCase.args.ctx, "ketchup", testCase.args.ketchupToNotify, testCase.args.userStatuses)

			failed := false

			if testCase.wantErr == nil && gotErr != nil {
				failed = true
			} else if testCase.wantErr != nil && gotErr == nil {
				failed = true
			} else if testCase.wantErr != nil && !strings.Contains(gotErr.Error(), testCase.wantErr.Error()) {
				failed = true
			}

			if failed {
				t.Errorf("sendNotification() = `%s`, want `%s`", gotErr, testCase.wantErr)
			}
		})
	}
}
