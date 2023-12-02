package notifier

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"sort"
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

type Service struct {
	repository model.RepositoryService
	ketchup    model.KetchupService
	user       user.Service
	mailer     model.Mailer
	helm       model.HelmProvider
	clock      GetNow
	pushURL    string
}

type Config struct {
	PushURL string
}

func Flags(fs *flag.FlagSet, prefix string) *Config {
	var config Config

	flags.New("PushUrl", "Pushgateway URL").Prefix(prefix).DocPrefix("notifier").StringVar(fs, &config.PushURL, "", nil)

	return &config
}

func New(config *Config, repositoryService model.RepositoryService, ketchupService model.KetchupService, userService user.Service, mailerService model.Mailer, helmService model.HelmProvider) Service {
	return Service{
		pushURL:    config.PushURL,
		clock:      time.Now,
		repository: repositoryService,
		ketchup:    ketchupService,
		user:       userService,
		mailer:     mailerService,
		helm:       helmService,
	}
}

func (s Service) Notify(ctx context.Context, meterProvider metric.MeterProvider) error {
	if err := s.repository.Clean(ctx); err != nil {
		return fmt.Errorf("clean repository before starting: %w", err)
	}

	newReleases, repoCount, err := s.getNewReleases(ctx)
	if err != nil {
		return fmt.Errorf("get new releases: %w", err)
	}

	sort.Sort(model.ReleaseByRepositoryIDAndPattern(newReleases))
	if err := s.updateRepositories(ctx, newReleases); err != nil {
		return fmt.Errorf("update repositories: %w", err)
	}

	ketchupsToNotify, err := s.getKetchupToNotify(ctx, newReleases)
	if err != nil {
		return fmt.Errorf("get ketchup to notify: %w", err)
	}

	if err := s.sendNotification(ctx, "ketchup", ketchupsToNotify); err != nil {
		return fmt.Errorf("send notification: %w", err)
	}

	if meterProvider != nil {
		meter := meterProvider.Meter("github.com/ViBiOh/ketchup/pkg/notifier")

		_, err := meter.Int64ObservableGauge("ketchup.metrics", metric.WithInt64Callback(func(ctx context.Context, io metric.Int64Observer) error {
			userCount, err := s.user.Count(ctx)
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
			slog.ErrorContext(ctx, "create ketchup counter", "err", err)
		}
	}

	return nil
}

func (s Service) updateRepositories(ctx context.Context, releases []model.Release) error {
	if len(releases) == 0 {
		return nil
	}

	repo := releases[0].Repository

	for _, release := range releases {
		if release.Repository.ID != repo.ID {
			if err := s.repository.Update(ctx, repo); err != nil {
				return fmt.Errorf("update repository `%s`: %w", repo.Name, err)
			}

			repo = release.Repository
		}

		repo.Versions[release.Pattern] = release.Version.Name
	}

	if err := s.repository.Update(ctx, repo); err != nil {
		return fmt.Errorf("update repository `%s`: %w", repo.Name, err)
	}

	return nil
}

func (s Service) getKetchupToNotify(ctx context.Context, releases []model.Release) (map[model.User][]model.Release, error) {
	repositories := make([]model.Repository, len(releases))
	for index, release := range releases {
		repositories[index] = release.Repository
	}

	ketchups, err := s.ketchup.ListForRepositories(ctx, repositories, model.Daily, model.None)
	if err != nil {
		return nil, fmt.Errorf("get ketchups for repositories: %w", err)
	}

	slog.InfoContext(ctx, "Daily ketchups updates", "count", len(ketchups))

	userToNotify := s.syncReleasesByUser(ctx, releases, ketchups)

	if s.clock().Weekday() == time.Monday {
		weeklyKetchups, err := s.ketchup.ListOutdated(ctx)
		if err != nil {
			return nil, fmt.Errorf("get weekly ketchups: %w", err)
		}

		slog.InfoContext(ctx, "Weekly ketchups updates", "count", len(weeklyKetchups))
		s.appendKetchupsToUser(ctx, userToNotify, weeklyKetchups)
	}

	slog.InfoContext(ctx, "Users to notify", "count", len(userToNotify))

	return userToNotify, nil
}

func releaseKey(r model.Release) []byte {
	return []byte(fmt.Sprintf("%10d|%s", r.Repository.ID, r.Pattern))
}

func ketchupKey(k model.Ketchup) []byte {
	return []byte(fmt.Sprintf("%10d|%s", k.Repository.ID, k.Pattern))
}

func (s Service) syncReleasesByUser(ctx context.Context, releases []model.Release, ketchups []model.Ketchup) map[model.User][]model.Release {
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
				s.handleKetchupNotification(ctx, usersToNotify, ketchup, release)
			}
			return nil
		})
	if err != nil {
		slog.ErrorContext(ctx, "synchronise releases and ketchups", "err", err)
	}

	return usersToNotify
}

