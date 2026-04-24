package domain

import (
	"context"
	"time"
)

type InboxRepository interface {
	Migrate(ctx context.Context) error
	Create(ctx context.Context, inbox *Inbox) error
	Get(ctx context.Context, id string) (*Inbox, error)
	List(ctx context.Context, q InboxListQuery) ([]*Inbox, error)
	Update(ctx context.Context, inbox *Inbox) error
	Delete(ctx context.Context, id string) error
	ConvertFromSomeday(ctx context.Context, somedayID string, title, description *string, now time.Time) (*Inbox, error)
}
