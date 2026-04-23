// Package application contains use-cases for Project.
package application

import (
	"context"
	"time"

	"github.com/anyvoxel/multivac/internal/id"
	"github.com/anyvoxel/multivac/pkg/domain"
)

// Service implements Project use-cases.
type ProjectService struct {
	repo domain.ProjectRepository
	now  func() time.Time
}

// Option customizes Service.
type ProjectOption func(*ProjectService)

// WithNow overrides time source (useful for tests).
func WithProjectNow(f func() time.Time) ProjectOption {
	return func(s *ProjectService) {
		s.now = f
	}
}

// NewService creates the Project application service.
func NewProjectService(repo domain.ProjectRepository, opts ...ProjectOption) *ProjectService {
	s := &ProjectService{repo: repo, now: time.Now}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// CreateProjectCmd describes input for creating a project.
type CreateProjectCmd struct {
	Title       string
	Goals       []domain.ProjectGoal
	Description string
	References  []domain.ProjectReference
}

// Migrate ensures persistence schema is ready.
func (s *ProjectService) Migrate(ctx context.Context) error {
	return s.repo.Migrate(ctx)
}

// Create creates a new project.
func (s *ProjectService) Create(ctx context.Context, cmd CreateProjectCmd) (*domain.Project, error) {
	now := s.now().UTC()
	id, err := id.NewULIDAt(now)
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
func (s *ProjectService) Get(ctx context.Context, id string) (*domain.Project, error) {
	return s.repo.Get(ctx, id)
}

// List returns projects, optionally filtered by status.
func (s *ProjectService) List(ctx context.Context, q domain.ProjectListQuery) ([]*domain.Project, error) {
	return s.repo.List(ctx, q)
}

// UpdateProjectCmd describes input for updating a project.
type UpdateProjectCmd struct {
	Title       string
	Goals       []domain.ProjectGoal
	Description string
	References  []domain.ProjectReference
}

// Update updates textual fields of a project.
func (s *ProjectService) Update(ctx context.Context, id string, cmd UpdateProjectCmd) (*domain.Project, error) {
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
func (s *ProjectService) SetStatus(ctx context.Context, id string, status domain.Status) (*domain.Project, error) {
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
func (s *ProjectService) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
