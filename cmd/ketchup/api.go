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

func initAuth(db db.App, tracerProvider trace.TracerProvider) (authService.App, authMiddleware.App) {
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

	telemetryApp, err := telemetry.New(ctx, telemetryConfig)
	if err != nil {
		slog.Error("create telemetry", "err", err)
		os.Exit(1)
	}

	defer telemetryApp.Close(ctx)
	request.AddOpenTelemetryToDefaultClient(telemetryApp.MeterProvider(), telemetryApp.TracerProvider())

	go func() {
		fmt.Println(http.ListenAndServe("localhost:9999", http.DefaultServeMux))
	}()

	appServer := server.New(appServerConfig)

	ketchupDb, err := db.New(ctx, dbConfig, telemetryApp.TracerProvider())
	if err != nil {
		slog.Error("create database", "err", err)
		os.Exit(1)
	}

	defer ketchupDb.Close()

	redisApp, err := redis.New(redisConfig, telemetryApp.MeterProvider(), telemetryApp.TracerProvider())
	if err != nil {
		slog.Error("create redis", "err", err)
		os.Exit(1)
	}

	defer redisApp.Close()

	healthApp := health.New(healthConfig, ketchupDb.Ping, redisApp.Ping)

	authServiceApp, authMiddlewareApp := initAuth(ketchupDb, telemetryApp.TracerProvider())

	userServiceApp := userService.New(userStore.New(ketchupDb), &authServiceApp)
	githubApp := github.New(githubConfig, redisApp, telemetryApp.MeterProvider(), telemetryApp.TracerProvider())
	dockerApp := docker.New(dockerConfig)
	helmApp := helm.New()
	npmApp := npm.New()
	pypiApp := pypi.New()
	repositoryServiceApp := repositoryService.New(repositoryStore.New(ketchupDb), githubApp, helmApp, dockerApp, npmApp, pypiApp)
	ketchupServiceApp := ketchupService.New(ketchupStore.New(ketchupDb), repositoryServiceApp)

	mailerApp, err := mailer.New(mailerConfig, telemetryApp.MeterProvider(), telemetryApp.TracerProvider())
	if err != nil {
		slog.Error("create mailer", "err", err)
		os.Exit(1)
	}

	defer mailerApp.Close()

	publicRendererApp, err := renderer.New(rendererConfig, content, ketchup.FuncMap, telemetryApp.MeterProvider(), telemetryApp.TracerProvider())
	if err != nil {
		slog.Error("create renderer", "err", err)
		os.Exit(1)
	}

	notifierApp := notifier.New(notifierConfig, repositoryServiceApp, ketchupServiceApp, userServiceApp, mailerApp, helmApp)
	schedulerApp := scheduler.New(schedulerConfig, notifierApp, redisApp, telemetryApp.MeterProvider(), telemetryApp.TracerProvider())
	ketchupApp := ketchup.New(publicRendererApp, ketchupServiceApp, userServiceApp, repositoryServiceApp, redisApp, telemetryApp.TracerProvider())

	publicHandler := publicRendererApp.Handler(ketchupApp.PublicTemplateFunc)
	signupHandler := http.StripPrefix(signupPath, ketchupApp.Signup())
	protectedhandler := authMiddlewareApp.Middleware(middleware.New(userServiceApp).Middleware(http.StripPrefix(appPath, ketchupApp.Handler())))

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

	doneCtx := healthApp.Done(ctx)

	go githubApp.Start(doneCtx)
	if schedulerApp != nil {
		go schedulerApp.Start(doneCtx)
	}

	endCtx := healthApp.End(ctx)

	go appServer.Start(endCtx, "http", httputils.Handler(appHandler, healthApp, recoverer.Middleware, telemetryApp.Middleware("http"), owasp.New(owaspConfig).Middleware, cors.New(corsConfig).Middleware))

	healthApp.WaitForTermination(appServer.Done())
	server.GracefulWait(appServer.Done())
}
