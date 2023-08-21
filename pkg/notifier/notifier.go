package notifier

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"sort"
	"strings"
	"time"

	"github.com/ViBiOh/flags"
	"github.com/ViBiOh/httputils/v4/pkg/breaksync"
	"github.com/ViBiOh/ketchup/pkg/model"
	"github.com/ViBiOh/ketchup/pkg/semver"
	"github.com/ViBiOh/ketchup/pkg/service/user"
	mailerModel "github.com/ViBiOh/mailer/pkg/model"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

type GetNow func() time.Time

var pageSize = uint(20)

type App struct {
	repositoryService model.RepositoryService
	ketchupService    model.KetchupService
	userService       user.App
	mailerApp         model.Mailer
	helmApp           model.HelmProvider
	clock             GetNow
	pushURL           string
}

type Config struct {
	pushURL *string
}

func Flags(fs *flag.FlagSet, prefix string) Config {
	return Config{
		pushURL: flags.New("PushUrl", "Pushgateway URL").Prefix(prefix).DocPrefix("notifier").String(fs, "", nil),
	}
}

func New(config Config, repositoryService model.RepositoryService, ketchupService model.KetchupService, userService user.App, mailerApp model.Mailer, helmApp model.HelmProvider) App {
	return App{
		pushURL:           strings.TrimSpace(*config.pushURL),
		clock:             time.Now,
		repositoryService: repositoryService,
		ketchupService:    ketchupService,
		userService:       userService,
		mailerApp:         mailerApp,
		helmApp:           helmApp,
	}
}

func (a App) Notify(ctx context.Context, meterProvider metric.MeterProvider) error {
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
		return fmt.Errorf("send notification: %w", err)
	}

	if meterProvider != nil {
		meter := meterProvider.Meter("github.com/ViBiOh/ketchup/pkg/notifier")

		_, err := meter.Int64ObservableGauge("ketchup_metrics", metric.WithInt64Callback(func(ctx context.Context, io metric.Int64Observer) error {
			userCount, err := a.userService.Count(ctx)
			if err != nil {
				return fmt.Errorf("get users count: %w", err)
			}

			io.Observe(int64(userCount), metric.WithAttributes(attribute.String("type", "users")))
			io.Observe(int64(repoCount), metric.WithAttributes(attribute.String("type", "repositories")))
			io.Observe(int64(len(newReleases)), metric.WithAttributes(attribute.String("type", "releases")))
			io.Observe(int64(len(ketchupsToNotify)), metric.WithAttributes(attribute.String("type", "notifications")))

			return nil
		}))
		if err != nil {
			slog.Error("create ketchup counter", "err", err)
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
				return fmt.Errorf("update repository `%s`: %w", repo.Name, err)
			}

			repo = release.Repository
		}

		repo.Versions[release.Pattern] = release.Version.Name
	}

	if err := a.repositoryService.Update(ctx, repo); err != nil {
		return fmt.Errorf("update repository `%s`: %w", repo.Name, err)
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

	slog.Info("Daily ketchups updates", "count", len(ketchups))

	userToNotify := a.syncReleasesByUser(ctx, releases, ketchups)

	if a.clock().Weekday() == time.Monday {
		weeklyKetchups, err := a.ketchupService.ListOutdated(ctx)
		if err != nil {
			return nil, fmt.Errorf("get weekly ketchups: %w", err)
		}

		slog.Info("Weekly ketchups updates", "count", len(weeklyKetchups))
		a.appendKetchupsToUser(ctx, userToNotify, weeklyKetchups)
	}

	slog.Info("Users to notify", "count", len(userToNotify))

	return userToNotify, nil
}

func releaseKey(r model.Release) []byte {
	return []byte(fmt.Sprintf("%10d|%s", r.Repository.ID, r.Pattern))
}

func ketchupKey(k model.Ketchup) []byte {
	return []byte(fmt.Sprintf("%10d|%s", k.Repository.ID, k.Pattern))
}

