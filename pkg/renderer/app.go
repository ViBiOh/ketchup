package renderer

import (
	"net/http"

	"github.com/ViBiOh/httputils/v3/pkg/httperror"
	"github.com/ViBiOh/httputils/v3/pkg/templates"
	"github.com/ViBiOh/ketchup/pkg/model"
)

func (a app) getData(r *http.Request) (interface{}, error) {
	ketchups, _, err := a.ketchupService.List(r.Context(), 1, 100)

	return ketchups, err
}

func (a app) appHandler(w http.ResponseWriter, r *http.Request, status int, message model.Message) {
	ketchups, err := a.getData(r)
	if err != nil {
		a.errorHandler(w, http.StatusInternalServerError, err)
		return
	}

	content := map[string]interface{}{
		"Version":  a.version,
		"Ketchups": ketchups,
	}

	if len(message.Content) > 0 {
		content["Message"] = message
	}

	if err := templates.ResponseHTMLTemplate(a.tpl.Lookup("app"), w, content, status); err != nil {
		httperror.InternalServerError(w, err)
	}
}
