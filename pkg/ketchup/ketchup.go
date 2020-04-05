package ketchup

import (
	"database/sql"
	"flag"
	"strings"
	"time"

	"github.com/ViBiOh/httputils/v3/pkg/cron"
	"github.com/ViBiOh/httputils/v3/pkg/flags"
	"github.com/ViBiOh/httputils/v3/pkg/logger"
	"github.com/ViBiOh/ketchup/pkg/github"
)

// App of package
type App interface {
	Start()
}

// Config of package
type Config struct {
	emailTo  *string
	timezone *string
}

type app struct {
	emailTo  string
	timezone string

	db        *sql.DB
	githubApp github.App
}

// Flags adds flags for configuring package
func Flags(fs *flag.FlagSet, prefix string) Config {
	return Config{
		emailTo:  flags.New(prefix, "ketchup").Name("To").Default("").Label("Email to send notification").ToString(fs),
		timezone: flags.New(prefix, "ketchup").Name("Timezone").Default("Europe/Paris").Label("Timezone").ToString(fs),
	}
}

// New creates new App from Config
func New(config Config, db *sql.DB, githubApp github.App) App {
	return &app{
		emailTo:  strings.TrimSpace(*config.emailTo),
		timezone: strings.TrimSpace(*config.timezone),

		db:        db,
		githubApp: githubApp,
	}
}

func (a app) Start() {
	cron.New().At("08:00").In(a.timezone).Days().Start(a.checkUpdates, func(err error) {
		logger.Error("error while running cron: %s", err)
	})
}

func (a app) checkUpdates(_ time.Time) error {
	release, err := a.githubApp.LastRelease("vibioh", "viws")
	if err != nil {
		return err
	}

	logger.Info("%+v", release)
	return nil
}
