package domain

import (
	"context"
	"time"
)

type Repository interface {
	Migrate(ctx context.Context) error
	Create(ctx context.Context, action *Action) error
	Get(ctx context.Context, id string) (*Action, error)
	List(ctx context.Context, q ListQuery) ([]*Action, error)
	Update(ctx context.Context, action *Action) error
	Delete(ctx context.Context, id string) error
	ConvertFromInbox(ctx context.Context, inboxID string, title, description *string, kind *Kind, projectID *string, contextIDs []string, labels []Label, attributes *Attributes, now time.Time) (*Action, error)
}
