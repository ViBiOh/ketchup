package github

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/ViBiOh/flags"
	"github.com/ViBiOh/httputils/v4/pkg/cron"
	"github.com/ViBiOh/httputils/v4/pkg/httpjson"
	"github.com/ViBiOh/httputils/v4/pkg/logger"
	httpModel "github.com/ViBiOh/httputils/v4/pkg/model"
	"github.com/ViBiOh/httputils/v4/pkg/request"
	"github.com/ViBiOh/httputils/v4/pkg/tracer"
	"github.com/ViBiOh/ketchup/pkg/model"
	"github.com/ViBiOh/ketchup/pkg/semver"
	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel/trace"
)

var (
	apiURL = "https://api.github.com"

	httpClient = request.CreateClient(30*time.Second, func(r *http.Request, via []*http.Request) error {
		logger.Warn("Redirect from %s to %s", via[len(via)-1].URL.Path, r.URL.Path)
		return nil
	})
)

type redis interface {
	Ping(context.Context) error
	Exclusive(context.Context, string, time.Duration, func(context.Context) error) (bool, error)
}

// Tag describes a GitHub Tag
type Tag struct {
	Name string `json:"name"`
}

// RateLimit describes a rate limit on given ressource
type RateLimit struct {
	Remaining uint64 `json:"remaining"`
}

// RateLimitResponse describes the rate_limit response
type RateLimitResponse struct {
	Resources map[string]RateLimit `json:"resources"`
}

// App of package
type App interface {
	Start(context.Context)
	LatestVersions(context.Context, string, []string) (map[string]semver.Version, error)
}

// Config of package
type Config struct {
	token *string
}

type app struct {
	tracer   trace.Tracer
	redisApp redis
	metrics  prometheus.Gauge
	token    string
}

// Flags adds flags for configuring package
func Flags(fs *flag.FlagSet, prefix string) Config {
	return Config{
		token: flags.String(fs, prefix, "github", "Token", "OAuth Token", "", nil),
	}
}

// New creates new App from Config
func New(config Config, redisApp redis, registerer prometheus.Registerer, tracerApp tracer.App) App {
	httpClient = tracer.AddTracerToClient(httpClient, tracerApp.GetProvider())

	var metrics prometheus.Gauge
	if !httpModel.IsNil(registerer) {
		metrics := prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "ketchup",
			Name:      "github_rate_limit_remainings",
		})
		registerer.MustRegister(metrics)
	}

	return app{
		token:    strings.TrimSpace(*config.token),
		redisApp: redisApp,
		tracer:   tracerApp.GetTracer("github"),
		metrics:  metrics,
	}
}

func (a app) newClient() request.Request {
	return request.New().Header("Authorization", fmt.Sprintf("token %s", a.token)).WithClient(httpClient)
}

func (a app) Start(ctx context.Context) {
	if httpModel.IsNil(a.metrics) {
		return
	}

	cron.New().Now().Each(time.Minute).WithTracer(a.tracer).OnError(func(err error) {
		logger.Error("get rate limit metrics: %s", err)
	}).Exclusive(a.redisApp, "ketchup:github_rate_limit_metrics", 15*time.Second).Start(ctx, func(ctx context.Context) error {
		value, err := a.getRateLimit(ctx)
		if err != nil {
			return err
		}

		a.metrics.Set(float64(value))
		return nil
	})
}

func (a app) LatestVersions(ctx context.Context, repository string, patterns []string) (map[string]semver.Version, error) {
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

func (a app) getRateLimit(ctx context.Context) (uint64, error) {
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
