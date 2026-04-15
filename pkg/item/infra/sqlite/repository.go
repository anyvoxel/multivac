package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"

	"github.com/anyvoxel/multivac/pkg/item/domain"
)

type Repository struct{ db *sqlx.DB }

func NewRepository(db *sqlx.DB) *Repository { return &Repository{db: db} }

func (r *Repository) Migrate(ctx context.Context) error {
	const ddl = `
PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXISTS items (
  id TEXT PRIMARY KEY,
  kind TEXT NOT NULL,
  bucket TEXT NOT NULL,
  project_id TEXT NULL,
  title TEXT NOT NULL,
  description TEXT NOT NULL,
  context TEXT NOT NULL,
  details TEXT NOT NULL,
  task_status TEXT NOT NULL,
  priority TEXT NOT NULL,
  waiting_for TEXT NOT NULL,
  expected_at DATETIME NULL,
  due_at DATETIME NULL,
  created_at DATETIME NOT NULL,
  updated_at DATETIME NOT NULL,
  FOREIGN KEY(project_id) REFERENCES projects(id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_items_bucket ON items(bucket);
CREATE INDEX IF NOT EXISTS idx_items_kind ON items(kind);
CREATE INDEX IF NOT EXISTS idx_items_project_id ON items(project_id);
CREATE INDEX IF NOT EXISTS idx_items_task_status ON items(task_status);
CREATE INDEX IF NOT EXISTS idx_items_due_at ON items(due_at);
CREATE INDEX IF NOT EXISTS idx_items_expected_at ON items(expected_at);

CREATE TABLE IF NOT EXISTS item_events (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  item_id TEXT NOT NULL,
  event_type TEXT NOT NULL,
  payload TEXT NOT NULL,
  created_at DATETIME NOT NULL,
  FOREIGN KEY(item_id) REFERENCES items(id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_item_events_item_id ON item_events(item_id);
`
	if _, err := r.db.ExecContext(ctx, ddl); err != nil {
		return err
	}
	legacyMigrations := []struct {
		table string
		sql   string
	}{
		{
			table: "inboxes",
			sql: `INSERT OR IGNORE INTO items (id, kind, bucket, project_id, title, description, context, details, task_status, priority, waiting_for, expected_at, due_at, created_at, updated_at)
SELECT id, 'Inbox', 'Inbox', NULL, name, description, '', '', '', '', '', NULL, NULL, created_at, updated_at FROM inboxes;`,
		},
		{
			table: "somedays",
			sql: `INSERT OR IGNORE INTO items (id, kind, bucket, project_id, title, description, context, details, task_status, priority, waiting_for, expected_at, due_at, created_at, updated_at)
SELECT id, 'SomedayMaybe', 'SomedayMaybe', NULL, name, description, '', '', '', '', '', NULL, NULL, created_at, updated_at FROM somedays;`,
		},
		{
			table: "waiting_lists",
			sql: `INSERT OR IGNORE INTO items (id, kind, bucket, project_id, title, description, context, details, task_status, priority, waiting_for, expected_at, due_at, created_at, updated_at)
SELECT id, 'WaitingFor', 'WaitingFor', NULL, name, '', '', details, '', '', owner, expected_at, NULL, created_at, updated_at FROM waiting_lists;`,
		},
		{
			table: "tasks",
			sql: `INSERT OR IGNORE INTO items (id, kind, bucket, project_id, title, description, context, details, task_status, priority, waiting_for, expected_at, due_at, created_at, updated_at)
SELECT id, 'Task', CASE status WHEN 'Done' THEN 'Completed' WHEN 'Canceled' THEN 'Dropped' ELSE 'NextAction' END, NULLIF(project_id, ''), name, description, context, details, status, priority, '', NULL, due_at, created_at, updated_at FROM tasks;`,
		},
	}
	for _, migration := range legacyMigrations {
		var exists int
		if err := r.db.GetContext(ctx, &exists, `SELECT COUNT(1) FROM sqlite_master WHERE type = 'table' AND name = ?;`, migration.table); err != nil {
			return err
		}
		if exists == 0 {
			continue
		}
		if _, err := r.db.ExecContext(ctx, migration.sql); err != nil {
			return err
		}
	}
	return nil
}

