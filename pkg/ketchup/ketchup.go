package ketchup

import (
	"html/template"
	"net/http"
	"strings"

	"github.com/ViBiOh/httputils/v4/pkg/renderer"
	"github.com/ViBiOh/ketchup/pkg/redis"
	"github.com/ViBiOh/ketchup/pkg/service/ketchup"
	"github.com/ViBiOh/ketchup/pkg/service/repository"
	"github.com/ViBiOh/ketchup/pkg/service/user"
)

const (
	ketchupsPath = "/ketchups"
	appPath      = "/app"
)

var (
	// FuncMap for template rendering
	FuncMap = template.FuncMap{}
)

// App of package
type App interface {
	Handler() http.Handler
	Signup() http.Handler
	PublicTemplateFunc(*http.Request) (string, int, map[string]interface{}, error)
	AppTemplateFunc(*http.Request) (string, int, map[string]interface{}, error)
}

type app struct {
	rendererApp       renderer.App
	ketchupService    ketchup.App
	userService       user.App
	repositoryService repository.App
	redisApp          redis.App
}

// New creates new App from Config
func New(rendererApp renderer.App, ketchupService ketchup.App, userService user.App, repositoryService repository.App, redisApp redis.App) App {
	return app{
		rendererApp:       rendererApp,
		ketchupService:    ketchupService,
		userService:       userService,
		repositoryService: repositoryService,
		redisApp:          redisApp,
	}
}

// Handler for request. Should be use with net/http
func (a app) Handler() http.Handler {
	rendererHandler := a.rendererApp.Handler(a.AppTemplateFunc)
	ketchupHandler := http.StripPrefix(ketchupsPath, a.ketchups())

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, ketchupsPath) {
			ketchupHandler.ServeHTTP(w, r)
			return
		}

		rendererHandler.ServeHTTP(w, r)
	})
}
