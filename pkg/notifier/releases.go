package notifier

import (
	"context"
	"fmt"

	"github.com/ViBiOh/httputils/v4/pkg/logger"
	"github.com/ViBiOh/ketchup/pkg/model"
)

func (a app) getNewReleases(ctx context.Context) ([]model.Release, uint64, error) {
	var newReleases []model.Release
	count := uint64(0)
	page := uint(1)

	for {
		repositories, totalCount, err := a.repositoryService.List(ctx, page, pageSize)
		if err != nil {
			return nil, count, fmt.Errorf("unable to fetch page %d of repositories: %s", page, err)
		}

		for _, repo := range repositories {
			count++
			newReleases = append(newReleases, a.getNewRepositoryReleases(repo)...)
		}

		if uint64(page*pageSize) < totalCount {
			page++
		} else {
			logger.Info("%d repositories checked, %d new releases", count, len(newReleases))
			return newReleases, count, nil
		}
	}
}
