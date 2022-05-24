package notifier

import (
	"context"
	"fmt"
	"sync"

	"github.com/ViBiOh/httputils/v4/pkg/concurrent"
	"github.com/ViBiOh/httputils/v4/pkg/logger"
	"github.com/ViBiOh/ketchup/pkg/model"
	"github.com/ViBiOh/ketchup/pkg/semver"
)

func (a App) getNewReleases(ctx context.Context) ([]model.Release, uint64, error) {
	var releases, helmReleases []model.Release
	var releasesCount, helmCount uint64

	wg := concurrent.NewLimited(2)

	wg.Go(func() {
		var err error

		releases, releasesCount, err = a.getNewStandardReleases(ctx)
		if err != nil {
			logger.Error("unable to fetch standard releases: %s", err)
		}
	})

	wg.Go(func() {
		var err error

		helmReleases, helmCount, err = a.getNewHelmReleases(ctx)

		if err != nil {
			logger.Error("unable to fetch helm releases: %s", err)
		}
	})

	wg.Wait()

	return append(releases, helmReleases...), releasesCount + helmCount, nil
}

func (a App) getNewStandardReleases(ctx context.Context) ([]model.Release, uint64, error) {
	var newReleases []model.Release
	var count uint64
	var last string

	workerCount := uint64(4)
	done := make(chan struct{})
	wg := concurrent.NewLimited(workerCount)

	workerOutput := make(chan []model.Release, workerCount)
	closeWorker := func() {
		close(workerOutput)
	}

	var end sync.Once
	defer end.Do(closeWorker)

	go func() {
		defer close(done)

		for newRelease := range workerOutput {
			newReleases = append(newReleases, newRelease...)
		}
	}()

	for {
		repositories, _, err := a.repositoryService.ListByKinds(ctx, pageSize, last, model.Github, model.Docker, model.NPM, model.Pypi)
		if err != nil {
			return nil, count, fmt.Errorf("unable to fetch standard repositories: %s", err)
		}

		repoCount := len(repositories)
		if repoCount == 0 {
			break
		}

		for _, repo := range repositories {
			repo := repo
			count++

			wg.Go(func() {
				workerOutput <- a.getNewRepositoryReleases(ctx, repo)
			})
		}

		if repoCount < int(pageSize) {
			break
		}

		lastRepo := repositories[len(repositories)-1]
		last = fmt.Sprintf("%s|%s", lastRepo.Name, lastRepo.Part)
	}

	wg.Wait()
	end.Do(closeWorker)
	<-done

	logger.Info("%d standard repositories checked, %d new releases", count, len(newReleases))
	return newReleases, count, nil
}

func (a App) getNewRepositoryReleases(ctx context.Context, repo model.Repository) []model.Release {
	versions, err := a.repositoryService.LatestVersions(ctx, repo)
	if err != nil {
		logger.Error("unable to get latest versions of %s `%s`: %s", repo.Kind, repo.Name, err)
		return nil
	}

	var releases []model.Release

	for pattern, version := range versions {
		releases = appendVersion(releases, version, repo, pattern, repo.Versions[pattern])
	}

	return releases
}

func (a App) getNewHelmReleases(ctx context.Context) ([]model.Release, uint64, error) {
	var newReleases []model.Release
	var count uint64
	var last string

	for {
		repositories, _, err := a.repositoryService.ListByKinds(ctx, pageSize, last, model.Helm)
		if err != nil {
			return nil, count, fmt.Errorf("unable to fetch Helm repositories: %s", err)
		}

		repoCount := len(repositories)
		if repoCount == 0 {
			break
		}

		repoName := repositories[0].Name
		repoWithNames := make(map[string]model.Repository)

		for _, repo := range repositories {
			count++

			if repo.Name != repoName {
				newReleases = append(newReleases, a.getFetchHelmSources(ctx, repoWithNames)...)

				repoName = repo.Name
				repoWithNames = make(map[string]model.Repository)
			}

			repoWithNames[repo.Part] = repo
		}

		newReleases = append(newReleases, a.getFetchHelmSources(ctx, repoWithNames)...)

		if repoCount < int(pageSize) {
			break
		}

		lastRepo := repositories[len(repositories)-1]
		last = fmt.Sprintf("%s|%s", lastRepo.Name, lastRepo.Part)
	}

	logger.Info("%d Helm repositories checked, %d new releases", count, len(newReleases))
	return newReleases, count, nil
}

func (a App) getFetchHelmSources(ctx context.Context, repos map[string]model.Repository) []model.Release {
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

	values, err := a.helmApp.FetchIndex(ctx, url, charts)
	if err != nil {
		logger.WithField("url", url).Error("unable to fetch helm index: %s", err)
		return nil
	}

	var releases []model.Release

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
