// Package sqlite provides sqlite implementation for task repository.
package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/jmoiron/sqlx"
	// sqlite3 driver
	_ "github.com/mattn/go-sqlite3"

	"github.com/anyvoxel/multivac/pkg/task/domain"
)

func applyTaskFilters(base string, q domain.ListQuery, args []any) (string, []any) {
	if q.ProjectID != "" {
		base += " AND project_id = ?"
		args = append(args, q.ProjectID)
	}
	if q.Status != nil {
		base += " AND status = ?"
		args = append(args, string(*q.Status))
	}
	return base, args
}

func taskOrderBy(q domain.ListQuery) (string, error) {
	priorityOrder := "CASE priority WHEN 'P0' THEN 0 WHEN 'High' THEN 1 WHEN 'Medium' THEN 2 WHEN 'Low' THEN 3 ELSE 99 END"
	orderByMap := map[domain.SortBy]string{
		domain.TaskSortByCreatedAt: "created_at",
		domain.TaskSortByUpdatedAt: "updated_at",
		domain.TaskSortByDueAt:     "due_at",
		domain.TaskSortByPriority:  priorityOrder,
	}
	orderParts := make([]string, 0, len(q.Sorts))
	for _, s := range q.Sorts {
		expr, ok := orderByMap[s.By]
		if !ok {
			return "", domain.ErrInvalidArg
		}
		dir := "ASC"
		if s.Dir == domain.SortDesc {
			dir = "DESC"
		}
		orderParts = append(orderParts, expr+" "+dir)
	}
	return " ORDER BY " + strings.Join(orderParts, ", "), nil
}

func applyTaskLimitOffset(base string, q domain.ListQuery, args []any) (string, []any) {
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

// NewRepository creates a sqlite-backed task repository.
func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

// Migrate creates required tables and indices.
func (r *Repository) Migrate(ctx context.Context) error {
	const ddl = `
PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXISTS tasks (
  id TEXT PRIMARY KEY,
  project_id TEXT NOT NULL,
  name TEXT NOT NULL,
  description TEXT NOT NULL,
  context TEXT NOT NULL,
  details TEXT NOT NULL,
  status TEXT NOT NULL,
  priority TEXT NOT NULL,
  due_at DATETIME NULL,
  created_at DATETIME NOT NULL,
  updated_at DATETIME NOT NULL,
  FOREIGN KEY(project_id) REFERENCES projects(id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_tasks_project_id ON tasks(project_id);
CREATE INDEX IF NOT EXISTS idx_tasks_project_status ON tasks(project_id, status);
CREATE INDEX IF NOT EXISTS idx_tasks_project_priority ON tasks(project_id, priority);
`
	_, err := r.db.ExecContext(ctx, ddl)
	return err
}

type taskRow struct {
	ID          string       `db:"id"`
	ProjectID   string       `db:"project_id"`
	Name        string       `db:"name"`
	Description string       `db:"description"`
	Context     string       `db:"context"`
	Details     string       `db:"details"`
	Status      string       `db:"status"`
	Priority    string       `db:"priority"`
	DueAt       sql.NullTime `db:"due_at"`
	CreatedAt   sql.NullTime `db:"created_at"`
	UpdatedAt   sql.NullTime `db:"updated_at"`
}

func toDomain(row taskRow) (*domain.Task, error) {
	st, ok := domain.ParseStatus(row.Status)
	if !ok {
		return nil, domain.ErrInvalidArg
	}
	pr, ok := domain.ParsePriority(row.Priority)
	if !ok {
		return nil, domain.ErrInvalidArg
	}

	t := &domain.Task{
		ID:          row.ID,
		ProjectID:   row.ProjectID,
		Name:        row.Name,
		Description: row.Description,
		Context:     row.Context,
		Details:     row.Details,
		Status:      st,
		Priority:    pr,
	}
	if row.DueAt.Valid {
		d := row.DueAt.Time
		t.DueAt = &d
	}
	if row.CreatedAt.Valid {
		t.CreatedAt = row.CreatedAt.Time
	}
	if row.UpdatedAt.Valid {
		t.UpdatedAt = row.UpdatedAt.Time
	}
	return t, nil
}

// Create inserts a new task.
func (r *Repository) Create(ctx context.Context, t *domain.Task) error {
	const q = `
INSERT INTO tasks (
  id, project_id, name, description, context, details, status, priority,
  due_at, created_at, updated_at
) VALUES (
  :id, :project_id, :name, :description, :context, :details, :status, :priority,
  :due_at, :created_at, :updated_at
);
`
	params := map[string]any{
		"id":          t.ID,
		"project_id":  t.ProjectID,
		"name":        t.Name,
		"description": t.Description,
		"context":     t.Context,
		"details":     t.Details,
		"status":      string(t.Status),
		"priority":    string(t.Priority),
		"due_at":      t.DueAt,
		"created_at":  t.CreatedAt,
		"updated_at":  t.UpdatedAt,
	}
	_, err := r.db.NamedExecContext(ctx, q, params)
	return err
}

// Get loads a task by id.
func (r *Repository) Get(ctx context.Context, id string) (*domain.Task, error) {
	const q = `
SELECT id, project_id, name, description, context, details, status, priority,
       due_at, created_at, updated_at
FROM tasks
WHERE id = ?;
`
	var row taskRow
	if err := r.db.GetContext(ctx, &row, q, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return toDomain(row)
}

// List returns tasks by query.
func (r *Repository) List(ctx context.Context, q domain.ListQuery) ([]*domain.Task, error) {
	if err := (&q).Validate(); err != nil {
		return nil, err
	}

	base := `
SELECT id, project_id, name, description, context, details, status, priority,
       due_at, created_at, updated_at
FROM tasks
WHERE 1=1
`
	args := []any{}
	base, args = applyTaskFilters(base, q, args)
	orderBy, err := taskOrderBy(q)
	if err != nil {
		return nil, err
	}
	base += orderBy
	base, args = applyTaskLimitOffset(base, q, args)

	rows := make([]taskRow, 0)
	if err := r.db.SelectContext(ctx, &rows, base, args...); err != nil {
		return nil, err
	}
	out := make([]*domain.Task, 0, len(rows))
	for _, row := range rows {
		t, err := toDomain(row)
		if err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, nil
}

// Update updates a task by id.
func (r *Repository) Update(ctx context.Context, t *domain.Task) error {
	const q = `
UPDATE tasks SET
  name = :name,
  description = :description,
  context = :context,
  details = :details,
  status = :status,
  priority = :priority,
  due_at = :due_at,
  updated_at = :updated_at
WHERE id = :id;
`
	params := map[string]any{
		"id":          t.ID,
		"name":        t.Name,
		"description": t.Description,
		"context":     t.Context,
		"details":     t.Details,
		"status":      string(t.Status),
		"priority":    string(t.Priority),
		"due_at":      t.DueAt,
		"updated_at":  t.UpdatedAt,
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

// Delete removes a task by id.
func (r *Repository) Delete(ctx context.Context, id string) error {
	const q = `DELETE FROM tasks WHERE id = ?;`
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
