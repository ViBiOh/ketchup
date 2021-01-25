package scheduler

import (
	"context"
	"flag"
	"fmt"
	"sort"
	"strings"
	"syscall"
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
	mailerModel "github.com/ViBiOh/mailer/pkg/model"
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
	defer logger.Info("Ending ketchup notifier")

	ctx := authModel.StoreUser(context.Background(), authModel.NewUser(a.loginID, "scheduler"))

	if err := a.repositoryService.Clean(ctx); err != nil {
		return fmt.Errorf("unable to clean repository before starting: %w", err)
	}

	newReleases, err := a.getNewReleases(ctx)
	if err != nil {
		return fmt.Errorf("unable to get new releases: %w", err)
	}

	ketchupsToNotify, err := a.getKetchupToNotify(ctx, newReleases)
	if err != nil {
		return fmt.Errorf("unable to get ketchup to notify: %w", err)
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

			releases := a.checkRepositoryVersion(repo)
			if len(releases) == 0 {
				continue
			}

			newReleases = append(newReleases, releases...)
			for _, release := range releases {
				repo.Versions[release.Pattern] = release.Version.Name
			}

			if err := a.repositoryService.Update(ctx, repo); err != nil {
				return nil, fmt.Errorf("unable to update repo `%s`: %s", repo.Name, err)
			}
		}

		if uint64(page*pageSize) < totalCount {
			page++
		} else {
			logger.Info("%d repositories checked, %d new releases", count, len(newReleases))
			return newReleases, nil
		}
	}
}

func (a app) checkRepositoryVersion(repo model.Repository) []model.Release {
	versions, err := a.repositoryService.LatestVersions(repo)
	if err != nil {
		logger.Error("unable to get latest versions of %s: %s", repo.Name, err)
		return nil
	}

	releases := make([]model.Release, 0)

	for pattern, version := range versions {
		repositoryVersionName := repo.Versions[pattern]

		if version.Name == repositoryVersionName {
			continue
		}

		repositoryVersion, err := semver.Parse(repositoryVersionName)
		if err != nil {
			continue
		}

		if !version.IsGreater(repositoryVersion) {
			continue
		}

		logger.Info("New `%s` version available for %s: %s", pattern, repo.Name, version.Name)

		releases = append(releases, model.NewRelease(repo, pattern, version))
	}

	return releases
}

func (a app) getKetchupToNotify(ctx context.Context, releases []model.Release) (map[model.User][]model.Release, error) {
	repositories := make([]model.Repository, len(releases))
	for index, release := range releases {
		repositories[index] = release.Repository
	}

	ketchups, err := a.ketchupService.ListForRepositories(ctx, repositories)
	if err != nil {
		return nil, fmt.Errorf("unable to get ketchups for repositories: %w", err)
	}

	userToNotify := make(map[model.User][]model.Release)

	sort.Sort(model.ReleaseByRepositoryID(releases))
	sort.Sort(model.KetchupByRepositoryID(ketchups))

	index := 0
	size := len(ketchups)

	for _, release := range releases {
		for index < size {
			current := ketchups[index]
			if release.Repository.ID < current.Repository.ID || release.Pattern < current.Pattern {
				break
			}

			index++

			if current.Pattern != release.Pattern {
				continue
			}

			if current.Version != release.Version.Name {
				if userToNotify[current.User] != nil {
					userToNotify[current.User] = append(userToNotify[current.User], release)
				} else {
					userToNotify[current.User] = []model.Release{release}
				}
			}
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

		mailRequest := mailerModel.NewMailRequest().Template("ketchup").From("ketchup@vibioh.fr").As("Ketchup").To(user.Email).Data(payload)
		subject := fmt.Sprintf("Ketchup - %d new release", len(releases))
		if len(releases) > 1 {
			subject += "s"
		}
		mailRequest.WithSubject(subject)

		if err := a.mailerApp.Send(ctx, *mailRequest); err != nil {
			return fmt.Errorf("unable to send email to %s: %s", user.Email, err)
		}
	}

	return nil
}
