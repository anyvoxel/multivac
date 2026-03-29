// Package domain contains core business concepts for Project.
package domain

import "errors"

var (
	// ErrNotFound is returned when a Project does not exist.
	ErrNotFound = errors.New("project not found")
	// ErrInvalidArg is returned when input violates domain constraints.
	ErrInvalidArg = errors.New("invalid argument")
)
