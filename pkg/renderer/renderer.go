package renderer

import (
	"fmt"
	"html/template"
	"net/http"
	"os"
	"strings"

	"github.com/ViBiOh/httputils/v3/pkg/query"
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

	return app{
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

		if query.IsRoot(r) {
			a.uiHandler(w, r, http.StatusOK, model.Message{
				Level:   r.URL.Query().Get("messageLevel"),
				Content: r.URL.Query().Get("messageContent"),
			})
			return
		}

		targetsHandler.ServeHTTP(w, r)
	})
}
