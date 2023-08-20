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

	telemetryApp, err := telemetry.New(ctx, telemetryConfig)
	if err != nil {
		slog.Error("create telemetry", "err", err)
		os.Exit(1)
	}

	defer telemetryApp.Close(ctx)
	request.AddOpenTelemetryToDefaultClient(telemetryApp.GetMeterProvider(), telemetryApp.GetTraceProvider())

	ketchupDb, err := db.New(ctx, dbConfig, telemetryApp.GetTracer("database"))
	if err != nil {
		slog.Error("create database", "err", err)
		os.Exit(1)
	}

	defer ketchupDb.Close()

	mailerApp, err := mailer.New(mailerConfig, nil, telemetryApp.GetTracer("mailer"))
	if err != nil {
		slog.Error("create mailer", "err", err)
		os.Exit(1)
	}

	defer mailerApp.Close()

	helmApp := helm.New()
	npmApp := npm.New()
	pypiApp := pypi.New()
	repositoryServiceApp := repositoryService.New(repositoryStore.New(ketchupDb), github.New(githubConfig, nil, nil, telemetryApp.GetTraceProvider()), helmApp, docker.New(dockerConfig), npmApp, pypiApp)
	ketchupServiceApp := ketchupService.New(ketchupStore.New(ketchupDb), repositoryServiceApp)
	userServiceApp := userService.New(userStore.New(ketchupDb), nil)

	notifierApp := notifier.New(notifierConfig, repositoryServiceApp, ketchupServiceApp, userServiceApp, mailerApp, helmApp)

	slog.Info("Starting notifier...")

	ctx, end := telemetry.StartSpan(ctx, telemetryApp.GetTracer("notifier"), "notifier")
	defer end(&err)

	switch *notificationType {
	case "daily":
		if err = notifierApp.Notify(ctx); err != nil {
			slog.Error("notify", "err", err)
			os.Exit(1)
		}
	case "reminder":
		if err = notifierApp.Remind(ctx); err != nil {
			slog.Error("remind", "err", err)
			os.Exit(1)
		}
	default:
		slog.Error("unknown notification type", "type", *notificationType)
		os.Exit(1)
	}

	slog.Info("Notifier ended!")
}
