package main

import (
	"context"
	"fmt"

	"github.com/ViBiOh/ketchup/pkg/notifier"
	"github.com/ViBiOh/ketchup/pkg/provider/docker"
	"github.com/ViBiOh/ketchup/pkg/provider/github"
	"github.com/ViBiOh/ketchup/pkg/provider/helm"
	"github.com/ViBiOh/ketchup/pkg/provider/npm"
	"github.com/ViBiOh/ketchup/pkg/provider/pypi"
	ketchupService "github.com/ViBiOh/ketchup/pkg/service/ketchup"
	repositoryService "github.com/ViBiOh/ketchup/pkg/service/repository"
	userService "github.com/ViBiOh/ketchup/pkg/service/user"
	ketchupStore "github.com/ViBiOh/ketchup/pkg/store/ketchup"
	repositoryStore "github.com/ViBiOh/ketchup/pkg/store/repository"
	userStore "github.com/ViBiOh/ketchup/pkg/store/user"
	mailer "github.com/ViBiOh/mailer/pkg/client"
)

type services struct {
	notifier notifier.Service
	mailer   mailer.Service
}

func newServices(ctx context.Context, config configuration, clients clients) (services, error) {
	var output services
	var err error

	githubService := github.New(config.github, nil, clients.telemetry.MeterProvider(), clients.telemetry.TracerProvider())
	helmService := helm.New()
	dockerService := docker.New(config.docker)
	npmService := npm.New()
	pypiService := pypi.New()

	repositoryService := repositoryService.New(repositoryStore.New(clients.db), githubService, helmService, dockerService, npmService, pypiService)
	ketchupService := ketchupService.New(ketchupStore.New(clients.db), repositoryService)
	userService := userService.New(userStore.New(clients.db), nil)

	output.mailer, err = mailer.New(ctx, config.mailer, clients.telemetry.MeterProvider(), clients.telemetry.TracerProvider())
	if err != nil {
		return output, fmt.Errorf("mailer: %w", err)
	}

	output.notifier = notifier.New(config.notifier, repositoryService, ketchupService, userService, output.mailer, helmService)

	return output, nil
}

func (c services) Close(ctx context.Context) {
	c.mailer.Close(ctx)
}
