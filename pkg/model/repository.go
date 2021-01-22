package model

import (
	"fmt"
	"strings"
)

// RepositoryKind defines constant for repository types
type RepositoryKind int

const (
	githubURL = "https://github.com"
)

const (
	// Github repository kind
	Github RepositoryKind = iota
	// Helm repository kind
	Helm
)

var (
	// RepositoryKindValues string values
	RepositoryKindValues = []string{"github", "helm"}
)

var (
	// NoneRepository is an undefined repository
	NoneRepository = Repository{}
)

func (r RepositoryKind) String() string {
	return RepositoryKindValues[r]
}

// Repository of app
type Repository struct {
	Name     string            `json:"name"`
	ID       uint64            `json:"id"`
	Kind     RepositoryKind    `json:"kind"`
	Versions map[string]string `json:"versions"`
}

// URL format the URL of given repository with current version
func (r Repository) URL() string {
	if r.Kind == Helm {
		parts := strings.SplitN(r.Name, "@", 2)
		if len(parts) > 1 {
			return parts[1]
		}
		return r.Name
	}

	return fmt.Sprintf("%s/%s/releases/tag/%s", githubURL, r.Name, r.Versions["stable"])
}

// CompareURL format the URL of given repository compared against given version
func (r Repository) CompareURL(version string) string {
	if r.Kind == Helm {
		return r.URL()
	}

	return fmt.Sprintf("%s/%s/compare/%s...%s", githubURL, r.Name, version, r.Versions["stable"])
}

// ParseRepositoryKind parse raw string into a RepositoryKind
func ParseRepositoryKind(value string) (RepositoryKind, error) {
	for i, short := range RepositoryKindValues {
		if strings.EqualFold(short, value) {
			return RepositoryKind(i), nil
		}
	}

	return Github, fmt.Errorf("invalid value `%s` for repository kind", value)
}

// RepositoryByID sort repository by ID
type RepositoryByID []Repository

func (a RepositoryByID) Len() int           { return len(a) }
func (a RepositoryByID) Less(i, j int) bool { return a[i].ID < a[j].ID }
func (a RepositoryByID) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
