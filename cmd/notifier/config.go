package main

import (
	"flag"
	"os"

	"github.com/ViBiOh/flags"
	"github.com/ViBiOh/httputils/v4/pkg/db"
	"github.com/ViBiOh/httputils/v4/pkg/logger"
	"github.com/ViBiOh/httputils/v4/pkg/telemetry"
	"github.com/ViBiOh/ketchup/pkg/notifier"
	"github.com/ViBiOh/ketchup/pkg/provider/docker"
	"github.com/ViBiOh/ketchup/pkg/provider/github"
	mailer "github.com/ViBiOh/mailer/pkg/client"
)

type configuration struct {
	logger    *logger.Config
	telemetry *telemetry.Config

	db *db.Config

	github   *github.Config
	docker   *docker.Config
	mailer   *mailer.Config
	notifier *notifier.Config
}

func newConfig() configuration {
	fs := flag.NewFlagSet("ketchup", flag.ExitOnError)
	fs.Usage = flags.Usage(fs)

	config := configuration{
		logger:    logger.Flags(fs, "logger"),
		telemetry: telemetry.Flags(fs, "telemetry"),

		db: db.Flags(fs, "db"),

		github:   github.Flags(fs, "github"),
		docker:   docker.Flags(fs, "docker"),
		mailer:   mailer.Flags(fs, "mailer"),
		notifier: notifier.Flags(fs, "notifier"),
	}

	_ = fs.Parse(os.Args[1:])

	return config
}
