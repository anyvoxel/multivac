package domain

import "context"

type ContextRepository interface {
	Migrate(ctx context.Context) error
	Create(ctx context.Context, context *Context) error
	Get(ctx context.Context, id string) (*Context, error)
	List(ctx context.Context, q ContextListQuery) ([]*Context, error)
	Update(ctx context.Context, context *Context) error
	Delete(ctx context.Context, id string) error
}
