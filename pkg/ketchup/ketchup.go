package ketchup

import (
	"html/template"
	"net/http"
	"strings"

	"github.com/ViBiOh/httputils/v4/pkg/redis"
	"github.com/ViBiOh/httputils/v4/pkg/renderer"
	"github.com/ViBiOh/ketchup/pkg/model"
	"github.com/ViBiOh/ketchup/pkg/service/ketchup"
	"github.com/ViBiOh/ketchup/pkg/service/repository"
	"github.com/ViBiOh/ketchup/pkg/service/user"
)

const (
	ketchupsPath = "/ketchups"
	appPath      = "/app"
)

// FuncMap for template rendering
var FuncMap = template.FuncMap{
	"frequencyImage": func(frequency model.KetchupFrequency) string {
		switch frequency {
		case model.None:
			return "bell-slash"
		case model.Weekly:
			return "calendar"
		default:
			return "clock"
		}
	},
}

// App of package
type App struct {
	userService       user.App
	ketchupService    ketchup.App
	repositoryService repository.App
	redisApp          redis.App
	rendererApp       renderer.App
}

// New creates new App from Config
func New(rendererApp renderer.App, ketchupService ketchup.App, userService user.App, repositoryService repository.App, redisApp redis.App) App {
	return App{
		rendererApp:       rendererApp,
		ketchupService:    ketchupService,
		userService:       userService,
		repositoryService: repositoryService,
		redisApp:          redisApp,
	}
}

// Handler for request. Should be use with net/http
func (a App) Handler() http.Handler {
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
