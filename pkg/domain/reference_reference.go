package domain

import (
	"net/url"
	"strings"
	"time"
)

type ReferenceLink struct {
	Title string `json:"title"`
	URL   string `json:"url"`
}

type Reference struct {
	ID          string
	Title       string
	Description string
	References  []ReferenceLink
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func NewReference(id, title, description string, references []ReferenceLink, now time.Time) (*Reference, error) {
	now = now.UTC()
	reference := &Reference{
		ID:          strings.TrimSpace(id),
		Title:       strings.TrimSpace(title),
		Description: description,
		References:  normalizeReferenceLinks(references),
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := reference.Validate(); err != nil {
		return nil, err
	}
	return reference, nil
}

func (r *Reference) Validate() error {
	if r == nil {
		return ErrInvalidArg
	}
	if err := requiredFieldsError(requiredReferenceField(r.ID, "id"), requiredReferenceField(r.Title, "title")); err != nil {
		return err
	}
	if err := validateReferenceLinks(r.References); err != nil {
		return err
	}
	return nil
}

func (r *Reference) UpdateDetails(title, description string, references []ReferenceLink, now time.Time) error {
	if r == nil {
		return ErrInvalidArg
	}
	r.Title = strings.TrimSpace(title)
	r.Description = description
	r.References = normalizeReferenceLinks(references)
	r.UpdatedAt = now.UTC()
	return r.Validate()
}

func validateReferenceLinks(references []ReferenceLink) error {
	for i, reference := range references {
		if strings.TrimSpace(reference.Title) == "" {
			return &ValidationError{Problems: []string{"references[" + itoaReference(i) + "].title: Required value"}}
		}
		if strings.TrimSpace(reference.URL) == "" {
			return &ValidationError{Problems: []string{"references[" + itoaReference(i) + "].url: Required value"}}
		}
		u, err := url.ParseRequestURI(reference.URL)
		if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
			return invalidFieldValueError("references["+itoaReference(i)+"].url", reference.URL)
		}
	}
	return nil
}

func normalizeReferenceLinks(references []ReferenceLink) []ReferenceLink {
	if len(references) == 0 {
		return nil
	}
	out := make([]ReferenceLink, 0, len(references))
	for _, reference := range references {
		out = append(out, ReferenceLink{
			Title: strings.TrimSpace(reference.Title),
			URL:   strings.TrimSpace(reference.URL),
		})
	}
	return out
}

func requiredReferenceField(value, field string) string {
	if strings.TrimSpace(value) == "" {
		return field
	}
	return ""
}

func itoaReference(v int) string {
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
