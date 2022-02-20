package main

import (
	"embed"
	"flag"
	"fmt"
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
	"github.com/ViBiOh/httputils/v4/pkg/prometheus"
	"github.com/ViBiOh/httputils/v4/pkg/recoverer"
	"github.com/ViBiOh/httputils/v4/pkg/redis"
	"github.com/ViBiOh/httputils/v4/pkg/renderer"
	"github.com/ViBiOh/httputils/v4/pkg/request"
	"github.com/ViBiOh/httputils/v4/pkg/server"
	"github.com/ViBiOh/httputils/v4/pkg/tracer"
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
)

const (
	appPath    = "/app"
	signupPath = "/signup"
)

//go:embed templates static
var content embed.FS

func initAuth(db db.App, tracerApp tracer.App) (authService.App, authMiddleware.App) {
	authProvider := authStore.New(db)
	identProvider := authIdent.New(authProvider, "ketchup")

	return authService.New(authProvider, authProvider), authMiddleware.New(authProvider, tracerApp, identProvider)
}

func main() {
	fs := flag.NewFlagSet("ketchup", flag.ExitOnError)

	appServerConfig := server.Flags(fs, "")
	promServerConfig := server.Flags(fs, "prometheus", flags.NewOverride("Port", 9090), flags.NewOverride("IdleTimeout", "10s"), flags.NewOverride("ShutdownTimeout", "5s"))
	healthConfig := health.Flags(fs, "")

	alcotestConfig := alcotest.Flags(fs, "")
	loggerConfig := logger.Flags(fs, "logger")
	tracerConfig := tracer.Flags(fs, "tracer")
	prometheusConfig := prometheus.Flags(fs, "prometheus", flags.NewOverride("Gzip", false))
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

	logger.Fatal(fs.Parse(os.Args[1:]))

	alcotest.DoAndExit(alcotestConfig)
	logger.Global(logger.New(loggerConfig))
	defer logger.Close()

	tracerApp, err := tracer.New(tracerConfig)
	logger.Fatal(err)
	defer tracerApp.Close()
	request.AddTracerToDefaultClient(tracerApp.GetProvider())

	go func() {
		fmt.Println(http.ListenAndServe("localhost:9999", http.DefaultServeMux))
	}()

	appServer := server.New(appServerConfig)
	promServer := server.New(promServerConfig)
	prometheusApp := prometheus.New(prometheusConfig)

	ketchupDb, err := db.New(dbConfig, tracerApp)
	logger.Fatal(err)
	defer ketchupDb.Close()

	redisApp := redis.New(redisConfig, prometheusApp.Registerer(), tracerApp)

	healthApp := health.New(healthConfig, ketchupDb.Ping, redisApp.Ping)

	authServiceApp, authMiddlewareApp := initAuth(ketchupDb, tracerApp)

	userServiceApp := userService.New(userStore.New(ketchupDb), &authServiceApp)
	githubApp := github.New(githubConfig, redisApp, tracerApp)
	dockerApp := docker.New(dockerConfig)
	helmApp := helm.New()
	npmApp := npm.New()
	pypiApp := pypi.New()
	repositoryServiceApp := repositoryService.New(repositoryStore.New(ketchupDb), githubApp, helmApp, dockerApp, npmApp, pypiApp)
	ketchupServiceApp := ketchupService.New(ketchupStore.New(ketchupDb), repositoryServiceApp)

	mailerApp, err := mailer.New(mailerConfig, prometheusApp.Registerer())
	logger.Fatal(err)
	defer mailerApp.Close()

	publicRendererApp, err := renderer.New(rendererConfig, content, ketchup.FuncMap, tracerApp)
	logger.Fatal(err)

	notifierApp := notifier.New(notifierConfig, repositoryServiceApp, ketchupServiceApp, userServiceApp, mailerApp, helmApp)
	schedulerApp := scheduler.New(schedulerConfig, notifierApp, redisApp, tracerApp)
	ketchupApp := ketchup.New(publicRendererApp, ketchupServiceApp, userServiceApp, repositoryServiceApp, redisApp)

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

	go githubApp.Start(prometheusApp.Registerer(), healthApp.Done())
	if schedulerApp != nil {
		go schedulerApp.Start(healthApp.Done())
	}

	go promServer.Start("prometheus", healthApp.End(), prometheusApp.Handler())
	go appServer.Start("http", healthApp.End(), httputils.Handler(appHandler, healthApp, recoverer.Middleware, prometheusApp.Middleware, tracerApp.Middleware, owasp.New(owaspConfig).Middleware, cors.New(corsConfig).Middleware))

	healthApp.WaitForTermination(appServer.Done())
	server.GracefulWait(appServer.Done(), promServer.Done())
}
