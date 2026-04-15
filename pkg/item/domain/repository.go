package domain

import "context"

type Repository interface {
	Migrate(ctx context.Context) error
	Create(ctx context.Context, item *Item) error
	Get(ctx context.Context, id string) (*Item, error)
	List(ctx context.Context, q ListQuery) ([]*Item, error)
	Update(ctx context.Context, item *Item) error
	Delete(ctx context.Context, id string) error
}
