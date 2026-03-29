package domain

import "strings"

// Priority represents the urgency/importance of a Task.
type Priority string

const (
	// PriorityLow is the lowest urgency.
	PriorityLow Priority = "Low"
	// PriorityMedium is the default urgency.
	PriorityMedium Priority = "Medium"
	// PriorityHigh is high urgency.
	PriorityHigh Priority = "High"
	// PriorityP0 is the highest urgency.
	PriorityP0 Priority = "P0"
)

// Valid reports whether the priority is one of the known values.
func (p Priority) Valid() bool {
	switch p {
	case PriorityLow, PriorityMedium, PriorityHigh, PriorityP0:
		return true
	default:
		return false
	}
}

// ParsePriority parses a priority string.
// It accepts both canonical (CamelCase) and legacy (lowercase) values.
func ParsePriority(v string) (Priority, bool) {
	switch strings.ToLower(v) {
	case "low":
		return PriorityLow, true
	case "medium":
		return PriorityMedium, true
	case "high":
		return PriorityHigh, true
	case "p0":
		return PriorityP0, true
	default:
		return "", false
	}
}
