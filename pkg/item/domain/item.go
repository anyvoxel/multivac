package domain

import "time"

type Item struct {
	ID          string
	Kind        Kind
	Bucket      Bucket
	ProjectID   string
	Title       string
	Description string
	Labels      []Label
	Context     string
	Details     string
	TaskStatus  string
	Priority    string
	WaitingFor  string
	ExpectedAt  *time.Time
	DueAt       *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func NewItem(id string, kind Kind, bucket Bucket, now time.Time) (*Item, error) {
	item := &Item{ID: id, Kind: kind, Bucket: bucket, CreatedAt: now, UpdatedAt: now}
	if err := requiredFieldsError(requiredField(item.ID, "id")); err != nil {
		return nil, err
	}
	if !item.Kind.Valid() {
		return nil, InvalidKind(string(item.Kind))
	}
	if !item.Bucket.Valid() {
		return nil, InvalidBucket(string(item.Bucket))
	}
	return item, nil
}

func (i *Item) Validate() error {
	if i == nil {
		return ErrInvalidArg
	}
	if err := requiredFieldsError(requiredField(i.ID, "id"), requiredField(i.Title, "title")); err != nil {
		return err
	}
	if !i.Kind.Valid() {
		return InvalidKind(string(i.Kind))
	}
	if !i.Bucket.Valid() {
		return InvalidBucket(string(i.Bucket))
	}
	return nil
}

func (i *Item) UpdateBucket(bucket Bucket, now time.Time) error {
	if i == nil {
		return ErrInvalidArg
	}
	if !bucket.Valid() {
		return InvalidBucket(string(bucket))
	}
	i.Bucket = bucket
	i.UpdatedAt = now
	return nil
}

func requiredField(value, field string) string {
	if value == "" {
		return field
	}
	return ""
}
