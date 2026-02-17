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

func (s Service) getNewReleases(ctx context.Context) ([]model.Release, error) {
	var releases, helmReleases []model.Release

	wg := concurrent.NewLimiter(2)

	wg.Go(func() {
		var err error

		releases, err = s.getNewStandardReleases(ctx)
		if err != nil {
			slog.LogAttrs(ctx, slog.LevelError, "fetch standard releases", slog.Any("error", err))
		}
	})

	wg.Go(func() {
		var err error

		helmReleases, err = s.getNewHelmReleases(ctx)
		if err != nil {
			slog.LogAttrs(ctx, slog.LevelError, "fetch helm releases", slog.Any("error", err))
		}
	})

	wg.Wait()

	return append(releases, helmReleases...), nil
}

func (s Service) getNewStandardReleases(ctx context.Context) ([]model.Release, error) {
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
		repositories, err := s.repository.ListByKinds(ctx, pageSize, last, model.Github, model.Docker, model.NPM, model.Pypi)
		if err != nil {
			return nil, fmt.Errorf("fetch standard repositories: %w", err)
		}

		repoCount := len(repositories)
		if repoCount == 0 {
			break
		}

		for _, repo := range repositories {
			count++

			wg.Go(func() {
				workerOutput <- s.getNewRepositoryReleases(ctx, repo)
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

	slog.LogAttrs(ctx, slog.LevelInfo, "Standard repositories checked", slog.Uint64("count", count), slog.Int("new", len(newReleases)))
	return newReleases, nil
}

func (s Service) getNewRepositoryReleases(ctx context.Context, repo model.Repository) []model.Release {
	versions, err := s.repository.LatestVersions(ctx, repo)
	if err != nil {
		slog.LogAttrs(ctx, slog.LevelError, "get latest versions", slog.String("name", repo.Name), slog.String("kind", repo.Kind.String()), slog.Any("error", err))
		return nil
	}

	var releases []model.Release

	for pattern, version := range versions {
		releases = appendVersion(ctx, releases, version, repo, pattern, repo.Versions[pattern])
	}

	return releases
}

func (s Service) getNewHelmReleases(ctx context.Context) ([]model.Release, error) {
	var newReleases []model.Release
	var count uint64
	var last string

	for {
		repositories, err := s.repository.ListByKinds(ctx, pageSize, last, model.Helm)
		if err != nil {
			return nil, fmt.Errorf("fetch Helm repositories: %w", err)
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
				newReleases = append(newReleases, s.getFetchHelmSources(ctx, repoWithNames)...)

				repoName = repo.Name
				repoWithNames = make(map[string]model.Repository)
			}

			repoWithNames[repo.Part] = repo
		}

		newReleases = append(newReleases, s.getFetchHelmSources(ctx, repoWithNames)...)

		if repoCount < int(pageSize) {
			break
		}

		lastRepo := repositories[len(repositories)-1]
		last = fmt.Sprintf("%s|%s", lastRepo.Name, lastRepo.Part)
	}

	slog.LogAttrs(ctx, slog.LevelInfo, "Helm repositories checked", slog.Uint64("count", count), slog.Int("new", len(newReleases)))
	return newReleases, nil
}

func (s Service) getFetchHelmSources(ctx context.Context, repos map[string]model.Repository) []model.Release {
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

	values, err := s.helm.FetchIndex(ctx, url, charts)
	if err != nil {
		slog.LogAttrs(ctx, slog.LevelError, "fetch helm index", slog.String("url", url), slog.Any("error", err))
		return nil
	}

	var releases []model.Release

	for chartName, patterns := range values {
		repo := repos[chartName]

		for repoPattern, repoVersionName := range repo.Versions {
			releases = appendVersion(ctx, releases, patterns[repoPattern], repo, repoPattern, repoVersionName)
		}
	}

	return releases
}

func appendVersion(ctx context.Context, releases []model.Release, upstreamVersion semver.Version, repo model.Repository, repoPattern, repoVersionName string) []model.Release {
	if upstreamVersion.Name == repoVersionName {
		return releases
	}

	compiledPattern, err := semver.ParsePattern(repoPattern)
	if err != nil {
		slog.LogAttrs(ctx, slog.LevelError, "parse pattern", slog.String("pattern", repoPattern), slog.Any("error", err))
		return releases
	}

	repositoryVersion, err := semver.Parse(repoVersionName, semver.ExtractName(repo.Name))
	if err != nil {
		slog.LogAttrs(ctx, slog.LevelError, "parse version", slog.String("version", repoVersionName), slog.String("repo", semver.ExtractName(repo.Name)), slog.Any("error", err))
		return releases
	}

	if !compiledPattern.Check(upstreamVersion) || !upstreamVersion.IsGreater(repositoryVersion) {
		return releases
	}

	slog.LogAttrs(ctx, slog.LevelInfo, "New version available", slog.String("pattern", repoPattern), slog.String("repo", repo.String()), slog.String("version", upstreamVersion.Name))

	return append(releases, model.NewRelease(repo, repoPattern, upstreamVersion))
}
