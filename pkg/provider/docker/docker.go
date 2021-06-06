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

	"github.com/ViBiOh/httputils/v4/pkg/flags"
	"github.com/ViBiOh/httputils/v4/pkg/httpjson"
	"github.com/ViBiOh/httputils/v4/pkg/request"
	"github.com/ViBiOh/ketchup/pkg/model"
	"github.com/ViBiOh/ketchup/pkg/semver"
)

const (
	registryURL = "https://index.docker.io"
	authURL     = "https://auth.docker.io/token"
	nextLink    = `rel="next"`
)

type authResponse struct {
	AccessToken string `json:"access_token"`
}

// App of package
type App interface {
	LatestVersions(string, []string) (map[string]semver.Version, error)
}

// Config of package
type Config struct {
	username *string
	password *string
}

type app struct {
	username string
	password string
}

// Flags adds flags for configuring package
func Flags(fs *flag.FlagSet, prefix string, overrides ...flags.Override) Config {
	return Config{
		username: flags.New(prefix, "docker").Name("Username").Default(flags.Default("Username", "", overrides)).Label("Registry Username").ToString(fs),
		password: flags.New(prefix, "docker").Name("Password").Default(flags.Default("Password", "", overrides)).Label("Registry Password").ToString(fs),
	}
}

// New creates new App from Config
func New(config Config) App {
	return app{
		username: strings.TrimSpace(*config.username),
		password: strings.TrimSpace(*config.password),
	}
}

func (a app) LatestVersions(repository string, patterns []string) (map[string]semver.Version, error) {
	ctx := context.Background()

	versions, compiledPatterns, err := model.PreparePatternMatching(patterns)
	if err != nil {
		return nil, fmt.Errorf("unable to prepare pattern matching: %s", err)
	}

	registry, repository, auth, err := a.getImageDetails(ctx, repository)
	if err != nil {
		return nil, fmt.Errorf("unable to compute image details: %s", err)
	}

	url := fmt.Sprintf("%s/v2/%s/tags/list", registry, repository)

	for len(url) != 0 {
		req := request.New().Get(url)

		if len(auth) != 0 {
			req.Header("Authorization", auth)
		}

		resp, err := req.Send(ctx, nil)
		if err != nil {
			return nil, fmt.Errorf("unable to fetch tags: %s", err)
		}

		if err := browseRegistryTagsList(resp.Body, versions, compiledPatterns); err != nil {
			return nil, err
		}

		url = getNextURL(resp.Header, registry)
	}

	return versions, nil
}

func (a app) getImageDetails(ctx context.Context, repository string) (string, string, string, error) {
	parts := strings.Split(repository, "/")
	if len(parts) > 2 {
		var token string

		if parts[0] == "ghcr.io" {
			token = "token"
		}

		return fmt.Sprintf("https://%s", parts[0]), strings.Join(parts[1:], "/"), token, nil
	}

	if len(parts) == 1 {
		repository = fmt.Sprintf("library/%s", repository)
	}

	bearerToken, err := a.login(ctx, repository)
	if err != nil {
		return "", "", "", fmt.Errorf("unable to authenticate to docker hub: %s", err)
	}

	return registryURL, repository, fmt.Sprintf("Bearer %s", bearerToken), nil
}

func (a app) login(ctx context.Context, repository string) (string, error) {
	values := url.Values{}
	values.Set("grant_type", "password")
	values.Set("service", "registry.docker.io")
	values.Set("client_id", "ketchup")
	values.Set("scope", fmt.Sprintf("repository:%s:pull", repository))
	values.Set("username", a.username)
	values.Set("password", a.password)

	resp, err := request.New().Post(authURL).Form(ctx, values)
	if err != nil {
		return "", fmt.Errorf("unable to authenticate to `%s`: %s", authURL, err)
	}

	var authContent authResponse
	if err := httpjson.Read(resp, &authContent); err != nil {
		return "", fmt.Errorf("unable to read auth token: %s", err)
	}

	return authContent.AccessToken, nil
}

func browseRegistryTagsList(body io.ReadCloser, versions map[string]semver.Version, patterns map[string]semver.Pattern) error {
	done := make(chan struct{})
	versionsStream := make(chan interface{}, runtime.NumCPU())

	go func() {
		defer close(done)

		for tag := range versionsStream {
			tagVersion, err := semver.Parse(*(tag.(*string)))
			if err != nil {
				continue
			}

			model.CheckPatternsMatching(versions, patterns, tagVersion)
		}
	}()

	if err := httpjson.Stream(body, func() interface{} {
		return new(string)
	}, versionsStream, "tags"); err != nil {
		return fmt.Errorf("unable to read tags: %s", err)
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
