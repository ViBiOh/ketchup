package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/ViBiOh/flags"
	"github.com/ViBiOh/httputils/v4/pkg/db"
	"github.com/ViBiOh/httputils/v4/pkg/logger"
	"github.com/ViBiOh/httputils/v4/pkg/request"
	"github.com/ViBiOh/httputils/v4/pkg/tracer"
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
	tracerConfig := tracer.Flags(fs, "tracer")

	dbConfig := db.Flags(fs, "db")
	mailerConfig := mailer.Flags(fs, "mailer")
	githubConfig := github.Flags(fs, "github")
	dockerConfig := docker.Flags(fs, "docker")
	notifierConfig := notifier.Flags(fs, "notifier")

	notificationType := fs.String("notification", "daily", "Notification type. \"daily\" or \"reminder\"")

	logger.Fatal(fs.Parse(os.Args[1:]))

	logger.Global(logger.New(loggerConfig))
	defer logger.Close()

	ctx := context.Background()

	ketchupDb, err := db.New(ctx, dbConfig, nil)
	logger.Fatal(err)
	defer ketchupDb.Close()

	tracerApp, err := tracer.New(ctx, tracerConfig)
	logger.Fatal(err)
	defer tracerApp.Close(ctx)
	request.AddTracerToDefaultClient(tracerApp.GetProvider())

	mailerApp, err := mailer.New(mailerConfig, nil, tracerApp.GetTracer("mailer"))
	logger.Fatal(err)
	defer mailerApp.Close()

	helmApp := helm.New()
	npmApp := npm.New()
	pypiApp := pypi.New()
	repositoryServiceApp := repositoryService.New(repositoryStore.New(ketchupDb), github.New(githubConfig, nil, nil, tracerApp), helmApp, docker.New(dockerConfig), npmApp, pypiApp)
	ketchupServiceApp := ketchupService.New(ketchupStore.New(ketchupDb), repositoryServiceApp)
	userServiceApp := userService.New(userStore.New(ketchupDb), nil)

	notifierApp := notifier.New(notifierConfig, repositoryServiceApp, ketchupServiceApp, userServiceApp, mailerApp, helmApp)

	logger.Info("Starting notifier...")

	ctx, end := tracer.StartSpan(ctx, tracerApp.GetTracer("notifier"), "notifier")
	defer end(&err)

	switch *notificationType {
	case "daily":
		if err = notifierApp.Notify(ctx); err != nil {
			logger.Fatal(err)
		}
	case "reminder":
		if err = notifierApp.Remind(ctx); err != nil {
			logger.Fatal(err)
		}
	default:
		logger.Fatal(fmt.Errorf("unknown notification type `%s`", *notificationType))
	}

	logger.Info("Notifier ended!")
}
