package github

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/ViBiOh/flags"
	"github.com/ViBiOh/httputils/v4/pkg/cron"
	"github.com/ViBiOh/httputils/v4/pkg/httpjson"
	httpModel "github.com/ViBiOh/httputils/v4/pkg/model"
	"github.com/ViBiOh/httputils/v4/pkg/request"
	"github.com/ViBiOh/httputils/v4/pkg/telemetry"
	"github.com/ViBiOh/ketchup/pkg/model"
	"github.com/ViBiOh/ketchup/pkg/semver"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

var (
	apiURL = "https://api.github.com"

	httpClient = request.CreateClient(30*time.Second, func(r *http.Request, via []*http.Request) error {
		slog.Warn("Redirect", "from", via[len(via)-1].URL.Path, "to", r.URL.Path)
		return nil
	})
)

type redis interface {
	Ping(context.Context) error
	Exclusive(context.Context, string, time.Duration, func(context.Context) error) (bool, error)
}

type Tag struct {
	Name string `json:"name"`
}

type RateLimit struct {
	Remaining uint64 `json:"remaining"`
}

type RateLimitResponse struct {
	Resources map[string]RateLimit `json:"resources"`
}

type App interface {
	Start(context.Context)
	LatestVersions(context.Context, string, []string) (map[string]semver.Version, error)
}

type Config struct {
	token *string
}

type app struct {
	traceProvider  trace.TracerProvider
	redisApp       redis
	token          string
	rateLimitValue uint64
	mutex          sync.RWMutex
}

func Flags(fs *flag.FlagSet, prefix string) Config {
	return Config{
		token: flags.New("Token", "OAuth Token").Prefix(prefix).DocPrefix("github").String(fs, "", nil),
	}
}

func New(config Config, redisApp redis, meterProvider metric.MeterProvider, traceProvider trace.TracerProvider) App {
	httpClient = telemetry.AddOpenTelemetryToClient(httpClient, meterProvider, traceProvider)

	app := &app{
		token:         strings.TrimSpace(*config.token),
		redisApp:      redisApp,
		traceProvider: traceProvider,
	}

	if !httpModel.IsNil(meterProvider) {
		_, err := meterProvider.Meter("github.com/ViBiOh/ketchup/pkg/provider/github").
			Int64ObservableGauge("github.rate_limit.remainings",
				metric.WithInt64Callback(func(ctx context.Context, io metric.Int64Observer) error {
					app.mutex.RLock()
					defer app.mutex.RUnlock()

					io.Observe(int64(app.rateLimitValue))
					return nil
				}))
		if err != nil {
			slog.Error("create observable gauge", "err", err)
		}
	}

	return app
}

func (a *app) newClient() request.Request {
	return request.New().Header("Authorization", fmt.Sprintf("token %s", a.token)).WithClient(httpClient)
}

func (a *app) Start(ctx context.Context) {
	cron.New().Now().Each(time.Minute).WithTracerProvider(a.traceProvider).OnError(func(err error) {
		slog.Error("get rate limit metrics", "err", err)
	}).Exclusive(a.redisApp, "ketchup:github_rate_limit_metrics", 15*time.Second).Start(ctx, func(ctx context.Context) error {
		value, err := a.getRateLimit(ctx)
		if err != nil {
			return err
		}

		a.mutex.Lock()
		defer a.mutex.Unlock()

		a.rateLimitValue = value

		return nil
	})
}

func (a *app) LatestVersions(ctx context.Context, repository string, patterns []string) (map[string]semver.Version, error) {
	versions, compiledPatterns, err := model.PreparePatternMatching(patterns)
	if err != nil {
		return nil, fmt.Errorf("prepare pattern matching: %w", err)
	}

	page := 1
	req := a.newClient()
	for {
		resp, err := req.Get(fmt.Sprintf("%s/repos/%s/tags?per_page=100&page=%d", apiURL, repository, page)).Send(ctx, nil)
		if err != nil {
			return nil, fmt.Errorf("list page %d of tags: %w", page, err)
		}

		var tags []Tag
		if err := httpjson.Read(resp, &tags); err != nil {
			return nil, fmt.Errorf("read tags page #%d: %w", page, err)
		}

		for _, tag := range tags {
			tagVersion, err := semver.Parse(tag.Name)
			if err != nil {
				continue
			}

			model.CheckPatternsMatching(versions, compiledPatterns, tagVersion)
		}

		if !hasNext(resp) {
			break
		}

		page++
	}

	return versions, nil
}

func (a *app) getRateLimit(ctx context.Context) (uint64, error) {
	resp, err := a.newClient().Get(fmt.Sprintf("%s/rate_limit", apiURL)).Send(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("get rate limit: %w", err)
	}

	var rateLimits RateLimitResponse
	if err := httpjson.Read(resp, &rateLimits); err != nil {
		return 0, fmt.Errorf("read rate limit: %w", err)
	}

	return rateLimits.Resources["core"].Remaining, nil
}

func hasNext(resp *http.Response) bool {
	for _, value := range resp.Header.Values("Link") {
		if strings.Contains(value, `rel="next"`) {
			return true
		}
	}

	return false
}
