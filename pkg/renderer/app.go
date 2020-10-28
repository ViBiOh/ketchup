package renderer

import (
	"net/http"

	"github.com/ViBiOh/httputils/v3/pkg/httperror"
	"github.com/ViBiOh/httputils/v3/pkg/logger"
	"github.com/ViBiOh/httputils/v3/pkg/templates"
	"github.com/ViBiOh/ketchup/pkg/model"
)

const (
	suggestThresold = uint64(10)
	suggestCount    = uint64(5)
)

func min(a, b uint64) uint64 {
	if a < b {
		return a
	}
	return b
}

func (a app) getData(r *http.Request) (interface{}, error) {
	ketchups, _, err := a.ketchupService.List(r.Context(), 1, 100)
	if err != nil {
		return nil, err
	}

	datas := map[string]interface{}{
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
			datas["Suggests"] = suggests
		}
	}

	return datas, nil
}

func (a app) appHandler(w http.ResponseWriter, r *http.Request, message model.Message) {
	ketchups, err := a.getData(r)
	if err != nil {
		a.errorHandler(w, http.StatusInternalServerError, err)
		return
	}

	content := map[string]interface{}{
		"Version": a.version,
		"Data":    ketchups,
		"Root":    "app/",
	}

	if len(message.Content) > 0 {
		content["Message"] = message
	}

	if err := templates.ResponseHTMLTemplate(a.tpl.Lookup("app"), w, content, http.StatusOK); err != nil {
		httperror.InternalServerError(w, err)
	}
}
