package domain

import "time"

// Inbox is the aggregate root for inbox management.
type Inbox struct {
	ID          string
	Name        string
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// NewInbox creates an inbox.
func NewInbox(id, name, description string, now time.Time) (*Inbox, error) {
	if err := requiredFieldsError(
		requiredField(id, "id"),
		requiredField(name, "name"),
	); err != nil {
		return nil, err
	}

	return &Inbox{
		ID:          id,
		Name:        name,
		Description: description,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

// UpdateDetails replaces inbox core fields.
func (i *Inbox) UpdateDetails(name, description string, now time.Time) error {
	if i == nil {
		return ErrInvalidArg
	}
	if err := requiredFieldsError(requiredField(name, "name")); err != nil {
		return err
	}

	i.Name = name
	i.Description = description
	i.UpdatedAt = now
	return nil
}

func requiredField(value, field string) string {
	if value == "" {
		return field
	}
	return ""
}
