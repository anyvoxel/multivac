package domain

import "context"

// Repository abstracts persistence for Task aggregate.
type Repository interface {
	Migrate(ctx context.Context) error
	Create(ctx context.Context, t *Task) error
	Get(ctx context.Context, id string) (*Task, error)
	List(ctx context.Context, q ListQuery) ([]*Task, error)
	Update(ctx context.Context, t *Task) error
	Delete(ctx context.Context, id string) error
}
