package scheduler

import (
	"context"
	"flag"
	"log/slog"
	"syscall"
	"time"

	"github.com/ViBiOh/flags"
	"github.com/ViBiOh/httputils/v4/pkg/cron"
	"github.com/ViBiOh/httputils/v4/pkg/redis"
	"github.com/ViBiOh/ketchup/pkg/notifier"
	"go.opentelemetry.io/otel/trace"
)

type Service interface {
	Start(context.Context)
}

type Config struct {
	timezone string
	hour     string
	enabled  bool
}

type service struct {
	tracerProvider trace.TracerProvider
	timezone       string
	hour           string
	redis          redis.Client
	notifier       notifier.Service
}

func Flags(fs *flag.FlagSet, prefix string) *Config {
	var config Config

	flags.New("Enabled", "Enable cron job").Prefix(prefix).DocPrefix("scheduler").BoolVar(fs, &config.enabled, true, nil)
	flags.New("Timezone", "Timezone").Prefix(prefix).DocPrefix("scheduler").StringVar(fs, &config.timezone, "Europe/Paris", nil)
	flags.New("Hour", "Hour of cron, 24-hour format").Prefix(prefix).DocPrefix("scheduler").StringVar(fs, &config.hour, "08:00", nil)

	return &config
}

func New(config *Config, notifierService notifier.Service, redisClient redis.Client, tracerProvider trace.TracerProvider) Service {
	if !config.enabled {
		return nil
	}

	return service{
		timezone:       config.timezone,
		hour:           config.hour,
		notifier:       notifierService,
		redis:          redisClient,
		tracerProvider: tracerProvider,
	}
}

func (s service) Start(ctx context.Context) {
	cron.New().At(s.hour).In(s.timezone).Days().WithTracerProvider(s.tracerProvider).OnError(func(ctx context.Context, err error) {
		slog.ErrorContext(ctx, "error while running ketchup notify", "error", err)
	}).OnSignal(syscall.SIGUSR1).Exclusive(s.redis, "ketchup:notify", 10*time.Minute).Start(ctx, func(ctx context.Context) error {
		slog.InfoContext(ctx, "Starting ketchup notifier")
		defer slog.InfoContext(ctx, "Ending ketchup notifier")

		return s.notifier.Notify(ctx)
	})
}
