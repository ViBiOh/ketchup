package main

import (
	"database/sql"
	"flag"
	"net/http"
	"os"
	"strings"

	authIdent "github.com/ViBiOh/auth/v2/pkg/ident/basic"
	authMiddleware "github.com/ViBiOh/auth/v2/pkg/middleware"
	authService "github.com/ViBiOh/auth/v2/pkg/service"
	authStore "github.com/ViBiOh/auth/v2/pkg/store/db"
	"github.com/ViBiOh/httputils/v3/pkg/alcotest"
	"github.com/ViBiOh/httputils/v3/pkg/cors"
	"github.com/ViBiOh/httputils/v3/pkg/db"
	"github.com/ViBiOh/httputils/v3/pkg/httputils"
	"github.com/ViBiOh/httputils/v3/pkg/logger"
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
	prometheusConfig := prometheus.Flags(fs, "prometheus")
	owaspConfig := owasp.Flags(fs, "")
	corsConfig := cors.Flags(fs, "cors")

	dbConfig := db.Flags(fs, "db")
	mailerConfig := mailer.Flags(fs, "mailer")
	githubConfig := github.Flags(fs, "github")
	rendererConfig := renderer.Flags(fs, "ui")
	schedulerConfig := scheduler.Flags(fs, "scheduler")

	logger.Fatal(fs.Parse(os.Args[1:]))

	alcotest.DoAndExit(alcotestConfig)

	server := httputils.New(serverConfig)
	server.Middleware(prometheus.New(prometheusConfig).Middleware)
	server.Middleware(owasp.New(owaspConfig).Middleware)
	server.Middleware(cors.New(corsConfig).Middleware)

	ketchupDb, err := db.New(dbConfig)
	logger.Fatal(err)
	server.Health(ketchupDb.Ping)

	authServiceApp, authMiddleware := initAuth(ketchupDb)

	githubApp := github.New(githubConfig)
	mailerApp := mailer.New(mailerConfig)

	userServiceApp := userService.New(userStore.New(ketchupDb), authServiceApp)
	repositoryServiceApp := repositoryService.New(repositoryStore.New(ketchupDb), githubApp)
	ketchupServiceApp := ketchupService.New(ketchupStore.New(ketchupDb), repositoryServiceApp)

	schedulerApp := scheduler.New(schedulerConfig, repositoryServiceApp, ketchupServiceApp, githubApp, mailerApp)

	rendererApp, err := renderer.New(rendererConfig, ketchupServiceApp, userServiceApp)
	logger.Fatal(err)

	publicHandler := rendererApp.PublicHandler()
	protectedhandler := httputils.ChainMiddlewares(http.StripPrefix(appPath, rendererApp.Handler()), authMiddleware.Middleware, middleware.New(userServiceApp).Middleware)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, appPath) {
			protectedhandler.ServeHTTP(w, r)
			return
		}

		publicHandler.ServeHTTP(w, r)
	})

	go schedulerApp.Start()
	go rendererApp.Start()

	server.ListenServeWait(handler)
}
