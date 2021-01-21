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
	Name string         `json:"name"`
	ID   uint64         `json:"id"`
	Kind RepositoryKind `json:"kind"`
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
