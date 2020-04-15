package ui

import (
	"fmt"
	"html/template"
	"net/http"
	"os"

	"github.com/ViBiOh/httputils/v3/pkg/httperror"
	"github.com/ViBiOh/httputils/v3/pkg/templates"
	"github.com/ViBiOh/ketchup/pkg/model"
	"github.com/ViBiOh/ketchup/pkg/target"
)

// App of package
type App interface {
	Handler() http.Handler
}

type app struct {
	tpl *template.Template

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
		targetApp: targetApp,
	}, nil
}

// Handler for request. Should be use with net/http
func (a app) Handler() http.Handler {
	version := os.Getenv("VERSION")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		content := map[string]interface{}{
			"Version": version,
		}
		status := http.StatusOK

		targets, _, err := a.targetApp.List(r.Context(), 1, 100, "", false, nil)
		if err != nil {
			content["Message"] = model.Message{
				Level:   "error",
				Content: err.Error(),
			}
			status = http.StatusInternalServerError
		} else {
			content["Targets"] = targets
		}

		if err := templates.ResponseHTMLTemplate(a.tpl.Lookup("ketchup"), w, content, status); err != nil {
			httperror.InternalServerError(w, err)
		}
	})
}
