package main

import (
	"database/sql"
	"flag"
	"net/http"
	"os"
	"path"
	"strings"

	auth "github.com/ViBiOh/auth/v2/pkg/auth"
	authIdent "github.com/ViBiOh/auth/v2/pkg/ident/basic"
	authMiddleware "github.com/ViBiOh/auth/v2/pkg/middleware"
	authService "github.com/ViBiOh/auth/v2/pkg/service"
	authStore "github.com/ViBiOh/auth/v2/pkg/store/db"
	"github.com/ViBiOh/httputils/v3/pkg/alcotest"
	"github.com/ViBiOh/httputils/v3/pkg/cors"
	"github.com/ViBiOh/httputils/v3/pkg/crud"
	"github.com/ViBiOh/httputils/v3/pkg/db"
	"github.com/ViBiOh/httputils/v3/pkg/httputils"
	"github.com/ViBiOh/httputils/v3/pkg/logger"
	"github.com/ViBiOh/httputils/v3/pkg/owasp"
	"github.com/ViBiOh/httputils/v3/pkg/prometheus"
	"github.com/ViBiOh/httputils/v3/pkg/swagger"
	"github.com/ViBiOh/ketchup/pkg/github"
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
	faviconPath  = "/favicon"
	apiPath      = "/api"
	usersPath    = apiPath + "/users"
	ketchupsPath = apiPath + "/ketchups"
)

func initAuth(db *sql.DB) (authService.App, auth.Provider, authMiddleware.App) {
	authProvider := authStore.New(db)
	identProvider := authIdent.New(authProvider)

	return authService.New(authProvider, authProvider), authProvider, authMiddleware.New(authProvider, identProvider)
}

func main() {
	fs := flag.NewFlagSet("ketchup", flag.ExitOnError)

	serverConfig := httputils.Flags(fs, "")
	alcotestConfig := alcotest.Flags(fs, "")
	prometheusConfig := prometheus.Flags(fs, "prometheus")
	owaspConfig := owasp.Flags(fs, "")
	corsConfig := cors.Flags(fs, "cors")
	swaggerConfig := swagger.Flags(fs, "swagger")

	dbConfig := db.Flags(fs, "db")
	mailerConfig := mailer.Flags(fs, "mailer")
	githubConfig := github.Flags(fs, "github")
	schedulerConfig := scheduler.Flags(fs, "scheduler")

	crudUserConfig := crud.GetConfiguredFlags(usersPath, "Users")(fs, "users")
	crudKetchupConfig := crud.GetConfiguredFlags(ketchupsPath, "Ketchup")(fs, "ketchups")

	logger.Fatal(fs.Parse(os.Args[1:]))

	alcotest.DoAndExit(alcotestConfig)

	server := httputils.New(serverConfig)
	server.Middleware(prometheus.New(prometheusConfig).Middleware)
	server.Middleware(owasp.New(owaspConfig).Middleware)
	server.Middleware(cors.New(corsConfig).Middleware)

	ketchupDb, err := db.New(dbConfig)
	logger.Fatal(err)
	server.Health(ketchupDb.Ping)

	authService, identProvider, authMiddleware := initAuth(ketchupDb)

	githubApp := github.New(githubConfig)
	mailerApp := mailer.New(mailerConfig)

	userServiceApp := userService.New(userStore.New(ketchupDb), authService, identProvider)
	repositoryApp := repositoryService.New(repositoryStore.New(ketchupDb), githubApp)
	ketchupApp := ketchupService.New(ketchupStore.New(ketchupDb), repositoryApp, userServiceApp)

	schedulerApp := scheduler.New(schedulerConfig, repositoryApp, ketchupApp, githubApp, mailerApp)

	rendererApp, err := renderer.New(ketchupApp)
	logger.Fatal(err)

	/* Crud and Swagger related things */
	crudUserApp, err := crud.New(crudUserConfig, userServiceApp)
	logger.Fatal(err)

	crudKetchupApp, err := crud.New(crudKetchupConfig, ketchupApp)
	logger.Fatal(err)

	swaggerApp, err := swagger.New(swaggerConfig, server.Swagger, crudUserApp.Swagger, crudKetchupApp.Swagger)
	logger.Fatal(err)

	swaggerHandler := http.StripPrefix(apiPath, swaggerApp.Handler())
	userHandler := http.StripPrefix(usersPath, crudUserApp.Handler())

	protectedUserHandler := httputils.ChainMiddlewares(userHandler, authMiddleware.Middleware)
	protectedKetchupHandler := httputils.ChainMiddlewares(http.StripPrefix(ketchupsPath, crudKetchupApp.Handler()), authMiddleware.Middleware)
	/* Crud and Swagger related things */

	rendererHandler := httputils.ChainMiddlewares(rendererApp.Handler(), authMiddleware.Middleware)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.TrimSuffix(r.URL.Path, "/") == usersPath && r.Method == http.MethodPost {
			userHandler.ServeHTTP(w, r)
			return
		}

		if strings.HasPrefix(r.URL.Path, usersPath) {
			protectedUserHandler.ServeHTTP(w, r)
			return
		}

		if strings.HasPrefix(r.URL.Path, ketchupsPath) {
			protectedKetchupHandler.ServeHTTP(w, r)
			return
		}

		if strings.HasPrefix(r.URL.Path, apiPath) {
			swaggerHandler.ServeHTTP(w, r)
			return
		}

		if strings.HasPrefix(r.URL.Path, faviconPath) {
			http.ServeFile(w, r, path.Join("static", r.URL.Path))
		}

		rendererHandler.ServeHTTP(w, r)
	})

	go schedulerApp.Start()

	server.ListenServeWait(handler)
}
