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

// Release describes a Github Release
type Release struct {
	Repository string `json:"repository"`
	TagName    string `json:"tag_name"`
}

// Tag describes a Github Tag
type Tag struct {
	Name string `json:"name"`
}

// App of package
type App interface {
	LastRelease(repository string) (Release, error)
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

func (a app) LastRelease(repository string) (Release, error) {
	release, latestErr := a.latestRelease(repository)
	if latestErr == nil {
		return release, nil
	}

	release, tagsErr := a.parseTags(repository)
	if tagsErr == nil {
		return release, nil
	}

	return Release{}, fmt.Errorf("unable to retrieve release: %s then %s", latestErr, tagsErr)
}

func (a app) latestRelease(repository string) (Release, error) {
	var release Release

	req := a.newClient()
	resp, err := req.Get(fmt.Sprintf("%s/repos/%s/releases/latest", apiURL, repository)).Send(context.Background(), nil)
	if err != nil {
		return release, fmt.Errorf("unable to get latest release for %s: %s", repository, err)
	}

	payload, err := request.ReadBodyResponse(resp)
	if err != nil {
		return release, fmt.Errorf("unable to read release body for %s: %s", repository, err)
	}

	if err := json.Unmarshal(payload, &release); err != nil {
		return release, fmt.Errorf("unable to parse release body for %s: %s", repository, err)
	}

	release.Repository = repository

	return release, err
}

func (a app) parseTags(repository string) (Release, error) {
	release := Release{
		Repository: repository,
	}

	page := 1
	version := semver.Version{}

	req := a.newClient()
	for {
		resp, err := req.Get(fmt.Sprintf("%s/repos/%s/tags?per_page=100&page=%d", apiURL, repository, page)).Send(context.Background(), nil)
		if err != nil {
			return release, fmt.Errorf("unable to get tags %s: %s", repository, err)
		}

		payload, err := request.ReadBodyResponse(resp)
		if err != nil {
			return release, fmt.Errorf("unable to read tags body for %s: %s", repository, err)
		}

		var tags []Tag
		if err := json.Unmarshal(payload, &tags); err != nil {
			return release, fmt.Errorf("unable to parse tags body for %s: %s", repository, err)
		}

		for _, tag := range tags {
			tagVersion, err := semver.Parse(tag.Name)
			if err == nil && tagVersion.IsGreater(version) {
				version = tagVersion
				release.TagName = tag.Name
			}
		}

		if !isLastPage(resp) {
			break
		}

		page++
	}

	if len(release.TagName) == 0 {
		return release, fmt.Errorf("unable to find semver in tags for %s", repository)
	}

	return release, nil
}

func isLastPage(resp *http.Response) bool {
	for _, link := range resp.Header.Values("Link") {
		if strings.Contains(link, `rel="next"`) {
			return true
		}
	}

	return false
}
