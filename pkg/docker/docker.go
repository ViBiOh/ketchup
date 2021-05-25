package docker

import (
	"context"
	"flag"
	"fmt"
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
	registryURL *string
	authURL     *string
	username    *string
	password    *string
}

type app struct {
	registryURL string
	authURL     string
	username    string
	password    string
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
		registryURL: strings.TrimSpace(*config.registryURL),
		authURL:     strings.TrimSpace(*config.authURL),
		username:    strings.TrimSpace(*config.username),
		password:    strings.TrimSpace(*config.password),
	}
}

func (a app) LatestVersions(repository string, patterns []string) (map[string]semver.Version, error) {
	ctx := context.Background()

	versions, compiledPatterns, err := model.PreparePatternMatching(patterns)
	if err != nil {
		return nil, fmt.Errorf("unable to prepare pattern matching: %s", err)
	}

	var registry, bearerToken string

	parts := strings.Split(repository, "/")
	if len(parts) < 3 {
		if len(parts) == 1 {
			repository = fmt.Sprintf("library/%s", repository)
		}

		token, err := a.login(ctx, repository)
		if err != nil {
			return nil, fmt.Errorf("unable to authenticate to docker hub: %s", err)
		}

		bearerToken = token
		registry = registryURL
	} else {
		registry = fmt.Sprintf("https://%s", parts[0])
		repository = strings.Join(parts[1:], "/")
	}

	url := fmt.Sprintf("%s/v2/%s/tags/list", registry, repository)

	for len(url) != 0 {
		req := request.New().Get(url)

		if registry == registryURL {
			req.Header("Authorization", fmt.Sprintf("Bearer %s", bearerToken))
		}

		resp, err := req.Send(ctx, nil)
		if err != nil {
			return nil, fmt.Errorf("unable to fetch tags: %s", err)
		}

		done := make(chan struct{})
		versionsStream := make(chan interface{}, runtime.NumCPU())

		go func() {
			defer close(done)

			for tag := range versionsStream {
				tagVersion, err := semver.Parse(*(tag.(*string)))
				if err != nil {
					continue
				}

				model.CheckPatternsMatching(versions, compiledPatterns, tagVersion)
			}
		}()

		if err := httpjson.Stream(resp.Body, func() interface{} {
			return new(string)
		}, versionsStream, "tags"); err != nil {
			return nil, fmt.Errorf("unable to read tags: %s", err)
		}

		<-done

		url = getNextURL(resp, registry)
	}

	return versions, nil
}

func getNextURL(resp *http.Response, registry string) string {
	link := resp.Header.Get("link")
	if len(link) == 0 {
		return ""
	}

	parts := strings.Split(link, ";")
	path := strings.Trim(strings.Trim(parts[0], "<"), ">")

	return fmt.Sprintf("%s%s", registry, path)
}

func (a app) login(ctx context.Context, repository string) (string, error) {
	values := url.Values{}
	values.Set("grant_type", "password")
	values.Set("service", "registry.docker.io")
	values.Set("client_id", "ketchup")
	values.Set("scope", fmt.Sprintf("repository:%s:pull", repository))
	values.Set("username", a.username)
	values.Set("password", a.password)

	resp, err := request.New().Post(a.authURL).Form(ctx, values)
	if err != nil {
		return "", fmt.Errorf("unable to authenticate to `%s`: %s", a.authURL, err)
	}

	var authContent authResponse
	if err := httpjson.Read(resp, &authContent); err != nil {
		return "", fmt.Errorf("unable to read auth token: %s", err)
	}

	return authContent.AccessToken, nil
}
