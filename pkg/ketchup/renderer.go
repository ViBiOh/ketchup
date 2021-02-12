package ketchup

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"net/http"
	"time"

	"github.com/ViBiOh/httputils/v3/pkg/logger"
)

const (
	suggestThresold = uint64(10)
	suggestCount    = uint64(5)
)

func (a app) generateToken() (string, int64, error) {
	questionID, err := rand.Int(rand.Reader, big.NewInt(int64(len(colors))))
	if err != nil {
		return "", 0, fmt.Errorf("unable to generate random int: %w", err)
	}

	return a.tokenStore.Store(questionID, time.Minute*5), questionID.Int64(), nil
}

func (a app) PublicTemplateFunc(r *http.Request) (string, int, map[string]interface{}, error) {
	token, questionID, err := a.generateToken()
	if err != nil {
		return "", http.StatusInternalServerError, nil, err
	}

	suggests, err := a.repositoryService.Suggest(r.Context(), []uint64{0}, 3)
	if err != nil {
		logger.Warn("unable to get publics suggestions: %s", err)
	}

	return "public", http.StatusOK, map[string]interface{}{
		"Token":    token,
		"Question": colors[questionID].Question,
		"Suggests": suggests,
	}, nil
}

func min(a, b uint64) uint64 {
	if a < b {
		return a
	}
	return b
}

func (a app) AppTemplateFunc(r *http.Request) (string, int, map[string]interface{}, error) {
	ketchups, _, err := a.ketchupService.List(r.Context(), 1, 100)
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
