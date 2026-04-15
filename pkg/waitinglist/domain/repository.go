package domain

import "context"

// Repository abstracts persistence for Waiting List aggregate.
type Repository interface {
	Migrate(ctx context.Context) error
	Create(ctx context.Context, item *WaitingList) error
	Get(ctx context.Context, id string) (*WaitingList, error)
	List(ctx context.Context, q ListQuery) ([]*WaitingList, error)
	Update(ctx context.Context, item *WaitingList) error
	Delete(ctx context.Context, id string) error
}
