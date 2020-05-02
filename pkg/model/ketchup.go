package model

import "github.com/ViBiOh/ketchup/pkg/github"

var (
	// NoneKetchup is an undefined ketchup
	NoneKetchup = Ketchup{}
)

// Ketchup of app
type Ketchup struct {
	Version    string
	Semver     string
	Repository Repository
	User       User
}

// KetchupByRepositoryID sort ketchup by repository ID
type KetchupByRepositoryID []Ketchup

func (a KetchupByRepositoryID) Len() int           { return len(a) }
func (a KetchupByRepositoryID) Less(i, j int) bool { return a[i].Repository.ID < a[j].Repository.ID }
func (a KetchupByRepositoryID) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

// Release is when new version is out
type Release struct {
	Repository Repository
	Release    github.Release
}

// ReleaseByRepositoryID sort release by repository ID
type ReleaseByRepositoryID []Release

func (a ReleaseByRepositoryID) Len() int           { return len(a) }
func (a ReleaseByRepositoryID) Less(i, j int) bool { return a[i].Repository.ID < a[j].Repository.ID }
func (a ReleaseByRepositoryID) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

// NewRelease creates a new release from its objects
func NewRelease(repository Repository, release github.Release) Release {
	return Release{
		Repository: repository,
		Release:    release,
	}
}
