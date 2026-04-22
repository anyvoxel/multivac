// Package application contains use-cases for Project.
package application

import (
	"context"
	"time"

	"github.com/anyvoxel/multivac/internal/id"
	"github.com/anyvoxel/multivac/pkg/project/domain"
)

// Service implements Project use-cases.
type Service struct {
	repo domain.Repository
	now  func() time.Time
}

// Option customizes Service.
type Option func(*Service)

// WithNow overrides time source (useful for tests).
func WithNow(f func() time.Time) Option {
	return func(s *Service) {
		s.now = f
	}
}

// NewService creates the Project application service.
func NewService(repo domain.Repository, opts ...Option) *Service {
	s := &Service{repo: repo, now: time.Now}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// CreateProjectCmd describes input for creating a project.
type CreateProjectCmd struct {
	Title       string
	Goals       []domain.Goal
	Description string
	References  []domain.Reference
}

// Migrate ensures persistence schema is ready.
func (s *Service) Migrate(ctx context.Context) error {
	return s.repo.Migrate(ctx)
}

// Create creates a new project.
func (s *Service) Create(ctx context.Context, cmd CreateProjectCmd) (*domain.Project, error) {
	now := s.now()
	id, err := id.New("p")
	if err != nil {
		return nil, err
	}
	p, err := domain.NewProject(id, cmd.Title, cmd.Goals, cmd.Description, cmd.References, now)
	if err != nil {
		return nil, err
	}
	if err := s.repo.Create(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

// Get returns a project by id.
func (s *Service) Get(ctx context.Context, id string) (*domain.Project, error) {
	return s.repo.Get(ctx, id)
}

// List returns projects, optionally filtered by status.
func (s *Service) List(ctx context.Context, q domain.ListQuery) ([]*domain.Project, error) {
	return s.repo.List(ctx, q)
}

// UpdateProjectCmd describes input for updating a project.
type UpdateProjectCmd struct {
	Title       string
	Goals       []domain.Goal
	Description string
	References  []domain.Reference
}

// Update updates textual fields of a project.
func (s *Service) Update(ctx context.Context, id string, cmd UpdateProjectCmd) (*domain.Project, error) {
	p, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := p.UpdateDetails(cmd.Title, cmd.Goals, cmd.Description, cmd.References, s.now()); err != nil {
		return nil, err
	}
	if err := s.repo.Update(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

// SetStatus changes a project's status.
func (s *Service) SetStatus(ctx context.Context, id string, status domain.Status) (*domain.Project, error) {
	p, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := p.SetStatus(status, s.now()); err != nil {
		return nil, err
	}
	if err := s.repo.Update(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

// Delete deletes a project.
func (s *Service) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
