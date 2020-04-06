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
	"github.com/ViBiOh/ketchup/pkg/target"
)

// App of package
type App interface {
	Start()
}

// Config of package
type Config struct {
	emailTo  *string
	timezone *string
}

type app struct {
	emailTo  string
	timezone string

	targetApp target.App
	githubApp github.App
}

// Flags adds flags for configuring package
func Flags(fs *flag.FlagSet, prefix string) Config {
	return Config{
		emailTo:  flags.New(prefix, "ketchup").Name("To").Default("").Label("Email to send notification").ToString(fs),
		timezone: flags.New(prefix, "ketchup").Name("Timezone").Default("Europe/Paris").Label("Timezone").ToString(fs),
	}
}

// New creates new App from Config
func New(config Config, targetApp target.App, githubApp github.App) App {
	return &app{
		emailTo:  strings.TrimSpace(*config.emailTo),
		timezone: strings.TrimSpace(*config.timezone),

		targetApp: targetApp,
		githubApp: githubApp,
	}
}

func (a app) Start() {
	cron.New().At("08:00").In(a.timezone).Days().Start(a.checkUpdates, func(err error) {
		logger.Error("error while running cron: %s", err)
	})
}

func (a app) checkUpdates(_ time.Time) error {
	targets, _, err := a.targetApp.List(context.Background(), 1, 100, "", false, nil)
	if err != nil {
		return fmt.Errorf("unable to get targets: %s", err)
	}

	for _, o := range targets {
		target := o.(target.Target)

		release, err := a.githubApp.LastRelease(target.Owner, target.Repository)
		if err != nil {
			return err
		}

		if release.TagName != target.Version {
			logger.Info("New version available for %s/%s", target.Owner, target.Repository)
		} else {
			logger.Info("%s/%s is up-to-date!", target.Owner, target.Repository)
		}
	}

	return nil
}
