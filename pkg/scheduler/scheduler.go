package scheduler

import (
	"flag"
	"strings"
	"syscall"
	"time"

	"github.com/ViBiOh/httputils/v4/pkg/cron"
	"github.com/ViBiOh/httputils/v4/pkg/flags"
	"github.com/ViBiOh/httputils/v4/pkg/logger"
	"github.com/ViBiOh/ketchup/pkg/notifier"
)

// App of package
type App interface {
	Start(<-chan struct{})
}

// Config of package
type Config struct {
	enabled  *bool
	timezone *string
	hour     *string
}

type app struct {
	notifierApp notifier.App

	timezone string
	hour     string
}

// Flags adds flags for configuring package
func Flags(fs *flag.FlagSet, prefix string) Config {
	return Config{
		enabled:  flags.New(prefix, "scheduler").Name("Enabled").Default(true).Label("Enable cron job").ToBool(fs),
		timezone: flags.New(prefix, "scheduler").Name("Timezone").Default("Europe/Paris").Label("Timezone").ToString(fs),
		hour:     flags.New(prefix, "scheduler").Name("Hour").Default("08:00").Label("Hour of cron, 24-hour format").ToString(fs),
	}
}

// New creates new App from Config
func New(config Config, notifierApp notifier.App) App {
	if !*config.enabled {
		return nil
	}

	return app{
		timezone:    strings.TrimSpace(*config.timezone),
		hour:        strings.TrimSpace(*config.hour),
		notifierApp: notifierApp,
	}
}

func (a app) Start(done <-chan struct{}) {
	cron.New().At(a.hour).In(a.timezone).Days().OnError(func(err error) {
		logger.Error("error while running ketchup notify: %s", err)
	}).OnSignal(syscall.SIGUSR1).Start(func(_ time.Time) error {
		logger.Info("Starting ketchup notifier")
		defer logger.Info("Ending ketchup notifier")
		return a.notifierApp.Notify()
	}, done)
}
