package ketchup

import (
	"context"
	"html/template"
	"net/http"
	"time"

	"github.com/ViBiOh/httputils/v4/pkg/cache"
	"github.com/ViBiOh/httputils/v4/pkg/redis"
	"github.com/ViBiOh/httputils/v4/pkg/renderer"
	"github.com/ViBiOh/ketchup/pkg/cap"
	"github.com/ViBiOh/ketchup/pkg/model"
	"github.com/ViBiOh/ketchup/pkg/service/ketchup"
	"github.com/ViBiOh/ketchup/pkg/service/repository"
	"github.com/ViBiOh/ketchup/pkg/service/user"
	"go.opentelemetry.io/otel/trace"
)

const appPath = "/app"

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

type LogoutService interface {
	Logout(w http.ResponseWriter, r *http.Request)
}

type Service struct {
	repository repository.Service
	user       user.Service
	ketchup    ketchup.Service
	redis      redis.Client
	logout     LogoutService
	cache      *cache.Cache[model.User, []model.Repository]
	renderer   *renderer.Service
	cap        cap.Service
}

func New(ctx context.Context, renderer *renderer.Service, ketchup ketchup.Service, user user.Service, repository repository.Service, cap cap.Service, logout LogoutService, redis redis.Client, traceProvider trace.TracerProvider) Service {
	service := Service{
		renderer:   renderer,
		cap:        cap,
		logout:     logout,
		ketchup:    ketchup,
		user:       user,
		repository: repository,
		redis:      redis,
	}

	service.cache = cache.New(redis, suggestCacheKey, func(ctx context.Context, user model.User) ([]model.Repository, error) {
		return service.repository.Suggest(ctx, ignoresIdsFromCtx(ctx), countFromCtx(ctx))
	}, traceProvider).
		WithTTL(time.Hour*24).
		WithMaxConcurrency(6).
		WithClientSideCaching(ctx, "ketchup_suggests", 10)

	return service
}
