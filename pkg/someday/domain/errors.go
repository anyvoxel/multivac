// Package domain contains core business concepts for Someday.
package domain

import (
	"errors"
	"fmt"
	"strings"
)

var (
	// ErrNotFound is returned when a Someday does not exist.
	ErrNotFound = errors.New("someday not found")
	// ErrInvalidArg is returned when input violates domain constraints.
	ErrInvalidArg = errors.New("invalid argument")
)

// ValidationError describes domain validation failures.
type ValidationError struct {
	Problems []string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: [%s]", ErrInvalidArg.Error(), strings.Join(e.Problems, ", "))
}

func (e *ValidationError) Unwrap() error {
	return ErrInvalidArg
}

func requiredFieldsError(fields ...string) error {
	problems := make([]string, 0, len(fields))
	for _, field := range fields {
		if field != "" {
			problems = append(problems, fmt.Sprintf("%s: Required value", field))
		}
	}
	if len(problems) == 0 {
		return nil
	}
	return &ValidationError{Problems: problems}
}

func invalidFieldValueError(field, value string) error {
	return &ValidationError{Problems: []string{fmt.Sprintf("%s: Unsupported value: %q", field, value)}}
}

func InvalidSortBy(value string) error {
	return invalidFieldValueError("sortBy", value)
}

func InvalidSortDir(value string) error {
	return invalidFieldValueError("sortDir", value)
}

func InvalidPaginationValue(field, value string) error {
	return invalidFieldValueError(field, value)
}
