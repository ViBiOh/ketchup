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
	User       User
	Semver     string
	Version    string
	Repository Repository
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

	if first.Repository.Type != second.Repository.Type {
		return first.Repository.Type < second.Repository.Type
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
