package domain

import "time"

type Inbox struct {
	ID          string
	Title       string
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func NewInbox(id, title, description string, now time.Time) (*Inbox, error) {
	now = now.UTC()
	inbox := &Inbox{ID: id, Title: title, Description: description, CreatedAt: now, UpdatedAt: now}
	if err := inbox.Validate(); err != nil {
		return nil, err
	}
	return inbox, nil
}

func (i *Inbox) Validate() error {
	if i == nil {
		return ErrInvalidArg
	}
	return requiredFieldsError(requiredInboxField(i.ID, "id"), requiredInboxField(i.Title, "title"))
}

func (i *Inbox) UpdateDetails(title, description string, now time.Time) error {
	if i == nil {
		return ErrInvalidArg
	}
	i.Title = title
	i.Description = description
	i.UpdatedAt = now.UTC()
	return i.Validate()
}

func requiredInboxField(value, field string) string {
	if value == "" {
		return field
	}
	return ""
}
