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
	dryRun     bool
}

type Config struct {
	DryRun bool
}

func Flags(fs *flag.FlagSet, prefix string) *Config {
	var config Config

	flags.New("DryRun", "Run in dry-run").Prefix(prefix).DocPrefix("notifier").BoolVar(fs, &config.DryRun, false, nil)

	return &config
}

func New(config *Config, repositoryService model.RepositoryService, ketchupService model.KetchupService, userService user.Service, mailerService model.Mailer, helmService model.HelmProvider) Service {
	return Service{
		clock:      time.Now,
		repository: repositoryService,
		ketchup:    ketchupService,
		user:       userService,
		mailer:     mailerService,
		helm:       helmService,
		dryRun:     config.DryRun,
	}
}

func (s Service) Notify(ctx context.Context) error {
	if !s.dryRun {
		if err := s.repository.Clean(ctx); err != nil {
			return fmt.Errorf("clean repository before starting: %w", err)
		}
	}

	newReleases, err := s.getNewReleases(ctx)
	if err != nil {
		return fmt.Errorf("get new releases: %w", err)
	}

	sort.Sort(model.ReleaseByRepositoryIDAndPattern(newReleases))

	if !s.dryRun {
		if err := s.updateRepositories(ctx, newReleases); err != nil {
			return fmt.Errorf("update repositories: %w", err)
		}
	}

	ketchupsToNotify, weeklyUsers, err := s.getKetchupToNotify(ctx, newReleases)
	if err != nil {
		return fmt.Errorf("get ketchup to notify: %w", err)
	}

	if !s.dryRun {
		if err := s.sendNotification(ctx, "ketchup", ketchupsToNotify, weeklyUsers); err != nil {
			return fmt.Errorf("send notification: %w", err)
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

func (s Service) getKetchupToNotify(ctx context.Context, releases []model.Release) (map[model.User][]model.Release, map[model.Identifier]uint8, error) {
	repositories := make([]model.Repository, len(releases))
	for index, release := range releases {
		repositories[index] = release.Repository
	}

	ketchups, err := s.ketchup.ListForRepositories(ctx, repositories, model.Daily, model.None)
	if err != nil {
		return nil, nil, fmt.Errorf("get ketchups for repositories: %w", err)
	}

	slog.LogAttrs(ctx, slog.LevelInfo, "Daily ketchups updates", slog.Int("count", len(ketchups)))

	userToNotify, userStatuses := s.syncReleasesByUser(ctx, releases, ketchups)

	if s.clock().Weekday() == time.Monday {
		weeklyKetchups, err := s.ketchup.ListOutdated(ctx)
		if err != nil {
			return nil, nil, fmt.Errorf("get weekly ketchups: %w", err)
		}

		slog.LogAttrs(ctx, slog.LevelInfo, "Weekly ketchups updates", slog.Int("count", len(weeklyKetchups)))

		s.appendWeeklyKetchupsToUsers(ctx, userToNotify, userStatuses, weeklyKetchups)
	}

	slog.LogAttrs(ctx, slog.LevelInfo, "Users to notify", slog.Int("count", len(userToNotify)))

	return userToNotify, userStatuses, nil
}

func releaseKey(r model.Release) []byte {
	return []byte(fmt.Sprintf("%10d|%s", r.Repository.ID, r.Pattern))
}

func ketchupKey(k model.Ketchup) []byte {
	return []byte(fmt.Sprintf("%10d|%s", k.Repository.ID, k.Pattern))
}

func (s Service) syncReleasesByUser(ctx context.Context, releases []model.Release, ketchups []model.Ketchup) (map[model.User][]model.Release, map[model.Identifier]uint8) {
	usersToNotify := make(map[model.User][]model.Release)
	userStatuses := make(map[model.Identifier]uint8)

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
				s.handleKetchupNotification(ctx, usersToNotify, userStatuses, ketchup, release)
			}
			return nil
		})
	if err != nil {
		slog.LogAttrs(ctx, slog.LevelError, "synchronise releases and ketchups", slog.Any("error", err))
	}

	return usersToNotify, userStatuses
}

func (s Service) appendWeeklyKetchupsToUsers(ctx context.Context, usersToNotify map[model.User][]model.Release, userStatuses map[model.Identifier]uint8, ketchups []model.Ketchup) {
	for _, ketchup := range ketchups {
		ketchupVersion, err := semver.Parse(ketchup.Version, semver.ExtractName(ketchup.Repository.Name))
		if err != nil {
			slog.LogAttrs(ctx, slog.LevelError, "parse version of ketchup", slog.String("version", ketchup.Version), slog.Any("error", err))
			continue
		}

		s.handleKetchupNotification(ctx, usersToNotify, userStatuses, ketchup, model.NewRelease(ketchup.Repository, ketchup.Pattern, ketchupVersion))
	}
}

func (s Service) handleKetchupNotification(ctx context.Context, usersToNotify map[model.User][]model.Release, userStatuses map[model.Identifier]uint8, ketchup model.Ketchup, release model.Release) {
	release = s.handleUpdateWhenNotify(ctx, ketchup, release)

	if ketchup.Frequency == model.None {
		return
	}

	userStatuses[ketchup.User.ID] |= uint8(ketchup.Frequency)

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

	if !s.dryRun {
		log.InfoContext(ctx, "Auto-updating ketchup", "version", release.Version.Name)
		if err := s.ketchup.UpdateVersion(ctx, ketchup.User.ID, ketchup.Repository.ID, ketchup.Pattern, release.Version.Name); err != nil {
			log.LogAttrs(ctx, slog.LevelError, "update ketchup", slog.Any("error", err))
			return release.SetUpdated(1)
		}
	}

	return release.SetUpdated(2)
}

func (s Service) sendNotification(ctx context.Context, template string, ketchupToNotify map[model.User][]model.Release, userStatuses map[model.Identifier]uint8) error {
	if len(ketchupToNotify) == 0 {
		return nil
	}

	if s.mailer == nil || !s.mailer.Enabled() {
		slog.WarnContext(ctx, "mailer is not configured")
		return nil
	}

	for ketchupUser, releases := range ketchupToNotify {
		slog.LogAttrs(ctx, slog.LevelInfo, "Sending email", slog.String("to", ketchupUser.Email), slog.Int("count", len(releases)))

		sort.Sort(model.ReleaseByKindAndName(releases))

		payload := map[string]any{
			"releases": releases,
		}

		var subject string

		switch userStatuses[ketchupUser.ID] {
		case uint8(model.Daily | model.Weekly):
			subject = "Ketchup - Daily & Weekly notification"
		case uint8(model.Weekly):
			subject = "Ketchup - Weekly notification"
		default:
			subject = "Ketchup - Daily notification"
		}

		mr := mailerModel.NewMailRequest().
			Template(template).
			From("ketchup@vibioh.fr").
			As("Ketchup").
			To(ketchupUser.Email).
			Data(payload).
			WithSubject(subject)

		if err := s.mailer.Send(ctx, mr); err != nil {
			return fmt.Errorf("send email to %s: %w", ketchupUser.Email, err)
		}
	}

	return nil
}
