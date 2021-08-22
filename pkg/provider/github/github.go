package github

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/ViBiOh/httputils/v4/pkg/cron"
	"github.com/ViBiOh/httputils/v4/pkg/flags"
	"github.com/ViBiOh/httputils/v4/pkg/httpjson"
	"github.com/ViBiOh/httputils/v4/pkg/logger"
	"github.com/ViBiOh/httputils/v4/pkg/request"
	"github.com/ViBiOh/ketchup/pkg/model"
	"github.com/ViBiOh/ketchup/pkg/semver"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	apiURL = "https://api.github.com"

	httpClient = &http.Client{
		Timeout: 30 * time.Second,
		CheckRedirect: func(r *http.Request, via []*http.Request) error {
			logger.Info("Redirect from %s to %s", via[len(via)-1].URL.Path, r.URL.Path)
			return nil
		},
	}
)

type redis interface {
	Ping() error
	Exclusive(context.Context, string, time.Duration, func(context.Context) error) error
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
	Start(prometheus.Registerer, <-chan struct{})
	LatestVersions(string, []string) (map[string]semver.Version, error)
}

// Config of package
type Config struct {
	token *string
}

type app struct {
	redisApp redis
	token    string
}

// Flags adds flags for configuring package
func Flags(fs *flag.FlagSet, prefix string) Config {
	return Config{
		token: flags.New(prefix, "github", "Token").Default("", nil).Label("OAuth Token").ToString(fs),
	}
}

// New creates new App from Config
func New(config Config, redisApp redis) App {
	return app{
		token:    strings.TrimSpace(*config.token),
		redisApp: redisApp,
	}
}

func (a app) newClient() request.Request {
	return request.New().Header("Authorization", fmt.Sprintf("token %s", a.token)).WithClient(httpClient)
}

func (a app) Start(registerer prometheus.Registerer, done <-chan struct{}) {
	metrics := prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "ketchup",
		Name:      "github_rate_limit_remainings",
	})
	registerer.MustRegister(metrics)

	cron.New().Now().Each(time.Minute).OnError(func(err error) {
		logger.Error("unable to get rate limit metrics: %s", err)
	}).Exclusive(a.redisApp, "github_rate_limit_metrics", 45*time.Second).Start(func(ctx context.Context) error {
		value, err := a.getRateLimit(ctx)
		if err != nil {
			return err
		}

		metrics.Set(float64(value))
		return nil
	}, done)
}

func (a app) LatestVersions(repository string, patterns []string) (map[string]semver.Version, error) {
	versions, compiledPatterns, err := model.PreparePatternMatching(patterns)
	if err != nil {
		return nil, fmt.Errorf("unable to prepare pattern matching: %s", err)
	}

	page := 1
	req := a.newClient()
	for {
		resp, err := req.Get(fmt.Sprintf("%s/repos/%s/tags?per_page=100&page=%d", apiURL, repository, page)).Send(context.Background(), nil)
		if err != nil {
			return nil, fmt.Errorf("unable to list page %d of tags: %s", page, err)
		}

		var tags []Tag
		if err := httpjson.Read(resp, &tags); err != nil {
			return nil, fmt.Errorf("unable to read tags page #%d: %s", page, err)
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
		return 0, fmt.Errorf("unable to get rate limit: %s", err)
	}

	var rateLimits RateLimitResponse
	if err := httpjson.Read(resp, &rateLimits); err != nil {
		return 0, fmt.Errorf("unable to read rate limit: %s", err)
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
