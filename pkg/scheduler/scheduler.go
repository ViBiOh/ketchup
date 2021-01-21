package scheduler

import (
	"context"
	"flag"
	"fmt"
	"strings"
	"syscall"
	"time"

	authModel "github.com/ViBiOh/auth/v2/pkg/model"
	"github.com/ViBiOh/httputils/v3/pkg/cron"
	"github.com/ViBiOh/httputils/v3/pkg/flags"
	"github.com/ViBiOh/httputils/v3/pkg/logger"
	"github.com/ViBiOh/ketchup/pkg/model"
	"github.com/ViBiOh/ketchup/pkg/service/ketchup"
	"github.com/ViBiOh/ketchup/pkg/service/repository"
	mailer "github.com/ViBiOh/mailer/pkg/client"
)

var (
	pageSize = uint(20)
)

// App of package
type App interface {
	Start(<-chan struct{})
}

// Config of package
type Config struct {
	timezone *string
	hour     *string
	loginID  *uint
}

type app struct {
	repositoryService repository.App
	ketchupService    ketchup.App
	mailerApp         mailer.App

	timezone string
	hour     string
	loginID  uint64
}

// Flags adds flags for configuring package
func Flags(fs *flag.FlagSet, prefix string) Config {
	return Config{
		timezone: flags.New(prefix, "scheduler").Name("Timezone").Default("Europe/Paris").Label("Timezone").ToString(fs),
		hour:     flags.New(prefix, "scheduler").Name("Hour").Default("08:00").Label("Hour of cron, 24-hour format").ToString(fs),
		loginID:  flags.New(prefix, "scheduler").Name("LoginID").Default(1).Label("Scheduler user ID").ToUint(fs),
	}
}

// New creates new App from Config
func New(config Config, repositoryService repository.App, ketchupService ketchup.App, mailerApp mailer.App) App {
	return app{
		timezone: strings.TrimSpace(*config.timezone),
		hour:     strings.TrimSpace(*config.hour),
		loginID:  uint64(*config.loginID),

		repositoryService: repositoryService,
		ketchupService:    ketchupService,
		mailerApp:         mailerApp,
	}
}

func (a app) Start(done <-chan struct{}) {
	cron.New().At(a.hour).In(a.timezone).Days().OnError(func(err error) {
		logger.Error("error while running ketchup notify: %s", err)
	}).OnSignal(syscall.SIGUSR1).Start(a.ketchupNotify, done)
}

func (a app) ketchupNotify(_ time.Time) error {
	logger.Info("Starting ketchup notifier")

	ctx := authModel.StoreUser(context.Background(), authModel.NewUser(a.loginID, "scheduler"))

	if err := a.repositoryService.Clean(ctx); err != nil {
		return fmt.Errorf("unable to clean repository before starting: %s", err)
	}

	if err := a.notifyUser(ctx); err != nil {
		return fmt.Errorf("unable to notify user: %s", err)
	}

	return nil
}

func (a app) notifyUser(ctx context.Context) error {
	var repositories []model.Repository
	var err error
	totalCount := uint64(0)

	for page := uint(1); err == nil && uint64(page*pageSize) <= totalCount; page++ {
		repositories, totalCount, err = a.repositoryService.List(ctx, page, pageSize)
		if err != nil {
			return fmt.Errorf("unable to fetch page %d of repositories: %s", page, err)
		}

		for _, repo := range repositories {
			kinds, err := a.ketchupService.ListKindsByRepositoryID(ctx, repo)
			if err != nil {
				return fmt.Errorf("unable to list kinds for repository `%s`: %s", repo.Name, err)
			}
			fmt.Println(kinds)
		}
	}

	return nil
}
