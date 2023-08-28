package ketchup

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/ViBiOh/httputils/v4/pkg/query"
	"github.com/ViBiOh/httputils/v4/pkg/renderer"
	"github.com/ViBiOh/ketchup/pkg/model"
)

const suggestThresold = uint64(5)

func (s Service) PublicTemplateFunc(_ http.ResponseWriter, r *http.Request) (renderer.Page, error) {
	securityPayload, err := s.generateToken(r.Context())
	if err != nil {
		return renderer.NewPage("", http.StatusInternalServerError, nil), err
	}

	return renderer.NewPage("public", http.StatusOK, map[string]any{
		"Security": securityPayload,
		"Suggests": s.suggests(r.Context(), nil, 3),
		"Root":     "/",
	}), nil
}

func min(a, b uint64) uint64 {
	if a < b {
		return a
	}
	return b
}

func (s Service) TemplateFunc(_ http.ResponseWriter, r *http.Request) (renderer.Page, error) {
	pagination, err := query.ParsePagination(r, 100, 100)
	if err != nil {
		return renderer.NewPage("", http.StatusBadRequest, nil), err
	}

	ketchups, _, err := s.ketchup.List(r.Context(), pagination.PageSize, pagination.Last)
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

		content["Suggests"] = s.suggests(r.Context(), ketchupIds, min(suggestThresold-ketchupsCount, suggestThresold))
	}

	return renderer.NewPage("ketchup", http.StatusOK, content), nil
}

func (s Service) suggests(ctx context.Context, ignoreIds []model.Identifier, count uint64) []model.Repository {
	user := model.ReadUser(ctx)
	if user.IsZero() {
		ignoreIds = []model.Identifier{0}
	}

	items, err := s.cache.Get(countToCtx(ignoresIdsToCtx(ctx, ignoreIds), count), user)
	if err != nil {
		slog.Warn("get suggests", "err", err)
		return nil
	}

	return items
}
