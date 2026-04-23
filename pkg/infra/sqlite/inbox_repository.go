package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"

	"github.com/anyvoxel/multivac/pkg/domain"
)

type InboxRepository struct{ db *sqlx.DB }

func NewInboxRepository(db *sqlx.DB) *InboxRepository { return &InboxRepository{db: db} }

func (r *InboxRepository) Migrate(ctx context.Context) error {
	const ddl = `
		PRAGMA foreign_keys = ON;

		CREATE TABLE IF NOT EXISTS inboxes (
		  id TEXT PRIMARY KEY,
		  title TEXT NOT NULL,
		  description TEXT NOT NULL,
		  created_at DATETIME NOT NULL,
		  updated_at DATETIME NOT NULL
		);
		CREATE INDEX IF NOT EXISTS idx_inboxes_created_at ON inboxes(created_at);
		CREATE INDEX IF NOT EXISTS idx_inboxes_updated_at ON inboxes(updated_at);
		`
	_, err := r.db.ExecContext(ctx, ddl)
	return err
}

type inboxRepositoryRow struct {
	ID          string       `db:"id"`
	Title       string       `db:"title"`
	Description string       `db:"description"`
	CreatedAt   sql.NullTime `db:"created_at"`
	UpdatedAt   sql.NullTime `db:"updated_at"`
}

func toInboxDomain(row inboxRepositoryRow) *domain.Inbox {
	inbox := &domain.Inbox{ID: row.ID, Title: row.Title, Description: row.Description}
	if row.CreatedAt.Valid {
		inbox.CreatedAt = row.CreatedAt.Time
	}
	if row.UpdatedAt.Valid {
		inbox.UpdatedAt = row.UpdatedAt.Time
	}
	return inbox
}

func applyInboxFilters(base string, q domain.InboxListQuery, args []any) (string, []any) {
	base += " WHERE 1=1"
	if q.Search != "" {
		like := "%" + strings.ToLower(q.Search) + "%"
		base += " AND (LOWER(title) LIKE ? OR LOWER(description) LIKE ?)"
		args = append(args, like, like)
	}
	return base, args
}

func inboxOrderBy(q domain.InboxListQuery) (string, error) {
	orderByMap := map[domain.SortBy]string{
		domain.InboxSortByCreatedAt: "created_at",
		domain.InboxSortByUpdatedAt: "updated_at",
		domain.InboxSortByTitle:     "title",
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

func applyInboxLimitOffset(base string, q domain.InboxListQuery, args []any) (string, []any) {
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

func (r *InboxRepository) Create(ctx context.Context, inbox *domain.Inbox) error {
	if err := inbox.Validate(); err != nil {
		return err
	}
	const q = `INSERT INTO inboxes (id, title, description, created_at, updated_at) VALUES (:id, :title, :description, :created_at, :updated_at);`
	_, err := r.db.NamedExecContext(ctx, q, map[string]any{"id": inbox.ID, "title": inbox.Title, "description": inbox.Description, "created_at": inbox.CreatedAt, "updated_at": inbox.UpdatedAt})
	return err
}

func (r *InboxRepository) Get(ctx context.Context, id string) (*domain.Inbox, error) {
	const q = `SELECT id, title, description, created_at, updated_at FROM inboxes WHERE id = ?;`
	var row inboxRepositoryRow
	if err := r.db.GetContext(ctx, &row, q, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return toInboxDomain(row), nil
}

func (r *InboxRepository) List(ctx context.Context, q domain.InboxListQuery) ([]*domain.Inbox, error) {
	if err := (&q).Validate(); err != nil {
		return nil, err
	}
	base := `SELECT id, title, description, created_at, updated_at FROM inboxes`
	args := []any{}
	base, args = applyInboxFilters(base, q, args)
	ob, err := inboxOrderBy(q)
	if err != nil {
		return nil, err
	}
	base += ob
	base, args = applyInboxLimitOffset(base, q, args)
	rows := make([]inboxRepositoryRow, 0)
	if err := r.db.SelectContext(ctx, &rows, base, args...); err != nil {
		return nil, err
	}
	out := make([]*domain.Inbox, 0, len(rows))
	for _, row := range rows {
		out = append(out, toInboxDomain(row))
	}
	return out, nil
}

func (r *InboxRepository) Update(ctx context.Context, inbox *domain.Inbox) error {
	if err := inbox.Validate(); err != nil {
		return err
	}
	const q = `UPDATE inboxes SET title = :title, description = :description, updated_at = :updated_at WHERE id = :id;`
	res, err := r.db.NamedExecContext(ctx, q, map[string]any{"id": inbox.ID, "title": inbox.Title, "description": inbox.Description, "updated_at": inbox.UpdatedAt})
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

func (r *InboxRepository) Delete(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM inboxes WHERE id = ?;`, id)
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
