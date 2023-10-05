package main

import (
	"context"
	"flag"
	"log"
	"log/slog"
	"os"

	"github.com/ViBiOh/flags"
	"github.com/ViBiOh/httputils/v4/pkg/db"
	"github.com/ViBiOh/httputils/v4/pkg/logger"
	"github.com/ViBiOh/httputils/v4/pkg/request"
	"github.com/ViBiOh/httputils/v4/pkg/telemetry"
	"github.com/ViBiOh/ketchup/pkg/notifier"
	"github.com/ViBiOh/ketchup/pkg/provider/docker"
	"github.com/ViBiOh/ketchup/pkg/provider/github"
	"github.com/ViBiOh/ketchup/pkg/provider/helm"
	"github.com/ViBiOh/ketchup/pkg/provider/npm"
	"github.com/ViBiOh/ketchup/pkg/provider/pypi"
	ketchupService "github.com/ViBiOh/ketchup/pkg/service/ketchup"
	repositoryService "github.com/ViBiOh/ketchup/pkg/service/repository"
	userService "github.com/ViBiOh/ketchup/pkg/service/user"
	ketchupStore "github.com/ViBiOh/ketchup/pkg/store/ketchup"
	repositoryStore "github.com/ViBiOh/ketchup/pkg/store/repository"
	userStore "github.com/ViBiOh/ketchup/pkg/store/user"
	mailer "github.com/ViBiOh/mailer/pkg/client"
	"go.opentelemetry.io/otel/sdk/metric"
)

func main() {
	fs := flag.NewFlagSet("ketchup", flag.ExitOnError)
	fs.Usage = flags.Usage(fs)

	loggerConfig := logger.Flags(fs, "logger")
	telemetryConfig := telemetry.Flags(fs, "telemetry")

	dbConfig := db.Flags(fs, "db")
	mailerConfig := mailer.Flags(fs, "mailer")
	githubConfig := github.Flags(fs, "github")
	dockerConfig := docker.Flags(fs, "docker")
	notifierConfig := notifier.Flags(fs, "notifier")

	notificationType := fs.String("notification", "daily", "Notification type. \"daily\" or \"reminder\"")

	if err := fs.Parse(os.Args[1:]); err != nil {
		log.Fatal(err)
	}

	logger.Init(loggerConfig)

	ctx := context.Background()

	telemetryService, err := telemetry.New(ctx, telemetryConfig)
	if err != nil {
		slog.Error("create telemetry", "err", err)
		os.Exit(1)
	}

	defer telemetryService.Close(ctx)
	request.AddOpenTelemetryToDefaultClient(telemetryService.MeterProvider(), telemetryService.TracerProvider())

	ketchupDb, err := db.New(ctx, dbConfig, telemetryService.TracerProvider())
	if err != nil {
		slog.Error("create database", "err", err)
		os.Exit(1)
	}

	defer ketchupDb.Close()

	mailerService, err := mailer.New(mailerConfig, telemetryService.MeterProvider(), telemetryService.TracerProvider())
	if err != nil {
		slog.Error("create mailer", "err", err)
		os.Exit(1)
	}

	defer mailerService.Close()

	helmService := helm.New()
	npmService := npm.New()
	pypiService := pypi.New()
	repositoryServiceService := repositoryService.New(repositoryStore.New(ketchupDb), github.New(githubConfig, nil, nil, telemetryService.TracerProvider()), helmService, docker.New(dockerConfig), npmService, pypiService)
	ketchupServiceService := ketchupService.New(ketchupStore.New(ketchupDb), repositoryServiceService)
	userServiceService := userService.New(userStore.New(ketchupDb), nil)

	notifierService := notifier.New(notifierConfig, repositoryServiceService, ketchupServiceService, userServiceService, mailerService, helmService)

	slog.Info("Starting notifier...")

	ctx, end := telemetry.StartSpan(ctx, telemetryService.TracerProvider().Tracer("notifier"), "notifier")
	defer end(&err)

	switch *notificationType {
	case "daily":
		if err = notifierService.Notify(ctx, telemetryService.MeterProvider()); err != nil {
			slog.Error("notify", "err", err)
			os.Exit(1)
		}
	case "reminder":
		if err = notifierService.Remind(ctx); err != nil {
			slog.Error("remind", "err", err)
			os.Exit(1)
		}
	default:
		slog.Error("unknown notification type", "type", *notificationType)
		os.Exit(1)
	}

	meterProvider := telemetryService.MeterProvider()
	if flushableProvider, ok := meterProvider.(*metric.MeterProvider); ok {
		if err := flushableProvider.ForceFlush(ctx); err != nil {
			slog.Error("flush meter provider", "err", err)
		}
	}

	slog.Info("Notifier ended!")
}
