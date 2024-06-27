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
	"github.com/ViBiOh/httputils/v4/pkg/db"
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
	"go.opentelemetry.io/otel/trace"
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
	authService, authMiddlewareService := initAuth(clients.db, clients.telemetry.TracerProvider())

	userService := userService.New(userStore.New(clients.db), &authService)

	githubService := github.New(config.github, clients.redis, clients.telemetry.MeterProvider(), clients.telemetry.TracerProvider())
	dockerService := docker.New(config.docker)
	helmService := helm.New()
	npmService := npm.New()
	pypiService := pypi.New()

	repositoryService := repositoryService.New(repositoryStore.New(clients.db), githubService, helmService, dockerService, npmService, pypiService)

	ketchupService := ketchupService.New(ketchupStore.New(clients.db), repositoryService)

	rendererService, err := renderer.New(ctx, config.renderer, content, ketchup.FuncMap, clients.telemetry.MeterProvider(), clients.telemetry.TracerProvider())
	if err != nil {
		return services{}, fmt.Errorf("renderer: %w", err)
	}

	ketchupApp := ketchup.New(ctx, rendererService, ketchupService, userService, repositoryService, clients.redis, clients.telemetry.TracerProvider())

	return services{
		server:   server.New(config.server),
		owasp:    owasp.New(config.owasp),
		cors:     cors.New(config.cors),
		renderer: rendererService,

		authMiddleware: authMiddlewareService,
		user:           userService,
		ketchup:        ketchupApp,
	}, nil
}

func initAuth(db db.Service, tracerProvider trace.TracerProvider) (authService.Service, authMiddleware.Service) {
	authProvider := authStore.New(db)
	identProvider := authIdent.New(authProvider, "ketchup")

	return authService.New(authProvider, authProvider), authMiddleware.New(authProvider, tracerProvider, identProvider)
}
