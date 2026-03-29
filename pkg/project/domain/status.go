package domain

import "strings"

// Status represents the lifecycle status of a Project.
type Status string

const (
	// StatusDraft is the initial status.
	StatusDraft Status = "Draft"
	// StatusActive means the project is in progress.
	StatusActive Status = "Active"
	// StatusCompleted means the project is finished.
	StatusCompleted Status = "Completed"
	// StatusArchived means the project is no longer active.
	StatusArchived Status = "Archived"
)

// Valid reports whether the status is one of the known values.
func (s Status) Valid() bool {
	switch s {
	case StatusDraft, StatusActive, StatusCompleted, StatusArchived:
		return true
	default:
		return false
	}
}

// ParseStatus parses a status string.
// It accepts both canonical (CamelCase) and legacy (lowercase) values.
func ParseStatus(v string) (Status, bool) {
	switch strings.ToLower(v) {
	case "draft":
		return StatusDraft, true
	case "active":
		return StatusActive, true
	case "completed":
		return StatusCompleted, true
	case "archived":
		return StatusArchived, true
	default:
		return "", false
	}
}
