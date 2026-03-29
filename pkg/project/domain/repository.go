package domain

import "context"

// Repository abstracts persistence for Project aggregate.
type Repository interface {
	Migrate(ctx context.Context) error
	Create(ctx context.Context, p *Project) error
	Get(ctx context.Context, id string) (*Project, error)
	List(ctx context.Context, q ListQuery) ([]*Project, error)
	Update(ctx context.Context, p *Project) error
	Delete(ctx context.Context, id string) error
}
