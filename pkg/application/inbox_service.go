package application

import (
	"context"
	"time"

	"github.com/anyvoxel/multivac/internal/id"
	"github.com/anyvoxel/multivac/pkg/domain"
)

type InboxService struct {
	repo domain.InboxRepository
	now  func() time.Time
}

type InboxOption func(*InboxService)

func WithInboxNow(f func() time.Time) InboxOption {
	return func(s *InboxService) { s.now = f }
}

func NewInboxService(repo domain.InboxRepository, opts ...InboxOption) *InboxService {
	s := &InboxService{repo: repo, now: time.Now}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

type CreateInboxCmd struct {
	Title       string
	Description string
}

type UpdateInboxCmd = CreateInboxCmd

func (s *InboxService) Migrate(ctx context.Context) error { return s.repo.Migrate(ctx) }

func (s *InboxService) Create(ctx context.Context, cmd CreateInboxCmd) (*domain.Inbox, error) {
	now := s.now().UTC()
	inboxID, err := id.NewULIDAt(now)
	if err != nil {
		return nil, err
	}
	inbox, err := domain.NewInbox(inboxID, cmd.Title, cmd.Description, now)
	if err != nil {
		return nil, err
	}
	if err := s.repo.Create(ctx, inbox); err != nil {
		return nil, err
	}
	return inbox, nil
}

func (s *InboxService) Get(ctx context.Context, id string) (*domain.Inbox, error) {
	return s.repo.Get(ctx, id)
}

func (s *InboxService) List(ctx context.Context, q domain.InboxListQuery) ([]*domain.Inbox, error) {
	return s.repo.List(ctx, q)
}

func (s *InboxService) Update(ctx context.Context, id string, cmd UpdateInboxCmd) (*domain.Inbox, error) {
	inbox, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := inbox.UpdateDetails(cmd.Title, cmd.Description, s.now().UTC()); err != nil {
		return nil, err
	}
	if err := s.repo.Update(ctx, inbox); err != nil {
		return nil, err
	}
	return inbox, nil
}

func (s *InboxService) Delete(ctx context.Context, id string) error { return s.repo.Delete(ctx, id) }
