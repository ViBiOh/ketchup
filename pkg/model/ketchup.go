package model

import (
	"fmt"
	"strings"

	"github.com/ViBiOh/ketchup/pkg/semver"
)

// KetchupFrequency defines constant for ketchup frequency
type KetchupFrequency int

const (
	// None frequency
	None KetchupFrequency = iota
	// Daily frequency
	Daily
	// Weekly frequency (on Monday)
	Weekly
)

var (
	// KetchupFrequencyValues string values
	KetchupFrequencyValues = []string{"None", "Daily", "Weekly"}

	// NoneKetchup is an undefined ketchup
	NoneKetchup = Ketchup{}
)

// ParseKetchupFrequency parse raw string into a KetchupFrequency
func ParseKetchupFrequency(value string) (KetchupFrequency, error) {
	for i, short := range KetchupFrequencyValues {
		if strings.EqualFold(short, value) {
			return KetchupFrequency(i), nil
		}
	}

	return Daily, fmt.Errorf("invalid value `%s` for ketchup frequency", value)
}

func (r KetchupFrequency) String() string {
	return KetchupFrequencyValues[r]
}

// Ketchup of app
type Ketchup struct {
	Semver     string
	Pattern    string
	Version    string
	User       User
	Repository Repository
	Frequency  KetchupFrequency
}

// NewKetchup creates new instance
func NewKetchup(pattern, version string, frequency KetchupFrequency, repo Repository) Ketchup {
	return Ketchup{
		Pattern:    pattern,
		Version:    version,
		Frequency:  frequency,
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
		if first.Repository.Name != second.Repository.Name {
			return first.Repository.Name < second.Repository.Name
		}

		if first.Repository.Name != second.Repository.Name {
			return first.Repository.Part < second.Repository.Part
		}
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
