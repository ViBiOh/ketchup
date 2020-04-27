package ketchup

import (
	"context"
	"flag"
	"fmt"
	"strings"
	"time"

	"github.com/ViBiOh/httputils/v3/pkg/cron"
	"github.com/ViBiOh/httputils/v3/pkg/flags"
	"github.com/ViBiOh/httputils/v3/pkg/logger"
	"github.com/ViBiOh/ketchup/pkg/github"
	"github.com/ViBiOh/ketchup/pkg/store"
	mailer "github.com/ViBiOh/mailer/pkg/client"
)

// App of package
type App interface {
	Start()
}

// Config of package
type Config struct {
	emailTo  *string
	timezone *string
	hour     *string
}

type app struct {
	emailTo  string
	timezone string
	hour     string

	repositoryStore store.RepositoryStore
	githubApp       github.App
	mailerApp       mailer.App
}

// Flags adds flags for configuring package
func Flags(fs *flag.FlagSet, prefix string) Config {
	return Config{
		emailTo:  flags.New(prefix, "ketchup").Name("To").Default("").Label("Email to send notification").ToString(fs),
		timezone: flags.New(prefix, "ketchup").Name("Timezone").Default("Europe/Paris").Label("Timezone").ToString(fs),
		hour:     flags.New(prefix, "ketchup").Name("Hour").Default("08:00").Label("Hour of cron, 24-hour format").ToString(fs),
	}
}

// New creates new App from Config
func New(config Config, repositoryStore store.RepositoryStore, githubApp github.App, mailerApp mailer.App) App {
	return app{
		emailTo:  strings.TrimSpace(*config.emailTo),
		timezone: strings.TrimSpace(*config.timezone),
		hour:     strings.TrimSpace(*config.hour),

		repositoryStore: repositoryStore,
		githubApp:       githubApp,
		mailerApp:       mailerApp,
	}
}

func (a app) Start() {
	cron.New().At(a.hour).In(a.timezone).Days().Start(a.checkUpdates, func(err error) {
		logger.Error("error while running cron: %s", err)
	})
}

func (a app) checkUpdates(_ time.Time) error {
	ctx := context.Background()

	repositories, _, err := a.repositoryStore.List(ctx, 1, 100, "", false)
	if err != nil {
		return fmt.Errorf("unable to get repositories: %s", err)
	}

	newReleases := make([]github.Release, 0)

	for _, repository := range repositories {
		release, err := a.githubApp.LastRelease(repository.Name)
		if err != nil {
			return err
		}

		if release.TagName != repository.Version {
			logger.Info("New version available for %s: %s", repository.Name, release.TagName)
			repository.Version = release.TagName

			newReleases = append(newReleases, release)

			if err = a.repositoryStore.Update(ctx, repository); err != nil {
				return fmt.Errorf("unable to update repository %s: %s", repository.Name, err)
			}
		} else if release.TagName == repository.Version {
			logger.Info("%s is up-to-date!", repository.Name)
		} else {
			logger.Info("%s still need to be updated to %s!", repository.Name, release.TagName)
		}
	}

	if err := a.sendNotification(ctx, newReleases); err != nil {
		return err
	}

	return nil
}

func (a app) sendNotification(ctx context.Context, releases []github.Release) error {
	if len(releases) == 0 {
		return nil
	}

	if a.mailerApp == nil || !a.mailerApp.Enabled() {
		logger.Warn("mailer is not configured")
		return nil
	}

	payload := map[string]interface{}{
		"targets": releases,
	}

	email := mailer.NewEmail(a.mailerApp).Template("ketchup").From("ketchup@vibioh.fr").As("Ketchup").To(a.emailTo).Data(payload)
	if len(releases) > 1 {
		email.WithSubject("Ketchup - New releases")
	} else {
		email.WithSubject("Ketchup - New release")
	}

	if err := email.Send(ctx); err != nil {
		return fmt.Errorf("unable to send email to %s: %s", a.emailTo, err)
	}

	return nil
}
