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

type Service struct {
	user       user.Service
	ketchup    ketchup.Service
	repository repository.Service
	cache      *cache.Cache[model.User, []model.Repository]
	redis      redis.Client
	renderer   *renderer.Service
}

func New(ctx context.Context, rendererService *renderer.Service, ketchupService ketchup.Service, userService user.Service, repositoryService repository.Service, redisClient redis.Client, traceProvider trace.TracerProvider) Service {
	service := Service{
		renderer:   rendererService,
		ketchup:    ketchupService,
		user:       userService,
		repository: repositoryService,
		redis:      redisClient,
	}

	service.cache = cache.New(redisClient, suggestCacheKey, func(ctx context.Context, user model.User) ([]model.Repository, error) {
		return service.repository.Suggest(ctx, ignoresIdsFromCtx(ctx), countFromCtx(ctx))
	}, traceProvider).
		WithTTL(time.Hour*24).
		WithMaxConcurrency(6).
		WithClientSideCaching(ctx, "ketchup_suggests")

	return service
}

func (s Service) Handler() http.Handler {
	rendererHandler := s.renderer.Handler(s.TemplateFunc)
	ketchupHandler := http.StripPrefix(ketchupsPath, s.ketchups())

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, ketchupsPath) {
			ketchupHandler.ServeHTTP(w, r)
			return
		}

		rendererHandler.ServeHTTP(w, r)
	})
}
