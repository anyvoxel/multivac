// Package sqlite provides sqlite implementation for project repository.
package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/jmoiron/sqlx"
	// sqlite3 driver
	_ "github.com/mattn/go-sqlite3"

	"github.com/anyvoxel/multivac/pkg/project/domain"
)

func applyProjectFilters(base string, q domain.ListQuery, args []any) (string, []any) {
	base += " WHERE 1=1"
	if q.Status != nil {
		base += " AND status = ?"
		args = append(args, string(*q.Status))
	}
	if q.Search != "" {
		like := "%" + strings.ToLower(q.Search) + "%"
		base += " AND (LOWER(name) LIKE ? OR LOWER(goal) LIKE ? OR LOWER(principles) LIKE ? OR LOWER(vision_result) LIKE ? OR LOWER(description) LIKE ?)"
		args = append(args, like, like, like, like, like)
	}
	return base, args
}

func projectOrderBy(q domain.ListQuery) (string, error) {
	orderByMap := map[domain.SortBy]string{
		domain.ProjectSortByCreatedAt: "created_at",
		domain.ProjectSortByUpdatedAt: "updated_at",
		domain.ProjectSortByName:      "name",
	}
	orderParts := make([]string, 0, len(q.Sorts))
	for _, s := range q.Sorts {
		col, ok := orderByMap[s.By]
		if !ok {
			return "", domain.InvalidSortBy(string(s.By))
		}
		dir := "ASC"
		if s.Dir == domain.SortDesc {
			dir = "DESC"
		}
		orderParts = append(orderParts, col+" "+dir)
	}
	return " ORDER BY " + strings.Join(orderParts, ", "), nil
}

func applyLimitOffset(base string, q domain.ListQuery, args []any) (string, []any) {
	if q.Limit > 0 {
		base += " LIMIT ?"
		args = append(args, q.Limit)
	}
	if q.Offset > 0 {
		base += " OFFSET ?"
		args = append(args, q.Offset)
	}
	return base, args
}

// Repository implements domain.Repository using sqlite.
type Repository struct {
	db *sqlx.DB
}

// NewRepository creates a sqlite-backed project repository.
func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

// Migrate creates required tables and indices.
func (r *Repository) Migrate(ctx context.Context) error {
	const ddl = `
PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXISTS projects (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  goal TEXT NOT NULL,
  principles TEXT NOT NULL,
  vision_result TEXT NOT NULL,
  description TEXT NOT NULL,
  status TEXT NOT NULL,
  started_at DATETIME NULL,
  completed_at DATETIME NULL,
  created_at DATETIME NOT NULL,
  updated_at DATETIME NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_projects_status ON projects(status);
`
	_, err := r.db.ExecContext(ctx, ddl)
	return err
}

type projectRow struct {
	ID           string       `db:"id"`
	Name         string       `db:"name"`
	Goal         string       `db:"goal"`
	Principles   string       `db:"principles"`
	VisionResult string       `db:"vision_result"`
	Description  string       `db:"description"`
	Status       string       `db:"status"`
	StartedAt    sql.NullTime `db:"started_at"`
	CompletedAt  sql.NullTime `db:"completed_at"`
	CreatedAt    sql.NullTime `db:"created_at"`
	UpdatedAt    sql.NullTime `db:"updated_at"`
}

func toDomain(row projectRow) (*domain.Project, error) {
	status, ok := domain.ParseStatus(row.Status)
	if !ok {
		return nil, domain.ErrInvalidArg
	}

	p := &domain.Project{
		ID:           row.ID,
		Name:         row.Name,
		Goal:         row.Goal,
		Principles:   row.Principles,
		VisionResult: row.VisionResult,
		Description:  row.Description,
		Status:       status,
	}
	if row.StartedAt.Valid {
		t := row.StartedAt.Time
		p.StartedAt = &t
	}
	if row.CompletedAt.Valid {
		t := row.CompletedAt.Time
		p.CompletedAt = &t
	}
	if row.CreatedAt.Valid {
		p.CreatedAt = row.CreatedAt.Time
	}
	if row.UpdatedAt.Valid {
		p.UpdatedAt = row.UpdatedAt.Time
	}
	return p, nil
}

// Create inserts a new project.
func (r *Repository) Create(ctx context.Context, p *domain.Project) error {
	const q = `
INSERT INTO projects (
  id, name, goal, principles, vision_result, description, status,
  started_at, completed_at, created_at, updated_at
) VALUES (
  :id, :name, :goal, :principles, :vision_result, :description, :status,
  :started_at, :completed_at, :created_at, :updated_at
);
`
	params := map[string]any{
		"id":            p.ID,
		"name":          p.Name,
		"goal":          p.Goal,
		"principles":    p.Principles,
		"vision_result": p.VisionResult,
		"description":   p.Description,
		"status":        string(p.Status),
		"started_at":    p.StartedAt,
		"completed_at":  p.CompletedAt,
		"created_at":    p.CreatedAt,
		"updated_at":    p.UpdatedAt,
	}
	_, err := r.db.NamedExecContext(ctx, q, params)
	return err
}

// Get loads a project by id.
func (r *Repository) Get(ctx context.Context, id string) (*domain.Project, error) {
	const q = `
SELECT id, name, goal, principles, vision_result, description, status,
       started_at, completed_at, created_at, updated_at
FROM projects
WHERE id = ?;
`
	var row projectRow
	if err := r.db.GetContext(ctx, &row, q, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return toDomain(row)
}

// List returns projects by query.
func (r *Repository) List(ctx context.Context, q domain.ListQuery) ([]*domain.Project, error) {
	if err := (&q).Validate(); err != nil {
		return nil, err
	}

	base := `
SELECT id, name, goal, principles, vision_result, description, status,
       started_at, completed_at, created_at, updated_at
FROM projects
`
	args := []any{}
	base, args = applyProjectFilters(base, q, args)
	orderBy, err := projectOrderBy(q)
	if err != nil {
		return nil, err
	}
	base += orderBy
	base, args = applyLimitOffset(base, q, args)

	rows := make([]projectRow, 0)
	if err := r.db.SelectContext(ctx, &rows, base, args...); err != nil {
		return nil, err
	}
	out := make([]*domain.Project, 0, len(rows))
	for _, row := range rows {
		p, err := toDomain(row)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, nil
}

// Update updates a project by id.
func (r *Repository) Update(ctx context.Context, p *domain.Project) error {
	const q = `
UPDATE projects SET
  name = :name,
  goal = :goal,
  principles = :principles,
  vision_result = :vision_result,
  description = :description,
  status = :status,
  started_at = :started_at,
  completed_at = :completed_at,
  updated_at = :updated_at
WHERE id = :id;
`
	params := map[string]any{
		"id":            p.ID,
		"name":          p.Name,
		"goal":          p.Goal,
		"principles":    p.Principles,
		"vision_result": p.VisionResult,
		"description":   p.Description,
		"status":        string(p.Status),
		"started_at":    p.StartedAt,
		"completed_at":  p.CompletedAt,
		"updated_at":    p.UpdatedAt,
	}
	res, err := r.db.NamedExecContext(ctx, q, params)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// Delete removes a project by id.
func (r *Repository) Delete(ctx context.Context, id string) error {
	const q = `DELETE FROM projects WHERE id = ?;`
	res, err := r.db.ExecContext(ctx, q, id)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return domain.ErrNotFound
	}
	return nil
}
