package semver

import (
	"fmt"
	"regexp"
	"strconv"
)

var (
	semverMatcher = regexp.MustCompile(`(?i)^[a-zA-Z]*([0-9]+)\.([0-9]+)(?:\.([0-9]+))?`)

	// NoneVersion is the empty semver
	NoneVersion = Version{}
)

// Version describe a semantic version
type Version struct {
	Major uint64
	Minor uint64
	Patch uint64
}

// IsGreater check if current version is greater than other
func (s Version) IsGreater(other Version) bool {
	if s.Major > other.Major {
		return true
	}

	if s.Major == other.Major && s.Minor > other.Minor {
		return true
	}

	if s.Major == other.Major && s.Minor == other.Minor && s.Patch > other.Patch {
		return true
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

	return ""
}

// Parse given version string into a version
func Parse(version string) (Version, error) {
	matches := semverMatcher.FindStringSubmatch(version)
	if len(matches) == 0 {
		return NoneVersion, fmt.Errorf("unable to parse version: %s", version)
	}

	semver := Version{}
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

	return semver, nil
}
