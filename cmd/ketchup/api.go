package main

import (
	"flag"
	"net/http"
	"os"

	"github.com/ViBiOh/httputils/v3/pkg/alcotest"
	"github.com/ViBiOh/httputils/v3/pkg/cors"
	"github.com/ViBiOh/httputils/v3/pkg/httputils"
	"github.com/ViBiOh/httputils/v3/pkg/logger"
	"github.com/ViBiOh/httputils/v3/pkg/owasp"
	"github.com/ViBiOh/httputils/v3/pkg/prometheus"
	"github.com/ViBiOh/httputils/v3/pkg/swagger"
	"github.com/ViBiOh/ketchup/pkg/github"
)

const (
	helloPath = "/hello"
	dumpPath  = "/dump"
)

func main() {
	fs := flag.NewFlagSet("ketchup", flag.ExitOnError)

	serverConfig := httputils.Flags(fs, "")
	alcotestConfig := alcotest.Flags(fs, "")
	prometheusConfig := prometheus.Flags(fs, "prometheus")
	owaspConfig := owasp.Flags(fs, "")
	corsConfig := cors.Flags(fs, "cors")
	swaggerConfig := swagger.Flags(fs, "swagger")

	githubConfig := github.Flags(fs, "github")

	logger.Fatal(fs.Parse(os.Args[1:]))

	alcotest.DoAndExit(alcotestConfig)

	server := httputils.New(serverConfig)
	server.Middleware(prometheus.New(prometheusConfig).Middleware)
	server.Middleware(owasp.New(owaspConfig).Middleware)
	server.Middleware(cors.New(corsConfig).Middleware)

	githubApp := github.New(githubConfig)

	viwsRelease, err := githubApp.LastRelease("vibioh", "viws")
	if err != nil {
		logger.Fatal(err)
	}
	logger.Info("%+v", viwsRelease)

	swaggerApp, err := swagger.New(swaggerConfig, server.Swagger)
	logger.Fatal(err)

	swaggerHandler := swaggerApp.Handler()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		swaggerHandler.ServeHTTP(w, r)
	})

	server.ListenServeWait(handler)
}
