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
	"github.com/ViBiOh/ketchup/pkg/github"
	"github.com/ViBiOh/ketchup/pkg/model"
	"github.com/ViBiOh/ketchup/pkg/semver"
	"github.com/ViBiOh/ketchup/pkg/service/ketchup"
	"github.com/ViBiOh/ketchup/pkg/service/repository"
	mailer "github.com/ViBiOh/mailer/pkg/client"
)

const (
	pageSize = 20
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
	timezone string
	hour     string
	loginID  uint64

	repositoryService repository.App
	ketchupService    ketchup.App
	githubApp         github.App
	mailerApp         mailer.App
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
func New(config Config, repositoryService repository.App, ketchupService ketchup.App, githubApp github.App, mailerApp mailer.App) App {
	return app{
		timezone: strings.TrimSpace(*config.timezone),
		hour:     strings.TrimSpace(*config.hour),
		loginID:  uint64(*config.loginID),

		repositoryService: repositoryService,
		ketchupService:    ketchupService,
		githubApp:         githubApp,
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
	newReleases := make([]model.Release, 0)
	count := 0
	page := uint(1)

	for {
		repositories, totalCount, err := a.repositoryService.List(ctx, page, pageSize)
		if err != nil {
			return nil, fmt.Errorf("unable to fetch page %d of repositories: %s", page, err)
		}

		for _, repository := range repositories {
			count++

			release, err := a.githubApp.LastRelease(repository.Name)
			if err != nil {
				return nil, err
			}

			if release.TagName == repository.Version {
				continue
			}

			releaseVersion, releaseErr := semver.Parse(release.TagName)
			repositoryVersion, repositoryErr := semver.Parse(repository.Version)
			if releaseErr == nil && repositoryErr == nil && repositoryVersion.IsGreater(releaseVersion) {
				continue
			}

			logger.Info("New version available for %s: %s", repository.Name, release.TagName)
			repository.Version = release.TagName

			if err := a.repositoryService.Update(ctx, repository); err != nil {
				return nil, fmt.Errorf("unable to update repository %s: %s", repository.Name, err)
			}

			newReleases = append(newReleases, model.NewRelease(repository, release))
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

	userToNotify := make(map[model.User][]model.Release, 0)

	sort.Sort(model.ReleaseByRepositoryID(releases))
	sort.Sort(model.KetchupByRepositoryID(ketchups))

	ketchupsIndex := 0
	ketchupsSize := len(ketchups)

	for _, release := range releases {
		for ketchupsIndex < ketchupsSize {
			ketchup := ketchups[ketchupsIndex]
			if release.Repository.ID < ketchup.Repository.ID {
				break
			}

			if ketchup.Version != release.Release.TagName {
				if userToNotify[ketchup.User] != nil {
					userToNotify[ketchup.User] = append(userToNotify[ketchup.User], release)
				} else {
					userToNotify[ketchup.User] = []model.Release{release}
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

		githubReleases := make([]github.Release, len(releases))
		for index, release := range releases {
			githubReleases[index] = release.Release
		}

		payload := map[string]interface{}{
			"releases": githubReleases,
		}

		email := mailer.NewEmail(a.mailerApp).Template("ketchup").From("ketchup@vibioh.fr").As("Ketchup").To(user.Email).Data(payload)
		if len(releases) > 1 {
			email.WithSubject("Ketchup - New releases")
		} else {
			email.WithSubject("Ketchup - New release")
		}

		if err := email.Send(ctx); err != nil {
			return fmt.Errorf("unable to send email to %s: %s", user.Email, err)
		}
	}

	return nil
}
