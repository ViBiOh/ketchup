package notifier

import (
	"context"
	"flag"
	"fmt"
	"sort"
	"strings"
	"time"

	authModel "github.com/ViBiOh/auth/v2/pkg/model"
	"github.com/ViBiOh/httputils/v4/pkg/clock"
	"github.com/ViBiOh/httputils/v4/pkg/flags"
	"github.com/ViBiOh/httputils/v4/pkg/logger"
	"github.com/ViBiOh/ketchup/pkg/model"
	"github.com/ViBiOh/ketchup/pkg/semver"
	"github.com/ViBiOh/ketchup/pkg/service/user"
	mailerModel "github.com/ViBiOh/mailer/pkg/model"
	"github.com/prometheus/client_golang/prometheus/push"
)

var (
	pageSize = uint(20)
)

// App of package
type App struct {
	repositoryService model.RepositoryService
	ketchupService    model.KetchupService
	userService       user.App
	mailerApp         model.Mailer
	helmApp           model.HelmProvider

	clock *clock.Clock

	pushURL string
	loginID uint64
}

// Config of package
type Config struct {
	loginID *uint
	pushURL *string
}

// Flags adds flags for configuring package
func Flags(fs *flag.FlagSet, prefix string) Config {
	return Config{
		loginID: flags.New(prefix, "notifier", "LoginID").Default(1, nil).Label("Scheduler user ID").ToUint(fs),
		pushURL: flags.New(prefix, "notifier", "PushUrl").Default("", nil).Label("Pushgateway URL").ToString(fs),
	}
}

// New creates new App from Config
func New(config Config, repositoryService model.RepositoryService, ketchupService model.KetchupService, userService user.App, mailerApp model.Mailer, helmApp model.HelmProvider) App {
	return App{
		loginID: uint64(*config.loginID),
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

	if err := a.sendNotification(userCtx, "ketchup", ketchupsToNotify); err != nil {
		return fmt.Errorf("unable to send notification: %s", err)
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

func (a App) updateRepositories(ctx context.Context, releases []model.Release) error {
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

func (a App) getKetchupToNotify(ctx context.Context, releases []model.Release) (map[model.User][]model.Release, error) {
	repositories := make([]model.Repository, len(releases))
	for index, release := range releases {
		repositories[index] = release.Repository
	}

	ketchups, err := a.ketchupService.ListForRepositories(ctx, repositories, model.Daily)
	if err != nil {
		return nil, fmt.Errorf("unable to get ketchups for repositories: %w", err)
	}

	userToNotify := a.syncReleasesByUser(releases, ketchups)
	logger.Info("%d daily ketchups to notify", len(ketchups))

	if a.clock.Now().Weekday() == time.Monday {
		weeklyKetchups, err := a.ketchupService.ListOutdatedByFrequency(ctx, model.Weekly)
		if err != nil {
			return nil, fmt.Errorf("unable to get weekly ketchups: %w", err)
		}

		logger.Info("%d weekly ketchups to notify", len(weeklyKetchups))
		a.groupKetchupsToUsers(weeklyKetchups, userToNotify)
	}

	logger.Info("%d users to notify", len(userToNotify))

	return userToNotify, nil
}

func (a App) syncReleasesByUser(releases []model.Release, ketchups []model.Ketchup) map[model.User][]model.Release {
	usersToNotify := make(map[model.User][]model.Release)

	sort.Sort(model.ReleaseByRepositoryIDAndPattern(releases))
	sort.Sort(model.KetchupByRepositoryIDAndPattern(ketchups))

	index := 0
	size := len(ketchups)

	for _, release := range releases {
		for index < size {
			current := ketchups[index]

			if release.Repository.ID < current.Repository.ID || (release.Repository.ID == current.Repository.ID && release.Pattern < current.Pattern) {
				break // release is out of sync, we need to go forward release
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

				a.handleKetchupNotification(current, release.Version.Name)
			}
		}
	}

	return usersToNotify
}

func (a App) groupKetchupsToUsers(ketchups []model.Ketchup, usersToNotify map[model.User][]model.Release) {
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

		a.handleKetchupNotification(ketchup, release.Version.Name)
	}
}

func (a App) handleKetchupNotification(ketchup model.Ketchup, version string) {
	if !ketchup.UpdateWhenNotify {
		return
	}

	log := logger.WithField("repository", ketchup.Repository.ID).WithField("user", ketchup.User.ID).WithField("pattern", ketchup.Pattern)

	log.Info("Auto-updating ketchup to %s", version)
	if err := a.ketchupService.UpdateVersion(context.Background(), ketchup.User.ID, ketchup.Repository.ID, ketchup.Pattern, version); err != nil {
		logger.Error("unable to update ketchup user=%d repository=%d: %s", ketchup.User.ID, ketchup.Repository.ID, err)
	}
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

		payload := map[string]interface{}{
			"releases": releases,
		}

		mr := mailerModel.NewMailRequest().Template(template).From("ketchup@vibioh.fr").As("Ketchup").To(user.Email).Data(payload)
		subject := fmt.Sprintf("Ketchup - %d new release", len(releases))
		if len(releases) > 1 {
			subject += "s"
		}
		mr = mr.WithSubject(subject)

		if err := a.mailerApp.Send(ctx, mr); err != nil {
			return fmt.Errorf("unable to send email to %s: %s", user.Email, err)
		}
	}

	return nil
}

// Remind users for new ketchup
func (a App) Remind(ctx context.Context) error {
	userCtx := authModel.StoreUser(ctx, authModel.NewUser(a.loginID, "scheduler"))

	usersToRemind, err := a.userService.ListReminderUsers(userCtx)
	if err != nil {
		return fmt.Errorf("unable to get reminder users: %s", err)
	}

	remindKetchups, err := a.ketchupService.ListOutdatedByFrequency(userCtx, model.Daily, usersToRemind...)
	if err != nil {
		return fmt.Errorf("unable to get daily ketchups to remind: %s", err)
	}

	if len(remindKetchups) == 0 {
		return nil
	}

	usersToNotify := make(map[model.User][]model.Release)
	a.groupKetchupsToUsers(remindKetchups, usersToNotify)

	if err := a.sendNotification(userCtx, "ketchup_remind", usersToNotify); err != nil {
		return fmt.Errorf("unable to send remind notification: %s", err)
	}

	return nil
}
