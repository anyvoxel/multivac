package domain

import "time"

// Task is the aggregate root for task management.
type Task struct {
	ID        string
	ProjectID string

	Name        string
	Description string
	Context     string
	Details     string
	Status      Status
	Priority    Priority
	DueAt       *time.Time

	CreatedAt time.Time
	UpdatedAt time.Time
}

// NewTask creates a Task in todo status.
func NewTask(
	id, projectID, name, description, contextCategory, details string,
	priority Priority,
	dueAt *time.Time,
	now time.Time,
) (*Task, error) {
	if id == "" || projectID == "" || name == "" {
		return nil, ErrInvalidArg
	}
	if description == "" || details == "" {
		return nil, ErrInvalidArg
	}
	if contextCategory == "" {
		return nil, ErrInvalidArg
	}
	if !priority.Valid() {
		return nil, ErrInvalidArg
	}

	return &Task{
		ID:          id,
		ProjectID:   projectID,
		Name:        name,
		Description: description,
		Context:     contextCategory,
		Details:     details,
		Status:      StatusTodo,
		Priority:    priority,
		DueAt:       dueAt,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

// UpdateDetails replaces editable fields.
func (t *Task) UpdateDetails(
	name, description, contextCategory, details string,
	priority Priority,
	dueAt *time.Time,
	now time.Time,
) error {
	if t == nil {
		return ErrInvalidArg
	}
	if name == "" || description == "" || contextCategory == "" || details == "" {
		return ErrInvalidArg
	}
	if !priority.Valid() {
		return ErrInvalidArg
	}
	if dueAt != nil && dueAt.IsZero() {
		return ErrInvalidArg
	}

	t.Name = name
	t.Description = description
	t.Context = contextCategory
	t.Details = details
	t.Priority = priority
	t.DueAt = dueAt
	t.UpdatedAt = now
	return nil
}

// SetStatus updates status.
func (t *Task) SetStatus(status Status, now time.Time) error {
	if t == nil {
		return ErrInvalidArg
	}
	if !status.Valid() {
		return ErrInvalidArg
	}
	t.Status = status
	t.UpdatedAt = now
	return nil
}
