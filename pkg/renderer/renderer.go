package renderer

import (
	"flag"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"strings"

	"github.com/ViBiOh/httputils/v3/pkg/flags"
	"github.com/ViBiOh/httputils/v3/pkg/query"
	"github.com/ViBiOh/httputils/v3/pkg/templates"
	"github.com/ViBiOh/ketchup/pkg/model"
	"github.com/ViBiOh/ketchup/pkg/service/ketchup"
	"github.com/ViBiOh/ketchup/pkg/service/user"
)

const (
	svgPath      = "/svg"
	signupPath   = "/signup"
	ketchupsPath = "/ketchups"
)

// App of package
type App interface {
	Handler() http.Handler
	PublicHandler() http.Handler
}

// Config of package
type Config struct {
	publicPath *string
}

type app struct {
	tpl        *template.Template
	publicPath string
	version    string

	ketchupService ketchup.App
	userService    user.App
}

// Flags adds flags for configuring package
func Flags(fs *flag.FlagSet, prefix string) Config {
	return Config{
		publicPath: flags.New(prefix, "ui").Name("PublicPath").Default("https://ketchup.vibioh.fr").Label("Public path of UI").ToString(fs),
	}
}

// New creates new App from Config
func New(config Config, ketchupService ketchup.App, userService user.App) (App, error) {
	filesTemplates, err := templates.GetTemplates("templates", ".html")
	if err != nil {
		return nil, fmt.Errorf("unable to get templates: %s", err)
	}

	return app{
		tpl:        template.Must(template.New("ketchup").ParseFiles(filesTemplates...)),
		publicPath: strings.TrimSpace(*config.publicPath),
		version:    os.Getenv("VERSION"),

		ketchupService: ketchupService,
		userService:    userService,
	}, nil
}

// Handler for request. Should be use with net/http
func (a app) Handler() http.Handler {
	ketchupsHandler := http.StripPrefix(ketchupsPath, a.ketchups())

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if query.IsRoot(r) {
			a.uiHandler(w, r, http.StatusOK, model.Message{
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
