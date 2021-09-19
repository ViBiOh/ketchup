package model

import (
	"fmt"

	"github.com/ViBiOh/ketchup/pkg/semver"
)

// PreparePatternMatching prepare maps for doing pattern matching
func PreparePatternMatching(patterns []string) (map[string]semver.Version, map[string]semver.Pattern, error) {
	versions := make(map[string]semver.Version)
	compiledPatterns := make(map[string]semver.Pattern)

	for _, pattern := range patterns {
		p, err := semver.ParsePattern(pattern)
		if err != nil {
			return nil, nil, fmt.Errorf("unable to parse pattern: %s", err)
		}

		versions[pattern] = semver.Version{}
		compiledPatterns[pattern] = p
	}

	return versions, compiledPatterns, nil
}

// CheckPatternsMatching check given version against pattern matching
func CheckPatternsMatching(versions map[string]semver.Version, patterns map[string]semver.Pattern, version semver.Version) {
	for pattern, patternVersion := range versions {
		if patterns[pattern].Check(version) && version.IsGreater(patternVersion) {
			versions[pattern] = version
		}
	}
}
