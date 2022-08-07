package notifier

import (
	"context"
	"flag"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/ViBiOh/flags"
	"github.com/ViBiOh/httputils/v4/pkg/breaksync"
	"github.com/ViBiOh/httputils/v4/pkg/clock"
	"github.com/ViBiOh/httputils/v4/pkg/logger"
	"github.com/ViBiOh/ketchup/pkg/model"
	"github.com/ViBiOh/ketchup/pkg/semver"
	"github.com/ViBiOh/ketchup/pkg/service/user"
	mailerModel "github.com/ViBiOh/mailer/pkg/model"
	"github.com/prometheus/client_golang/prometheus/push"
)

var pageSize = uint(20)

// App of package
type App struct {
	repositoryService model.RepositoryService
	ketchupService    model.KetchupService
	userService       user.App
	mailerApp         model.Mailer
	helmApp           model.HelmProvider

	clock clock.Clock

	pushURL string
}

// Config of package
type Config struct {
	pushURL *string
}

// Flags adds flags for configuring package
func Flags(fs *flag.FlagSet, prefix string) Config {
	return Config{
		pushURL: flags.String(fs, prefix, "notifier", "PushUrl", "Pushgateway URL", "", nil),
	}
}

// New creates new App from Config
func New(config Config, repositoryService model.RepositoryService, ketchupService model.KetchupService, userService user.App, mailerApp model.Mailer, helmApp model.HelmProvider) App {
	return App{
		pushURL: strings.TrimSpace(*config.pushURL),

		repositoryService: repositoryService,
		ketchupService:    ketchupService,
		userService:       userService,
		mailerApp:         mailerApp,
		helmApp:           helmApp,
	}
}

// Notify users for new ketchup
func (a App) Notify(ctx context.Context) error {
	if err := a.repositoryService.Clean(ctx); err != nil {
		return fmt.Errorf("clean repository before starting: %w", err)
	}

	newReleases, repoCount, err := a.getNewReleases(ctx)
	if err != nil {
		return fmt.Errorf("get new releases: %w", err)
	}

	sort.Sort(model.ReleaseByRepositoryIDAndPattern(newReleases))
	if err := a.updateRepositories(ctx, newReleases); err != nil {
		return fmt.Errorf("update repositories: %w", err)
	}

	ketchupsToNotify, err := a.getKetchupToNotify(ctx, newReleases)
	if err != nil {
		return fmt.Errorf("get ketchup to notify: %w", err)
	}

	if err := a.sendNotification(ctx, "ketchup", ketchupsToNotify); err != nil {
		return fmt.Errorf("send notification: %s", err)
	}

	if len(a.pushURL) != 0 {
		registry, metrics := configurePrometheus()

		userCount, err := a.userService.Count(ctx)
		if err != nil {
			logger.Error("get users count: %s", err)
		} else {
			metrics.WithLabelValues("users").Set(float64(userCount))
		}

		metrics.WithLabelValues("repositories").Set(float64(repoCount))
		metrics.WithLabelValues("releases").Set(float64(len(newReleases)))
		metrics.WithLabelValues("notifications").Set(float64(len(ketchupsToNotify)))

		if err := push.New(a.pushURL, "ketchup").Gatherer(registry).Push(); err != nil {
			logger.Error("push metrics: %s", err)
		}
	}

	return nil
}

func (a App) updateRepositories(ctx context.Context, releases []model.Release) error {
	if len(releases) == 0 {
		return nil
	}

	repo := releases[0].Repository

	for _, release := range releases {
		if release.Repository.ID != repo.ID {
			if err := a.repositoryService.Update(ctx, repo); err != nil {
				return fmt.Errorf("update repository `%s`: %s", repo.Name, err)
			}

			repo = release.Repository
		}

		repo.Versions[release.Pattern] = release.Version.Name
	}

	if err := a.repositoryService.Update(ctx, repo); err != nil {
		return fmt.Errorf("update repository `%s`: %s", repo.Name, err)
	}

	return nil
}

func (a App) getKetchupToNotify(ctx context.Context, releases []model.Release) (map[model.User][]model.Release, error) {
	repositories := make([]model.Repository, len(releases))
	for index, release := range releases {
		repositories[index] = release.Repository
	}

	ketchups, err := a.ketchupService.ListForRepositories(ctx, repositories, model.Daily, model.None)
	if err != nil {
		return nil, fmt.Errorf("get ketchups for repositories: %w", err)
	}

	logger.Info("%d daily ketchups updates", len(ketchups))

	userToNotify := a.syncReleasesByUser(ctx, releases, ketchups)

	if a.clock.Now().Weekday() == time.Monday {
		weeklyKetchups, err := a.ketchupService.ListOutdated(ctx)
		if err != nil {
			return nil, fmt.Errorf("get weekly ketchups: %w", err)
		}

		logger.Info("%d weekly ketchups updates", len(weeklyKetchups))
		a.appendKetchupsToUser(ctx, userToNotify, weeklyKetchups)
	}

	logger.Info("%d users to notify", len(userToNotify))

	return userToNotify, nil
}

