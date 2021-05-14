package ketchup

import (
	"net/http"
	"strings"

	"github.com/ViBiOh/httputils/v4/pkg/logger"
	"github.com/ViBiOh/httputils/v4/pkg/query"
)

const (
	suggestThresold = uint64(10)
	suggestCount    = uint64(5)
)

func (a app) PublicTemplateFunc(r *http.Request) (string, int, map[string]interface{}, error) {
	securityPayload, err := a.generateToken(r.Context())
	if err != nil {
		return "", http.StatusInternalServerError, nil, err
	}

	suggests, err := a.repositoryService.Suggest(r.Context(), []uint64{0}, 3)
	if err != nil {
		logger.Warn("unable to get publics suggestions: %s", err)
	}

	return "public", http.StatusOK, map[string]interface{}{
		"Security": securityPayload,
		"Suggests": suggests,
		"Root":     "/",
	}, nil
}

func min(a, b uint64) uint64 {
	if a < b {
		return a
	}
	return b
}

func (a app) AppTemplateFunc(r *http.Request) (string, int, map[string]interface{}, error) {
	pagination, err := query.ParsePagination(r, 1, 100, 100)
	if err != nil {
		return "", http.StatusBadRequest, nil, err
	}

	ketchups, _, err := a.ketchupService.List(r.Context(), pagination.PageSize, strings.TrimSpace(r.URL.Query().Get("lastKey")))
	if err != nil {
		return "", http.StatusInternalServerError, nil, err
	}

	content := map[string]interface{}{
		"Root":     appPath,
		"Ketchups": ketchups,
	}

	ketchupsCount := uint64(len(ketchups))

	if ketchupsCount <= suggestThresold {
		ketchupIds := make([]uint64, ketchupsCount)
		for index, ketchup := range ketchups {
			ketchupIds[index] = ketchup.Repository.ID
		}

		suggests, err := a.repositoryService.Suggest(r.Context(), ketchupIds, min(suggestThresold-ketchupsCount, suggestCount))
		if err != nil {
			logger.Warn("unable to get suggest repositories: %s", err)
		} else {
			content["Suggests"] = suggests
		}
	}

	return "ketchup", http.StatusOK, content, nil
}
