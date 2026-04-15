package domain

import "context"

// Repository abstracts persistence for Someday aggregate.
type Repository interface {
	Migrate(ctx context.Context) error
	Create(ctx context.Context, someday *Someday) error
	Get(ctx context.Context, id string) (*Someday, error)
	List(ctx context.Context, q ListQuery) ([]*Someday, error)
	Update(ctx context.Context, someday *Someday) error
	Delete(ctx context.Context, id string) error
}
