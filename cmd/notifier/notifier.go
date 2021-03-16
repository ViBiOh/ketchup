package main

import (
	"flag"
	"os"

	"github.com/ViBiOh/httputils/v4/pkg/db"
	"github.com/ViBiOh/httputils/v4/pkg/logger"
	"github.com/ViBiOh/ketchup/pkg/github"
	"github.com/ViBiOh/ketchup/pkg/helm"
	"github.com/ViBiOh/ketchup/pkg/notifier"
	ketchupService "github.com/ViBiOh/ketchup/pkg/service/ketchup"
	repositoryService "github.com/ViBiOh/ketchup/pkg/service/repository"
	ketchupStore "github.com/ViBiOh/ketchup/pkg/store/ketchup"
	repositoryStore "github.com/ViBiOh/ketchup/pkg/store/repository"
	mailer "github.com/ViBiOh/mailer/pkg/client"
)

func main() {
	fs := flag.NewFlagSet("ketchup", flag.ExitOnError)

	loggerConfig := logger.Flags(fs, "logger")

	dbConfig := db.Flags(fs, "db")
	mailerConfig := mailer.Flags(fs, "mailer")
	githubConfig := github.Flags(fs, "github")
	notifierConfig := notifier.Flags(fs, "notifier")

	logger.Fatal(fs.Parse(os.Args[1:]))

	logger.Global(logger.New(loggerConfig))
	defer logger.Close()

	ketchupDb, err := db.New(dbConfig)
	logger.Fatal(err)
	defer func() {
		if err := ketchupDb.Close(); err != nil {
			logger.Error("error while closing database connection: %s", err)
		}
	}()

	mailerApp, err := mailer.New(mailerConfig)
	logger.Fatal(err)
	defer mailerApp.Close()

	repositoryServiceApp := repositoryService.New(repositoryStore.New(ketchupDb), github.New(githubConfig), helm.New())
	ketchupServiceApp := ketchupService.New(ketchupStore.New(ketchupDb), repositoryServiceApp)

	notifierApp := notifier.New(notifierConfig, repositoryServiceApp, ketchupServiceApp, mailerApp)
	if err := notifierApp.Notify(); err != nil {
		logger.Fatal(err)
	}
}
