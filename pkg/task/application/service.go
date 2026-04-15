// Package application contains use-cases for Task.
package application

import (
	"context"
	"time"

	"github.com/anyvoxel/multivac/internal/id"
	"github.com/anyvoxel/multivac/pkg/task/domain"
)

// Service implements Task use-cases.
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

// NewService creates the Task application service.
func NewService(repo domain.Repository, opts ...Option) *Service {
	s := &Service{repo: repo, now: time.Now}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// CreateTaskCmd describes input for creating a task.
type CreateTaskCmd struct {
	ProjectID   string
	Name        string
	Description string
	Context     string
	Details     string
	Priority    domain.Priority
	DueAt       *time.Time
}

// UpdateTaskCmd describes input for updating a task.
type UpdateTaskCmd struct {
	ProjectID   string
	Name        string
	Description string
	Context     string
	Details     string
	Priority    domain.Priority
	DueAt       *time.Time
}

// Migrate ensures persistence schema is ready.
func (s *Service) Migrate(ctx context.Context) error {
	return s.repo.Migrate(ctx)
}

// Create creates a new task.
func (s *Service) Create(ctx context.Context, cmd CreateTaskCmd) (*domain.Task, error) {
	now := s.now()
	id, err := id.New("t")
	if err != nil {
		return nil, err
	}
	t, err := domain.NewTask(
		id,
		cmd.ProjectID,
		cmd.Name,
		cmd.Description,
		cmd.Context,
		cmd.Details,
		cmd.Priority,
		cmd.DueAt,
		now,
	)
	if err != nil {
		return nil, err
	}
	if err := s.repo.Create(ctx, t); err != nil {
		return nil, err
	}
	return t, nil
}

// Get returns a task by id.
func (s *Service) Get(ctx context.Context, id string) (*domain.Task, error) {
	return s.repo.Get(ctx, id)
}

// List returns tasks optionally filtered by project and status.
func (s *Service) List(ctx context.Context, q domain.ListQuery) ([]*domain.Task, error) {
	return s.repo.List(ctx, q)
}

// ListByProject returns tasks of a project.
func (s *Service) ListByProject(ctx context.Context, projectID string, status *domain.Status) ([]*domain.Task, error) {
	return s.repo.List(ctx, domain.ListQuery{ProjectID: projectID, Status: status})
}

// Update updates editable fields.
func (s *Service) Update(ctx context.Context, id string, cmd UpdateTaskCmd) (*domain.Task, error) {
	t, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := t.UpdateDetails(
		cmd.ProjectID,
		cmd.Name,
		cmd.Description,
		cmd.Context,
		cmd.Details,
		cmd.Priority,
		cmd.DueAt,
		s.now(),
	); err != nil {
		return nil, err
	}
	if err := s.repo.Update(ctx, t); err != nil {
		return nil, err
	}
	return t, nil
}

// SetStatus changes status.
func (s *Service) SetStatus(ctx context.Context, id string, status domain.Status) (*domain.Task, error) {
	t, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := t.SetStatus(status, s.now()); err != nil {
		return nil, err
	}
	if err := s.repo.Update(ctx, t); err != nil {
		return nil, err
	}
	return t, nil
}

// Delete deletes a task.
func (s *Service) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
