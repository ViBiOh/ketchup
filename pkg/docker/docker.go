package docker

import (
	"context"
	"flag"
	"fmt"
	"net/url"
	"runtime"
	"strings"
	"sync"

	"github.com/ViBiOh/httputils/v4/pkg/flags"
	"github.com/ViBiOh/httputils/v4/pkg/httpjson"
	"github.com/ViBiOh/httputils/v4/pkg/request"
	"github.com/ViBiOh/ketchup/pkg/model"
	"github.com/ViBiOh/ketchup/pkg/semver"
)

type authResponse struct {
	AccessToken string `json:"access_token"`
}

type tagsResponse struct {
	Tags []string `json:"tags"`
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
		registryURL: flags.New(prefix, "docker").Name("Registry").Default(flags.Default("Registry", "https://index.docker.io/v2/", overrides)).Label("Registry API URL").ToString(fs),
		authURL:     flags.New(prefix, "docker").Name("OAuth URL").Default(flags.Default("Registry", "https://auth.docker.io/token", overrides)).Label("Registry OAuth URL").ToString(fs),
		username:    flags.New(prefix, "docker").Name("Username").Default(flags.Default("Username", "", overrides)).Label("Registry Username").ToString(fs),
		password:    flags.New(prefix, "docker").Name("Password").Default(flags.Default("Password", "", overrides)).Label("Registry Password").ToString(fs),
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

	if !strings.Contains(repository, "/") {
		repository = fmt.Sprintf("library/%s", repository)
	}

	bearerToken, err := a.login(ctx, repository)
	if err != nil {
		return nil, fmt.Errorf("unable to login to registry: %s", err)
	}

	resp, err := request.New().Get(fmt.Sprintf("%s%s/tags/list", a.registryURL, repository)).Header("Authorization", fmt.Sprintf("Bearer %s", bearerToken)).Send(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("unable to fetch tags: %s", err)
	}

	var wg sync.WaitGroup
	versionsStream := make(chan interface{}, runtime.NumCPU())

	wg.Add(1)
	go func() {
		defer wg.Done()

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

	wg.Wait()

	return versions, nil
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
