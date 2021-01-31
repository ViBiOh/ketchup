package semver

import (
	"errors"
	"fmt"
)

type operation int

const (
	equal operation = iota
	notEqual
	greaterThan
	greaterOrEqual
	lowerThan
	lowerOrEqual
)

var (
	// NonePattern for empty response
	NonePattern = Pattern{}

	operationSigns = []string{"=", "!=", ">", ">=", "<", "<="}
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
	constraints []constraint
}

// NewPattern creates new pattern instance
func NewPattern(constraints ...constraint) Pattern {
	return Pattern{
		constraints: constraints,
	}
}

// ParsePattern parse given constraint to extract pattern matcher
func ParsePattern(pattern string) (Pattern, error) {
	if len(pattern) == 0 {
		return NonePattern, errors.New("pattern is empty")
	}

	if pattern == "latest" {
		return NewPattern(newConstraint(safeParse("0.0.0-0"), greaterOrEqual)), nil
	}

	if pattern == "stable" {
		return NewPattern(newConstraint(safeParse("0.0.0"), greaterOrEqual)), nil
	}

	return NonePattern, fmt.Errorf("unable to parse pattern `%s`", pattern)
}

// Check verifies is given version match Pattern
func (p Pattern) Check(version Version) bool {
	for _, constraint := range p.constraints {
		if constraint.version.suffix == -1 && version.suffix != -1 {
			return false
		}

		switch constraint.comparator {
		case equal:
			if !version.Equals(constraint.version) {
				return false
			}
		case notEqual:
			if version.Equals(constraint.version) {
				return false
			}
		case greaterThan:
			if !version.IsGreater(constraint.version) {
				return false
			}
		case greaterOrEqual:
			if !version.IsGreater(constraint.version) && !version.Equals(constraint.version) {
				return false
			}
		case lowerThan:
			if version.IsGreater(constraint.version) {
				return false
			}
		case lowerOrEqual:
			if version.IsGreater(constraint.version) || !version.Equals(constraint.version) {
				return false
			}
		default:
			fmt.Printf("comparator `%d` is not implemented\n", constraint.comparator)
			return false
		}
	}

	return true
}

func safeParsePattern(pattern string) Pattern {
	output, err := ParsePattern(pattern)
	if err != nil {
		fmt.Println(err)
	}
	return output
}
