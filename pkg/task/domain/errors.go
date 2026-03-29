// Package domain contains core business concepts for Task.
package domain

import "errors"

var (
	// ErrNotFound is returned when a Task does not exist.
	ErrNotFound = errors.New("task not found")
	// ErrInvalidArg is returned when input violates domain constraints.
	ErrInvalidArg = errors.New("invalid argument")
)
