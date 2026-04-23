package application

import (
	"context"
	"time"

	"github.com/anyvoxel/multivac/internal/id"
	"github.com/anyvoxel/multivac/pkg/someday/domain"
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

type CreateSomedayCmd struct {
	Title       string
	Description string
}

type UpdateSomedayCmd = CreateSomedayCmd

type ConvertFromInboxCmd struct {
	Title       *string
	Description *string
}

func (s *Service) Migrate(ctx context.Context) error { return s.repo.Migrate(ctx) }

func (s *Service) Create(ctx context.Context, cmd CreateSomedayCmd) (*domain.Someday, error) {
	now := s.now().UTC()
	somedayID, err := id.NewULIDAt(now)
	if err != nil {
		return nil, err
	}
	someday, err := domain.NewSomeday(somedayID, cmd.Title, cmd.Description, now)
	if err != nil {
		return nil, err
	}
	if err := s.repo.Create(ctx, someday); err != nil {
		return nil, err
	}
	return someday, nil
}

func (s *Service) ConvertFromInbox(ctx context.Context, inboxID string, cmd ConvertFromInboxCmd) (*domain.Someday, error) {
	return s.repo.ConvertFromInbox(ctx, inboxID, cmd.Title, cmd.Description, s.now().UTC())
}

func (s *Service) Get(ctx context.Context, id string) (*domain.Someday, error) {
	return s.repo.Get(ctx, id)
}

func (s *Service) List(ctx context.Context, q domain.ListQuery) ([]*domain.Someday, error) {
	return s.repo.List(ctx, q)
}

func (s *Service) Update(ctx context.Context, id string, cmd UpdateSomedayCmd) (*domain.Someday, error) {
	someday, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := someday.UpdateDetails(cmd.Title, cmd.Description, s.now().UTC()); err != nil {
		return nil, err
	}
	if err := s.repo.Update(ctx, someday); err != nil {
		return nil, err
	}
	return someday, nil
}

func (s *Service) Delete(ctx context.Context, id string) error { return s.repo.Delete(ctx, id) }
