package domain

import (
	"context"
	"time"
)

type SomedayRepository interface {
	Migrate(ctx context.Context) error
	Create(ctx context.Context, someday *Someday) error
	Get(ctx context.Context, id string) (*Someday, error)
	List(ctx context.Context, q SomedayListQuery) ([]*Someday, error)
	Update(ctx context.Context, someday *Someday) error
	Delete(ctx context.Context, id string) error
	ConvertFromInbox(ctx context.Context, inboxID string, title, description *string, now time.Time) (*Someday, error)
	ConvertFromAction(ctx context.Context, actionID string, title, description *string, now time.Time) (*Someday, error)
}
