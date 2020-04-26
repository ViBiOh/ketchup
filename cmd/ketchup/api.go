package main

import (
	"flag"
	"net/http"
	"os"
	"path"
	"strings"

	authHandler "github.com/ViBiOh/auth/v2/pkg/handler"
	basicIdent "github.com/ViBiOh/auth/v2/pkg/ident/basic"
	basicProvider "github.com/ViBiOh/auth/v2/pkg/provider/db"
	authService "github.com/ViBiOh/auth/v2/pkg/service"
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
	"github.com/ViBiOh/ketchup/pkg/target"
	mailer "github.com/ViBiOh/mailer/pkg/client"
)

const (
	faviconPath = "/favicon"
	apiPath     = "/api"
	targetPath  = apiPath + "/targets"
	usersPath   = apiPath + "/users"
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
	crudTargetConfig := crud.GetConfiguredFlags(targetPath, "Target of Ketchup")(fs, "targets")
	crudUserConfig := crud.GetConfiguredFlags(usersPath, "Users of Ketchup")(fs, "users")

	logger.Fatal(fs.Parse(os.Args[1:]))

	alcotest.DoAndExit(alcotestConfig)

	server := httputils.New(serverConfig)
	server.Middleware(prometheus.New(prometheusConfig).Middleware)
	server.Middleware(owasp.New(owaspConfig).Middleware)
	server.Middleware(cors.New(corsConfig).Middleware)

	ketchupDb, err := db.New(dbConfig)
	logger.Fatal(err)
	server.Health(ketchupDb.Ping)

	/* Basic auth management */
	basicApp := basicProvider.New(ketchupDb)
	basicProvider := basicIdent.New(basicApp)
	handlerApp := authHandler.New(basicApp, basicProvider)
	server.Middleware(handlerApp.Middleware)

	crudUserApp, err := crud.New(crudUserConfig, authService.New(ketchupDb, basicApp))
	logger.Fatal(err)
	/* Basic auth management */

	githubApp := github.New(githubConfig)
	mailerApp := mailer.New(mailerConfig)
	targetApp := target.New(ketchupDb, githubApp)
	ketchupAp := ketchup.New(ketchupConfig, targetApp, githubApp, mailerApp)

	rendererApp, err := renderer.New(targetApp)
	logger.Fatal(err)

	crudTargetApp, err := crud.New(crudTargetConfig, targetApp)
	logger.Fatal(err)

	swaggerApp, err := swagger.New(swaggerConfig, server.Swagger, crudTargetApp.Swagger, crudUserApp.Swagger)
	logger.Fatal(err)

	swaggerHandler := http.StripPrefix(apiPath, swaggerApp.Handler())
	crudTargetHandler := http.StripPrefix(targetPath, crudTargetApp.Handler())
	crudUserHandler := http.StripPrefix(usersPath, crudUserApp.Handler())
	rendererHandler := rendererApp.Handler()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, targetPath) {
			crudTargetHandler.ServeHTTP(w, r)
			return
		}

		if strings.HasPrefix(r.URL.Path, usersPath) {
			crudUserHandler.ServeHTTP(w, r)
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
