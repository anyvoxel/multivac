package application

import (
	"context"
	"time"

	"github.com/anyvoxel/multivac/internal/id"
	"github.com/anyvoxel/multivac/pkg/action/domain"
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

type CreateActionCmd struct {
	Title       string
	Description string
	ProjectID   *string
	Kind        domain.Kind
	ContextIDs  []string
	Labels      []domain.Label
	Attributes  domain.Attributes
}

type UpdateActionCmd = CreateActionCmd

type ConvertFromInboxCmd struct {
	Title       *string
	Description *string
	ProjectID   *string
	Kind        *domain.Kind
	ContextIDs  []string
	Labels      []domain.Label
	Attributes  *domain.Attributes
}

func (s *Service) Migrate(ctx context.Context) error { return s.repo.Migrate(ctx) }

func (s *Service) Create(ctx context.Context, cmd CreateActionCmd) (*domain.Action, error) {
	now := s.now().UTC()
	actionID, err := id.NewULIDAt(now)
	if err != nil {
		return nil, err
	}
	kind := cmd.Kind
	if kind == "" {
		kind = domain.KindTask
	}
	action, err := domain.NewAction(actionID, cmd.Title, cmd.Description, cmd.ProjectID, kind, cmd.ContextIDs, cmd.Labels, cmd.Attributes, now)
	if err != nil {
		return nil, err
	}
	if err := s.repo.Create(ctx, action); err != nil {
		return nil, err
	}
	return action, nil
}

func (s *Service) ConvertFromInbox(ctx context.Context, inboxID string, cmd ConvertFromInboxCmd) (*domain.Action, error) {
	kind := cmd.Kind
	if kind == nil {
		v := domain.KindTask
		kind = &v
	}
	attributes := cmd.Attributes
	if attributes == nil {
		v := domain.Attributes{Task: &domain.TaskAttributes{}}
		attributes = &v
	}
	return s.repo.ConvertFromInbox(ctx, inboxID, cmd.Title, cmd.Description, kind, cmd.ProjectID, cmd.ContextIDs, cmd.Labels, attributes, s.now().UTC())
}

func (s *Service) Get(ctx context.Context, id string) (*domain.Action, error) {
	return s.repo.Get(ctx, id)
}

func (s *Service) List(ctx context.Context, q domain.ListQuery) ([]*domain.Action, error) {
	return s.repo.List(ctx, q)
}

func (s *Service) Update(ctx context.Context, id string, cmd UpdateActionCmd) (*domain.Action, error) {
	action, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	kind := cmd.Kind
	if kind == "" {
		kind = action.Kind
	}
	if err := action.UpdateDetails(cmd.Title, cmd.Description, cmd.ProjectID, kind, cmd.ContextIDs, cmd.Labels, cmd.Attributes, s.now().UTC()); err != nil {
		return nil, err
	}
	if err := s.repo.Update(ctx, action); err != nil {
		return nil, err
	}
	return action, nil
}

func (s *Service) Delete(ctx context.Context, id string) error { return s.repo.Delete(ctx, id) }
