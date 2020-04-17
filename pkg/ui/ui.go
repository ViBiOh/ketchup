package ui

import (
	"fmt"
	"html/template"
	"net/http"
	"os"
	"strings"

	"github.com/ViBiOh/httputils/v3/pkg/crud"
	"github.com/ViBiOh/httputils/v3/pkg/httperror"
	"github.com/ViBiOh/httputils/v3/pkg/logger"
	"github.com/ViBiOh/httputils/v3/pkg/templates"
	"github.com/ViBiOh/ketchup/pkg/model"
	"github.com/ViBiOh/ketchup/pkg/target"
)

const (
	svgPath     = "/svg"
	targetsPath = "/targets"
)

// App of package
type App interface {
	Handler() http.Handler
}

type app struct {
	tpl     *template.Template
	version string

	targetApp target.App
}

// New creates new App from Config
func New(targetApp target.App) (App, error) {
	filesTemplates, err := templates.GetTemplates("templates", ".html")
	if err != nil {
		return nil, fmt.Errorf("unable to get templates: %s", err)
	}

	return &app{
		tpl:       template.Must(template.New("ketchup").ParseFiles(filesTemplates...)),
		version:   os.Getenv("VERSION"),
		targetApp: targetApp,
	}, nil
}

// Handler for request. Should be use with net/http
func (a app) Handler() http.Handler {
	svgHandler := http.StripPrefix(svgPath, a.svg())
	targetsHandler := http.StripPrefix(targetsPath, a.targets())

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, svgPath) {
			svgHandler.ServeHTTP(w, r)
			return
		}

		if strings.HasPrefix(r.URL.Path, targetsPath) {
			targetsHandler.ServeHTTP(w, r)
			return
		}

		targets, _, err := a.targetApp.List(r.Context(), 1, 100, "", false, nil)
		if err != nil {
			a.handleError(w, http.StatusInternalServerError, err, nil)
			return
		}

		content := map[string]interface{}{
			"Version": a.version,
			"Targets": targets,
		}

		if err := templates.ResponseHTMLTemplate(a.tpl.Lookup("app"), w, content, http.StatusOK); err != nil {
			httperror.InternalServerError(w, err)
		}
	})
}

func (a app) handleError(w http.ResponseWriter, status int, err error, errors []crud.Error) {
	logger.Error("%s", err)

	content := map[string]interface{}{
		"Version": a.version,
		"Message": model.Message{
			Level:   "error",
			Content: err.Error(),
		},
		"Errors": errors,
	}

	if err := templates.ResponseHTMLTemplate(a.tpl.Lookup("error"), w, content, status); err != nil {
		httperror.InternalServerError(w, err)
		return
	}
}