func (s Service) appendKetchupsToUser(ctx context.Context, usersToNotify map[model.User][]model.Release, ketchups []model.Ketchup) {
	for _, ketchup := range ketchups {
		ketchupVersion, err := semver.Parse(ketchup.Version)
		if err != nil {
			slog.ErrorContext(ctx, "parse version of ketchup", "err", err, "version", ketchup.Version)
			continue
		}

		s.handleKetchupNotification(ctx, usersToNotify, ketchup, model.NewRelease(ketchup.Repository, ketchup.Pattern, ketchupVersion))
	}
}

func (s Service) handleKetchupNotification(ctx context.Context, usersToNotify map[model.User][]model.Release, ketchup model.Ketchup, release model.Release) {
	release = s.handleUpdateWhenNotify(ctx, ketchup, release)

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

func (s Service) handleUpdateWhenNotify(ctx context.Context, ketchup model.Ketchup, release model.Release) model.Release {
	if !ketchup.UpdateWhenNotify {
		return release
	}

	log := slog.With("repository", ketchup.Repository.ID).With("user", ketchup.User.ID).With("pattern", ketchup.Pattern)

	log.InfoContext(ctx, "Auto-updating ketchup", "version", release.Version.Name)
	if err := s.ketchup.UpdateVersion(ctx, ketchup.User.ID, ketchup.Repository.ID, ketchup.Pattern, release.Version.Name); err != nil {
		log.ErrorContext(ctx, "update ketchup", "err", err)
		return release.SetUpdated(1)
	}

	return release.SetUpdated(2)
}

func (s Service) sendNotification(ctx context.Context, template string, ketchupToNotify map[model.User][]model.Release) error {
	if len(ketchupToNotify) == 0 {
		return nil
	}

	if s.mailer == nil || !s.mailer.Enabled() {
		slog.WarnContext(ctx, "mailer is not configured")
		return nil
	}

	for ketchupUser, releases := range ketchupToNotify {
		slog.InfoContext(ctx, "Sending email", "to", ketchupUser.Email, "count", len(releases))

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

		if err := s.mailer.Send(ctx, mr); err != nil {
			return fmt.Errorf("send email to %s: %w", ketchupUser.Email, err)
		}
	}

	return nil
}

func (s Service) Remind(ctx context.Context) error {
	usersToRemind, err := s.user.ListReminderUsers(ctx)
	if err != nil {
		return fmt.Errorf("get reminder users: %w", err)
	}

	remindKetchups, err := s.ketchup.ListOutdated(ctx, usersToRemind...)
	if err != nil {
		return fmt.Errorf("get daily ketchups to remind: %w", err)
	}

	if len(remindKetchups) == 0 {
		return nil
	}

	usersToNotify := make(map[model.User][]model.Release)
	s.appendKetchupsToUser(ctx, usersToNotify, remindKetchups)

	if err := s.sendNotification(ctx, "ketchup_remind", usersToNotify); err != nil {
		return fmt.Errorf("send remind notification: %w", err)
	}

	return nil
}
