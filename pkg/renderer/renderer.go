package renderer

import (
	"flag"
	"fmt"
	"html/template"
	"math/rand"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/ViBiOh/httputils/v3/pkg/cron"
	"github.com/ViBiOh/httputils/v3/pkg/flags"
	"github.com/ViBiOh/httputils/v3/pkg/logger"
	"github.com/ViBiOh/httputils/v3/pkg/query"
	"github.com/ViBiOh/httputils/v3/pkg/templates"
	"github.com/ViBiOh/ketchup/pkg/model"
	"github.com/ViBiOh/ketchup/pkg/service/ketchup"
	"github.com/ViBiOh/ketchup/pkg/service/user"
)

const (
	faviconPath  = "/favicon"
	svgPath      = "/svg"
	signupPath   = "/signup"
	ketchupsPath = "/ketchups"
	appPath      = "/app"
)

var (
	staticDir    = "static"
	templatesDir = "templates"
)

// App of package
type App interface {
	Start()
	Handler() http.Handler
	PublicHandler() http.Handler
}

// Config of package
type Config struct {
	uiPath *string
}

type app struct {
	tpl        *template.Template
	rand       *rand.Rand
	tokenStore TokenStore
	uiPath     string
	version    string

	ketchupService ketchup.App
	userService    user.App
}

// Flags adds flags for configuring package
func Flags(fs *flag.FlagSet, prefix string) Config {
	return Config{
		uiPath: flags.New(prefix, "ui").Name("PublicPath").Default("ketchup.vibioh.fr").Label("Public path").ToString(fs),
	}
}

// New creates new App from Config
func New(config Config, ketchupService ketchup.App, userService user.App) (App, error) {
	filesTemplates, err := templates.GetTemplates(templatesDir, ".html")
	if err != nil {
		return nil, fmt.Errorf("unable to get templates: %s", err)
	}

	return app{
		tpl:        template.Must(template.New("ketchup").ParseFiles(filesTemplates...)),
		uiPath:     strings.TrimSpace(*config.uiPath),
		version:    os.Getenv("VERSION"),
		rand:       rand.New(rand.NewSource(time.Now().UnixNano())),
		tokenStore: NewTokenStore(),

		ketchupService: ketchupService,
		userService:    userService,
	}, nil
}

func (a app) Start() {
	cron.New().Each(time.Hour).Start(a.tokenStore.Clean, func(err error) {
		logger.Error("error while running ketchup notify: %s", err)
	})
}

// Handler for request. Should be use with net/http
func (a app) Handler() http.Handler {
	ketchupsHandler := http.StripPrefix(ketchupsPath, a.ketchups())

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if query.IsRoot(r) {
			a.appHandler(w, r, model.Message{
				Level:   r.URL.Query().Get("messageLevel"),
				Content: r.URL.Query().Get("messageContent"),
			})
			return
		}

		ketchupsHandler.ServeHTTP(w, r)
	})
}

// PublicHandler for public requests. Should be use with net/http
func (a app) PublicHandler() http.Handler {
	svgHandler := http.StripPrefix(svgPath, a.svg())

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, faviconPath) {
			http.ServeFile(w, r, path.Join(staticDir, r.URL.Path))
			return
		}

		if strings.HasPrefix(r.URL.Path, svgPath) {
			svgHandler.ServeHTTP(w, r)
			return
		}

		if r.URL.Path == signupPath {
			a.signup(w, r)
			return
		}

		a.publicHandler(w, r, http.StatusOK, model.Message{
			Level:   r.URL.Query().Get("messageLevel"),
			Content: r.URL.Query().Get("messageContent"),
		})
	})
}
