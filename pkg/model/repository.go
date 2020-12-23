package model

import (
	"fmt"
	"strings"
)

// RepositoryType defines constant for repository types
type RepositoryType int

const (
	// Github repository type
	Github RepositoryType = iota
	// Helm repository type
	Helm
)

var (
	// RepositoryTypeValues string values
	RepositoryTypeValues = []string{"github", "helm"}
)

var (
	// NoneRepository is an undefined repository
	NoneRepository = Repository{}
)

func (r RepositoryType) String() string {
	return RepositoryTypeValues[r]
}

// Repository of app
type Repository struct {
	Name    string         `json:"name"`
	Version string         `json:"version"`
	ID      uint64         `json:"id"`
	Type    RepositoryType `json:"type"`
}

// ParseRepositoryType parse raw string into a RepositoryType
func ParseRepositoryType(value string) (RepositoryType, error) {
	for i, short := range RepositoryTypeValues {
		if strings.EqualFold(short, value) {
			return RepositoryType(i), nil
		}
	}

	return Github, fmt.Errorf("invalid value `%s` for repository type", value)
}
