package notifier

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/ViBiOh/ketchup/pkg/mocks"
	"github.com/ViBiOh/ketchup/pkg/model"
	"github.com/ViBiOh/ketchup/pkg/semver"
	"go.uber.org/mock/gomock"
)

func TestGetNewStandardReleases(t *testing.T) {
	t.Parallel()

	type args struct {
		ctx context.Context
	}

	cases := map[string]struct {
		instance Service
		args     args
		want     []model.Release
		wantErr  error
	}{
		"list error": {
			Service{},
			args{
				ctx: context.TODO(),
			},
			nil,
			errors.New("failed"),
		},
		"github error": {
			Service{},
			args{
				ctx: context.TODO(),
			},
			nil,
			nil,
		},
		"same version": {
			Service{},
			args{
				ctx: context.TODO(),
			},
			nil,
			nil,
		},
		"success": {
			Service{},
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
		t.Run(intention, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			mockRepositoryService := mocks.NewRepositoryService(ctrl)

			testCase.instance.repository = mockRepositoryService

			switch intention {
			case "list error":
				mockRepositoryService.EXPECT().ListByKinds(gomock.Any(), gomock.Any(), gomock.Any(), model.Github, model.Docker, model.NPM, model.Pypi).Return(nil, errors.New("failed"))
			case "github error":
				mockRepositoryService.EXPECT().ListByKinds(gomock.Any(), gomock.Any(), gomock.Any(), model.Github, model.Docker, model.NPM, model.Pypi).Return([]model.Repository{
					model.NewGithubRepository(model.Identifier(1), repositoryName).AddVersion(model.DefaultPattern, repositoryVersion),
				}, nil)
				mockRepositoryService.EXPECT().LatestVersions(gomock.Any(), gomock.Any()).Return(nil, errors.New("failed"))
			case "same version":
				mockRepositoryService.EXPECT().ListByKinds(gomock.Any(), gomock.Any(), gomock.Any(), model.Github, model.Docker, model.NPM, model.Pypi).Return([]model.Repository{
					model.NewGithubRepository(model.Identifier(1), repositoryName).AddVersion(model.DefaultPattern, repositoryVersion),
				}, nil)
				mockRepositoryService.EXPECT().LatestVersions(gomock.Any(), gomock.Any()).Return(map[string]semver.Version{
					model.DefaultPattern: {
						Name: "1.1.0",
					},
				}, nil)
			case "success":
				mockRepositoryService.EXPECT().ListByKinds(gomock.Any(), gomock.Any(), gomock.Any(), model.Github, model.Docker, model.NPM, model.Pypi).Return([]model.Repository{
					model.NewGithubRepository(model.Identifier(1), repositoryName).AddVersion(model.DefaultPattern, repositoryVersion),
				}, nil)
				mockRepositoryService.EXPECT().LatestVersions(gomock.Any(), gomock.Any()).Return(map[string]semver.Version{
					model.DefaultPattern: safeParse("1.1.0"),
					"1.0":                safeParse("1.0"),
				}, nil)
			}

			got, gotErr := testCase.instance.getNewStandardReleases(testCase.args.ctx)

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
