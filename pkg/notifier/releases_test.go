package notifier

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/ViBiOh/ketchup/pkg/model"
	"github.com/ViBiOh/ketchup/pkg/semver"
	"github.com/ViBiOh/ketchup/pkg/service/repository/repositorytest"
)

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
				repositoryService: repositorytest.New().SetList([]model.Repository{
					model.NewRepository(1, model.Github, repositoryName).AddVersion(model.DefaultPattern, repositoryVersion)}, 0, nil).SetLatestVersions(nil, errors.New("failed")),
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
					model.NewRepository(1, model.Github, repositoryName).AddVersion(model.DefaultPattern, repositoryVersion),
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
			"success",
			app{
				repositoryService: repositorytest.New().SetList([]model.Repository{
					model.NewRepository(1, model.Github, repositoryName).AddVersion(model.DefaultPattern, repositoryVersion),
				}, 0, nil).SetLatestVersions(map[string]semver.Version{
					model.DefaultPattern: safeParse("1.1.0"),
				}, nil).SetUpdate(nil),
			},
			args{
				ctx: context.Background(),
			},
			[]model.Release{model.NewRelease(
				model.NewRepository(1, model.Github, repositoryName).AddVersion(model.DefaultPattern, repositoryVersion),
				model.DefaultPattern,
				safeParse("1.1.0"),
			)},
			nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.intention, func(t *testing.T) {
			pageSize = 1
			got, _, gotErr := tc.instance.getNewReleases(tc.args.ctx)
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
