package renderer

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/ViBiOh/httputils/v3/pkg/httperror"
	"github.com/ViBiOh/httputils/v3/pkg/logger"
	"github.com/ViBiOh/httputils/v3/pkg/templates"
	"github.com/ViBiOh/ketchup/pkg/model"
	"github.com/ViBiOh/ketchup/pkg/service"
)

func redirectWithMessage(w http.ResponseWriter, r *http.Request, path, message string) {
	http.Redirect(w, r, fmt.Sprintf("%s?messageContent=%s", path, url.QueryEscape(message)), http.StatusFound)
}

func (a app) errorHandler(w http.ResponseWriter, status int, err error) {
	logger.Error("%s", err)

	content := map[string]interface{}{
		"Version": a.version,
	}

	if err != nil {
		message := err.Error()
		subMessages := ""

		if errors.Is(err, service.ErrInvalid) {
			subMessages = strings.TrimSuffix(message, ": "+service.ErrInvalid.Error())
			status = http.StatusBadRequest
			message = "Invalid form"
		} else if errors.Is(err, service.ErrInternalError) {
			status = http.StatusInternalServerError
			message = "Oops! Something went wrong."
		} else if errors.Is(err, service.ErrNotFound) {
			status = http.StatusNotFound
			message = strings.TrimSuffix(message, ": "+service.ErrNotFound.Error())
		} else if strings.HasPrefix(message, "invalid method") {
			status = http.StatusMethodNotAllowed
		}

		content["Message"] = model.NewErrorMessage(message)

		if len(subMessages) > 0 {
			content["Errors"] = strings.Split(subMessages, ", ")
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
