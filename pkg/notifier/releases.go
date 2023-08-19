package notifier

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/ViBiOh/httputils/v4/pkg/concurrent"
	"github.com/ViBiOh/ketchup/pkg/model"
	"github.com/ViBiOh/ketchup/pkg/semver"
)

func (a App) getNewReleases(ctx context.Context) ([]model.Release, uint64, error) {
	var releases, helmReleases []model.Release
	var releasesCount, helmCount uint64

	wg := concurrent.NewLimiter(2)

	wg.Go(func() {
		var err error

		releases, releasesCount, err = a.getNewStandardReleases(ctx)
		if err != nil {
			slog.Error("fetch standard releases", "err", err)
		}
	})

	wg.Go(func() {
		var err error

		helmReleases, helmCount, err = a.getNewHelmReleases(ctx)

		if err != nil {
			slog.Error("fetch helm releases", "err", err)
		}
	})

	wg.Wait()

	return append(releases, helmReleases...), releasesCount + helmCount, nil
}

func (a App) getNewStandardReleases(ctx context.Context) ([]model.Release, uint64, error) {
	var newReleases []model.Release
	var count uint64
	var last string

	workerCount := 4
	done := make(chan struct{})
	wg := concurrent.NewLimiter(workerCount)

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
			return nil, count, fmt.Errorf("fetch standard repositories: %w", err)
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

	slog.Info("Standard repositories checked", "count", count, "new", len(newReleases))
	return newReleases, count, nil
}

func (a App) getNewRepositoryReleases(ctx context.Context, repo model.Repository) []model.Release {
	versions, err := a.repositoryService.LatestVersions(ctx, repo)
	if err != nil {
		slog.Error("get latest versions", "err", err, "name", repo.Name, "kind", repo.Kind)
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
			return nil, count, fmt.Errorf("fetch Helm repositories: %w", err)
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

	slog.Info("Helm repositories checked", "count", count, "new", len(newReleases))
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
		slog.Error("fetch helm index", "err", err, "url", url)
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
		slog.Error("parse pattern", "err", err)
		return releases
	}

	repositoryVersion, err := semver.Parse(repoVersionName)
	if err != nil {
		slog.Error("parse version", "err", err)
		return releases
	}

	if !compiledPattern.Check(upstreamVersion) || !upstreamVersion.IsGreater(repositoryVersion) {
		return releases
	}

	slog.Info("Newversion available", "pattern", repoPattern, "repo", repo.String(), "version", upstreamVersion.Name)

	return append(releases, model.NewRelease(repo, repoPattern, upstreamVersion))
}