type itemRow struct {
	ID          string         `db:"id"`
	Kind        string         `db:"kind"`
	Bucket      string         `db:"bucket"`
	ProjectID   sql.NullString `db:"project_id"`
	Title       string         `db:"title"`
	Description string         `db:"description"`
	Context     string         `db:"context"`
	Details     string         `db:"details"`
	TaskStatus  string         `db:"task_status"`
	Priority    string         `db:"priority"`
	WaitingFor  string         `db:"waiting_for"`
	ExpectedAt  sql.NullTime   `db:"expected_at"`
	DueAt       sql.NullTime   `db:"due_at"`
	CreatedAt   sql.NullTime   `db:"created_at"`
	UpdatedAt   sql.NullTime   `db:"updated_at"`
}

func toDomain(row itemRow) (*domain.Item, error) {
	kind, ok := domain.ParseKind(row.Kind)
	if !ok {
		return nil, domain.ErrInvalidArg
	}
	bucket, ok := domain.ParseBucket(row.Bucket)
	if !ok {
		return nil, domain.ErrInvalidArg
	}
	item := &domain.Item{ID: row.ID, Kind: kind, Bucket: bucket, ProjectID: row.ProjectID.String, Title: row.Title, Description: row.Description, Context: row.Context, Details: row.Details, TaskStatus: row.TaskStatus, Priority: row.Priority, WaitingFor: row.WaitingFor}
	if row.ExpectedAt.Valid {
		v := row.ExpectedAt.Time
		item.ExpectedAt = &v
	}
	if row.DueAt.Valid {
		v := row.DueAt.Time
		item.DueAt = &v
	}
	if row.CreatedAt.Valid {
		item.CreatedAt = row.CreatedAt.Time
	}
	if row.UpdatedAt.Valid {
		item.UpdatedAt = row.UpdatedAt.Time
	}
	return item, nil
}

func applyFilters(base string, q domain.ListQuery, args []any) (string, []any) {
	base += " WHERE 1=1"
	if q.Bucket != nil {
		base += " AND bucket = ?"
		args = append(args, string(*q.Bucket))
	}
	if q.Kind != nil {
		base += " AND kind = ?"
		args = append(args, string(*q.Kind))
	}
	if q.ProjectID != "" {
		base += " AND project_id = ?"
		args = append(args, q.ProjectID)
	}
	if q.TaskStatus != "" {
		base += " AND task_status = ?"
		args = append(args, q.TaskStatus)
	}
	if q.Search != "" {
		like := "%" + strings.ToLower(q.Search) + "%"
		base += " AND (LOWER(title) LIKE ? OR LOWER(description) LIKE ? OR LOWER(context) LIKE ? OR LOWER(details) LIKE ? OR LOWER(waiting_for) LIKE ?)"
		args = append(args, like, like, like, like, like)
	}
	return base, args
}

