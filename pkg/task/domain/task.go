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
	if err := requiredFieldsError(
		requiredField(id, "id"),
		requiredField(name, "name"),
		requiredField(contextCategory, "context"),
	); err != nil {
		return nil, err
	}
	if !priority.Valid() {
		return nil, invalidFieldValueError("priority", string(priority))
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
	projectID, name, description, contextCategory, details string,
	priority Priority,
	dueAt *time.Time,
	now time.Time,
) error {
	if t == nil {
		return ErrInvalidArg
	}
	if err := requiredFieldsError(
		requiredField(name, "name"),
		requiredField(contextCategory, "context"),
	); err != nil {
		return err
	}
	if !priority.Valid() {
		return invalidFieldValueError("priority", string(priority))
	}
	if dueAt != nil && dueAt.IsZero() {
		return ErrInvalidArg
	}

	t.ProjectID = projectID
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
		return invalidFieldValueError("status", string(status))
	}
	t.Status = status
	t.UpdatedAt = now
	return nil
}

func requiredField(value, field string) string {
	if value == "" {
		return field
	}
	return ""
}