func releaseKey(r model.Release) string {
	return fmt.Sprintf("%10d|%s", r.Repository.ID, r.Pattern)
}

func ketchupKey(k model.Ketchup) string {
	return fmt.Sprintf("%10d|%s", k.Repository.ID, k.Pattern)
}

func (a App) syncReleasesByUser(ctx context.Context, releases []model.Release, ketchups []model.Ketchup) map[model.User][]model.Release {
	usersToNotify := make(map[model.User][]model.Release)

	sort.Sort(model.ReleaseByRepositoryIDAndPattern(releases))
	sort.Sort(model.KetchupByRepositoryIDAndPattern(ketchups))

	err := breaksync.NewSynchronization().
		AddSources(
			breaksync.NewSliceSource(releases, releaseKey, breaksync.NewRupture("release", breaksync.Identity)),
			breaksync.NewSliceSource(ketchups, ketchupKey, nil),
		).
		Run(func(synchronised uint64, items []any) error {
			if synchronised != 0 {
				return nil
			}

			release := items[0].(model.Release)
			ketchup := items[1].(model.Ketchup)

			if ketchup.Version != release.Version.Name {
				a.handleKetchupNotification(ctx, usersToNotify, ketchup, release)
			}
			return nil
		})
	if err != nil {
		logger.Error("synchronise releases and ketchups: %s", err)
	}

	return usersToNotify
}

func (a App) appendKetchupsToUser(ctx context.Context, usersToNotify map[model.User][]model.Release, ketchups []model.Ketchup) {
	for _, ketchup := range ketchups {
		ketchupVersion, err := semver.Parse(ketchup.Version)
		if err != nil {
			logger.WithField("version", ketchup.Version).Error("parse version of ketchup: %s", err)
			continue
		}

		a.handleKetchupNotification(ctx, usersToNotify, ketchup, model.NewRelease(ketchup.Repository, ketchup.Pattern, ketchupVersion))
	}
}

func (a App) handleKetchupNotification(ctx context.Context, usersToNotify map[model.User][]model.Release, ketchup model.Ketchup, release model.Release) {
	release = a.handleUpdateWhenNotify(ctx, ketchup, release)

	if ketchup.Frequency == model.None {
		return
	}

	if usersToNotify[ketchup.User] != nil {
		usersToNotify[ketchup.User] = append(usersToNotify[ketchup.User], release)
	} else {
		usersToNotify[ketchup.User] = []model.Release{release}
	}
}

func (a App) handleUpdateWhenNotify(ctx context.Context, ketchup model.Ketchup, release model.Release) model.Release {
	if !ketchup.UpdateWhenNotify {
		return release
	}

	log := logger.WithField("repository", ketchup.Repository.ID).WithField("user", ketchup.User.ID).WithField("pattern", ketchup.Pattern)

	log.Info("Auto-updating ketchup to %s", release.Version.Name)
	if err := a.ketchupService.UpdateVersion(ctx, ketchup.User.ID, ketchup.Repository.ID, ketchup.Pattern, release.Version.Name); err != nil {
		log.Error("update ketchup: %s", err)
		return release.SetUpdated(1)
	}

	return release.SetUpdated(2)
}

func (a App) sendNotification(ctx context.Context, template string, ketchupToNotify map[model.User][]model.Release) error {
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

		payload := map[string]any{
			"releases": releases,
		}

		mr := mailerModel.NewMailRequest().Template(template).From("ketchup@vibioh.fr").As("Ketchup").To(user.Email).Data(payload)
		subject := fmt.Sprintf("Ketchup - %d new release", len(releases))
		if len(releases) > 1 {
			subject += "s"
		}
		mr = mr.WithSubject(subject)

		if err := a.mailerApp.Send(ctx, mr); err != nil {
			return fmt.Errorf("send email to %s: %s", user.Email, err)
		}
	}

	return nil
}

// Remind users for new ketchup
func (a App) Remind(ctx context.Context) error {
	usersToRemind, err := a.userService.ListReminderUsers(ctx)
	if err != nil {
		return fmt.Errorf("get reminder users: %s", err)
	}

	remindKetchups, err := a.ketchupService.ListOutdated(ctx, usersToRemind...)
	if err != nil {
		return fmt.Errorf("get daily ketchups to remind: %s", err)
	}

	if len(remindKetchups) == 0 {
		return nil
	}

	usersToNotify := make(map[model.User][]model.Release)
	a.appendKetchupsToUser(ctx, usersToNotify, remindKetchups)

	if err := a.sendNotification(ctx, "ketchup_remind", usersToNotify); err != nil {
		return fmt.Errorf("send remind notification: %s", err)
	}

	return nil
}
