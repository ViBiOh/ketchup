package main

import (
	"context"
	"embed"
	"fmt"

	authIdent "github.com/ViBiOh/auth/v2/pkg/ident/basic"
	authMiddleware "github.com/ViBiOh/auth/v2/pkg/middleware"
	authService "github.com/ViBiOh/auth/v2/pkg/service"
	authStore "github.com/ViBiOh/auth/v2/pkg/store/db"
	"github.com/ViBiOh/httputils/v4/pkg/cors"
	"github.com/ViBiOh/httputils/v4/pkg/owasp"
	"github.com/ViBiOh/httputils/v4/pkg/renderer"
	"github.com/ViBiOh/httputils/v4/pkg/server"
	"github.com/ViBiOh/ketchup/pkg/ketchup"
	"github.com/ViBiOh/ketchup/pkg/provider/docker"
	"github.com/ViBiOh/ketchup/pkg/provider/github"
	"github.com/ViBiOh/ketchup/pkg/provider/helm"
	"github.com/ViBiOh/ketchup/pkg/provider/npm"
	"github.com/ViBiOh/ketchup/pkg/provider/pypi"
	ketchupService "github.com/ViBiOh/ketchup/pkg/service/ketchup"
	repositoryService "github.com/ViBiOh/ketchup/pkg/service/repository"
	"github.com/ViBiOh/ketchup/pkg/service/user"
	userService "github.com/ViBiOh/ketchup/pkg/service/user"
	ketchupStore "github.com/ViBiOh/ketchup/pkg/store/ketchup"
	repositoryStore "github.com/ViBiOh/ketchup/pkg/store/repository"
	userStore "github.com/ViBiOh/ketchup/pkg/store/user"
)

//go:embed templates static
var content embed.FS

type services struct {
	server   *server.Server
	owasp    owasp.Service
	cors     cors.Service
	renderer *renderer.Service

	ketchup        ketchup.Service
	user           user.Service
	authMiddleware authMiddleware.Service
}

func newServices(ctx context.Context, config configuration, clients clients) (services, error) {
	var output services
	var err error

	output.server = server.New(config.server)
	output.owasp = owasp.New(config.owasp)
	output.cors = cors.New(config.cors)

	authProvider := authStore.New(clients.db)
	identProvider := authIdent.New(authProvider, "ketchup")

	authService := authService.New(authProvider, authProvider)
	output.authMiddleware = authMiddleware.New(authProvider, clients.telemetry.TracerProvider(), identProvider)

	githubService := github.New(config.github, clients.redis, clients.telemetry.MeterProvider(), clients.telemetry.TracerProvider())
	helmService := helm.New()
	dockerService := docker.New(config.docker)
	npmService := npm.New()
	pypiService := pypi.New()

	repositoryService := repositoryService.New(repositoryStore.New(clients.db), githubService, helmService, dockerService, npmService, pypiService)

	ketchupService := ketchupService.New(ketchupStore.New(clients.db), repositoryService)

	output.renderer, err = renderer.New(ctx, config.renderer, content, ketchup.FuncMap, clients.telemetry.MeterProvider(), clients.telemetry.TracerProvider())
	if err != nil {
		return output, fmt.Errorf("renderer: %w", err)
	}

	output.user = userService.New(userStore.New(clients.db), &authService)
	output.ketchup = ketchup.New(ctx, output.renderer, ketchupService, output.user, repositoryService, clients.redis, clients.telemetry.TracerProvider())

	return output, nil
}
