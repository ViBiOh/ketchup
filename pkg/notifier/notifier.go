package notifier

import (
	"context"
	"flag"
	"fmt"
	"sort"
	"strings"
	"time"

	authModel "github.com/ViBiOh/auth/v2/pkg/model"
	"github.com/ViBiOh/httputils/v4/pkg/flags"
	"github.com/ViBiOh/httputils/v4/pkg/logger"
	"github.com/ViBiOh/ketchup/pkg/model"
	"github.com/ViBiOh/ketchup/pkg/provider/helm"
	"github.com/ViBiOh/ketchup/pkg/semver"
	"github.com/ViBiOh/ketchup/pkg/service/ketchup"
	"github.com/ViBiOh/ketchup/pkg/service/repository"
	"github.com/ViBiOh/ketchup/pkg/service/user"
	mailerModel "github.com/ViBiOh/mailer/pkg/model"
	"github.com/prometheus/client_golang/prometheus/push"
)

var (
	pageSize = uint(20)
)

// Mailer interface client
type Mailer interface {
	Enabled() bool
	Send(context.Context, mailerModel.MailRequest) error
}

// App of package
type App interface {
	Notify(context.Context) error
}

// Config of package
type Config struct {
	loginID *uint
	pushURL *string
}

type app struct {
	repositoryService repository.App
	ketchupService    ketchup.App
	userService       user.App
	mailerApp         Mailer
	helmApp           helm.App

	clock *Clock

	pushURL string
	loginID uint64
}

// Flags adds flags for configuring package
func Flags(fs *flag.FlagSet, prefix string) Config {
	return Config{
		loginID: flags.New(prefix, "notifier").Name("LoginID").Default(1).Label("Scheduler user ID").ToUint(fs),
		pushURL: flags.New(prefix, "notifier").Name("PushUrl").Default("").Label("Pushgateway URL").ToString(fs),
	}
}

// New creates new App from Config
func New(config Config, repositoryService repository.App, ketchupService ketchup.App, userService user.App, mailerApp Mailer, helmApp helm.App) App {
	return app{
		loginID: uint64(*config.loginID),
		pushURL: strings.TrimSpace(*config.pushURL),

		repositoryService: repositoryService,
		ketchupService:    ketchupService,
		userService:       userService,
		mailerApp:         mailerApp,
		helmApp:           helmApp,
	}
}

func (a app) Notify(ctx context.Context) error {
	userCtx := authModel.StoreUser(ctx, authModel.NewUser(a.loginID, "scheduler"))

	if err := a.repositoryService.Clean(userCtx); err != nil {
		return fmt.Errorf("unable to clean repository before starting: %w", err)
	}

	newReleases, repoCount, err := a.getNewReleases(userCtx)
	if err != nil {
		return fmt.Errorf("unable to get new releases: %w", err)
	}

	sort.Sort(model.ReleaseByRepositoryIDAndPattern(newReleases))
	if err := a.updateRepositories(userCtx, newReleases); err != nil {
		return fmt.Errorf("unable to update repositories: %w", err)
	}

	ketchupsToNotify, err := a.getKetchupToNotify(userCtx, newReleases)
	if err != nil {
		return fmt.Errorf("unable to get ketchup to notify: %w", err)
	}

	if err := a.sendNotification(userCtx, ketchupsToNotify); err != nil {
		return err
	}

	if len(a.pushURL) != 0 {
		registry, metrics := configurePrometheus()

		userCount, err := a.userService.Count(ctx)
		if err != nil {
			logger.Error("unable to get users count: %s", err)
		} else {
			metrics.WithLabelValues("users").Set(float64(userCount))
		}

		metrics.WithLabelValues("repositories").Set(float64(repoCount))
		metrics.WithLabelValues("releases").Set(float64(len(newReleases)))
		metrics.WithLabelValues("notifications").Set(float64(len(ketchupsToNotify)))

		if err := push.New(a.pushURL, "ketchup").Gatherer(registry).Push(); err != nil {
			logger.Error("unable to push metrics: %s", err)
		}
	}

	return nil
}

func (a app) updateRepositories(ctx context.Context, releases []model.Release) error {
	if len(releases) == 0 {
		return nil
	}

	repo := releases[0].Repository

	for _, release := range releases {
		if release.Repository.ID != repo.ID {
			if err := a.repositoryService.Update(ctx, repo); err != nil {
				return fmt.Errorf("unable to update repository `%s`: %s", repo.Name, err)
			}

			repo = release.Repository
		}

		repo.Versions[release.Pattern] = release.Version.Name
	}

	if err := a.repositoryService.Update(ctx, repo); err != nil {
		return fmt.Errorf("unable to update repository `%s`: %s", repo.Name, err)
	}

	return nil
}

func (a app) getKetchupToNotify(ctx context.Context, releases []model.Release) (map[model.User][]model.Release, error) {
	repositories := make([]model.Repository, len(releases))
	for index, release := range releases {
		repositories[index] = release.Repository
	}

	ketchups, err := a.ketchupService.ListForRepositories(ctx, repositories, model.Daily)
	if err != nil {
		return nil, fmt.Errorf("unable to get ketchups for repositories: %w", err)
	}

	userToNotify := syncReleasesByUser(releases, ketchups)
	logger.Info("%d daily ketchups to notify", len(ketchups))

	if a.clock.Now().Weekday() == time.Monday {
		weeklyKetchups, err := a.ketchupService.ListOutdatedByFrequency(ctx, model.Weekly)
		if err != nil {
			return nil, fmt.Errorf("unable to get weekly ketchups: %w", err)
		}

		logger.Info("%d weekly ketchups to notify", len(weeklyKetchups))
		addWeeklyKetchups(weeklyKetchups, userToNotify)
	}

	logger.Info("%d users to notify", len(userToNotify))

	return userToNotify, nil
}

func syncReleasesByUser(releases []model.Release, ketchups []model.Ketchup) map[model.User][]model.Release {
	usersToNotify := make(map[model.User][]model.Release)

	sort.Sort(model.ReleaseByRepositoryIDAndPattern(releases))
	sort.Sort(model.KetchupByRepositoryIDAndPattern(ketchups))

	index := 0
	size := len(ketchups)

	for _, release := range releases {
		for index < size {
			current := ketchups[index]

			if release.Repository.ID < current.Repository.ID || (release.Repository.ID == current.Repository.ID && release.Pattern < current.Pattern) {
				break // release is out of sync, we need to go foward release
			}

			index++

			if release.Repository.ID != current.Repository.ID || release.Pattern != current.Pattern {
				continue // ketchup is not sync with release, we need for go forward ketchup
			}

			if current.Version != release.Version.Name {
				if usersToNotify[current.User] != nil {
					usersToNotify[current.User] = append(usersToNotify[current.User], release)
				} else {
					usersToNotify[current.User] = []model.Release{release}
				}
			}
		}
	}

	return usersToNotify
}

func addWeeklyKetchups(ketchups []model.Ketchup, usersToNotify map[model.User][]model.Release) {
	for _, ketchup := range ketchups {
		ketchupVersion, err := semver.Parse(ketchup.Version)
		if err != nil {
			logger.WithField("version", ketchup.Version).Error("unable to parse version of ketchup: %s", err)
			continue
		}

		release := model.NewRelease(ketchup.Repository, ketchup.Pattern, ketchupVersion)

		if usersToNotify[ketchup.User] != nil {
			usersToNotify[ketchup.User] = append(usersToNotify[ketchup.User], release)
		} else {
			usersToNotify[ketchup.User] = []model.Release{release}
		}
	}
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

		sort.Sort(model.ReleaseByKindAndName(releases))

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
