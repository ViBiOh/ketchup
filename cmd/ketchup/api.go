package main

import (
	"context"
	"embed"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"strings"

	_ "net/http/pprof"

	authIdent "github.com/ViBiOh/auth/v2/pkg/ident/basic"
	authMiddleware "github.com/ViBiOh/auth/v2/pkg/middleware"
	authService "github.com/ViBiOh/auth/v2/pkg/service"
	authStore "github.com/ViBiOh/auth/v2/pkg/store/db"
	"github.com/ViBiOh/flags"
	"github.com/ViBiOh/httputils/v4/pkg/alcotest"
	"github.com/ViBiOh/httputils/v4/pkg/cors"
	"github.com/ViBiOh/httputils/v4/pkg/db"
	"github.com/ViBiOh/httputils/v4/pkg/health"
	"github.com/ViBiOh/httputils/v4/pkg/httputils"
	"github.com/ViBiOh/httputils/v4/pkg/logger"
	"github.com/ViBiOh/httputils/v4/pkg/owasp"
	"github.com/ViBiOh/httputils/v4/pkg/recoverer"
	"github.com/ViBiOh/httputils/v4/pkg/redis"
	"github.com/ViBiOh/httputils/v4/pkg/renderer"
	"github.com/ViBiOh/httputils/v4/pkg/request"
	"github.com/ViBiOh/httputils/v4/pkg/server"
	"github.com/ViBiOh/httputils/v4/pkg/telemetry"
	"github.com/ViBiOh/ketchup/pkg/ketchup"
	"github.com/ViBiOh/ketchup/pkg/middleware"
	"github.com/ViBiOh/ketchup/pkg/notifier"
	"github.com/ViBiOh/ketchup/pkg/provider/docker"
	"github.com/ViBiOh/ketchup/pkg/provider/github"
	"github.com/ViBiOh/ketchup/pkg/provider/helm"
	"github.com/ViBiOh/ketchup/pkg/provider/npm"
	"github.com/ViBiOh/ketchup/pkg/provider/pypi"
	"github.com/ViBiOh/ketchup/pkg/scheduler"
	ketchupService "github.com/ViBiOh/ketchup/pkg/service/ketchup"
	repositoryService "github.com/ViBiOh/ketchup/pkg/service/repository"
	userService "github.com/ViBiOh/ketchup/pkg/service/user"
	ketchupStore "github.com/ViBiOh/ketchup/pkg/store/ketchup"
	repositoryStore "github.com/ViBiOh/ketchup/pkg/store/repository"
	userStore "github.com/ViBiOh/ketchup/pkg/store/user"
	mailer "github.com/ViBiOh/mailer/pkg/client"
	"go.opentelemetry.io/otel/trace"
)

const (
	appPath    = "/app"
	signupPath = "/signup"
)

//go:embed templates static
var content embed.FS

func initAuth(db db.Service, tracerProvider trace.TracerProvider) (authService.Service, authMiddleware.Service) {
	authProvider := authStore.New(db)
	identProvider := authIdent.New(authProvider, "ketchup")

	return authService.New(authProvider, authProvider), authMiddleware.New(authProvider, tracerProvider, identProvider)
}

