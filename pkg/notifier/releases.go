package notifier

import (
	"context"
	"fmt"

	"github.com/ViBiOh/httputils/v4/pkg/logger"
	"github.com/ViBiOh/ketchup/pkg/model"
	"github.com/ViBiOh/ketchup/pkg/semver"
)

func (a app) getNewReleases(ctx context.Context) ([]model.Release, uint64, error) {
	githubReleases, githubCount, err := a.getNewGithubReleases(ctx)
	if err != nil {
		return nil, 0, err
	}

	helmReleases, helmCount, err := a.getNewHelmReleases(ctx)
	if err != nil {
		return nil, 0, err
	}

	return append(githubReleases, helmReleases...), githubCount + helmCount, nil
}

func (a app) getNewGithubReleases(ctx context.Context) ([]model.Release, uint64, error) {
	var newReleases []model.Release
	var count uint64
	page := uint(1)

	for {
		repositories, totalCount, err := a.repositoryService.ListByKind(ctx, page, pageSize, model.Github)
		if err != nil {
			return nil, count, fmt.Errorf("unable to fetch page %d of GitHub repositories: %s", page, err)
		}

		for _, repo := range repositories {
			count++
			newReleases = append(newReleases, a.getNewRepositoryReleases(repo)...)
		}

		if uint64(page*pageSize) < totalCount {
			page++
		} else {
			logger.Info("%d GitHub repositories checked, %d new releases", count, len(newReleases))
			return newReleases, count, nil
		}
	}
}

func (a app) getNewRepositoryReleases(repo model.Repository) []model.Release {
	versions, err := a.repositoryService.LatestVersions(repo)
	if err != nil {
		logger.Error("unable to get latest versions of %s: %s", repo.Name, err)
		return nil
	}

	releases := make([]model.Release, 0)

	for pattern, version := range versions {
		releases = appendVersion(releases, version, repo, pattern, repo.Versions[pattern])
	}

	return releases
}

func (a app) getNewHelmReleases(ctx context.Context) ([]model.Release, uint64, error) {
	var newReleases []model.Release
	var count uint64
	page := uint(1)

	for {
		repositories, totalCount, err := a.repositoryService.ListByKind(ctx, page, pageSize, model.Helm)
		if err != nil {
			return nil, 0, fmt.Errorf("unable to fetch page %d of Helm repositories: %s", page, err)
		}

		if len(repositories) == 0 {
			return nil, 0, nil
		}

		repoName := repositories[0].Name
		repoWithNames := make(map[string]model.Repository)

		for _, repo := range repositories {
			count++

			if repo.Name != repoName {
				newReleases = append(newReleases, a.getFetchHelmSources(repoWithNames)...)

				repoName = repo.Name
				repoWithNames = make(map[string]model.Repository)
			}

			repoWithNames[repo.Part] = repo
		}

		newReleases = append(newReleases, a.getFetchHelmSources(repoWithNames)...)

		if uint64(page*pageSize) < totalCount {
			page++
		} else {
			logger.Info("%d Helm repositories checked, %d new releases", count, len(newReleases))
			return newReleases, count, nil
		}
	}
}

func (a app) getFetchHelmSources(repos map[string]model.Repository) []model.Release {
	if len(repos) == 0 {
		return nil
	}

	var url string

	charts := make(map[string][]string)
	for _, repo := range repos {
		if len(url) == 0 {
			url = repo.Name
		}

		index := 0
		patterns := make([]string, len(repo.Versions))
		for pattern := range repo.Versions {
			patterns[index] = pattern
			index++
		}

		charts[repo.Part] = patterns
	}

	values, err := a.helmApp.FetchIndex(url, charts)
	if err != nil {
		logger.WithField("url", url).Error("unable to fetch helm index: %s", err)
		return nil
	}

	releases := make([]model.Release, 0)

	for chartName, patterns := range values {
		repo := repos[chartName]

		for repoPattern, repoVersionName := range repo.Versions {
			releases = appendVersion(releases, patterns[repoPattern], repo, repoPattern, repoVersionName)
		}
	}

	return releases
}

func appendVersion(releases []model.Release, upstreamVersion semver.Version, repo model.Repository, repoPattern, repoVersionName string) []model.Release {
	if upstreamVersion.Name == repoVersionName {
		return releases
	}

	compiledPattern, err := semver.ParsePattern(repoPattern)
	if err != nil {
		logger.Error("unable to parse pattern: %s", err)
		return releases
	}

	repositoryVersion, err := semver.Parse(repoVersionName)
	if err != nil {
		logger.Error("unable to parse version: %s", err)
		return releases
	}

	if !compiledPattern.Check(upstreamVersion) || !upstreamVersion.IsGreater(repositoryVersion) {
		return releases
	}

	logger.Info("New `%s` version available for `%s`: %s", repoPattern, repo.String(), upstreamVersion.Name)

	return append(releases, model.NewRelease(repo, repoPattern, upstreamVersion))
}
