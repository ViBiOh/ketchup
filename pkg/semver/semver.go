package semver

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

const (
	maxVersionNumber = 1 << 16
)

// NonFinalVersion is a detail for temporary version
type NonFinalVersion int

type Version struct {
	Name   string `json:"name"`
	major  uint64
	minor  uint64
	patch  uint64
	suffix NonFinalVersion
}

func (v Version) IsZero() bool {
	return len(v.Name) == 0
}

const (
	alpha NonFinalVersion = iota + 1
	beta
	canary
	rc
	test
)

var (
	// According to https://semver.org/#spec-item-11
	semverMatcher = regexp.MustCompile(`(?i)^[a-zA-Z]*([0-9]+)(?:\.([0-9]+))?(?:\.([0-9]+))?(-[a-zA-Z0-9.]+)?(?:\+[a-zA-Z0-9.]+)?$`)

	nonFinalVersions = []string{"alpha", "beta", "canary", "rc", "test"}
)

func (v Version) Equals(other Version) bool {
	return v.major == other.major && v.minor == other.minor && v.patch == other.patch && v.suffix == other.suffix
}

func (v Version) IsGreater(other Version) bool {
	if v.major != other.major {
		return v.major > other.major
	}

	if v.minor != other.minor {
		return v.minor > other.minor
	}

	if v.patch != other.patch {
		return v.patch > other.patch
	}

	if v.suffix != other.suffix {
		if v.suffix == -1 {
			return true
		}

		return v.suffix > other.suffix
	}

	return v.Name > other.Name
}

func (v Version) Compare(other Version) string {
	if v.major != other.major {
		return "Major"
	}

	if v.minor != other.minor {
		return "Minor"
	}

	if v.patch != other.patch {
		return "Patch"
	}

	if v.suffix != other.suffix {
		return "Suffix"
	}

	return ""
}

func Parse(version string) (Version, error) {
	matches := semverMatcher.FindStringSubmatch(version)
	if len(matches) == 0 {
		return Version{}, fmt.Errorf("parse version: %s", version)
	}

	semver := Version{
		Name:   version,
		suffix: parseNonFinalVersion(matches),
	}
	var err error

	if len(matches[1]) >= 8 {
		return Version{}, fmt.Errorf("version major looks like a date: %s", version)
	}

	semver.major, err = strconv.ParseUint(matches[1], 10, 64)
	if err != nil {
		return Version{}, fmt.Errorf("version major is not numeric")
	}

	if semver.major > maxVersionNumber {
		return Version{}, fmt.Errorf("version major seems a bit high")
	}

	if len(matches[2]) != 0 {
		semver.minor, err = strconv.ParseUint(matches[2], 10, 64)
		if err != nil {
			return Version{}, fmt.Errorf("version minor is not numeric")
		}
	}

	if len(matches[3]) != 0 {
		semver.patch, err = strconv.ParseUint(matches[3], 10, 64)
		if err != nil {
			return Version{}, fmt.Errorf("version patch is not numeric")
		}
	}

	return semver, nil
}

func parseNonFinalVersion(matches []string) NonFinalVersion {
	if len(matches[4]) == 0 {
		return -1
	}

	for index, nonFinalVersion := range nonFinalVersions {
		if strings.Contains(matches[4], nonFinalVersion) {
			return NonFinalVersion(index + 1)
		}
	}

	return 0
}

func safeParse(version string) Version {
	output, err := Parse(version)
	if err != nil {
		fmt.Println(err)
	}
	return output
}
