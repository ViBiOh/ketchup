package main

import (
	"flag"
	"net/http"
	"os"
	"path"
	"strings"

	authIdent "github.com/ViBiOh/auth/v2/pkg/ident/basic"
	"github.com/ViBiOh/auth/v2/pkg/middleware"
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
	"github.com/ViBiOh/ketchup/pkg/ketchup"
	"github.com/ViBiOh/ketchup/pkg/renderer"
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
	ketchupConfig := ketchup.Flags(fs, "ketchup")
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

	/* Auth related things */
	authProvider := authStore.New(ketchupDb)
	identProvider := authIdent.New(authProvider)
	authServiceApp := authService.New(authProvider, authProvider)
	authMiddleware := middleware.New(authProvider, identProvider)
	/* Auth related things */

	githubApp := github.New(githubConfig)
	mailerApp := mailer.New(mailerConfig)

	userStoreApp := userStore.New(ketchupDb)
	userServiceApp := userService.New(userStoreApp, authServiceApp, authProvider)

	repositoryStoreApp := repositoryStore.New(ketchupDb)
	repositoryServiceApp := repositoryService.New(repositoryStoreApp, githubApp)

	ketchupStoreApp := ketchupStore.New(ketchupDb)
	ketchupServiceApp := ketchupService.New(ketchupStoreApp, repositoryServiceApp)

	ketchupAp := ketchup.New(ketchupConfig, repositoryStoreApp, githubApp, mailerApp)

	rendererApp, err := renderer.New(ketchupServiceApp)
	logger.Fatal(err)

	/* Crud and Swagger related things */
	crudUserApp, err := crud.New(crudUserConfig, userServiceApp)
	logger.Fatal(err)

	crudKetchupApp, err := crud.New(crudKetchupConfig, ketchupServiceApp)
	logger.Fatal(err)

	swaggerApp, err := swagger.New(swaggerConfig, server.Swagger, crudUserApp.Swagger, crudKetchupApp.Swagger)
	logger.Fatal(err)

	swaggerHandler := http.StripPrefix(apiPath, swaggerApp.Handler())
	crudUserHandler := httputils.ChainMiddlewares(http.StripPrefix(usersPath, crudUserApp.Handler()), authMiddleware.Middleware)
	crudKetchupHandler := httputils.ChainMiddlewares(http.StripPrefix(ketchupsPath, crudKetchupApp.Handler()), authMiddleware.Middleware)
	/* Crud and Swagger related things */

	rendererHandler := httputils.ChainMiddlewares(rendererApp.Handler(), authMiddleware.Middleware)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, usersPath) {
			crudUserHandler.ServeHTTP(w, r)
			return
		}
		if strings.HasPrefix(r.URL.Path, ketchupsPath) {
			crudKetchupHandler.ServeHTTP(w, r)
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

	go ketchupAp.Start()

	server.ListenServeWait(handler)
}
