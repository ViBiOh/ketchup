package semver

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// NonFinalVersion is a detail for temporary version
type NonFinalVersion int

// Version describe a semantic version
type Version struct {
	Name   string          `json:"name"`
	Major  uint64          `json:"major"`
	Minor  uint64          `json:"minor"`
	Patch  uint64          `json:"patch"`
	Suffix NonFinalVersion `json:"-"`
}

const (
	alpha NonFinalVersion = iota + 1
	beta
	rc
	canary
	test
)

var (
	semverMatcher = regexp.MustCompile(`(?i)^[a-zA-Z]*([0-9]+)\.([0-9]+)(?:\.([0-9]+))?(?:$|(?:[+-](.*)))`)

	nonFinalVersions = []string{"alpha", "beta", "rc", "canary", "test"}

	// NoneVersion is the empty semver
	NoneVersion = Version{}
)

// IsGreater check if current version is greater than other
func (s Version) IsGreater(other Version) bool {
	if s.Major != other.Major {
		return s.Major > other.Major
	}

	if s.Minor != other.Minor {
		return s.Minor > other.Minor
	}

	if s.Patch != other.Patch {
		return s.Patch > other.Patch
	}

	if s.Suffix != other.Suffix {
		return s.Suffix > other.Suffix
	}

	return false
}

// Compare return version diff in semver nomenclture
func (s Version) Compare(other Version) string {
	if s.Major != other.Major {
		return "Major"
	}

	if s.Minor != other.Minor {
		return "Minor"
	}

	if s.Patch != other.Patch {
		return "Patch"
	}

	if s.Suffix != other.Suffix {
		return "Suffix"
	}

	return ""
}

// Parse given version string into a version
func Parse(version string) (Version, error) {
	matches := semverMatcher.FindStringSubmatch(version)
	if len(matches) == 0 {
		return NoneVersion, fmt.Errorf("unable to parse version: %s", version)
	}

	semver := Version{
		Name: version,
	}
	var err error

	semver.Major, err = strconv.ParseUint(matches[1], 10, 64)
	if err != nil {
		return NoneVersion, fmt.Errorf("version major is not numeric")
	}

	if len(matches[2]) != 0 {
		semver.Minor, err = strconv.ParseUint(matches[2], 10, 64)
		if err != nil {
			return NoneVersion, fmt.Errorf("version minor is not numeric")
		}
	}

	if len(matches[3]) != 0 {
		semver.Patch, err = strconv.ParseUint(matches[3], 10, 64)
		if err != nil {
			return NoneVersion, fmt.Errorf("version patch is not numeric")
		}
	}

	if len(matches[4]) != 0 {
		for index, nonFinalVersion := range nonFinalVersions {
			if strings.Contains(matches[4], nonFinalVersion) {
				semver.Suffix = NonFinalVersion(index + 1)
			}
		}
	}

	return semver, nil
}
