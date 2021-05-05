package notifier

import (
	"context"
	"flag"
	"fmt"
	"sort"
	"strings"

	authModel "github.com/ViBiOh/auth/v2/pkg/model"
	"github.com/ViBiOh/httputils/v4/pkg/flags"
	"github.com/ViBiOh/httputils/v4/pkg/logger"
	"github.com/ViBiOh/ketchup/pkg/helm"
	"github.com/ViBiOh/ketchup/pkg/model"
	"github.com/ViBiOh/ketchup/pkg/service/ketchup"
	"github.com/ViBiOh/ketchup/pkg/service/repository"
	mailer "github.com/ViBiOh/mailer/pkg/client"
	mailerModel "github.com/ViBiOh/mailer/pkg/model"
	"github.com/prometheus/client_golang/prometheus/push"
)

var (
	pageSize = uint(20)
)

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
	mailerApp         mailer.App
	helmApp           helm.App

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
func New(config Config, repositoryService repository.App, ketchupService ketchup.App, mailerApp mailer.App, helmApp helm.App) App {
	return app{
		loginID: uint64(*config.loginID),
		pushURL: strings.TrimSpace(*config.pushURL),

		repositoryService: repositoryService,
		ketchupService:    ketchupService,
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

	sort.Sort(model.ReleaseByRepositoryID(newReleases))
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

	userToNotify := make(map[model.User][]model.Release)

	sort.Sort(model.ReleaseByRepositoryID(releases))
	sort.Sort(model.KetchupByRepositoryID(ketchups))

	index := 0
	size := len(ketchups)

	for _, release := range releases {
		for index < size {
			current := ketchups[index]

			if release.Repository.ID < current.Repository.ID || (release.Repository.ID == current.Repository.ID && release.Pattern < current.Pattern) {
				break // release is out of sync, we need to advance
			}

			index++
			if release.Repository.ID != current.Repository.ID || release.Pattern != current.Pattern {
				continue // ketchup is not sync with release
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
