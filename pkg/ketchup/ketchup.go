package ketchup

import (
	"html/template"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/ViBiOh/httputils/v3/pkg/cron"
	"github.com/ViBiOh/httputils/v3/pkg/logger"
	"github.com/ViBiOh/httputils/v3/pkg/renderer"
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
	Start(<-chan struct{})
	Handler() http.Handler
	Signup() http.Handler
	PublicTemplateFunc(*http.Request) (string, int, map[string]interface{}, error)
	AppTemplateFunc(*http.Request) (string, int, map[string]interface{}, error)
}

type app struct {
	rand *rand.Rand

	rendererApp       renderer.App
	ketchupService    ketchup.App
	userService       user.App
	repositoryService repository.App
	tokenStore        TokenStore
}

// New creates new App from Config
func New(rendererApp renderer.App, ketchupService ketchup.App, userService user.App, repositoryService repository.App) App {
	return app{
		rand:       rand.New(rand.NewSource(time.Now().UnixNano())),
		tokenStore: NewTokenStore(),

		rendererApp:       rendererApp,
		ketchupService:    ketchupService,
		userService:       userService,
		repositoryService: repositoryService,
	}
}

func (a app) Start(done <-chan struct{}) {
	cron.New().Each(time.Hour).OnError(func(err error) {
		logger.Error("error while running token store cleanup: %s", err)
	}).Start(a.tokenStore.Clean, done)
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
