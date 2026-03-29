package domain

import "strings"

// Status represents the lifecycle status of a Task.
type Status string

const (
	// StatusTodo is the initial status.
	StatusTodo Status = "Todo"
	// StatusInProgress means the task is being worked on.
	StatusInProgress Status = "InProgress"
	// StatusDone means the task is finished.
	StatusDone Status = "Done"
	// StatusCanceled means the task is canceled.
	StatusCanceled Status = "Canceled"
)

// Valid reports whether the status is one of the known values.
func (s Status) Valid() bool {
	switch s {
	case StatusTodo, StatusInProgress, StatusDone, StatusCanceled:
		return true
	default:
		return false
	}
}

// ParseStatus parses a status string.
// It accepts both canonical (CamelCase) and legacy (lowercase/snake_case) values.
func ParseStatus(v string) (Status, bool) {
	switch strings.ToLower(strings.ReplaceAll(v, "_", "")) {
	case "todo":
		return StatusTodo, true
	case "inprogress":
		return StatusInProgress, true
	case "done":
		return StatusDone, true
	case "canceled":
		return StatusCanceled, true
	default:
		return "", false
	}
}
