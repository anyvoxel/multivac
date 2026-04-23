package application

import (
	"context"
	"time"

	"github.com/anyvoxel/multivac/internal/id"
	"github.com/anyvoxel/multivac/pkg/domain"
)

type ReferenceService struct {
	repo domain.ReferenceRepository
	now  func() time.Time
}

type ReferenceOption func(*ReferenceService)

func WithReferenceNow(f func() time.Time) ReferenceOption {
	return func(s *ReferenceService) { s.now = f }
}

func NewReferenceService(repo domain.ReferenceRepository, opts ...ReferenceOption) *ReferenceService {
	s := &ReferenceService{repo: repo, now: time.Now}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

type CreateReferenceCmd struct {
	Title       string
	Description string
	References  []domain.ReferenceLink
}

type UpdateReferenceCmd = CreateReferenceCmd

func (s *ReferenceService) Migrate(ctx context.Context) error { return s.repo.Migrate(ctx) }

func (s *ReferenceService) Create(ctx context.Context, cmd CreateReferenceCmd) (*domain.Reference, error) {
	now := s.now().UTC()
	referenceID, err := id.NewULIDAt(now)
	if err != nil {
		return nil, err
	}
	reference, err := domain.NewReference(referenceID, cmd.Title, cmd.Description, cmd.References, now)
	if err != nil {
		return nil, err
	}
	if err := s.repo.Create(ctx, reference); err != nil {
		return nil, err
	}
	return reference, nil
}

func (s *ReferenceService) Get(ctx context.Context, id string) (*domain.Reference, error) {
	return s.repo.Get(ctx, id)
}

func (s *ReferenceService) List(ctx context.Context, q domain.ReferenceListQuery) ([]*domain.Reference, error) {
	return s.repo.List(ctx, q)
}

func (s *ReferenceService) Update(ctx context.Context, id string, cmd UpdateReferenceCmd) (*domain.Reference, error) {
	reference, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := reference.UpdateDetails(cmd.Title, cmd.Description, cmd.References, s.now().UTC()); err != nil {
		return nil, err
	}
	if err := s.repo.Update(ctx, reference); err != nil {
		return nil, err
	}
	return reference, nil
}

func (s *ReferenceService) Delete(ctx context.Context, id string) error { return s.repo.Delete(ctx, id) }
