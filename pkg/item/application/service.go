package application

import (
	"context"
	"time"

	"github.com/anyvoxel/multivac/internal/id"
	"github.com/anyvoxel/multivac/pkg/item/domain"
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

type CreateItemCmd struct {
	Kind        domain.Kind
	Bucket      domain.Bucket
	ProjectID   string
	Title       string
	Description string
	Labels      []domain.Label
	Context     string
	Details     string
	TaskStatus  string
	Priority    string
	WaitingFor  string
	ExpectedAt  *time.Time
	DueAt       *time.Time
}

type UpdateItemCmd = CreateItemCmd

func (s *Service) Migrate(ctx context.Context) error { return s.repo.Migrate(ctx) }

func (s *Service) Create(ctx context.Context, cmd CreateItemCmd) (*domain.Item, error) {
	now := s.now()
	itemID, err := id.New("it")
	if err != nil {
		return nil, err
	}
	item, err := domain.NewItem(itemID, cmd.Kind, cmd.Bucket, now)
	if err != nil {
		return nil, err
	}
	item.ProjectID = cmd.ProjectID
	item.Title = cmd.Title
	item.Description = cmd.Description
	item.Labels = cmd.Labels
	item.Context = cmd.Context
	item.Details = cmd.Details
	item.TaskStatus = cmd.TaskStatus
	item.Priority = cmd.Priority
	item.WaitingFor = cmd.WaitingFor
	item.ExpectedAt = cmd.ExpectedAt
	item.DueAt = cmd.DueAt
	if err := item.Validate(); err != nil {
		return nil, err
	}
	if err := s.repo.Create(ctx, item); err != nil {
		return nil, err
	}
	return item, nil
}

func (s *Service) Get(ctx context.Context, id string) (*domain.Item, error) {
	return s.repo.Get(ctx, id)
}
func (s *Service) List(ctx context.Context, q domain.ListQuery) ([]*domain.Item, error) {
	return s.repo.List(ctx, q)
}

func (s *Service) Update(ctx context.Context, id string, cmd UpdateItemCmd) (*domain.Item, error) {
	item, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	item.Kind = cmd.Kind
	item.Bucket = cmd.Bucket
	item.ProjectID = cmd.ProjectID
	item.Title = cmd.Title
	item.Description = cmd.Description
	item.Labels = cmd.Labels
	item.Context = cmd.Context
	item.Details = cmd.Details
	item.TaskStatus = cmd.TaskStatus
	item.Priority = cmd.Priority
	item.WaitingFor = cmd.WaitingFor
	item.ExpectedAt = cmd.ExpectedAt
	item.DueAt = cmd.DueAt
	item.UpdatedAt = s.now()
	if err := item.Validate(); err != nil {
		return nil, err
	}
	if err := s.repo.Update(ctx, item); err != nil {
		return nil, err
	}
	return item, nil
}

func (s *Service) MoveBucket(ctx context.Context, id string, bucket domain.Bucket) (*domain.Item, error) {
	item, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := item.UpdateBucket(bucket, s.now()); err != nil {
		return nil, err
	}
	if err := s.repo.Update(ctx, item); err != nil {
		return nil, err
	}
	return item, nil
}

func (s *Service) Delete(ctx context.Context, id string) error { return s.repo.Delete(ctx, id) }
