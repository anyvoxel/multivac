// Package sqlite provides sqlite implementation for project repository.
package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"

	"github.com/anyvoxel/multivac/pkg/domain"
)

func applyProjectFilters(base string, q domain.ProjectListQuery, args []any) (string, []any) {
	base += " WHERE 1=1"
	if q.Status != nil {
		base += " AND status = ?"
		args = append(args, string(*q.Status))
	}
	if q.Search != "" {
		like := "%" + strings.ToLower(q.Search) + "%"
		base += " AND (LOWER(title) LIKE ? OR LOWER(description) LIKE ?)"
		args = append(args, like, like)
	}
	return base, args
}

func projectOrderBy(q domain.ProjectListQuery) (string, error) {
	orderByMap := map[domain.SortBy]string{
		domain.ProjectSortByCreatedAt: "created_at",
		domain.ProjectSortByUpdatedAt: "updated_at",
		domain.ProjectSortByTitle:     "title",
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

func applyProjectLimitOffset(base string, q domain.ProjectListQuery, args []any) (string, []any) {
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
type ProjectRepository struct {
	db *sqlx.DB
}

// NewRepository creates a sqlite-backed project repository.
func NewProjectRepository(db *sqlx.DB) *ProjectRepository {
	return &ProjectRepository{db: db}
}

// Migrate creates required tables and indices.
func (r *ProjectRepository) Migrate(ctx context.Context) error {
	const ddl = `
		CREATE TABLE IF NOT EXISTS projects (
		  id TEXT PRIMARY KEY,
		  title TEXT NOT NULL,
		  description TEXT NOT NULL,
		  goals JSONB NOT NULL DEFAULT (jsonb('[]')),
		  "references" JSONB NOT NULL DEFAULT (jsonb('[]')),
		  status TEXT NOT NULL,
		  start_at DATETIME NULL,
		  completed_at DATETIME NULL,
		  created_at DATETIME NOT NULL,
		  updated_at DATETIME NOT NULL
		);
		CREATE INDEX IF NOT EXISTS idx_projects_status ON projects(status);
		CREATE INDEX IF NOT EXISTS idx_projects_created_at ON projects(created_at);
		CREATE INDEX IF NOT EXISTS idx_projects_updated_at ON projects(updated_at);
	`
	_, err := r.db.ExecContext(ctx, ddl)
	return err
}

type projectRow struct {
	ID          string       `db:"id"`
	Title       string       `db:"title"`
	Description string       `db:"description"`
	Goals       []byte       `db:"goals"`
	References  []byte       `db:"project_references"`
	Status      string       `db:"status"`
	StartAt     sql.NullTime `db:"start_at"`
	CompletedAt sql.NullTime `db:"completed_at"`
	CreatedAt   sql.NullTime `db:"created_at"`
	UpdatedAt   sql.NullTime `db:"updated_at"`
}

func toProjectDomain(row projectRow) (*domain.Project, error) {
	status, ok := domain.ParseStatus(row.Status)
	if !ok {
		return nil, domain.ErrInvalidArg
	}
	goals, err := unmarshalProjectGoals(row.Goals)
	if err != nil {
		return nil, err
	}
	references, err := unmarshalProjectReferences(row.References)
	if err != nil {
		return nil, err
	}
	p := &domain.Project{
		ID:          row.ID,
		Title:       row.Title,
		Description: row.Description,
		Goals:       goals,
		References:  references,
		Status:      status,
	}
	if row.StartAt.Valid {
		t := row.StartAt.Time
		p.StartAt = &t
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

func marshalProjectGoals(goals []domain.ProjectGoal) (string, error) {
	if len(goals) == 0 {
		return "[]", nil
	}
	data, err := json.Marshal(goals)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func unmarshalProjectGoals(raw []byte) ([]domain.ProjectGoal, error) {
	raw = []byte(strings.TrimSpace(string(raw)))
	if len(raw) == 0 {
		return nil, nil
	}
	var goals []domain.ProjectGoal
	if err := json.Unmarshal(raw, &goals); err != nil {
		return nil, err
	}
	return goals, nil
}

func marshalProjectReferences(references []domain.ProjectReference) (string, error) {
	if len(references) == 0 {
		return "[]", nil
	}
	data, err := json.Marshal(references)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func unmarshalProjectReferences(raw []byte) ([]domain.ProjectReference, error) {
	raw = []byte(strings.TrimSpace(string(raw)))
	if len(raw) == 0 {
		return nil, nil
	}
	var references []domain.ProjectReference
	if err := json.Unmarshal(raw, &references); err != nil {
		return nil, err
	}
	return references, nil
}

// Create inserts a new project.
func (r *ProjectRepository) Create(ctx context.Context, p *domain.Project) error {
	goals, err := marshalProjectGoals(p.Goals)
	if err != nil {
		return err
	}
	references, err := marshalProjectReferences(p.References)
	if err != nil {
		return err
	}
	const q = `
		INSERT INTO projects (
		  id, title, description, goals, "references", status, start_at, completed_at, created_at, updated_at
		) VALUES (
		  :id, :title, :description, jsonb(:goals), jsonb(:references), :status, :start_at, :completed_at, :created_at, :updated_at
		);
	`
	_, err = r.db.NamedExecContext(ctx, q, map[string]any{
		"id":           p.ID,
		"title":        p.Title,
		"description":  p.Description,
		"goals":        goals,
		"references":   references,
		"status":       string(p.Status),
		"start_at":     p.StartAt,
		"completed_at": p.CompletedAt,
		"created_at":   p.CreatedAt,
		"updated_at":   p.UpdatedAt,
	})
	return err
}

// Get loads a project by id.
func (r *ProjectRepository) Get(ctx context.Context, id string) (*domain.Project, error) {
	const q = `
		SELECT id, title, description, json(goals) AS goals, json("references") AS project_references, status, start_at, completed_at, created_at, updated_at
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
	return toProjectDomain(row)
}

// List returns projects by query.
func (r *ProjectRepository) List(ctx context.Context, q domain.ProjectListQuery) ([]*domain.Project, error) {
	if err := (&q).Validate(); err != nil {
		return nil, err
	}
	base := `
		SELECT id, title, description, json(goals) AS goals, json("references") AS project_references, status, start_at, completed_at, created_at, updated_at
		FROM projects
	`
	args := []any{}
	base, args = applyProjectFilters(base, q, args)
	orderBy, err := projectOrderBy(q)
	if err != nil {
		return nil, err
	}
	base += orderBy
	base, args = applyProjectLimitOffset(base, q, args)
	rows := make([]projectRow, 0)
	if err := r.db.SelectContext(ctx, &rows, base, args...); err != nil {
		return nil, err
	}
	out := make([]*domain.Project, 0, len(rows))
	for _, row := range rows {
		p, err := toProjectDomain(row)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, nil
}

// Update updates a project by id.
func (r *ProjectRepository) Update(ctx context.Context, p *domain.Project) error {
	goals, err := marshalProjectGoals(p.Goals)
	if err != nil {
		return err
	}
	references, err := marshalProjectReferences(p.References)
	if err != nil {
		return err
	}
	const q = `
		UPDATE projects SET
		  title = :title,
		  description = :description,
		  goals = jsonb(:goals),
		  "references" = jsonb(:references),
		  status = :status,
		  start_at = :start_at,
		  completed_at = :completed_at,
		  updated_at = :updated_at
		WHERE id = :id;
	`
	res, err := r.db.NamedExecContext(ctx, q, map[string]any{
		"id":           p.ID,
		"title":        p.Title,
		"description":  p.Description,
		"goals":        goals,
		"references":   references,
		"status":       string(p.Status),
		"start_at":     p.StartAt,
		"completed_at": p.CompletedAt,
		"updated_at":   p.UpdatedAt,
	})
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
func (r *ProjectRepository) Delete(ctx context.Context, id string) error {
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
