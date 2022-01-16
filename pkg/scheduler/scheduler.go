package scheduler

import (
	"context"
	"flag"
	"strings"
	"syscall"
	"time"

	"github.com/ViBiOh/flags"
	"github.com/ViBiOh/httputils/v4/pkg/cron"
	"github.com/ViBiOh/httputils/v4/pkg/logger"
	"github.com/ViBiOh/httputils/v4/pkg/redis"
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
	timezone    string
	hour        string
	redisApp    redis.App
	notifierApp notifier.App
}

// Flags adds flags for configuring package
func Flags(fs *flag.FlagSet, prefix string) Config {
	return Config{
		enabled:  flags.New(prefix, "scheduler", "Enabled").Default(true, nil).Label("Enable cron job").ToBool(fs),
		timezone: flags.New(prefix, "scheduler", "Timezone").Default("Europe/Paris", nil).Label("Timezone").ToString(fs),
		hour:     flags.New(prefix, "scheduler", "Hour").Default("08:00", nil).Label("Hour of cron, 24-hour format").ToString(fs),
	}
}

// New creates new App from Config
func New(config Config, notifierApp notifier.App, redisApp redis.App) App {
	if !*config.enabled {
		return nil
	}

	return app{
		timezone:    strings.TrimSpace(*config.timezone),
		hour:        strings.TrimSpace(*config.hour),
		notifierApp: notifierApp,
		redisApp:    redisApp,
	}
}

func (a app) Start(done <-chan struct{}) {
	cron.New().At(a.hour).In(a.timezone).Days().OnError(func(err error) {
		logger.Error("error while running ketchup notify: %s", err)
	}).OnSignal(syscall.SIGUSR1).Exclusive(a.redisApp, "ketchup:notify", 10*time.Minute).Start(func(ctx context.Context) error {
		logger.Info("Starting ketchup notifier")
		defer logger.Info("Ending ketchup notifier")
		return a.notifierApp.Notify(ctx)
	}, done)
}
