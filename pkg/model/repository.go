package model

import (
	"bytes"
	"encoding/json"
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
	// NPM repository kind
	NPM
	// Pypi repository kind
	Pypi
)

// RepositoryKindValues string values
var RepositoryKindValues = []string{"github", "helm", "docker", "npm", "pypi"}

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

// MarshalJSON marshals the enum as a quoted json string
func (r RepositoryKind) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString(`"`)
	buffer.WriteString(r.String())
	buffer.WriteString(`"`)
	return buffer.Bytes(), nil
}

// UnmarshalJSON unmarshal JSOn
func (r *RepositoryKind) UnmarshalJSON(b []byte) error {
	var strValue string
	err := json.Unmarshal(b, &strValue)
	if err != nil {
		return fmt.Errorf("unable to unmarshal event type: %s", err)
	}

	value, err := ParseRepositoryKind(strValue)
	if err != nil {
		return fmt.Errorf("unable to parse event type: %s", err)
	}

	*r = value
	return nil
}

// Repository of app
type Repository struct {
	Versions map[string]string `json:"versions"`
	Name     string            `json:"name"`
	Part     string            `json:"part"`
	ID       Identifier        `json:"id"`
	Kind     RepositoryKind    `json:"kind"`
}

// NewRepository create new Repository with initialized values
func NewRepository(id Identifier, kind RepositoryKind, name, part string) Repository {
	return Repository{
		ID:       id,
		Kind:     kind,
		Name:     name,
		Part:     part,
		Versions: make(map[string]string),
	}
}

// NewEmptyRepository create an empty Repository
func NewEmptyRepository() Repository {
	return Repository{
		Versions: make(map[string]string),
	}
}

// NewGithubRepository create new Repository with initialized values
func NewGithubRepository(id Identifier, name string) Repository {
	return NewRepository(id, Github, name, "")
}

// NewHelmRepository create new Repository with initialized values
func NewHelmRepository(id Identifier, name, part string) Repository {
	return NewRepository(id, Helm, name, part)
}

// AddVersion adds given pattern to versions map
func (r Repository) AddVersion(pattern, version string) Repository {
	r.Versions[pattern] = version

	return r
}

// IsZero return false if instance is not initialized
func (r Repository) IsZero() bool {
	return r.ID.IsZero()
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

// URL format the URL of given repository with pattern version
func (r Repository) URL(pattern string) string {
	return r.VersionURL(r.Versions[pattern])
}

// VersionURL format the URL of given repository with given version
func (r Repository) VersionURL(version string) string {
	switch r.Kind {
	case Github:
		return fmt.Sprintf("%s/%s/releases/tag/%s", githubURL, r.Name, version)
	case Helm:
		parts := strings.SplitN(r.Name, "@", 2)
		if len(parts) > 1 {
			return parts[1]
		}
		return r.Name
	case Docker:
		switch strings.Count(r.Name, "/") {
		case 0:
			return fmt.Sprintf("https://hub.docker.com/_/%s?tab=tags&page=1&ordering=last_updated&name=%s", r.Name, version)
		case 1:
			return fmt.Sprintf("https://hub.docker.com/r/%s/tags?page=1&ordering=last_updated&name=%s", r.Name, version)
		default:
			return fmt.Sprintf("https://%s", r.Name)
		}
	case NPM:
		return fmt.Sprintf("https://www.npmjs.com/package/%s/v/%s", r.Name, version)
	case Pypi:
		return fmt.Sprintf("https://pypi.org/project/%s/%s/", r.Name, version)
	default:
		return "#"
	}
}

// CompareURL format the URL of given repository compared against given version
func (r Repository) CompareURL(version string, pattern string) string {
	switch r.Kind {
	case Github:
		return fmt.Sprintf("%s/%s/compare/%s...%s", githubURL, r.Name, version, r.Versions[pattern])
	default:
		return r.URL(pattern)
	}
}
