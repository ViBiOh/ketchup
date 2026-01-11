package model

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

const (
	DefaultPattern = "stable"

	githubURL = "https://github.com"
)

//go:generate stringer -type=RepositoryKind
type RepositoryKind int

const (
	Github RepositoryKind = iota
	Helm
	Docker
	NPM
	Pypi
)

var ErrUnknownRepositoryKind = errors.New("unknown repository kind")

func ParseRepositoryKind(value string) (RepositoryKind, error) {
	var previous, current uint8

	for i := 1; i < len(_RepositoryKind_index); i++ {
		current = _RepositoryKind_index[i]

		if strings.EqualFold(_RepositoryKind_name[previous:current], value) {
			return RepositoryKind(i - 1), nil
		}

		previous = current
	}

	return Github, ErrUnknownRepositoryKind
}

func (r RepositoryKind) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString(`"`)
	buffer.WriteString(r.String())
	buffer.WriteString(`"`)
	return buffer.Bytes(), nil
}

func (r *RepositoryKind) UnmarshalJSON(b []byte) error {
	var strValue string
	err := json.Unmarshal(b, &strValue)
	if err != nil {
		return fmt.Errorf("unmarshal repository kind: %w", err)
	}

	value, err := ParseRepositoryKind(strValue)
	if err != nil {
		return fmt.Errorf("parse repository kind: %w", err)
	}

	*r = value
	return nil
}

type Repository struct {
	Versions map[string]string `json:"versions"`
	Name     string            `json:"name"`
	Part     string            `json:"part"`
	ID       Identifier        `json:"id"`
	Kind     RepositoryKind    `json:"kind"`
}

func NewRepository(id Identifier, kind RepositoryKind, name, part string) Repository {
	return Repository{
		ID:       id,
		Kind:     kind,
		Name:     name,
		Part:     part,
		Versions: make(map[string]string),
	}
}

func NewEmptyRepository() Repository {
	return Repository{
		Versions: make(map[string]string),
	}
}

func NewGithubRepository(id Identifier, name string) Repository {
	return NewRepository(id, Github, name, "")
}

func NewHelmRepository(id Identifier, name, part string) Repository {
	return NewRepository(id, Helm, name, part)
}

func (r Repository) AddVersion(pattern, version string) Repository {
	r.Versions[pattern] = version

	return r
}

func (r Repository) IsZero() bool {
	return r.ID.IsZero()
}

func (r Repository) String() string {
	switch r.Kind {
	case Helm:
		return fmt.Sprintf("%s - %s", r.Name, r.Part)
	default:
		return r.Name
	}
}

func (r Repository) URL(pattern string) string {
	return r.VersionURL(r.Versions[pattern])
}

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

func (r Repository) CompareURL(version, pattern string) string {
	switch r.Kind {
	case Github:
		return fmt.Sprintf("%s/%s/compare/%s...%s", githubURL, r.Name, version, r.Versions[pattern])
	default:
		return r.URL(pattern)
	}
}
