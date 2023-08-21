package ketchup

import (
	"context"
	"html/template"
	"net/http"
	"strings"
	"time"

	"github.com/ViBiOh/httputils/v4/pkg/cache"
	"github.com/ViBiOh/httputils/v4/pkg/redis"
	"github.com/ViBiOh/httputils/v4/pkg/renderer"
	"github.com/ViBiOh/ketchup/pkg/model"
	"github.com/ViBiOh/ketchup/pkg/service/ketchup"
	"github.com/ViBiOh/ketchup/pkg/service/repository"
	"github.com/ViBiOh/ketchup/pkg/service/user"
	"go.opentelemetry.io/otel/trace"
)

const (
	ketchupsPath = "/ketchups"
	appPath      = "/app"
)

var FuncMap = template.FuncMap{
	"frequencyImage": func(frequency model.KetchupFrequency) string {
		switch frequency {
		case model.None:
			return "bell-slash"
		case model.Weekly:
			return "calendar"
		default:
			return "clock"
		}
	},
}

type App struct {
	userService       user.App
	ketchupService    ketchup.App
	repositoryService repository.App
	cacheApp          *cache.App[model.User, []model.Repository]
	redisApp          redis.Client
	rendererApp       *renderer.App
}

func New(rendererApp *renderer.App, ketchupService ketchup.App, userService user.App, repositoryService repository.App, redisApp redis.Client, traceProvider trace.TracerProvider) App {
	app := App{
		rendererApp:       rendererApp,
		ketchupService:    ketchupService,
		userService:       userService,
		repositoryService: repositoryService,
		redisApp:          redisApp,
	}

	app.cacheApp = cache.New(redisApp, suggestCacheKey, func(ctx context.Context, user model.User) ([]model.Repository, error) {
		return app.repositoryService.Suggest(ctx, ignoresIdsFromCtx(ctx), countFromCtx(ctx))
	}, traceProvider).WithTTL(time.Hour * 24).WithMaxConcurrency(6)

	return app
}

func (a App) Handler() http.Handler {
	rendererHandler := a.rendererApp.Handler(a.AppTemplateFunc)
	ketchupHandler := http.StripPrefix(ketchupsPath, a.ketchups())

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, ketchupsPath) {
			ketchupHandler.ServeHTTP(w, r)
			return
		}

		rendererHandler.ServeHTTP(w, r)
	})
}
