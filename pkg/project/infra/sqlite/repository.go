// Package sqlite provides sqlite implementation for project repository.
package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"time"

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
		base += " AND (LOWER(name) LIKE ? OR LOWER(goals) LIKE ? OR LOWER(description) LIKE ? OR LOWER(labels) LIKE ?)"
		args = append(args, like, like, like, like)
	}
	for _, context := range q.Contexts {
		base += " AND EXISTS (SELECT 1 FROM json_each(projects.contexts) WHERE LOWER(value) = ?)"
		args = append(args, strings.ToLower(context))
	}
	for _, tag := range q.Tags {
		base += " AND EXISTS (SELECT 1 FROM json_each(projects.tags) WHERE LOWER(value) = ?)"
		args = append(args, strings.ToLower(tag))
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
  goals TEXT NOT NULL DEFAULT '[]',
  principles TEXT NOT NULL,
  vision_result TEXT NOT NULL,
  description TEXT NOT NULL,
  labels TEXT NOT NULL DEFAULT '',
  contexts TEXT NOT NULL DEFAULT '[]',
  tags TEXT NOT NULL DEFAULT '[]',
  links TEXT NOT NULL DEFAULT '[]',
  status TEXT NOT NULL,
  started_at DATETIME NULL,
  completed_at DATETIME NULL,
  created_at DATETIME NOT NULL,
  updated_at DATETIME NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_projects_status ON projects(status);
`
	if _, err := r.db.ExecContext(ctx, ddl); err != nil {
		return err
	}
	if _, err := r.db.ExecContext(ctx, `ALTER TABLE projects ADD COLUMN labels TEXT NOT NULL DEFAULT '[]';`); err != nil && !strings.Contains(strings.ToLower(err.Error()), "duplicate column name") {
		return err
	}
	if _, err := r.db.ExecContext(ctx, `ALTER TABLE projects ADD COLUMN contexts TEXT NOT NULL DEFAULT '[]';`); err != nil && !strings.Contains(strings.ToLower(err.Error()), "duplicate column name") {
		return err
	}
	if _, err := r.db.ExecContext(ctx, `ALTER TABLE projects ADD COLUMN tags TEXT NOT NULL DEFAULT '[]';`); err != nil && !strings.Contains(strings.ToLower(err.Error()), "duplicate column name") {
		return err
	}
	if _, err := r.db.ExecContext(ctx, `ALTER TABLE projects ADD COLUMN links TEXT NOT NULL DEFAULT '[]';`); err != nil && !strings.Contains(strings.ToLower(err.Error()), "duplicate column name") {
		return err
	}
	if _, err := r.db.ExecContext(ctx, `ALTER TABLE projects ADD COLUMN goals TEXT NOT NULL DEFAULT '[]';`); err != nil && !strings.Contains(strings.ToLower(err.Error()), "duplicate column name") {
		return err
	}
	if _, err := r.db.ExecContext(ctx, `UPDATE projects
SET goals = json_array(json_object('text', goal, 'completed', 0, 'createdAt', created_at, 'completedAt', NULL))
WHERE TRIM(COALESCE(goal, '')) <> '' AND (goals IS NULL OR TRIM(goals) = '' OR TRIM(goals) = '[]');`); err != nil {
		return err
	}
	if _, err := r.db.ExecContext(ctx, `UPDATE projects
SET description = TRIM(
	COALESCE(NULLIF(description, ''), '') ||
	CASE WHEN TRIM(COALESCE(principles, '')) <> '' THEN
		CASE WHEN TRIM(COALESCE(description, '')) <> '' THEN '\n\n## Principles\n\n' ELSE '## Principles\n\n' END || TRIM(principles)
	ELSE '' END ||
	CASE WHEN TRIM(COALESCE(vision_result, '')) <> '' THEN
		CASE
			WHEN TRIM(COALESCE(description, '')) <> '' OR TRIM(COALESCE(principles, '')) <> '' THEN '\n\n## Vision Result\n\n'
			ELSE '## Vision Result\n\n'
		END || TRIM(vision_result)
	ELSE '' END
)
WHERE (description NOT LIKE '%## Principles%' AND description NOT LIKE '%## Vision Result%');`); err != nil {
		return err
	}
	return nil
}

type projectRow struct {
	ID          string       `db:"id"`
	Name        string       `db:"name"`
	Goals       string       `db:"goals"`
	Description string       `db:"description"`
	Labels      string       `db:"labels"`
	Contexts    string       `db:"contexts"`
	Tags        string       `db:"tags"`
	Links       string       `db:"links"`
	Status      string       `db:"status"`
	StartedAt   sql.NullTime `db:"started_at"`
	CompletedAt sql.NullTime `db:"completed_at"`
	CreatedAt   sql.NullTime `db:"created_at"`
	UpdatedAt   sql.NullTime `db:"updated_at"`
}

func toDomain(row projectRow) (*domain.Project, error) {
	status, ok := domain.ParseStatus(row.Status)
	if !ok {
		return nil, domain.ErrInvalidArg
	}

	links, err := unmarshalLinks(row.Links)
	if err != nil {
		return nil, err
	}
	goals, err := unmarshalGoals(row.Goals)
	if err != nil {
		return nil, err
	}

	labels, err := unmarshalLabels(row.Labels)
	if err != nil {
		return nil, err
	}
	p := &domain.Project{
		ID:          row.ID,
		Name:        row.Name,
		Goals:       goals,
		Description: row.Description,
		Labels:      labels,
		Links:       links,
		Status:      status,
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

func marshalLinks(links []domain.Link) (string, error) {
	if len(links) == 0 {
		return "[]", nil
	}
	data, err := json.Marshal(links)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func marshalLabels(labels []domain.Label) (string, error) {
	if len(labels) == 0 {
		return "[]", nil
	}
	data, err := json.Marshal(labels)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func marshalGoals(goals []domain.Goal) (string, error) {
	if len(goals) == 0 {
		return "[]", nil
	}
	data, err := json.Marshal(goals)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func unmarshalGoals(raw string) ([]domain.Goal, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}

	var goals []domain.Goal
	if err := json.Unmarshal([]byte(raw), &goals); err == nil {
		return goals, nil
	}

	type goalCompatRaw struct {
		Text        string          `json:"text"`
		Completed   json.RawMessage `json:"completed"`
		CreatedAt   string          `json:"createdAt"`
		CompletedAt *string         `json:"completedAt"`
	}
	var compat []goalCompatRaw
	if err := json.Unmarshal([]byte(raw), &compat); err != nil {
		return nil, err
	}
	out := make([]domain.Goal, 0, len(compat))
	for _, item := range compat {
		createdAt, err := parseGoalTime(item.CreatedAt)
		if err != nil {
			return nil, err
		}
		completed, err := parseGoalCompleted(item.Completed)
		if err != nil {
			return nil, err
		}
		goal := domain.Goal{
			Text:      item.Text,
			Completed: completed,
			CreatedAt: createdAt,
		}
		if item.CompletedAt != nil && strings.TrimSpace(*item.CompletedAt) != "" {
			completedAt, err := parseGoalTime(*item.CompletedAt)
			if err != nil {
				return nil, err
			}
			goal.CompletedAt = &completedAt
		}
		out = append(out, goal)
	}
	return out, nil
}

func parseGoalTime(raw string) (time.Time, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return time.Time{}, domain.ErrInvalidArg
	}
	layouts := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02 15:04:05.999999999Z07:00",
		"2006-01-02 15:04:05Z07:00",
		"2006-01-02 15:04:05.999999999",
		"2006-01-02 15:04:05",
	}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, raw); err == nil {
			return t, nil
		}
	}
	return time.Time{}, domain.ErrInvalidArg
}

func parseGoalCompleted(raw json.RawMessage) (bool, error) {
	trimmed := strings.TrimSpace(string(raw))
	if trimmed == "" || trimmed == "null" {
		return false, nil
	}
	var b bool
	if err := json.Unmarshal(raw, &b); err == nil {
		return b, nil
	}
	var n int
	if err := json.Unmarshal(raw, &n); err == nil {
		return n != 0, nil
	}
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		s = strings.ToLower(strings.TrimSpace(s))
		switch s {
		case "true", "1", "yes", "y":
			return true, nil
		case "false", "0", "no", "n", "":
			return false, nil
		}
	}
	return false, domain.ErrInvalidArg
}

func unmarshalLabels(raw string) ([]domain.Label, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	var labels []domain.Label
	if err := json.Unmarshal([]byte(raw), &labels); err == nil {
		return labels, nil
	}
	return parseLegacyLabels(raw), nil
}

func parseLegacyLabels(raw string) []domain.Label {
	seen := map[string]struct{}{}
	labels := make([]domain.Label, 0)
	for _, token := range strings.Fields(raw) {
		normalized := strings.ToLower(strings.TrimSpace(token))
		if normalized == "" {
			continue
		}
		label := domain.Label{Value: normalized, Kind: domain.LabelKindTag, Filterable: false}
		switch {
		case strings.HasPrefix(normalized, "@"):
			value := strings.TrimSpace(strings.TrimPrefix(normalized, "@"))
			if value == "" {
				continue
			}
			label.Value = value
			label.Kind = domain.LabelKindContext
			label.Filterable = true
		case strings.HasPrefix(normalized, "#"):
			value := strings.TrimSpace(strings.TrimPrefix(normalized, "#"))
			if value == "" {
				continue
			}
			label.Value = value
			label.Kind = domain.LabelKindTag
			label.Filterable = true
		}
		key := string(label.Kind) + ":" + label.Value
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		labels = append(labels, label)
	}
	if len(labels) == 0 {
		return nil
	}
	return labels
}

func parseLabels(labels []domain.Label) (contexts []string, tags []string) {
	if len(labels) == 0 {
		return nil, nil
	}
	contextSet := map[string]struct{}{}
	tagSet := map[string]struct{}{}
	for _, label := range labels {
		value := strings.ToLower(strings.TrimSpace(label.Value))
		if value == "" || !label.Filterable {
			continue
		}
		if label.Kind == domain.LabelKindContext {
			contextSet[value] = struct{}{}
			continue
		}
		tagSet[value] = struct{}{}
	}
	for v := range contextSet {
		contexts = append(contexts, v)
	}
	for v := range tagSet {
		tags = append(tags, v)
	}
	return contexts, tags
}

func unmarshalLinks(raw string) ([]domain.Link, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, nil
	}
	var links []domain.Link
	if err := json.Unmarshal([]byte(raw), &links); err != nil {
		return nil, err
	}
	return links, nil
}

// Create inserts a new project.
func (r *Repository) Create(ctx context.Context, p *domain.Project) error {
	links, err := marshalLinks(p.Links)
	if err != nil {
		return err
	}
	const q = `
INSERT INTO projects (
  id, name, goal, goals, principles, vision_result, description, labels, contexts, tags, links, status,
  started_at, completed_at, created_at, updated_at
) VALUES (
  :id, :name, '', :goals, '', '', :description, :labels, :contexts, :tags, :links, :status,
  :started_at, :completed_at, :created_at, :updated_at
);
`
	labelsJSON, err := marshalLabels(p.Labels)
	if err != nil {
		return err
	}
	goalsJSON, err := marshalGoals(p.Goals)
	if err != nil {
		return err
	}
	contexts, tags := parseLabels(p.Labels)
	contextsJSON, err := json.Marshal(contexts)
	if err != nil {
		return err
	}
	tagsJSON, err := json.Marshal(tags)
	if err != nil {
		return err
	}
	params := map[string]any{
		"id":            p.ID,
		"name":          p.Name,
		"goals":         goalsJSON,
		"description":   p.Description,
		"labels":        labelsJSON,
		"contexts":      string(contextsJSON),
		"tags":          string(tagsJSON),
		"links":         links,
		"status":        string(p.Status),
		"started_at":    p.StartedAt,
		"completed_at":  p.CompletedAt,
		"created_at":    p.CreatedAt,
		"updated_at":    p.UpdatedAt,
	}
	_, err = r.db.NamedExecContext(ctx, q, params)
	return err
}

// Get loads a project by id.
func (r *Repository) Get(ctx context.Context, id string) (*domain.Project, error) {
	const q = `
SELECT id, name, goals, description, labels, contexts, tags, links, status,
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
SELECT id, name, goals, description, labels, contexts, tags, links, status,
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
	links, err := marshalLinks(p.Links)
	if err != nil {
		return err
	}
	const q = `
UPDATE projects SET
  name = :name,
  goals = :goals,
  principles = '',
  vision_result = '',
  description = :description,
  labels = :labels,
  contexts = :contexts,
  tags = :tags,
  links = :links,
  status = :status,
  started_at = :started_at,
  completed_at = :completed_at,
  updated_at = :updated_at
WHERE id = :id;
`
	labelsJSON, err := marshalLabels(p.Labels)
	if err != nil {
		return err
	}
	goalsJSON, err := marshalGoals(p.Goals)
	if err != nil {
		return err
	}
	contexts, tags := parseLabels(p.Labels)
	contextsJSON, err := json.Marshal(contexts)
	if err != nil {
		return err
	}
	tagsJSON, err := json.Marshal(tags)
	if err != nil {
		return err
	}
	params := map[string]any{
		"id":            p.ID,
		"name":          p.Name,
		"goals":         goalsJSON,
		"description":   p.Description,
		"labels":        labelsJSON,
		"contexts":      string(contextsJSON),
		"tags":          string(tagsJSON),
		"links":         links,
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
