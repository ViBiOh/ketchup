package scheduler

import (
	"context"
	"flag"
	"log/slog"
	"strings"
	"syscall"
	"time"

	"github.com/ViBiOh/flags"
	"github.com/ViBiOh/httputils/v4/pkg/cron"
	"github.com/ViBiOh/httputils/v4/pkg/redis"
	"github.com/ViBiOh/ketchup/pkg/notifier"
	"go.opentelemetry.io/otel/trace"
)

type App interface {
	Start(context.Context)
}

type Config struct {
	enabled  *bool
	timezone *string
	hour     *string
}

type app struct {
	tracerProvider trace.TracerProvider
	timezone       string
	hour           string
	redisApp       redis.Client
	notifierApp    notifier.App
}

func Flags(fs *flag.FlagSet, prefix string) Config {
	return Config{
		enabled:  flags.New("Enabled", "Enable cron job").Prefix(prefix).DocPrefix("scheduler").Bool(fs, true, nil),
		timezone: flags.New("Timezone", "Timezone").Prefix(prefix).DocPrefix("scheduler").String(fs, "Europe/Paris", nil),
		hour:     flags.New("Hour", "Hour of cron, 24-hour format").Prefix(prefix).DocPrefix("scheduler").String(fs, "08:00", nil),
	}
}

func New(config Config, notifierApp notifier.App, redisApp redis.Client, tracerProvider trace.TracerProvider) App {
	if !*config.enabled {
		return nil
	}

	return app{
		timezone:       strings.TrimSpace(*config.timezone),
		hour:           strings.TrimSpace(*config.hour),
		notifierApp:    notifierApp,
		redisApp:       redisApp,
		tracerProvider: tracerProvider,
	}
}

func (a app) Start(ctx context.Context) {
	cron.New().At(a.hour).In(a.timezone).Days().WithTracerProvider(a.tracerProvider).OnError(func(err error) {
		slog.Error("error while running ketchup notify", "err", err)
	}).OnSignal(syscall.SIGUSR1).Exclusive(a.redisApp, "ketchup:notify", 10*time.Minute).Start(ctx, func(ctx context.Context) error {
		slog.Info("Starting ketchup notifier")
		defer slog.Info("Ending ketchup notifier")
		return a.notifierApp.Notify(ctx)
	})
}
