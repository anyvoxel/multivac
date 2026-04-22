package domain

import "context"

type Repository interface {
	Migrate(ctx context.Context) error
	Create(ctx context.Context, inbox *Inbox) error
	Get(ctx context.Context, id string) (*Inbox, error)
	List(ctx context.Context, q ListQuery) ([]*Inbox, error)
	Update(ctx context.Context, inbox *Inbox) error
	Delete(ctx context.Context, id string) error
}
