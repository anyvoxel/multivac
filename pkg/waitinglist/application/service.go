// Package application contains use-cases for Waiting List.
package application

import (
	"context"
	"time"

	"github.com/anyvoxel/multivac/internal/id"
	"github.com/anyvoxel/multivac/pkg/waitinglist/domain"
)

// Service implements Waiting List use-cases.
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

// NewService creates the Waiting List application service.
func NewService(repo domain.Repository, opts ...Option) *Service {
	s := &Service{repo: repo, now: time.Now}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// CreateWaitingListCmd describes input for creating a waiting list item.
type CreateWaitingListCmd struct {
	Name       string
	Details    string
	Owner      string
	ExpectedAt *time.Time
}

// UpdateWaitingListCmd describes input for updating a waiting list item.
type UpdateWaitingListCmd struct {
	Name       string
	Details    string
	Owner      string
	ExpectedAt *time.Time
}

// Migrate ensures persistence schema is ready.
func (s *Service) Migrate(ctx context.Context) error {
	return s.repo.Migrate(ctx)
}

// Create creates a new waiting list item.
func (s *Service) Create(ctx context.Context, cmd CreateWaitingListCmd) (*domain.WaitingList, error) {
	waitingID, err := id.New("w")
	if err != nil {
		return nil, err
	}
	now := s.now()
	item, err := domain.NewWaitingList(waitingID, cmd.Name, cmd.Details, cmd.Owner, cmd.ExpectedAt, now)
	if err != nil {
		return nil, err
	}
	if err := s.repo.Create(ctx, item); err != nil {
		return nil, err
	}
	return item, nil
}

// Get returns a waiting list item by id.
func (s *Service) Get(ctx context.Context, id string) (*domain.WaitingList, error) {
	return s.repo.Get(ctx, id)
}

// List returns waiting list items.
func (s *Service) List(ctx context.Context, q domain.ListQuery) ([]*domain.WaitingList, error) {
	return s.repo.List(ctx, q)
}

// Update updates waiting list fields.
func (s *Service) Update(ctx context.Context, id string, cmd UpdateWaitingListCmd) (*domain.WaitingList, error) {
	item, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := item.UpdateDetails(cmd.Name, cmd.Details, cmd.Owner, cmd.ExpectedAt, s.now()); err != nil {
		return nil, err
	}
	if err := s.repo.Update(ctx, item); err != nil {
		return nil, err
	}
	return item, nil
}

// Delete deletes a waiting list item.
func (s *Service) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
