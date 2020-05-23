package github

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"strings"

	"github.com/ViBiOh/httputils/v3/pkg/flags"
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
	LatestVersion(repository string) (semver.Version, error)
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
	return request.New().Header("Authorization", fmt.Sprintf("token %s", a.token))
}

func (a app) LatestVersion(repository string) (semver.Version, error) {
	version, err := a.parseTags(repository)
	if err != nil {
		return version, fmt.Errorf("unable to retrieve version: %s", err)
	}

	return version, nil
}

func (a app) parseTags(repository string) (semver.Version, error) {
	page := 1
	var version semver.Version

	req := a.newClient()
	for {
		resp, err := req.Get(fmt.Sprintf("%s/repos/%s/tags?per_page=100&page=%d", apiURL, repository, page)).Send(context.Background(), nil)
		if err != nil {
			return version, fmt.Errorf("unable to list page %d of tags: %s", page, err)
		}

		payload, err := request.ReadBodyResponse(resp)
		if err != nil {
			return version, fmt.Errorf("unable to read page %d tags body: %s", page, err)
		}

		var tags []Tag
		if err := json.Unmarshal(payload, &tags); err != nil {
			return version, fmt.Errorf("unable to parse page %d tags body: %s", page, err)
		}

		for _, tag := range tags {
			tagVersion, err := semver.Parse(tag.Name)
			if err == nil && tagVersion.IsGreater(version) {
				version = tagVersion
			}
		}

		if !hasNext(resp) {
			break
		}

		page++
	}

	if version == semver.NoneVersion {
		return version, fmt.Errorf("unable to find semver in tags for %s", repository)
	}

	return version, nil
}

func hasNext(resp *http.Response) bool {
	for _, value := range resp.Header.Values("Link") {
		if strings.Contains(value, `rel="next"`) {
			return true
		}
	}

	return false
}
