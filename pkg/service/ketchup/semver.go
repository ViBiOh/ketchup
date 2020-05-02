package ketchup

import (
	"regexp"

	"github.com/ViBiOh/ketchup/pkg/model"
)

var (
	semverMatcher = regexp.MustCompile(`(?i)(?:v)?([0-9]+)\.([0-9]+)\.([0-9]+)(?:\+.+)?`)
)

func computeSemver(item model.Ketchup) string {
	if item.Version == item.Repository.Version {
		return ""
	}

	repositorySemver := semverMatcher.FindStringSubmatch(item.Repository.Version)
	ketchupSemver := semverMatcher.FindStringSubmatch(item.Version)

	if len(ketchupSemver) != 4 || len(repositorySemver) != 4 {
		return ""
	}

	if repositorySemver[1] != ketchupSemver[1] {
		return "Major"
	}

	if repositorySemver[2] != ketchupSemver[2] {
		return "Minor"
	}

	return "Patch"
}
