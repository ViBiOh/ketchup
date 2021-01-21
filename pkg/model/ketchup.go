package model

import (
	"fmt"
	"strings"

	"github.com/ViBiOh/ketchup/pkg/semver"
)

var (
	// NoneKetchup is an undefined ketchup
	NoneKetchup = Ketchup{}
)

// Ketchup of app
type Ketchup struct {
	User       User
	Semver     string
	Kind       string
	Upstream   string
	Current    string
	Repository Repository
}

// URL format the URL of given repository with current version
func (k Ketchup) URL() string {
	if k.Repository.Kind == Helm {
		parts := strings.SplitN(k.Repository.Name, "@", 2)
		if len(parts) > 1 {
			return parts[1]
		}
		return k.Repository.Name
	}

	return fmt.Sprintf("%s/%s/releases/tag/%s", githubURL, k.Repository.Name, k.Current)
}

// CompareURL format the URL of given ketchup compared against upstream version
func (k Ketchup) CompareURL() string {
	if k.Repository.Kind == Helm {
		return k.URL()
	}

	return fmt.Sprintf("%s/%s/compare/%s...%s", githubURL, k.Repository.Name, k.Upstream, k.Current)
}

// KetchupByRepositoryID sort ketchup by repository ID
type KetchupByRepositoryID []Ketchup

func (a KetchupByRepositoryID) Len() int           { return len(a) }
func (a KetchupByRepositoryID) Less(i, j int) bool { return a[i].Repository.ID < a[j].Repository.ID }
func (a KetchupByRepositoryID) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

// KetchupByPriority sort ketchup by priority (outdated first, name then)
type KetchupByPriority []Ketchup

func (a KetchupByPriority) Len() int { return len(a) }
func (a KetchupByPriority) Less(i, j int) bool {
	first := a[i]
	second := a[j]

	if first.Repository.Kind != second.Repository.Kind {
		return first.Repository.Kind < second.Repository.Kind
	}

	if first.Semver == second.Semver {
		return first.Repository.Name < second.Repository.Name
	}

	if first.Semver != "" && second.Semver == "" {
		return true
	}

	if first.Semver == "" && second.Semver != "" {
		return false
	}

	return first.Semver < second.Semver
}
func (a KetchupByPriority) Swap(i, j int) { a[i], a[j] = a[j], a[i] }

// Release is when new version is out
type Release struct {
	Repository Repository     `json:"repository"`
	Version    semver.Version `json:"version"`
}

// ReleaseByRepositoryID sort release by repository ID
type ReleaseByRepositoryID []Release

func (a ReleaseByRepositoryID) Len() int           { return len(a) }
func (a ReleaseByRepositoryID) Less(i, j int) bool { return a[i].Repository.ID < a[j].Repository.ID }
func (a ReleaseByRepositoryID) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

// NewRelease creates a new version from its objects
func NewRelease(repository Repository, version semver.Version) Release {
	return Release{
		Repository: repository,
		Version:    version,
	}
}
