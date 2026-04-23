package domain

import (
	"context"
	"time"
)

type Repository interface {
	Migrate(ctx context.Context) error
	Create(ctx context.Context, someday *Someday) error
	Get(ctx context.Context, id string) (*Someday, error)
	List(ctx context.Context, q ListQuery) ([]*Someday, error)
	Update(ctx context.Context, someday *Someday) error
	Delete(ctx context.Context, id string) error
	ConvertFromInbox(ctx context.Context, inboxID string, title, description *string, now time.Time) (*Someday, error)
}
