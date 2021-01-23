package github

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/ViBiOh/httputils/v3/pkg/flags"
	"github.com/ViBiOh/httputils/v3/pkg/logger"
	"github.com/ViBiOh/httputils/v3/pkg/request"
	"github.com/ViBiOh/ketchup/pkg/semver"
)

var (
	apiURL = "https://api.github.com"
)

// Tag describes a Github Tag
type Tag struct {
	Name string `json:"name"`
}

// App of package
type App interface {
	LatestVersion(string, []string) (map[string]semver.Version, error)
}

// Config of package
type Config struct {
	token *string
}

type app struct {
	token string
}

// Flags adds flags for configuring package
func Flags(fs *flag.FlagSet, prefix string) Config {
	return Config{
		token: flags.New(prefix, "github").Name("Token").Default("").Label("OAuth Token").ToString(fs),
	}
}

// New creates new App from Config
func New(config Config) App {
	return app{
		token: strings.TrimSpace(*config.token),
	}
}

func (a app) newClient() *request.Request {
	return request.New().Header("Authorization", fmt.Sprintf("token %s", a.token)).WithClient(http.Client{
		Timeout: 30 * time.Second,
		CheckRedirect: func(r *http.Request, via []*http.Request) error {
			logger.Info("Redirect from %s to %s", via[len(via)-1].URL.Path, r.URL.Path)
			return nil
		},
	})
}

func (a app) LatestVersion(repository string, patterns []string) (map[string]semver.Version, error) {
	output := make(map[string]semver.Version)
	for _, pattern := range patterns {
		output[pattern] = semver.NoneVersion
	}

	page := 1
	req := a.newClient()
	for {
		resp, err := req.Get(fmt.Sprintf("%s/repos/%s/tags?per_page=100&page=%d", apiURL, repository, page)).Send(context.Background(), nil)
		if err != nil {
			return nil, fmt.Errorf("unable to list page %d of tags: %s", page, err)
		}

		payload, err := request.ReadBodyResponse(resp)
		if err != nil {
			return nil, fmt.Errorf("unable to read page %d tags body `%s`: %s", page, payload, err)
		}

		var tags []Tag
		if err := json.Unmarshal(payload, &tags); err != nil {
			return nil, fmt.Errorf("unable to parse page %d tags body: %s", page, err)
		}

		for _, tag := range tags {
			tagVersion, err := semver.Parse(tag.Name)
			if err != nil {
				continue
			}

			for pattern, patternVersion := range output {
				if tagVersion.Match(pattern) && tagVersion.IsGreater(patternVersion) {
					output[pattern] = tagVersion
				}
			}
		}

		if !hasNext(resp) {
			break
		}

		page++
	}

	return output, nil
}

func hasNext(resp *http.Response) bool {
	for _, value := range resp.Header.Values("Link") {
		if strings.Contains(value, `rel="next"`) {
			return true
		}
	}

	return false
}
