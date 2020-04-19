package renderer

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/ViBiOh/httputils/v3/pkg/httperror"
	"github.com/ViBiOh/httputils/v3/pkg/logger"
	"github.com/ViBiOh/httputils/v3/pkg/templates"
)

func (a app) getData(r *http.Request) (interface{}, error) {
	targets, _, err := a.targetApp.List(r.Context(), 1, 100, "", false, nil)

	return targets, err
}

func (a app) uiHandler(w http.ResponseWriter, r *http.Request, status int, message Message) {
	targets, err := a.getData(r)
	if err != nil {
		a.errorHandler(w, http.StatusInternalServerError, err, nil)
		return
	}

	content := map[string]interface{}{
		"Version": a.version,
		"Targets": targets,
	}

	if len(message.Content) > 0 {
		content["Message"] = message
	}

	if err := templates.ResponseHTMLTemplate(a.tpl.Lookup("app"), w, content, status); err != nil {
		httperror.InternalServerError(w, err)
	}
}

func (a app) errorHandler(w http.ResponseWriter, status int, errs ...error) {
	logger.Error("%s", errs)

	content := map[string]interface{}{
		"Version": a.version,
	}

	if len(errs) > 0 {
		content["Message"] = Message{
			Level:   "error",
			Content: errs[0].Error(),
		}

		if len(errs) > 1 {
			content["Errors"] = errs[1:]
		}
	}

	if err := templates.ResponseHTMLTemplate(a.tpl.Lookup("error"), w, content, status); err != nil {
		httperror.InternalServerError(w, err)
		return
	}
}

func (a app) svg() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tpl := a.tpl.Lookup(fmt.Sprintf("svg-%s", strings.Trim(r.URL.Path, "/")))
		if tpl == nil {
			httperror.NotFound(w)
			return
		}

		w.Header().Set("Content-Type", "image/svg+xml")
		if err := templates.WriteTemplate(tpl, w, r.URL.Query().Get("fill"), "text/xml"); err != nil {
			httperror.InternalServerError(w, err)
		}
	})
}
