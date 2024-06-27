package main

import (
	"flag"
	"os"

	"github.com/ViBiOh/flags"
	"github.com/ViBiOh/httputils/v4/pkg/alcotest"
	"github.com/ViBiOh/httputils/v4/pkg/cors"
	"github.com/ViBiOh/httputils/v4/pkg/db"
	"github.com/ViBiOh/httputils/v4/pkg/health"
	"github.com/ViBiOh/httputils/v4/pkg/logger"
	"github.com/ViBiOh/httputils/v4/pkg/owasp"
	"github.com/ViBiOh/httputils/v4/pkg/pprof"
	"github.com/ViBiOh/httputils/v4/pkg/redis"
	"github.com/ViBiOh/httputils/v4/pkg/renderer"
	"github.com/ViBiOh/httputils/v4/pkg/server"
	"github.com/ViBiOh/httputils/v4/pkg/telemetry"
	"github.com/ViBiOh/ketchup/pkg/provider/docker"
	"github.com/ViBiOh/ketchup/pkg/provider/github"
)

type configuration struct {
	logger    *logger.Config
	alcotest  *alcotest.Config
	telemetry *telemetry.Config
	pprof     *pprof.Config
	health    *health.Config

	server   *server.Config
	owasp    *owasp.Config
	cors     *cors.Config
	renderer *renderer.Config

	db    *db.Config
	redis *redis.Config

	github *github.Config
	docker *docker.Config
}

func newConfig() configuration {
	fs := flag.NewFlagSet("ketchup", flag.ExitOnError)
	fs.Usage = flags.Usage(fs)

	config := configuration{
		logger:    logger.Flags(fs, "logger"),
		alcotest:  alcotest.Flags(fs, ""),
		telemetry: telemetry.Flags(fs, "telemetry"),
		pprof:     pprof.Flags(fs, "pprof"),
		health:    health.Flags(fs, ""),

		server:   server.Flags(fs, ""),
		owasp:    owasp.Flags(fs, "", flags.NewOverride("Csp", "default-src 'self'; base-uri 'self'; script-src 'self' 'httputils-nonce'; style-src 'self' 'httputils-nonce'")),
		cors:     cors.Flags(fs, "cors"),
		renderer: renderer.Flags(fs, "", flags.NewOverride("Title", "Ketchup"), flags.NewOverride("PublicURL", "https://ketchup.vibioh.fr")),

		db:    db.Flags(fs, "db"),
		redis: redis.Flags(fs, "redis"),

		github: github.Flags(fs, "github"),
		docker: docker.Flags(fs, "docker"),
	}

	_ = fs.Parse(os.Args[1:])

	return config
}
