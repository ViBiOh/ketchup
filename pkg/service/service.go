package service

import (
	"errors"
	"strings"
)

var (
	// ErrInvalid occurs when checks fails
	ErrInvalid = errors.New("invalid")

	// ErrNotFound occurs when somehint is not found
	ErrNotFound = errors.New("not found")

	// ErrInternalError occurs when shit happens
	ErrInternalError = errors.New("internal error")
)

// ConcatError concat errors to a single string
func ConcatError(errs []error) error {
	if len(errs) == 0 {
		return nil
	}

	values := make([]string, len(errs))
	for index, err := range errs {
		values[index] = err.Error()
	}

	return errors.New(strings.Join(values, ", "))
}
