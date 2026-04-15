package domain

import "time"

// Someday is the aggregate root for someday/maybe management.
type Someday struct {
	ID          string
	Name        string
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// NewSomeday creates a someday item.
func NewSomeday(id, name, description string, now time.Time) (*Someday, error) {
	if err := requiredFieldsError(
		requiredField(id, "id"),
		requiredField(name, "name"),
	); err != nil {
		return nil, err
	}

	return &Someday{
		ID:          id,
		Name:        name,
		Description: description,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

// UpdateDetails replaces someday core fields.
func (s *Someday) UpdateDetails(name, description string, now time.Time) error {
	if s == nil {
		return ErrInvalidArg
	}
	if err := requiredFieldsError(requiredField(name, "name")); err != nil {
		return err
	}

	s.Name = name
	s.Description = description
	s.UpdatedAt = now
	return nil
}

func requiredField(value, field string) string {
	if value == "" {
		return field
	}
	return ""
}
