package main

import (
	"database/sql"
	"flag"
	"net/http"
	"os"
	"strings"
	"time"

	authIdent "github.com/ViBiOh/auth/v2/pkg/ident/basic"
	authMiddleware "github.com/ViBiOh/auth/v2/pkg/middleware"
	authService "github.com/ViBiOh/auth/v2/pkg/service"
	authStore "github.com/ViBiOh/auth/v2/pkg/store/db"
	"github.com/ViBiOh/httputils/v3/pkg/alcotest"
	"github.com/ViBiOh/httputils/v3/pkg/cors"
	"github.com/ViBiOh/httputils/v3/pkg/db"
	"github.com/ViBiOh/httputils/v3/pkg/flags"
	"github.com/ViBiOh/httputils/v3/pkg/httputils"
	"github.com/ViBiOh/httputils/v3/pkg/logger"
	"github.com/ViBiOh/httputils/v3/pkg/model"
	"github.com/ViBiOh/httputils/v3/pkg/owasp"
	"github.com/ViBiOh/httputils/v3/pkg/prometheus"
	"github.com/ViBiOh/ketchup/pkg/github"
	"github.com/ViBiOh/ketchup/pkg/middleware"
	"github.com/ViBiOh/ketchup/pkg/renderer"
	"github.com/ViBiOh/ketchup/pkg/scheduler"
	ketchupService "github.com/ViBiOh/ketchup/pkg/service/ketchup"
	repositoryService "github.com/ViBiOh/ketchup/pkg/service/repository"
	userService "github.com/ViBiOh/ketchup/pkg/service/user"
	ketchupStore "github.com/ViBiOh/ketchup/pkg/store/ketchup"
	repositoryStore "github.com/ViBiOh/ketchup/pkg/store/repository"
	userStore "github.com/ViBiOh/ketchup/pkg/store/user"
	mailer "github.com/ViBiOh/mailer/pkg/client"
)

const (
	appPath = "/app"
)

func initAuth(db *sql.DB) (authService.App, authMiddleware.App) {
	authProvider := authStore.New(db)
	identProvider := authIdent.New(authProvider)

	return authService.New(authProvider, authProvider), authMiddleware.New(authProvider, identProvider)
}

func main() {
	fs := flag.NewFlagSet("ketchup", flag.ExitOnError)

	serverConfig := httputils.Flags(fs, "")
	alcotestConfig := alcotest.Flags(fs, "")
	loggerConfig := logger.Flags(fs, "logger")
	prometheusConfig := prometheus.Flags(fs, "prometheus")
	owaspConfig := owasp.Flags(fs, "", flags.NewOverride("Csp", "default-src 'self'; base-uri 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'"))
	corsConfig := cors.Flags(fs, "cors")

	dbConfig := db.Flags(fs, "db")
	mailerConfig := mailer.Flags(fs, "mailer")
	githubConfig := github.Flags(fs, "github")
	rendererConfig := renderer.Flags(fs, "ui")
	schedulerConfig := scheduler.Flags(fs, "scheduler")

	logger.Fatal(fs.Parse(os.Args[1:]))

	alcotest.DoAndExit(alcotestConfig)
	logger.Global(logger.New(loggerConfig))
	defer logger.Close()

	logger.Info("time.Now: %s", time.Now())
	if loc, err := time.LoadLocation("Europe/Paris"); err != nil {
		logger.Error("%s", err)
	} else {
		logger.Info("Time in %s: %s", "Europe/Paris", time.Now().In(loc))
	}

	ketchupDb, err := db.New(dbConfig)
	logger.Fatal(err)

	authServiceApp, authMiddlewareApp := initAuth(ketchupDb)

	githubApp := github.New(githubConfig)
	mailerApp := mailer.New(mailerConfig)

	userServiceApp := userService.New(userStore.New(ketchupDb), authServiceApp)
	repositoryServiceApp := repositoryService.New(repositoryStore.New(ketchupDb), githubApp)
	ketchupServiceApp := ketchupService.New(ketchupStore.New(ketchupDb), repositoryServiceApp)

	schedulerApp := scheduler.New(schedulerConfig, repositoryServiceApp, ketchupServiceApp, githubApp, mailerApp)

	rendererApp, err := renderer.New(rendererConfig, ketchupServiceApp, userServiceApp, repositoryServiceApp)
	logger.Fatal(err)

	publicHandler := rendererApp.PublicHandler()
	protectedhandler := authMiddlewareApp.Middleware(middleware.New(userServiceApp).Middleware(http.StripPrefix(appPath, rendererApp.Handler())))

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, appPath) {
			protectedhandler.ServeHTTP(w, r)
			return
		}

		publicHandler.ServeHTTP(w, r)
	})

	go schedulerApp.Start()
	go rendererApp.Start()

	httputils.New(serverConfig).ListenAndServe(handler, []model.Pinger{ketchupDb.Ping}, prometheus.New(prometheusConfig).Middleware, owasp.New(owaspConfig).Middleware, cors.New(corsConfig).Middleware)
}
