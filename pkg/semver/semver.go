package semver

import (
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

const maxVersionNumber = 1 << 16

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
	edge
	rc
	test
)

var (
	// According to https://semver.org/#spec-item-11
	semverMatcher    = regexp.MustCompile(`(?i)^(?P<prefix>[a-zA-Z-]*)(?P<major>[0-9]+)(?:\.(?P<minor>[0-9]+))?(?:\.(?P<patch>[0-9]+))?(?P<prerelease>-[a-zA-Z0-9.]+)?(?P<build>\+[a-zA-Z0-9.]+)?$`)
	semverMatchNames = semverMatcher.SubexpNames()

	nonFinalVersions = []string{"alpha", "beta", "canary", "edge", "rc", "test", "preview"}

	allowedPrefixes = []string{"v", "stable-"}
	ignoredPrefixes = []string{"sealed-secrets-"} // manually ignore historic stuff

	stableBuild = []*regexp.Regexp{
		regexp.MustCompile(`k3s[0-9]`),
	}

	ErrPrefixInvalid = errors.New("invalid prefix")
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

func ExtractName(name string) string {
	parts := strings.Split(name, "/")

	return parts[len(parts)-1]
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

	if v.Name != other.Name {
		return "Version"
	}

	return ""
}

func parseSemver(version string) (map[string]string, error) {
	matches := semverMatcher.FindStringSubmatch(version)
	if len(matches) == 0 {
		return nil, fmt.Errorf("parse version: %s", version)
	}

	output := make(map[string]string)
	for index, name := range semverMatchNames {
		if index != 0 {
			output[name] = matches[index]
		}
	}

	return output, nil
}

func Parse(version, name string) (Version, error) {
	matches, err := parseSemver(version)
	if err != nil {
		return Version{}, err
	}

	if !isPrefixAllowed(matches, name) {
		return Version{}, ErrPrefixInvalid
	}

	semver := Version{
		Name:   version,
		suffix: parseNonFinalVersion(matches),
	}

	if len(matches["major"]) >= 8 {
		return Version{}, fmt.Errorf("version major looks like a date: %s", version)
	}

	semver.major, err = strconv.ParseUint(matches["major"], 10, 64)
	if err != nil {
		return Version{}, fmt.Errorf("version major is not numeric")
	}

	if semver.major > maxVersionNumber {
		return Version{}, fmt.Errorf("version major seems a bit high")
	}

	if len(matches["minor"]) != 0 {
		semver.minor, err = strconv.ParseUint(matches["minor"], 10, 64)
		if err != nil {
			return Version{}, fmt.Errorf("version minor is not numeric")
		}
	}

	if len(matches["patch"]) != 0 {
		semver.patch, err = strconv.ParseUint(matches["patch"], 10, 64)
		if err != nil {
			return Version{}, fmt.Errorf("version patch is not numeric")
		}
	}

	return semver, nil
}

func isPrefixAllowed(matches map[string]string, allowedPrefix string) bool {
	if len(matches["prefix"]) == 0 {
		return true
	}

	if slices.Contains(ignoredPrefixes, matches["prefix"]) {
		return false
	}

	if slices.Contains(allowedPrefixes, matches["prefix"]) {
		return true
	}

	return matches["prefix"] == allowedPrefix || matches["prefix"] == allowedPrefix+"-"
}

func parseNonFinalVersion(matches map[string]string) NonFinalVersion {
	if build := matches["build"]; len(build) > 0 && !isKnownBuild(build) {
		return 0
	}

	if len(matches["prerelease"]) != 0 {
		for index, nonFinalVersion := range nonFinalVersions {
			if strings.Contains(matches["prerelease"], nonFinalVersion) {
				return NonFinalVersion(index + 1)
			}
		}

		return 0
	}

	return -1
}

func isKnownBuild(build string) bool {
	for _, knownBuild := range stableBuild {
		if knownBuild.MatchString(build) {
			return true
		}
	}

	return false
}

func safeParse(version string) Version {
	output, err := Parse(version, "")
	if err != nil {
		fmt.Println(err)
	}
	return output
}
