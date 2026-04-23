package domain

import (
	"strings"
	"time"
)

type Someday struct {
	ID          string
	Title       string
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func NewSomeday(id, title, description string, now time.Time) (*Someday, error) {
	now = now.UTC()
	someday := &Someday{
		ID:          strings.TrimSpace(id),
		Title:       strings.TrimSpace(title),
		Description: description,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := someday.Validate(); err != nil {
		return nil, err
	}
	return someday, nil
}

func (s *Someday) Validate() error {
	if s == nil {
		return ErrInvalidArg
	}
	return requiredFieldsError(requiredSomedayField(s.ID, "id"), requiredSomedayField(s.Title, "title"))
}

func (s *Someday) UpdateDetails(title, description string, now time.Time) error {
	if s == nil {
		return ErrInvalidArg
	}
	s.Title = strings.TrimSpace(title)
	s.Description = description
	s.UpdatedAt = now.UTC()
	return s.Validate()
}

func requiredSomedayField(value, field string) string {
	if strings.TrimSpace(value) == "" {
		return field
	}
	return ""
}
