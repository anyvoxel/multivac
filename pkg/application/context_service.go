package application

import (
	"context"
	"time"

	"github.com/anyvoxel/multivac/internal/id"
	"github.com/anyvoxel/multivac/pkg/domain"
)

type ContextService struct {
	repo domain.ContextRepository
	now  func() time.Time
}

type ContextOption func(*ContextService)

func WithContextNow(f func() time.Time) ContextOption {
	return func(s *ContextService) { s.now = f }
}

func NewContextService(repo domain.ContextRepository, opts ...ContextOption) *ContextService {
	s := &ContextService{repo: repo, now: time.Now}
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

func (s *ContextService) Migrate(ctx context.Context) error { return s.repo.Migrate(ctx) }

func (s *ContextService) Create(ctx context.Context, cmd CreateContextCmd) (*domain.Context, error) {
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

func (s *ContextService) Get(ctx context.Context, id string) (*domain.Context, error) {
	return s.repo.Get(ctx, id)
}

func (s *ContextService) List(ctx context.Context, q domain.ContextListQuery) ([]*domain.Context, error) {
	return s.repo.List(ctx, q)
}

func (s *ContextService) Update(ctx context.Context, id string, cmd UpdateContextCmd) (*domain.Context, error) {
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

func (s *ContextService) Delete(ctx context.Context, id string) error { return s.repo.Delete(ctx, id) }
