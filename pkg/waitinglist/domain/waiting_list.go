package domain

import "time"

// WaitingList is the aggregate root for waiting list management.
type WaitingList struct {
	ID         string
	Name       string
	Details    string
	Owner      string
	ExpectedAt *time.Time
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// NewWaitingList creates a waiting list item.
func NewWaitingList(id, name, details, owner string, expectedAt *time.Time, now time.Time) (*WaitingList, error) {
	if err := requiredFieldsError(
		requiredField(id, "id"),
		requiredField(name, "name"),
	); err != nil {
		return nil, err
	}

	return &WaitingList{
		ID:         id,
		Name:       name,
		Details:    details,
		Owner:      owner,
		ExpectedAt: cloneTimePtr(expectedAt),
		CreatedAt:  now,
		UpdatedAt:  now,
	}, nil
}

// UpdateDetails replaces waiting list core fields.
func (w *WaitingList) UpdateDetails(name, details, owner string, expectedAt *time.Time, now time.Time) error {
	if w == nil {
		return ErrInvalidArg
	}
	if err := requiredFieldsError(requiredField(name, "name")); err != nil {
		return err
	}

	w.Name = name
	w.Details = details
	w.Owner = owner
	w.ExpectedAt = cloneTimePtr(expectedAt)
	w.UpdatedAt = now
	return nil
}

func requiredField(value, field string) string {
	if value == "" {
		return field
	}
	return ""
}

func cloneTimePtr(v *time.Time) *time.Time {
	if v == nil {
		return nil
	}
	cloned := *v
	return &cloned
}
