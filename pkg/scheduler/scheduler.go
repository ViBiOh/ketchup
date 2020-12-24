package scheduler

import (
	"context"
	"flag"
	"fmt"
	"sort"
	"strings"
	"time"

	authModel "github.com/ViBiOh/auth/v2/pkg/model"
	"github.com/ViBiOh/httputils/v3/pkg/cron"
	"github.com/ViBiOh/httputils/v3/pkg/flags"
	"github.com/ViBiOh/httputils/v3/pkg/logger"
	"github.com/ViBiOh/ketchup/pkg/model"
	"github.com/ViBiOh/ketchup/pkg/semver"
	"github.com/ViBiOh/ketchup/pkg/service/ketchup"
	"github.com/ViBiOh/ketchup/pkg/service/repository"
	mailer "github.com/ViBiOh/mailer/pkg/client"
)

var (
	pageSize = uint(20)
)

// App of package
type App interface {
	Start()
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

func (a app) Start() {
	cron.New().At(a.hour).In(a.timezone).Days().Start(a.ketchupNotify, func(err error) {
		logger.Error("error while running ketchup notify: %s", err)
	})
}

func (a app) ketchupNotify(_ time.Time) error {
	logger.Info("Starting ketchup notifier")

	ctx := authModel.StoreUser(context.Background(), authModel.NewUser(a.loginID, "scheduler"))

	if err := a.repositoryService.Clean(ctx); err != nil {
		return fmt.Errorf("unable to clean repository before starting: %s", err)
	}

	newReleases, err := a.getNewReleases(ctx)
	if err != nil {
		return fmt.Errorf("unable to get new releases: %s", err)
	}

	ketchupsToNotify, err := a.getKetchupToNotify(ctx, newReleases)
	if err != nil {
		return fmt.Errorf("unable to get ketchup to notify: %s", err)
	}

	if err := a.sendNotification(ctx, ketchupsToNotify); err != nil {
		return err
	}

	return nil
}

func (a app) getNewReleases(ctx context.Context) ([]model.Release, error) {
	var newReleases []model.Release
	count := 0
	page := uint(1)

	for {
		repositories, totalCount, err := a.repositoryService.List(ctx, page, pageSize)
		if err != nil {
			return nil, fmt.Errorf("unable to fetch page %d of repositories: %s", page, err)
		}

		for _, repo := range repositories {
			count++

			latestVersion, err := a.repositoryService.LatestVersion(repo)
			if err != nil {
				logger.Error("unable to get latest version of %s: %s", repo.Name, err)
				continue
			}

			if latestVersion.Name == repo.Version {
				continue
			}

			repositoryVersion, repositoryErr := semver.Parse(repo.Version)
			if repositoryErr == nil && repositoryVersion.IsGreater(latestVersion) {
				continue
			}

			logger.Info("New version available for %s: %s", repo.Name, latestVersion.Name)
			repo.Version = latestVersion.Name

			if err := a.repositoryService.Update(ctx, repo); err != nil {
				return nil, fmt.Errorf("unable to update repo %s: %s", repo.Name, err)
			}

			newReleases = append(newReleases, model.NewRelease(repo, latestVersion))
		}

		if uint64(page*pageSize) < totalCount {
			page++
		} else {
			logger.Info("%d repositories checked, %d new releases", count, len(newReleases))
			return newReleases, nil
		}
	}
}

func (a app) getKetchupToNotify(ctx context.Context, releases []model.Release) (map[model.User][]model.Release, error) {
	repositories := make([]model.Repository, len(releases))
	for index, release := range releases {
		repositories[index] = release.Repository
	}

	ketchups, err := a.ketchupService.ListForRepositories(ctx, repositories)
	if err != nil {
		return nil, fmt.Errorf("unable to get ketchups for repositories: %s", err)
	}

	userToNotify := make(map[model.User][]model.Release)

	sort.Sort(model.ReleaseByRepositoryID(releases))
	sort.Sort(model.KetchupByRepositoryID(ketchups))

	ketchupsIndex := 0
	ketchupsSize := len(ketchups)

	for _, release := range releases {
		for ketchupsIndex < ketchupsSize {
			current := ketchups[ketchupsIndex]
			if release.Repository.ID < current.Repository.ID {
				break
			}

			if current.Version != release.Version.Name {
				if userToNotify[current.User] != nil {
					userToNotify[current.User] = append(userToNotify[current.User], release)
				} else {
					userToNotify[current.User] = []model.Release{release}
				}
			}

			ketchupsIndex++
		}
	}

	logger.Info("%d ketchups for %d users to notify", len(ketchups), len(userToNotify))

	return userToNotify, nil
}

func (a app) sendNotification(ctx context.Context, ketchupToNotify map[model.User][]model.Release) error {
	if len(ketchupToNotify) == 0 {
		return nil
	}

	if a.mailerApp == nil || !a.mailerApp.Enabled() {
		logger.Warn("mailer is not configured")
		return nil
	}

	for user, releases := range ketchupToNotify {
		logger.Info("Sending email to %s for %d releases", user.Email, len(releases))

		payload := map[string]interface{}{
			"releases": releases,
		}

		email := mailer.NewEmail().Template("ketchup").From("ketchup@vibioh.fr").As("Ketchup").To(user.Email).Data(payload)
		if len(releases) > 1 {
			email.WithSubject("Ketchup - New releases")
		} else {
			email.WithSubject("Ketchup - New release")
		}

		if err := email.Send(ctx, a.mailerApp); err != nil {
			return fmt.Errorf("unable to send email to %s: %s", user.Email, err)
		}
	}

	return nil
}
