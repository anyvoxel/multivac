// Package application contains use-cases for Someday.
package application

import (
	"context"
	"time"

	"github.com/anyvoxel/multivac/internal/id"
	"github.com/anyvoxel/multivac/pkg/someday/domain"
)

// Service implements Someday use-cases.
type Service struct {
	repo domain.Repository
	now  func() time.Time
}

// Option customizes Service.
type Option func(*Service)

// WithNow overrides time source.
func WithNow(f func() time.Time) Option {
	return func(s *Service) {
		s.now = f
	}
}

// NewService creates the Someday application service.
func NewService(repo domain.Repository, opts ...Option) *Service {
	s := &Service{repo: repo, now: time.Now}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// CreateSomedayCmd describes input for creating a someday item.
type CreateSomedayCmd struct {
	Name        string
	Description string
}

// UpdateSomedayCmd describes input for updating a someday item.
type UpdateSomedayCmd struct {
	Name        string
	Description string
}

// Migrate ensures persistence schema is ready.
func (s *Service) Migrate(ctx context.Context) error {
	return s.repo.Migrate(ctx)
}

// Create creates a new someday item.
func (s *Service) Create(ctx context.Context, cmd CreateSomedayCmd) (*domain.Someday, error) {
	somedayID, err := id.New("s")
	if err != nil {
		return nil, err
	}
	now := s.now()
	someday, err := domain.NewSomeday(somedayID, cmd.Name, cmd.Description, now)
	if err != nil {
		return nil, err
	}
	if err := s.repo.Create(ctx, someday); err != nil {
		return nil, err
	}
	return someday, nil
}

// Get returns a someday item by id.
func (s *Service) Get(ctx context.Context, id string) (*domain.Someday, error) {
	return s.repo.Get(ctx, id)
}

// List returns someday items.
func (s *Service) List(ctx context.Context, q domain.ListQuery) ([]*domain.Someday, error) {
	return s.repo.List(ctx, q)
}

// Update updates someday fields.
func (s *Service) Update(ctx context.Context, id string, cmd UpdateSomedayCmd) (*domain.Someday, error) {
	someday, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := someday.UpdateDetails(cmd.Name, cmd.Description, s.now()); err != nil {
		return nil, err
	}
	if err := s.repo.Update(ctx, someday); err != nil {
		return nil, err
	}
	return someday, nil
}

// Delete deletes a someday item.
func (s *Service) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
