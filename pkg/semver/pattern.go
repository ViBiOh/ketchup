package semver

import (
	"errors"
	"fmt"
)

type operation int

const (
	greaterOrEqual operation = iota
	lowerThan
)

type constraint struct {
	version    Version
	comparator operation
}

func newConstraint(version Version, comparator operation) constraint {
	return constraint{
		version:    version,
		comparator: comparator,
	}
}

// Pattern describe a pattern constraint
type Pattern struct {
	Name        string
	constraints []constraint
}

// NewPattern creates new pattern instance
func NewPattern(name string, constraints ...constraint) Pattern {
	return Pattern{
		constraints: constraints,
	}
}

// Check verifies is given version match Pattern
func (p Pattern) Check(version Version) bool {
	for _, constraint := range p.constraints {
		if constraint.version.suffix == -1 && version.suffix != -1 {
			return false
		}

		switch constraint.comparator {
		case greaterOrEqual:
			if !version.IsGreater(constraint.version) && !version.Equals(constraint.version) {
				return false
			}
		case lowerThan:
			if version.IsGreater(constraint.version) || version.Equals(constraint.version) {
				return false
			}
		}
	}

	return true
}

// ParsePattern parse given constraint to extract pattern matcher
func ParsePattern(pattern string) (Pattern, error) {
	if len(pattern) < 2 {
		return Pattern{}, errors.New("pattern is invalid")
	}

	if pattern == "latest" {
		return NewPattern(pattern, newConstraint(safeParse("0.0-0"), greaterOrEqual)), nil
	}

	if pattern == "stable" {
		return NewPattern(pattern, newConstraint(safeParse("0.0"), greaterOrEqual)), nil
	}

	version, err := Parse(pattern[1:])
	if err != nil {
		return Pattern{}, fmt.Errorf("unable to parse version in pattern: %s", err)
	}

	constraintVersionSuffix := ""
	if version.suffix != -1 {
		constraintVersionSuffix = "-0"
	}

	if pattern[0] == '^' {
		return NewPattern(pattern, newConstraint(version, greaterOrEqual), newConstraint(safeParse(fmt.Sprintf("%d.0%s", version.major+1, constraintVersionSuffix)), lowerThan)), nil
	}

	return NewPattern(pattern, newConstraint(version, greaterOrEqual), newConstraint(safeParse(fmt.Sprintf("%d.%d%s", version.major, version.minor+1, constraintVersionSuffix)), lowerThan)), nil
}
