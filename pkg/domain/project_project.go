package domain

import (
	"fmt"
	"net/url"
	"strings"
	"time"
)

type ProjectReference struct {
	Title string `json:"title"`
	URL   string `json:"URL"`
}

type ProjectGoal struct {
	Title       string     `json:"title"`
	CreatedAt   time.Time  `json:"createdAt"`
	CompletedAt *time.Time `json:"completedAt,omitempty"`
}

// Project is the aggregate root for project management.
type Project struct {
	ID          string
	Title       string
	Goals       []ProjectGoal
	Description string
	References  []ProjectReference
	Status      Status

	StartAt     *time.Time
	CompletedAt *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// NewProject creates a Project in draft status.
func NewProject(id, title string, goals []ProjectGoal, description string, references []ProjectReference, now time.Time) (*Project, error) {
	if err := requiredFieldsError(
		requiredProjectField(id, "id"),
		requiredProjectField(title, "title"),
		requiredProjectField(description, "description"),
	); err != nil {
		return nil, err
	}
	if err := validateProjectGoals(goals); err != nil {
		return nil, err
	}
	if err := validateProjectReferences(references); err != nil {
		return nil, err
	}

	return &Project{
		ID:          id,
		Title:       title,
		Goals:       normalizeProjectGoals(goals, now),
		Description: description,
		References:  cloneProjectReferences(references),
		Status:      StatusDraft,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

// UpdateDetails replaces core textual fields.
func (p *Project) UpdateDetails(title string, goals []ProjectGoal, description string, references []ProjectReference, now time.Time) error {
	if p == nil {
		return ErrInvalidArg
	}
	if err := requiredFieldsError(
		requiredProjectField(title, "title"),
		requiredProjectField(description, "description"),
	); err != nil {
		return err
	}
	if err := validateProjectGoals(goals); err != nil {
		return err
	}
	if err := validateProjectReferences(references); err != nil {
		return err
	}
	p.Title = title
	p.Goals = normalizeProjectGoals(goals, now)
	p.Description = description
	p.References = cloneProjectReferences(references)
	p.UpdatedAt = now
	return nil
}

// SetStatus changes the lifecycle status and maintains timestamps.
func (p *Project) SetStatus(status Status, now time.Time) error {
	if p == nil {
		return ErrInvalidArg
	}
	if !status.Valid() {
		return invalidFieldValueError("status", string(status))
	}
	if !p.canTransitionTo(status) {
		return &ValidationError{Problems: []string{fmt.Sprintf("status: Unsupported transition: %q -> %q", p.Status, status)}}
	}

	switch status {
	case StatusActive:
		if p.StartAt == nil {
			p.StartAt = ptrTime(now)
		}
		p.CompletedAt = nil
	case StatusCompleted:
		if p.CompletedAt == nil {
			p.CompletedAt = ptrTime(now)
		}
	case StatusDraft, StatusHold:
		p.CompletedAt = nil
	}

	p.Status = status
	p.UpdatedAt = now
	return nil
}

func (p *Project) canTransitionTo(next Status) bool {
	switch p.Status {
	case StatusDraft:
		return next == StatusDraft || next == StatusActive || next == StatusHold
	case StatusActive:
		return next == StatusActive || next == StatusCompleted || next == StatusHold
	case StatusHold:
		return next == StatusHold || next == StatusActive
	case StatusCompleted:
		return next == StatusCompleted
	default:
		return false
	}
}

func ptrTime(t time.Time) *time.Time {
	return &t
}

func validateProjectReferences(references []ProjectReference) error {
	for i, reference := range references {
		if strings.TrimSpace(reference.Title) == "" {
			return &ValidationError{Problems: []string{`references[` + itoaProject(i) + `].title: Required value`}}
		}
		if reference.URL == "" {
			return &ValidationError{Problems: []string{`references[` + itoaProject(i) + `].url: Required value`}}
		}
		u, err := url.ParseRequestURI(reference.URL)
		if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
			return invalidFieldValueError(`references[`+itoaProject(i)+`].url`, reference.URL)
		}
	}
	return nil
}

func cloneProjectReferences(references []ProjectReference) []ProjectReference {
	if len(references) == 0 {
		return nil
	}
	out := make([]ProjectReference, len(references))
	copy(out, references)
	return out
}

func validateProjectGoals(goals []ProjectGoal) error {
	for i, goal := range goals {
		title := strings.TrimSpace(goal.Title)
		if title == "" {
			return &ValidationError{Problems: []string{`goals[` + itoaProject(i) + `].title: Required value`}}
		}
		if strings.Contains(title, "\n") || strings.Contains(title, "\r") {
			return invalidFieldValueError(`goals[`+itoaProject(i)+`].title`, goal.Title)
		}
	}
	return nil
}

func normalizeProjectGoals(goals []ProjectGoal, now time.Time) []ProjectGoal {
	if len(goals) == 0 {
		return nil
	}
	out := make([]ProjectGoal, 0, len(goals))
	for _, goal := range goals {
		title := strings.TrimSpace(goal.Title)
		g := ProjectGoal{Title: title}
		if goal.CreatedAt.IsZero() {
			g.CreatedAt = now
		} else {
			g.CreatedAt = goal.CreatedAt
		}
		if goal.CompletedAt != nil && !goal.CompletedAt.IsZero() {
			t := *goal.CompletedAt
			g.CompletedAt = &t
		}
		out = append(out, g)
	}
	return out
}

func itoaProject(v int) string {
	if v == 0 {
		return "0"
	}
	buf := [20]byte{}
	pos := len(buf)
	for v > 0 {
		pos--
		buf[pos] = byte('0' + v%10)
		v /= 10
	}
	return string(buf[pos:])
}

func requiredProjectField(value, field string) string {
	if value == "" {
		return field
	}
	return ""
}
