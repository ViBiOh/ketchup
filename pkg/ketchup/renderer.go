package ketchup

import (
	"context"
	"net/http"

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
		ketchupIds := make([]model.Identifier, ketchupsCount)
		for index, ketchup := range ketchups {
			ketchupIds[index] = ketchup.Repository.ID
		}

		content["Suggests"] = a.suggests(r.Context(), ketchupIds, min(suggestThresold-ketchupsCount, suggestThresold))
	}

	return renderer.NewPage("ketchup", http.StatusOK, content), nil
}

func (a App) suggests(ctx context.Context, ignoreIds []model.Identifier, count uint64) []model.Repository {
	user := model.ReadUser(ctx)
	if user.IsZero() {
		ignoreIds = []model.Identifier{0}
	}

	items, err := a.cacheApp.Get(countToCtx(ignoresIdsToCtx(ctx, ignoreIds), count), user)
	if err != nil {
		logger.Warn("get suggests: %s", err)
		return nil
	}

	return items
}
