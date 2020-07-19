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
	"github.com/ViBiOh/mailer/pkg/client/clienttest"
)

type testKetchupService struct{}

func (tks testKetchupService) List(_ context.Context, _, _ uint) ([]model.Ketchup, uint64, error) {
	return nil, 0, nil
}

func (tks testKetchupService) ListForRepositories(ctx context.Context, repositories []model.Repository) ([]model.Ketchup, error) {
	if ctx == context.TODO() {
		return nil, errors.New("invalid context")
	}

	if len(repositories) == 0 {
		return nil, nil
	}

	if repositories[0].Name == "vibioh/ketchup" {
		return []model.Ketchup{
			{Repository: repositories[0], User: model.User{ID: 1, Email: "nobody@localhost"}},
			{Repository: repositories[0], User: model.User{ID: 2, Email: "guest@nowhere"}},
		}, nil
	}

	if repositories[0].Name == "vibioh/viws" {
		return []model.Ketchup{
			{Repository: repositories[0], User: model.User{ID: 1, Email: "nobody@localhost"}},
			{Repository: repositories[0], User: model.User{ID: 2, Email: "guest@nowhere"}},
			{Repository: repositories[1], User: model.User{ID: 2, Email: "guest@nowhere"}},
		}, nil
	}

	return nil, nil
}

func (tks testKetchupService) Create(_ context.Context, _ model.Ketchup) (model.Ketchup, error) {
	return model.NoneKetchup, nil
}

func (tks testKetchupService) Update(_ context.Context, _ model.Ketchup) (model.Ketchup, error) {
	return model.NoneKetchup, nil
}

func (tks testKetchupService) Delete(_ context.Context, _ model.Ketchup) error {
	return nil
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

func TestGetKetchupToNotify(t *testing.T) {
	firstRelease := []model.Release{
		{
			Version: semver.Version{
				Name: "1.0.1",
			},
			Repository: model.Repository{
				ID:      1,
				Name:    "vibioh/ketchup",
				Version: "1.0.0",
			},
		},
	}

	secondRelease := []model.Release{
		{
			Version: semver.Version{
				Name: "1.0.1",
			},
			Repository: model.Repository{
				ID:      1,
				Name:    "vibioh/viws",
				Version: "1.0.0",
			},
		},
		{
			Version: semver.Version{
				Name: "1.0.1",
			},
			Repository: model.Repository{
				ID:      2,
				Name:    "vibioh/dotfiles",
				Version: "1.0.0",
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
			make(map[model.User][]model.Release, 0),
			nil,
		},
		{
			"one release, n ketchups",
			args{
				ctx:      context.Background(),
				releases: firstRelease,
			},
			map[model.User][]model.Release{
				{ID: 1, Email: "nobody@localhost"}: firstRelease,
				{ID: 2, Email: "guest@nowhere"}:    firstRelease,
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
			got, gotErr := app{ketchupService: testKetchupService{}}.getKetchupToNotify(tc.args.ctx, tc.args.releases)

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
								Name: "vibioh/ketchup",
							},
							Version: semver.Version{
								Name: "1.0.0",
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
								Name: "vibioh/ketchup",
							},
							Version: semver.Version{
								Name: "1.0.0",
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
								Name: "vibioh/ketchup",
							},
							Version: semver.Version{
								Name: "1.0.0",
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
								Name: "vibioh/ketchup",
							},
							Version: semver.Version{
								Name: "1.0.0",
							},
						},
						{
							Repository: model.Repository{
								Name: "vibioh/viws",
							},
							Version: semver.Version{
								Name: "1.0.0",
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