func orderBy(q domain.ListQuery) (string, error) {
	priorityOrder := "CASE priority WHEN 'P0' THEN 0 WHEN 'High' THEN 1 WHEN 'Medium' THEN 2 WHEN 'Low' THEN 3 ELSE 99 END"
	orderParts := make([]string, 0, len(q.Sorts))
	for _, s := range q.Sorts {
		dir := "ASC"
		if s.Dir == domain.SortDesc {
			dir = "DESC"
		}
		switch s.By {
		case domain.ItemSortByCreatedAt:
			orderParts = append(orderParts, "created_at "+dir)
		case domain.ItemSortByUpdatedAt:
			orderParts = append(orderParts, "updated_at "+dir)
		case domain.ItemSortByTitle:
			orderParts = append(orderParts, "title "+dir)
		case domain.ItemSortByPriority:
			orderParts = append(orderParts, priorityOrder+" "+dir)
		case domain.ItemSortByDueAt:
			orderParts = append(orderParts, "CASE WHEN due_at IS NULL THEN 1 ELSE 0 END ASC", "due_at "+dir)
		case domain.ItemSortByExpectedAt:
			orderParts = append(orderParts, "CASE WHEN expected_at IS NULL THEN 1 ELSE 0 END ASC", "expected_at "+dir)
		default:
			return "", domain.InvalidSortBy(string(s.By))
		}
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

func (r *Repository) Create(ctx context.Context, item *domain.Item) error {
	if err := item.Validate(); err != nil {
		return err
	}
	const q = `
INSERT INTO items (id, kind, bucket, project_id, title, description, context, details, task_status, priority, waiting_for, expected_at, due_at, created_at, updated_at)
VALUES (:id, :kind, :bucket, :project_id, :title, :description, :context, :details, :task_status, :priority, :waiting_for, :expected_at, :due_at, :created_at, :updated_at);`
	var projectID any
	if item.ProjectID != "" {
		projectID = item.ProjectID
	}
	_, err := r.db.NamedExecContext(ctx, q, map[string]any{"id": item.ID, "kind": string(item.Kind), "bucket": string(item.Bucket), "project_id": projectID, "title": item.Title, "description": item.Description, "context": item.Context, "details": item.Details, "task_status": item.TaskStatus, "priority": item.Priority, "waiting_for": item.WaitingFor, "expected_at": timePtrOrNil(item.ExpectedAt), "due_at": timePtrOrNil(item.DueAt), "created_at": item.CreatedAt, "updated_at": item.UpdatedAt})
	return err
}

func (r *Repository) Get(ctx context.Context, id string) (*domain.Item, error) {
	const q = `SELECT id, kind, bucket, project_id, title, description, context, details, task_status, priority, waiting_for, expected_at, due_at, created_at, updated_at FROM items WHERE id = ?;`
	var row itemRow
	if err := r.db.GetContext(ctx, &row, q, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return toDomain(row)
}

func (r *Repository) List(ctx context.Context, q domain.ListQuery) ([]*domain.Item, error) {
	if err := (&q).Validate(); err != nil {
		return nil, err
	}
	base := `SELECT id, kind, bucket, project_id, title, description, context, details, task_status, priority, waiting_for, expected_at, due_at, created_at, updated_at FROM items`
	args := []any{}
	base, args = applyFilters(base, q, args)
	ob, err := orderBy(q)
	if err != nil {
		return nil, err
	}
	base += ob
	base, args = applyLimitOffset(base, q, args)
	rows := make([]itemRow, 0)
	if err := r.db.SelectContext(ctx, &rows, base, args...); err != nil {
		return nil, err
	}
	out := make([]*domain.Item, 0, len(rows))
	for _, row := range rows {
		item, err := toDomain(row)
		if err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, nil
}

func (r *Repository) Update(ctx context.Context, item *domain.Item) error {
	if err := item.Validate(); err != nil {
		return err
	}
	const q = `UPDATE items SET kind = :kind, bucket = :bucket, project_id = :project_id, title = :title, description = :description, context = :context, details = :details, task_status = :task_status, priority = :priority, waiting_for = :waiting_for, expected_at = :expected_at, due_at = :due_at, updated_at = :updated_at WHERE id = :id;`
	var projectID any
	if item.ProjectID != "" {
		projectID = item.ProjectID
	}
	res, err := r.db.NamedExecContext(ctx, q, map[string]any{"id": item.ID, "kind": string(item.Kind), "bucket": string(item.Bucket), "project_id": projectID, "title": item.Title, "description": item.Description, "context": item.Context, "details": item.Details, "task_status": item.TaskStatus, "priority": item.Priority, "waiting_for": item.WaitingFor, "expected_at": timePtrOrNil(item.ExpectedAt), "due_at": timePtrOrNil(item.DueAt), "updated_at": item.UpdatedAt})
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

func (r *Repository) Delete(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM items WHERE id = ?;`, id)
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

func timePtrOrNil(v *time.Time) any {
	if v == nil {
		return nil
	}
	return *v
}
