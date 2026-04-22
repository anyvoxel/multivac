package domain

import (
	"regexp"
	"strings"
	"time"
)

var hexColorPattern = regexp.MustCompile(`^#[0-9a-f]{6}$`)

type Context struct {
	ID          string
	Title       string
	Description string
	Color       string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func NewContext(id, title, description, color string, now time.Time) (*Context, error) {
	now = now.UTC()
	context := &Context{
		ID:          strings.TrimSpace(id),
		Title:       strings.TrimSpace(title),
		Description: description,
		Color:       normalizeColor(color),
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := context.Validate(); err != nil {
		return nil, err
	}
	return context, nil
}

func (c *Context) Validate() error {
	if c == nil {
		return ErrInvalidArg
	}
	if err := requiredFieldsError(requiredField(c.ID, "id"), requiredField(c.Title, "title"), requiredField(c.Color, "color")); err != nil {
		return err
	}
	if !hexColorPattern.MatchString(c.Color) {
		return InvalidColor(c.Color)
	}
	return nil
}

func (c *Context) UpdateDetails(title, description, color string, now time.Time) error {
	if c == nil {
		return ErrInvalidArg
	}
	c.Title = strings.TrimSpace(title)
	c.Description = description
	c.Color = normalizeColor(color)
	c.UpdatedAt = now.UTC()
	return c.Validate()
}

func requiredField(value, field string) string {
	if strings.TrimSpace(value) == "" {
		return field
	}
	return ""
}

func normalizeColor(color string) string {
	return strings.ToLower(strings.TrimSpace(color))
}
