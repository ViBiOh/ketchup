package main

import (
	"flag"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/ViBiOh/httputils/v3/pkg/alcotest"
	"github.com/ViBiOh/httputils/v3/pkg/cors"
	"github.com/ViBiOh/httputils/v3/pkg/crud"
	"github.com/ViBiOh/httputils/v3/pkg/db"
	"github.com/ViBiOh/httputils/v3/pkg/httperror"
	"github.com/ViBiOh/httputils/v3/pkg/httputils"
	"github.com/ViBiOh/httputils/v3/pkg/logger"
	"github.com/ViBiOh/httputils/v3/pkg/owasp"
	"github.com/ViBiOh/httputils/v3/pkg/prometheus"
	"github.com/ViBiOh/httputils/v3/pkg/swagger"
	"github.com/ViBiOh/ketchup/pkg/github"
	"github.com/ViBiOh/ketchup/pkg/ketchup"
	service "github.com/ViBiOh/ketchup/pkg/service/user"
	"github.com/ViBiOh/ketchup/pkg/store"
	mailer "github.com/ViBiOh/mailer/pkg/client"
)

const (
	faviconPath = "/favicon"
	apiPath     = "/api"
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
	crudUserConfig := crud.GetConfiguredFlags(usersPath, "Users")(fs, "users")

	logger.Fatal(fs.Parse(os.Args[1:]))

	alcotest.DoAndExit(alcotestConfig)

	server := httputils.New(serverConfig)
	server.Middleware(prometheus.New(prometheusConfig).Middleware)
	server.Middleware(owasp.New(owaspConfig).Middleware)
	server.Middleware(cors.New(corsConfig).Middleware)

	ketchupDb, err := db.New(dbConfig)
	logger.Fatal(err)
	server.Health(ketchupDb.Ping)

	storeApp := store.New(ketchupDb)
	serviceApp := service.New(storeApp, nil, nil)
	githubApp := github.New(githubConfig)
	mailerApp := mailer.New(mailerConfig)

	ketchupAp := ketchup.New(ketchupConfig, storeApp, githubApp, mailerApp)

	crudUserApp, err := crud.New(crudUserConfig, serviceApp)
	logger.Fatal(err)

	swaggerApp, err := swagger.New(swaggerConfig, server.Swagger, crudUserApp.Swagger)
	logger.Fatal(err)

	swaggerHandler := http.StripPrefix(apiPath, swaggerApp.Handler())
	crudTargetHandler := http.StripPrefix(usersPath, crudUserApp.Handler())

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, usersPath) {
			crudTargetHandler.ServeHTTP(w, r)
			return
		}

		if strings.HasPrefix(r.URL.Path, apiPath) {
			swaggerHandler.ServeHTTP(w, r)
			return
		}

		if strings.HasPrefix(r.URL.Path, faviconPath) {
			http.ServeFile(w, r, path.Join("static", r.URL.Path))
		}

		httperror.NotFound(w)
	})

	go ketchupAp.Start()

	server.ListenServeWait(handler)
}
