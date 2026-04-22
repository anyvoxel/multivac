package application

import (
	"context"
	"time"

	"github.com/anyvoxel/multivac/internal/id"
	"github.com/anyvoxel/multivac/pkg/context/domain"
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

type CreateContextCmd struct {
	Title       string
	Description string
	Color       string
}

type UpdateContextCmd = CreateContextCmd

func (s *Service) Migrate(ctx context.Context) error { return s.repo.Migrate(ctx) }

func (s *Service) Create(ctx context.Context, cmd CreateContextCmd) (*domain.Context, error) {
	now := s.now().UTC()
	contextID, err := id.NewULIDAt(now)
	if err != nil {
		return nil, err
	}
	contextObj, err := domain.NewContext(contextID, cmd.Title, cmd.Description, cmd.Color, now)
	if err != nil {
		return nil, err
	}
	if err := s.repo.Create(ctx, contextObj); err != nil {
		return nil, err
	}
	return contextObj, nil
}

func (s *Service) Get(ctx context.Context, id string) (*domain.Context, error) {
	return s.repo.Get(ctx, id)
}

func (s *Service) List(ctx context.Context, q domain.ListQuery) ([]*domain.Context, error) {
	return s.repo.List(ctx, q)
}

func (s *Service) Update(ctx context.Context, id string, cmd UpdateContextCmd) (*domain.Context, error) {
	contextObj, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := contextObj.UpdateDetails(cmd.Title, cmd.Description, cmd.Color, s.now().UTC()); err != nil {
		return nil, err
	}
	if err := s.repo.Update(ctx, contextObj); err != nil {
		return nil, err
	}
	return contextObj, nil
}

func (s *Service) Delete(ctx context.Context, id string) error { return s.repo.Delete(ctx, id) }
