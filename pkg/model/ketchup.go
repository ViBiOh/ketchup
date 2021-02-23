package model

import (
	"github.com/ViBiOh/ketchup/pkg/semver"
)

var (
	// NoneKetchup is an undefined ketchup
	NoneKetchup = Ketchup{}
)

// Ketchup of app
type Ketchup struct {
	Semver     string
	Pattern    string
	Version    string
	User       User
	Repository Repository
}

// NewKetchup creates new instance
func NewKetchup(pattern, version string, repo Repository) Ketchup {
	return Ketchup{
		Pattern:    pattern,
		Version:    version,
		Repository: repo,
	}
}

// KetchupByRepositoryID sort ketchup by repository ID
type KetchupByRepositoryID []Ketchup

func (a KetchupByRepositoryID) Len() int      { return len(a) }
func (a KetchupByRepositoryID) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a KetchupByRepositoryID) Less(i, j int) bool {
	if a[i].Repository.ID == a[j].Repository.ID {
		return a[i].Pattern < a[j].Pattern
	}
	return a[i].Repository.ID < a[j].Repository.ID
}

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
	Pattern    string         `json:"pattern"`
	Version    semver.Version `json:"version"`
}

// NewRelease creates a new version from its objects
func NewRelease(repository Repository, pattern string, version semver.Version) Release {
	return Release{
		Repository: repository,
		Pattern:    pattern,
		Version:    version,
	}
}

// ReleaseByRepositoryID sort release by repository ID
type ReleaseByRepositoryID []Release

func (a ReleaseByRepositoryID) Len() int      { return len(a) }
func (a ReleaseByRepositoryID) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ReleaseByRepositoryID) Less(i, j int) bool {
	if a[i].Repository.ID == a[j].Repository.ID {
		return a[i].Pattern < a[j].Pattern
	}
	return a[i].Repository.ID < a[j].Repository.ID
}
