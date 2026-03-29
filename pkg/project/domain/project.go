package domain

import (
	"time"
)

// Project is the aggregate root for project management.
type Project struct {
	ID           string
	Name         string
	Goal         string
	Principles   string
	VisionResult string
	Description  string
	Status       Status

	StartedAt   *time.Time
	CompletedAt *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// NewProject creates a Project in draft status.
func NewProject(id, name, goal, principles, visionResult, description string, now time.Time) (*Project, error) {
	if id == "" || name == "" {
		return nil, ErrInvalidArg
	}
	if goal == "" {
		return nil, ErrInvalidArg
	}
	if principles == "" {
		return nil, ErrInvalidArg
	}
	if visionResult == "" {
		return nil, ErrInvalidArg
	}
	if description == "" {
		return nil, ErrInvalidArg
	}

	return &Project{
		ID:           id,
		Name:         name,
		Goal:         goal,
		Principles:   principles,
		VisionResult: visionResult,
		Description:  description,
		Status:       StatusDraft,
		CreatedAt:    now,
		UpdatedAt:    now,
	}, nil
}

// UpdateDetails replaces core textual fields.
func (p *Project) UpdateDetails(name, goal, principles, visionResult, description string, now time.Time) error {
	if p == nil {
		return ErrInvalidArg
	}
	if name == "" || goal == "" || principles == "" || visionResult == "" || description == "" {
		return ErrInvalidArg
	}
	p.Name = name
	p.Goal = goal
	p.Principles = principles
	p.VisionResult = visionResult
	p.Description = description
	p.UpdatedAt = now
	return nil
}

// SetStatus changes the lifecycle status and maintains timestamps.
func (p *Project) SetStatus(status Status, now time.Time) error {
	if p == nil {
		return ErrInvalidArg
	}
	if !status.Valid() {
		return ErrInvalidArg
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