func (a App) syncReleasesByUser(ctx context.Context, releases []model.Release, ketchups []model.Ketchup) map[model.User][]model.Release {
	usersToNotify := make(map[model.User][]model.Release)

	sort.Sort(model.ReleaseByRepositoryIDAndPattern(releases))
	sort.Sort(model.KetchupByRepositoryIDAndPattern(ketchups))

	err := breaksync.NewSynchronization().
		AddSources(
			breaksync.NewSliceSource(releases, releaseKey, breaksync.NewRupture("release", breaksync.RuptureIdentity)),
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
		slog.Error("synchronise releases and ketchups", "err", err)
	}

	return usersToNotify
}

func (a App) appendKetchupsToUser(ctx context.Context, usersToNotify map[model.User][]model.Release, ketchups []model.Ketchup) {
	for _, ketchup := range ketchups {
		ketchupVersion, err := semver.Parse(ketchup.Version)
		if err != nil {
			slog.Error("parse version of ketchup", "err", err, "version", ketchup.Version)
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
		for _, userRelease := range usersToNotify[ketchup.User] {
			if userRelease.Repository.ID == release.Repository.ID {
				return
			}
		}

		usersToNotify[ketchup.User] = append(usersToNotify[ketchup.User], release)
	} else {
		usersToNotify[ketchup.User] = []model.Release{release}
	}
}

func (a App) handleUpdateWhenNotify(ctx context.Context, ketchup model.Ketchup, release model.Release) model.Release {
	if !ketchup.UpdateWhenNotify {
		return release
	}

	log := slog.With("repository", ketchup.Repository.ID).With("user", ketchup.User.ID).With("pattern", ketchup.Pattern)

	log.Info("Auto-updating ketchup", "version", release.Version.Name)
	if err := a.ketchupService.UpdateVersion(ctx, ketchup.User.ID, ketchup.Repository.ID, ketchup.Pattern, release.Version.Name); err != nil {
		log.Error("update ketchup", "err", err)
		return release.SetUpdated(1)
	}

	return release.SetUpdated(2)
}

func (a App) sendNotification(ctx context.Context, template string, ketchupToNotify map[model.User][]model.Release) error {
	if len(ketchupToNotify) == 0 {
		return nil
	}

	if a.mailerApp == nil || !a.mailerApp.Enabled() {
		slog.Warn("mailer is not configured")
		return nil
	}

	for ketchupUser, releases := range ketchupToNotify {
		slog.Info("Sending email", "to", ketchupUser.Email, "count", len(releases))

		sort.Sort(model.ReleaseByKindAndName(releases))

		payload := map[string]any{
			"releases": releases,
		}

		mr := mailerModel.NewMailRequest().Template(template).From("ketchup@vibioh.fr").As("Ketchup").To(ketchupUser.Email).Data(payload)
		subject := fmt.Sprintf("Ketchup - %d new release", len(releases))
		if len(releases) > 1 {
			subject += "s"
		}
		mr = mr.WithSubject(subject)

		if err := a.mailerApp.Send(ctx, mr); err != nil {
			return fmt.Errorf("send email to %s: %w", ketchupUser.Email, err)
		}
	}

	return nil
}

func (a App) Remind(ctx context.Context) error {
	usersToRemind, err := a.userService.ListReminderUsers(ctx)
	if err != nil {
		return fmt.Errorf("get reminder users: %w", err)
	}

	remindKetchups, err := a.ketchupService.ListOutdated(ctx, usersToRemind...)
	if err != nil {
		return fmt.Errorf("get daily ketchups to remind: %w", err)
	}

	if len(remindKetchups) == 0 {
		return nil
	}

	usersToNotify := make(map[model.User][]model.Release)
	a.appendKetchupsToUser(ctx, usersToNotify, remindKetchups)

	if err := a.sendNotification(ctx, "ketchup_remind", usersToNotify); err != nil {
		return fmt.Errorf("send remind notification: %w", err)
	}

	return nil
}
