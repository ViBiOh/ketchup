package github

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/ViBiOh/flags"
	"github.com/ViBiOh/httputils/v4/pkg/httpjson"
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
		slog.LogAttrs(r.Context(), slog.LevelWarn, "Redirect", slog.String("from", via[len(via)-1].URL.Path), slog.String("to", r.URL.Path))
		return nil
	})
)

type Redis interface {
	Ping(context.Context) error
	Exclusive(context.Context, string, time.Duration, func(context.Context) error) (bool, error)
}

type Tag struct {
	Name string `json:"name"`
}

type Config struct {
	Token string
}

type Service struct {
	traceProvider trace.TracerProvider
	redis         Redis
	token         string
}

func Flags(fs *flag.FlagSet, prefix string) *Config {
	var config Config

	flags.New("Token", "OAuth Token").Prefix(prefix).DocPrefix("github").StringVar(fs, &config.Token, "", nil)

	return &config
}

func New(config *Config, redisClient Redis, meterProvider metric.MeterProvider, traceProvider trace.TracerProvider) Service {
	httpClient = telemetry.AddOpenTelemetryToClient(httpClient, meterProvider, traceProvider)

	return Service{
		token:         config.Token,
		redis:         redisClient,
		traceProvider: traceProvider,
	}
}

func (s Service) newClient() request.Request {
	return request.New().Header("Authorization", fmt.Sprintf("token %s", s.token)).WithClient(httpClient)
}

func (s Service) LatestVersions(ctx context.Context, repository string, patterns []string) (map[string]semver.Version, error) {
	versions, compiledPatterns, err := model.PreparePatternMatching(patterns)
	if err != nil {
		return nil, fmt.Errorf("prepare pattern matching: %w", err)
	}

	page := 1
	req := s.newClient()
	for {
		resp, err := req.Get(fmt.Sprintf("%s/repos/%s/tags?per_page=100&page=%d", apiURL, repository, page)).Send(ctx, nil)
		if err != nil {
			return nil, fmt.Errorf("list page %d of tags: %w", page, err)
		}

		tags, err := httpjson.Read[[]Tag](resp)
		if err != nil {
			return nil, fmt.Errorf("read tags page #%d: %w", page, err)
		}

		for _, tag := range tags {
			tagVersion, err := semver.Parse(tag.Name, semver.ExtractName(repository))
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

func hasNext(resp *http.Response) bool {
	for _, value := range resp.Header.Values("Link") {
		if strings.Contains(value, `rel="next"`) {
			return true
		}
	}

	return false
}
