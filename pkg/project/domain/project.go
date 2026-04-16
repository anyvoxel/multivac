package domain

import (
	"net/url"
	"time"
)

type Link struct {
	Label string
	URL   string
}

// Project is the aggregate root for project management.
type Project struct {
	ID           string
	Name         string
	Goal         string
	Principles   string
	VisionResult string
	Description  string
	Links        []Link
	Status       Status

	StartedAt   *time.Time
	CompletedAt *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// NewProject creates a Project in draft status.
func NewProject(id, name, goal, principles, visionResult, description string, links []Link, now time.Time) (*Project, error) {
	if err := requiredFieldsError(
		requiredField(id, "id"),
		requiredField(name, "name"),
		requiredField(goal, "goal"),
		requiredField(principles, "principles"),
		requiredField(visionResult, "visionResult"),
		requiredField(description, "description"),
	); err != nil {
		return nil, err
	}

	if err := validateLinks(links); err != nil {
		return nil, err
	}

	return &Project{
		ID:           id,
		Name:         name,
		Goal:         goal,
		Principles:   principles,
		VisionResult: visionResult,
		Description:  description,
		Links:        cloneLinks(links),
		Status:       StatusDraft,
		CreatedAt:    now,
		UpdatedAt:    now,
	}, nil
}

// UpdateDetails replaces core textual fields.
func (p *Project) UpdateDetails(name, goal, principles, visionResult, description string, links []Link, now time.Time) error {
	if p == nil {
		return ErrInvalidArg
	}
	if err := requiredFieldsError(
		requiredField(name, "name"),
		requiredField(goal, "goal"),
		requiredField(principles, "principles"),
		requiredField(visionResult, "visionResult"),
		requiredField(description, "description"),
	); err != nil {
		return err
	}
	if err := validateLinks(links); err != nil {
		return err
	}
	p.Name = name
	p.Goal = goal
	p.Principles = principles
	p.VisionResult = visionResult
	p.Description = description
	p.Links = cloneLinks(links)
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
	// Minimal domain rules: manage timestamps based on status.
	if status == StatusActive {
		if p.StartedAt == nil {
			p.StartedAt = ptrTime(now)
		}
		p.CompletedAt = nil
	}
	if status == StatusCompleted {
		if p.CompletedAt == nil {
			p.CompletedAt = ptrTime(now)
		}
	}
	p.Status = status
	p.UpdatedAt = now
	return nil
}

func ptrTime(t time.Time) *time.Time {
	return &t
}

func validateLinks(links []Link) error {
	for i, link := range links {
		if link.Label == "" {
			return &ValidationError{Problems: []string{`links[` + itoa(i) + `].label: Required value`}}
		}
		if link.URL == "" {
			return &ValidationError{Problems: []string{`links[` + itoa(i) + `].url: Required value`}}
		}
		u, err := url.ParseRequestURI(link.URL)
		if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
			return invalidFieldValueError(`links[`+itoa(i)+`].url`, link.URL)
		}
	}
	return nil
}

func cloneLinks(links []Link) []Link {
	if len(links) == 0 {
		return nil
	}
	out := make([]Link, len(links))
	copy(out, links)
	return out
}

func itoa(v int) string {
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

func requiredField(value, field string) string {
	if value == "" {
		return field
	}
	return ""
}
