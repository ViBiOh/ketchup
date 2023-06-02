package model

import (
	"fmt"
	"strings"

	"github.com/ViBiOh/httputils/v4/pkg/sha"
	"github.com/ViBiOh/ketchup/pkg/semver"
)

type KetchupFrequency int

const (
	None KetchupFrequency = iota
	Daily
	Weekly
)

var KetchupFrequencyValues = []string{"None", "Daily", "Weekly"}

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

type Ketchup struct {
	ID               string
	Semver           string
	Pattern          string
	Version          string
	User             User
	Repository       Repository
	Frequency        KetchupFrequency
	UpdateWhenNotify bool
}

func NewKetchup(pattern, version string, frequency KetchupFrequency, updateWhenNotify bool, repo Repository) Ketchup {
	return Ketchup{
		Pattern:          pattern,
		Version:          version,
		Frequency:        frequency,
		UpdateWhenNotify: updateWhenNotify,
		Repository:       repo,
	}
}

func (k Ketchup) WithID() Ketchup {
	k.ID = sha.New(k)[:8]

	return k
}

type KetchupByRepositoryIDAndPattern []Ketchup

func (a KetchupByRepositoryIDAndPattern) Len() int      { return len(a) }
func (a KetchupByRepositoryIDAndPattern) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a KetchupByRepositoryIDAndPattern) Less(i, j int) bool {
	if a[i].Repository.ID == a[j].Repository.ID {
		return a[i].Pattern < a[j].Pattern
	}
	return a[i].Repository.ID < a[j].Repository.ID
}

type KetchupByPriority []Ketchup

func (a KetchupByPriority) Len() int { return len(a) }
func (a KetchupByPriority) Less(i, j int) bool {
	first := a[i]
	second := a[j]

	if first.Repository.Kind != second.Repository.Kind {
		return first.Repository.Kind < second.Repository.Kind
	}

	if first.Semver == second.Semver {
		if first.Repository.Name == second.Repository.Name {
			return first.Repository.Part < second.Repository.Part
		}
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

type Release struct {
	Pattern    string         `json:"pattern"`
	URL        string         `json:"url"`
	Repository Repository     `json:"repository"`
	Version    semver.Version `json:"version"`
	Updated    uint           `json:"updated"`
}

func NewRelease(repository Repository, pattern string, version semver.Version) Release {
	return Release{
		Repository: repository,
		Pattern:    pattern,
		Version:    version,
		URL:        repository.VersionURL(version.Name),
	}
}

func (r Release) SetUpdated(status uint) Release {
	r.Updated = status

	return r
}

type ReleaseByRepositoryIDAndPattern []Release

func (a ReleaseByRepositoryIDAndPattern) Len() int      { return len(a) }
func (a ReleaseByRepositoryIDAndPattern) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ReleaseByRepositoryIDAndPattern) Less(i, j int) bool {
	if a[i].Repository.ID == a[j].Repository.ID {
		return a[i].Pattern < a[j].Pattern
	}
	return a[i].Repository.ID < a[j].Repository.ID
}

type ReleaseByKindAndName []Release

func (a ReleaseByKindAndName) Len() int      { return len(a) }
func (a ReleaseByKindAndName) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ReleaseByKindAndName) Less(i, j int) bool {
	if a[i].Repository.Kind == a[j].Repository.Kind {
		if a[i].Repository.Name == a[j].Repository.Name {
			return a[i].Repository.Part < a[j].Repository.Part
		}
		return a[i].Repository.Name < a[j].Repository.Name
	}
	return a[i].Repository.Kind < a[j].Repository.Kind
}
