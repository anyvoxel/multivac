package domain

import "context"

type Repository interface {
	Migrate(ctx context.Context) error
	Create(ctx context.Context, reference *Reference) error
	Get(ctx context.Context, id string) (*Reference, error)
	List(ctx context.Context, q ListQuery) ([]*Reference, error)
	Update(ctx context.Context, reference *Reference) error
	Delete(ctx context.Context, id string) error
}
