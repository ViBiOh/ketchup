package model

import (
	"fmt"
	"strings"
)

const (
	// DefaultPattern is the latest but non-beta version
	DefaultPattern = "stable"

	githubURL = "https://github.com"
)

// RepositoryKind defines constant for repository types
type RepositoryKind int

const (
	// Github repository kind
	Github RepositoryKind = iota
	// Helm repository kind
	Helm
	// Docker repository kind
	Docker
)

var (
	// RepositoryKindValues string values
	RepositoryKindValues = []string{"github", "helm", "docker"}

	// NoneRepository is an undefined repository
	NoneRepository = Repository{}
)

// ParseRepositoryKind parse raw string into a RepositoryKind
func ParseRepositoryKind(value string) (RepositoryKind, error) {
	for i, short := range RepositoryKindValues {
		if strings.EqualFold(short, value) {
			return RepositoryKind(i), nil
		}
	}

	return Github, fmt.Errorf("invalid value `%s` for repository kind", value)
}

func (r RepositoryKind) String() string {
	return RepositoryKindValues[r]
}

// Repository of app
type Repository struct {
	Versions map[string]string `json:"versions"`
	Name     string            `json:"name"`
	Part     string            `json:"part"`
	ID       uint64            `json:"id"`
	Kind     RepositoryKind    `json:"kind"`
}

// NewRepository create new Repository with initialized values
func NewRepository(id uint64, kind RepositoryKind, name, part string) Repository {
	return Repository{
		ID:       id,
		Kind:     kind,
		Name:     name,
		Part:     part,
		Versions: make(map[string]string),
	}
}

// NewGithubRepository create new Repository with initialized values
func NewGithubRepository(id uint64, name string) Repository {
	return NewRepository(id, Github, name, "")
}

// NewHelmRepository create new Repository with initialized values
func NewHelmRepository(id uint64, name, part string) Repository {
	return NewRepository(id, Helm, name, part)
}

// NewDockerRepository create new Repository with initialized values
func NewDockerRepository(id uint64, name string) Repository {
	return NewRepository(id, Docker, name, "")
}

// AddVersion adds given pattern to versions map
func (r Repository) AddVersion(pattern, version string) Repository {
	r.Versions[pattern] = version

	return r
}

// URL format the URL of given repository with current version
func (r Repository) String() string {
	switch r.Kind {
	case Helm:
		return fmt.Sprintf("%s - %s", r.Name, r.Part)
	default:
		return r.Name
	}
}

// URL format the URL of given repository with current version
func (r Repository) URL(pattern string) string {
	if r.Kind == Helm {
		parts := strings.SplitN(r.Name, "@", 2)
		if len(parts) > 1 {
			return parts[1]
		}
		return r.Name
	}

	return fmt.Sprintf("%s/%s/releases/tag/%s", githubURL, r.Name, r.Versions[pattern])
}

// CompareURL format the URL of given repository compared against given version
func (r Repository) CompareURL(version string, pattern string) string {
	if r.Kind == Helm {
		return r.URL(pattern)
	}

	return fmt.Sprintf("%s/%s/compare/%s...%s", githubURL, r.Name, r.Versions[pattern], version)
}

// RepositoryByID sort repository by ID
type RepositoryByID []Repository

func (a RepositoryByID) Len() int           { return len(a) }
func (a RepositoryByID) Less(i, j int) bool { return a[i].ID < a[j].ID }
func (a RepositoryByID) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

// RepositoryByName sort repository by Name
type RepositoryByName []Repository

func (a RepositoryByName) Len() int { return len(a) }
func (a RepositoryByName) Less(i, j int) bool {
	if a[i].Name == a[j].Name {
		return a[i].Part < a[j].Part
	}
	return a[i].Name < a[j].Name
}
func (a RepositoryByName) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
