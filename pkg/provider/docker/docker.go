package docker

import (
	"context"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"runtime"
	"strings"
	"time"

	"github.com/ViBiOh/flags"
	"github.com/ViBiOh/httputils/v4/pkg/httpjson"
	"github.com/ViBiOh/httputils/v4/pkg/request"
	"github.com/ViBiOh/ketchup/pkg/model"
	"github.com/ViBiOh/ketchup/pkg/semver"
)

const (
	registryURL  = "https://index.docker.io"
	authURL      = "https://auth.docker.io/token"
	ghcrLoginURL = "https://ghcr.io/token"
	nextLink     = `rel="next"`
)

type authResponse struct {
	AccessToken string `json:"access_token"`
	Token       string `json:"token"`
}

type Service struct {
	username string
	password string
}

type Config struct {
	Username string
	Password string
}

func Flags(fs *flag.FlagSet, prefix string, overrides ...flags.Override) *Config {
	var config Config

	flags.New("Username", "Registry Username").Prefix(prefix).DocPrefix("docker").StringVar(fs, &config.Username, "", overrides)
	flags.New("Password", "Registry Password").Prefix(prefix).DocPrefix("docker").StringVar(fs, &config.Password, "", overrides)

	return &config
}

func New(config *Config) Service {
	return Service{
		username: config.Username,
		password: config.Password,
	}
}

func (s Service) LatestVersions(ctx context.Context, repository string, patterns []string) (map[string]semver.Version, error) {
	versions, compiledPatterns, err := model.PreparePatternMatching(patterns)
	if err != nil {
		return nil, fmt.Errorf("prepare pattern matching: %w", err)
	}

	registry, repository, auth, err := s.getImageDetails(ctx, repository)
	if err != nil {
		return nil, fmt.Errorf("compute image details: %w", err)
	}

	tagsURL := fmt.Sprintf("%s/v2/%s/tags/list", registry, repository)

	for len(tagsURL) != 0 {
		req := request.Get(tagsURL).WithClient(request.CreateClient(15*time.Second, nil))

		if len(auth) != 0 {
			req = req.Header("Authorization", auth)
		}

		resp, err := req.Send(ctx, nil)
		if err != nil {
			return nil, fmt.Errorf("fetch tags: %w", err)
		}

		if err := browseRegistryTagsList(resp.Body, repository, versions, compiledPatterns); err != nil {
			return nil, err
		}

		tagsURL = getNextURL(resp.Header, registry)
	}

	return versions, nil
}

func (s Service) getImageDetails(ctx context.Context, repository string) (string, string, string, error) {
	parts := strings.Split(repository, "/")
	if len(parts) > 2 {
		var token string
		var err error

		if parts[0] == "ghcr.io" {
			token, err = s.ghcr(ctx, strings.Join(parts[1:], "/"))
		}

		return fmt.Sprintf("https://%s", parts[0]), strings.Join(parts[1:], "/"), token, err
	}

	if len(parts) == 1 {
		repository = fmt.Sprintf("library/%s", repository)
	}

	bearerToken, err := s.login(ctx, repository)
	if err != nil {
		return "", "", "", fmt.Errorf("authenticate to docker hub: %w", err)
	}

	return registryURL, repository, fmt.Sprintf("Bearer %s", bearerToken), nil
}

func (s Service) login(ctx context.Context, repository string) (string, error) {
	values := url.Values{}
	values.Add("grant_type", "password")
	values.Add("service", "registry.docker.io")
	values.Add("client_id", "ketchup")
	values.Add("scope", fmt.Sprintf("repository:%s:pull", repository))
	values.Add("username", s.username)
	values.Add("password", s.password)

	resp, err := request.Post(authURL).Form(ctx, values)
	if err != nil {
		return "", fmt.Errorf("authenticate to `%s`: %w", authURL, err)
	}

	var authContent authResponse
	if err := httpjson.Read(resp, &authContent); err != nil {
		return "", fmt.Errorf("read auth token: %w", err)
	}

	return authContent.AccessToken, nil
}

func (s Service) ghcr(ctx context.Context, repository string) (string, error) {
	values := url.Values{}
	values.Add("scope", fmt.Sprintf("repository:%s:pull", repository))

	resp, err := request.Get(fmt.Sprintf("%s?%s", ghcrLoginURL, values.Encode())).Send(ctx, nil)
	if err != nil {
		return "", fmt.Errorf("authenticate to `%s`: %w", ghcrLoginURL, err)
	}

	var authContent authResponse
	if err := httpjson.Read(resp, &authContent); err != nil {
		return "", fmt.Errorf("read auth token: %w", err)
	}

	return fmt.Sprintf("Bearer %s", authContent.Token), nil
}

func browseRegistryTagsList(body io.ReadCloser, name string, versions map[string]semver.Version, patterns map[string]semver.Pattern) error {
	done := make(chan struct{})
	versionsStream := make(chan string, runtime.NumCPU())

	go func() {
		defer close(done)

		for tag := range versionsStream {
			tagVersion, err := semver.Parse(tag, semver.ExtractName(name))
			if err != nil {
				continue
			}

			model.CheckPatternsMatching(versions, patterns, tagVersion)
		}
	}()

	if err := httpjson.Stream(body, versionsStream, "tags", true); err != nil {
		return fmt.Errorf("read tags: %w", err)
	}

	<-done

	return nil
}

func getNextURL(headers http.Header, registry string) string {
	links := headers.Values("link")
	if len(links) == 0 {
		return ""
	}

	for _, link := range links {
		if !strings.Contains(link, nextLink) {
			continue
		}

		for _, part := range strings.Split(link, ";") {
			if strings.Contains(part, nextLink) {
				continue
			}

			return fmt.Sprintf("%s%s", registry, strings.TrimSpace(strings.Trim(strings.Trim(part, "<"), ">")))
		}
	}

	return ""
}
