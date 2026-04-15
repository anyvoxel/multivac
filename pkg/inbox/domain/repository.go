package domain

import "context"

// Repository abstracts persistence for Inbox aggregate.
type Repository interface {
	Migrate(ctx context.Context) error
	Create(ctx context.Context, inbox *Inbox) error
	Get(ctx context.Context, id string) (*Inbox, error)
	List(ctx context.Context, q ListQuery) ([]*Inbox, error)
	Update(ctx context.Context, inbox *Inbox) error
	Delete(ctx context.Context, id string) error
}
