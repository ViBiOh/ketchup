package ketchup

import (
	"context"
	"net/http"
	"time"

	"github.com/ViBiOh/httputils/v4/pkg/cache"
	"github.com/ViBiOh/httputils/v4/pkg/logger"
	"github.com/ViBiOh/httputils/v4/pkg/query"
	"github.com/ViBiOh/httputils/v4/pkg/renderer"
	"github.com/ViBiOh/ketchup/pkg/model"
)

const (
	suggestThresold = uint64(5)
)

// PublicTemplateFunc rendering public GUI
func (a App) PublicTemplateFunc(_ http.ResponseWriter, r *http.Request) (renderer.Page, error) {
	securityPayload, err := a.generateToken(r.Context())
	if err != nil {
		return renderer.NewPage("", http.StatusInternalServerError, nil), err
	}

	return renderer.NewPage("public", http.StatusOK, map[string]any{
		"Security": securityPayload,
		"Suggests": a.suggests(r.Context(), nil, 3),
		"Root":     "/",
	}), nil
}

func min(a, b uint64) uint64 {
	if a < b {
		return a
	}
	return b
}

// AppTemplateFunc rendering private GUI
func (a App) AppTemplateFunc(_ http.ResponseWriter, r *http.Request) (renderer.Page, error) {
	pagination, err := query.ParsePagination(r, 100, 100)
	if err != nil {
		return renderer.NewPage("", http.StatusBadRequest, nil), err
	}

	ketchups, _, err := a.ketchupService.List(r.Context(), pagination.PageSize, pagination.Last)
	if err != nil {
		return renderer.NewPage("", http.StatusInternalServerError, nil), err
	}

	content := map[string]any{
		"Root":     appPath,
		"Ketchups": ketchups,
	}

	ketchupsCount := uint64(len(ketchups))

	if ketchupsCount <= suggestThresold {
		ketchupIds := make([]uint64, ketchupsCount)
		for index, ketchup := range ketchups {
			ketchupIds[index] = ketchup.Repository.ID
		}

		content["Suggests"] = a.suggests(r.Context(), ketchupIds, min(suggestThresold-ketchupsCount, suggestThresold))
	}

	return renderer.NewPage("ketchup", http.StatusOK, content), nil
}

func (a App) suggests(ctx context.Context, ignoreIds []uint64, count uint64) []model.Repository {
	user := model.ReadUser(ctx)
	if user.IsZero() {
		ignoreIds = []uint64{0}
	}

	var suggests []model.Repository
	items, err := cache.Retrieve(ctx, a.redisApp, suggestCacheKey(user), &suggests, func() (any, error) {
		return a.repositoryService.Suggest(ctx, ignoreIds, count)
	}, time.Hour*24)
	if err != nil {
		logger.Warn("unable to get suggests: %s", err)
		return nil
	}

	if repos, ok := items.([]model.Repository); ok {
		return repos
	}

	if repos, ok := items.(*[]model.Repository); ok {
		return *repos
	}

	return nil
}
