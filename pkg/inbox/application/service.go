package application

import (
	"context"
	"time"

	"github.com/anyvoxel/multivac/internal/id"
	"github.com/anyvoxel/multivac/pkg/inbox/domain"
)

type Service struct {
	repo domain.Repository
	now  func() time.Time
}

type Option func(*Service)

func WithNow(f func() time.Time) Option {
	return func(s *Service) { s.now = f }
}

func NewService(repo domain.Repository, opts ...Option) *Service {
	s := &Service{repo: repo, now: time.Now}
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

func (s *Service) Migrate(ctx context.Context) error { return s.repo.Migrate(ctx) }

func (s *Service) Create(ctx context.Context, cmd CreateInboxCmd) (*domain.Inbox, error) {
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

func (s *Service) Get(ctx context.Context, id string) (*domain.Inbox, error) {
	return s.repo.Get(ctx, id)
}

func (s *Service) List(ctx context.Context, q domain.ListQuery) ([]*domain.Inbox, error) {
	return s.repo.List(ctx, q)
}

func (s *Service) Update(ctx context.Context, id string, cmd UpdateInboxCmd) (*domain.Inbox, error) {
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

func (s *Service) Delete(ctx context.Context, id string) error { return s.repo.Delete(ctx, id) }
