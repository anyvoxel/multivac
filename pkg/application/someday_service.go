package application

import (
	"context"
	"time"

	"github.com/anyvoxel/multivac/internal/id"
	"github.com/anyvoxel/multivac/pkg/domain"
)

type SomedayService struct {
	repo domain.SomedayRepository
	now  func() time.Time
}

type SomedayOption func(*SomedayService)

func WithSomedayNow(f func() time.Time) SomedayOption {
	return func(s *SomedayService) { s.now = f }
}

func NewSomedayService(repo domain.SomedayRepository, opts ...SomedayOption) *SomedayService {
	s := &SomedayService{repo: repo, now: time.Now}
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

type ConvertSomedayFromInboxCmd struct {
	Title       *string
	Description *string
}

func (s *SomedayService) Migrate(ctx context.Context) error { return s.repo.Migrate(ctx) }

func (s *SomedayService) Create(ctx context.Context, cmd CreateSomedayCmd) (*domain.Someday, error) {
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

func (s *SomedayService) ConvertFromInbox(ctx context.Context, inboxID string, cmd ConvertSomedayFromInboxCmd) (*domain.Someday, error) {
	return s.repo.ConvertFromInbox(ctx, inboxID, cmd.Title, cmd.Description, s.now().UTC())
}

func (s *SomedayService) Get(ctx context.Context, id string) (*domain.Someday, error) {
	return s.repo.Get(ctx, id)
}

func (s *SomedayService) List(ctx context.Context, q domain.SomedayListQuery) ([]*domain.Someday, error) {
	return s.repo.List(ctx, q)
}

func (s *SomedayService) Update(ctx context.Context, id string, cmd UpdateSomedayCmd) (*domain.Someday, error) {
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

func (s *SomedayService) Delete(ctx context.Context, id string) error { return s.repo.Delete(ctx, id) }
