package github

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"strings"

	"github.com/ViBiOh/httputils/v3/pkg/flags"
	"github.com/ViBiOh/httputils/v3/pkg/request"
)

const (
	apiURL = "https://api.github.com"
)

// Release describes a Github Release
type Release struct {
	Repository string `json:"repository"`
	TagName    string `json:"tag_name"`
	Body       string `json:"body"`
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