func main() {
	fs := flag.NewFlagSet("ketchup", flag.ExitOnError)
	fs.Usage = flags.Usage(fs)

	appServerConfig := server.Flags(fs, "")
	healthConfig := health.Flags(fs, "")

	alcotestConfig := alcotest.Flags(fs, "")
	loggerConfig := logger.Flags(fs, "logger")
	telemetryConfig := telemetry.Flags(fs, "telemetry")
	owaspConfig := owasp.Flags(fs, "", flags.NewOverride("Csp", "default-src 'self'; base-uri 'self'; script-src 'self' 'httputils-nonce'; style-src 'self' 'httputils-nonce'"))
	corsConfig := cors.Flags(fs, "cors")
	rendererConfig := renderer.Flags(fs, "", flags.NewOverride("Title", "Ketchup"), flags.NewOverride("PublicURL", "https://ketchup.vibioh.fr"))

	dbConfig := db.Flags(fs, "db")
	redisConfig := redis.Flags(fs, "redis")
	mailerConfig := mailer.Flags(fs, "mailer")
	githubConfig := github.Flags(fs, "github")
	dockerConfig := docker.Flags(fs, "docker")
	notifierConfig := notifier.Flags(fs, "notifier")
	schedulerConfig := scheduler.Flags(fs, "scheduler")

	if err := fs.Parse(os.Args[1:]); err != nil {
		log.Fatal(err)
	}

	alcotest.DoAndExit(alcotestConfig)

	logger.Init(loggerConfig)

	ctx := context.Background()

	telemetryService, err := telemetry.New(ctx, telemetryConfig)
	if err != nil {
		slog.ErrorContext(ctx, "create telemetry", "error", err)
		os.Exit(1)
	}

	defer telemetryService.Close(ctx)

	logger.AddOpenTelemetryToDefaultLogger(telemetryService)
	request.AddOpenTelemetryToDefaultClient(telemetryService.MeterProvider(), telemetryService.TracerProvider())

	go func() {
		fmt.Println(http.ListenAndServe("localhost:9999", http.DefaultServeMux))
	}()

	appServer := server.New(appServerConfig)

	ketchupDb, err := db.New(ctx, dbConfig, telemetryService.TracerProvider())
	if err != nil {
		slog.ErrorContext(ctx, "create database", "error", err)
		os.Exit(1)
	}

	defer ketchupDb.Close()

	redisClient, err := redis.New(redisConfig, telemetryService.MeterProvider(), telemetryService.TracerProvider())
	if err != nil {
		slog.ErrorContext(ctx, "create redis", "error", err)
		os.Exit(1)
	}

	defer redisClient.Close()

	healthService := health.New(ctx, healthConfig, ketchupDb.Ping, redisClient.Ping)

	authServiceService, authMiddlewareApp := initAuth(ketchupDb, telemetryService.TracerProvider())

	userServiceService := userService.New(userStore.New(ketchupDb), &authServiceService)
	githubService := github.New(githubConfig, redisClient, telemetryService.MeterProvider(), telemetryService.TracerProvider())
	dockerService := docker.New(dockerConfig)
	helmService := helm.New()
	npmService := npm.New()
	pypiService := pypi.New()
	repositoryServiceService := repositoryService.New(repositoryStore.New(ketchupDb), githubService, helmService, dockerService, npmService, pypiService)
	ketchupServiceService := ketchupService.New(ketchupStore.New(ketchupDb), repositoryServiceService)

	mailerService, err := mailer.New(mailerConfig, telemetryService.MeterProvider(), telemetryService.TracerProvider())
	if err != nil {
		slog.ErrorContext(ctx, "create mailer", "error", err)
		os.Exit(1)
	}

	defer mailerService.Close()

	publicRendererService, err := renderer.New(rendererConfig, content, ketchup.FuncMap, telemetryService.MeterProvider(), telemetryService.TracerProvider())
	if err != nil {
		slog.ErrorContext(ctx, "create renderer", "error", err)
		os.Exit(1)
	}

	endCtx := healthService.EndCtx()

	notifierService := notifier.New(notifierConfig, repositoryServiceService, ketchupServiceService, userServiceService, mailerService, helmService)
	schedulerService := scheduler.New(schedulerConfig, notifierService, redisClient, telemetryService.MeterProvider(), telemetryService.TracerProvider())
	ketchupService := ketchup.New(endCtx, publicRendererService, ketchupServiceService, userServiceService, repositoryServiceService, redisClient, telemetryService.TracerProvider())

	publicHandler := publicRendererService.Handler(ketchupService.PublicTemplateFunc)
	signupHandler := http.StripPrefix(signupPath, ketchupService.Signup())
	protectedhandler := authMiddlewareApp.Middleware(middleware.New(userServiceService).Middleware(http.StripPrefix(appPath, ketchupService.Handler())))

	appHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, appPath) {
			protectedhandler.ServeHTTP(w, r)
			return
		}

		if strings.HasPrefix(r.URL.Path, signupPath) {
			signupHandler.ServeHTTP(w, r)
			return
		}

		publicHandler.ServeHTTP(w, r)
	})

	doneCtx := healthService.DoneCtx()

	go githubService.Start(doneCtx)
	if schedulerService != nil {
		go schedulerService.Start(doneCtx)
	}

	go appServer.Start(endCtx, httputils.Handler(appHandler, healthService, recoverer.Middleware, telemetryService.Middleware(appServerConfig.Name), owasp.New(owaspConfig).Middleware, cors.New(corsConfig).Middleware))

	healthService.WaitForTermination(appServer.Done())

	server.GracefulWait(appServer.Done())
}
