// Package application contains use-cases for Inbox.
package application

import (
	"context"
	"time"

	"github.com/anyvoxel/multivac/internal/id"
	"github.com/anyvoxel/multivac/pkg/inbox/domain"
)

// Service implements Inbox use-cases.
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

// NewService creates the Inbox application service.
func NewService(repo domain.Repository, opts ...Option) *Service {
	s := &Service{repo: repo, now: time.Now}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// CreateInboxCmd describes input for creating an inbox.
type CreateInboxCmd struct {
	Name        string
	Description string
}

// UpdateInboxCmd describes input for updating an inbox.
type UpdateInboxCmd struct {
	Name        string
	Description string
}

// Migrate ensures persistence schema is ready.
func (s *Service) Migrate(ctx context.Context) error {
	return s.repo.Migrate(ctx)
}

// Create creates a new inbox.
func (s *Service) Create(ctx context.Context, cmd CreateInboxCmd) (*domain.Inbox, error) {
	now := s.now()
	inboxID, err := id.New("i")
	if err != nil {
		return nil, err
	}
	inbox, err := domain.NewInbox(inboxID, cmd.Name, cmd.Description, now)
	if err != nil {
		return nil, err
	}
	if err := s.repo.Create(ctx, inbox); err != nil {
		return nil, err
	}
	return inbox, nil
}

// Get returns an inbox by id.
func (s *Service) Get(ctx context.Context, id string) (*domain.Inbox, error) {
	return s.repo.Get(ctx, id)
}

// List returns inboxes.
func (s *Service) List(ctx context.Context, q domain.ListQuery) ([]*domain.Inbox, error) {
	return s.repo.List(ctx, q)
}

// Update updates inbox fields.
func (s *Service) Update(ctx context.Context, id string, cmd UpdateInboxCmd) (*domain.Inbox, error) {
	inbox, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := inbox.UpdateDetails(cmd.Name, cmd.Description, s.now()); err != nil {
		return nil, err
	}
	if err := s.repo.Update(ctx, inbox); err != nil {
		return nil, err
	}
	return inbox, nil
}

// Delete deletes an inbox.
func (s *Service) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
