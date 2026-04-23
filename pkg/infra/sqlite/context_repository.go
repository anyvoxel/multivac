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

type ContextRepository struct{ db *sqlx.DB }

func NewContextRepository(db *sqlx.DB) *ContextRepository { return &ContextRepository{db: db} }

func (r *ContextRepository) Migrate(ctx context.Context) error {
	const ddl = `
		PRAGMA foreign_keys = ON;

		CREATE TABLE IF NOT EXISTS contexts (
		  id TEXT PRIMARY KEY,
		  title TEXT NOT NULL,
		  description TEXT NOT NULL,
		  color TEXT NOT NULL,
		  created_at DATETIME NOT NULL,
		  updated_at DATETIME NOT NULL
		);
		CREATE INDEX IF NOT EXISTS idx_contexts_created_at ON contexts(created_at);
		CREATE INDEX IF NOT EXISTS idx_contexts_updated_at ON contexts(updated_at);
		`
	_, err := r.db.ExecContext(ctx, ddl)
	return err
}

type contextRow struct {
	ID          string       `db:"id"`
	Title       string       `db:"title"`
	Description string       `db:"description"`
	Color       string       `db:"color"`
	CreatedAt   sql.NullTime `db:"created_at"`
	UpdatedAt   sql.NullTime `db:"updated_at"`
}

func toContextDomain(row contextRow) *domain.Context {
	contextObj := &domain.Context{ID: row.ID, Title: row.Title, Description: row.Description, Color: row.Color}
	if row.CreatedAt.Valid {
		contextObj.CreatedAt = row.CreatedAt.Time
	}
	if row.UpdatedAt.Valid {
		contextObj.UpdatedAt = row.UpdatedAt.Time
	}
	return contextObj
}

func applyContextFilters(base string, q domain.ContextListQuery, args []any) (string, []any) {
	base += " WHERE 1=1"
	if q.Search != "" {
		like := "%" + strings.ToLower(q.Search) + "%"
		base += " AND (LOWER(title) LIKE ? OR LOWER(description) LIKE ?)"
		args = append(args, like, like)
	}
	return base, args
}

func contextOrderBy(q domain.ContextListQuery) (string, error) {
	orderByMap := map[domain.SortBy]string{
		domain.ContextSortByCreatedAt: "created_at",
		domain.ContextSortByUpdatedAt: "updated_at",
		domain.ContextSortByTitle:     "title",
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

func applyContextLimitOffset(base string, q domain.ContextListQuery, args []any) (string, []any) {
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

func (r *ContextRepository) Create(ctx context.Context, contextObj *domain.Context) error {
	if err := contextObj.Validate(); err != nil {
		return err
	}
	const q = `INSERT INTO contexts (id, title, description, color, created_at, updated_at) VALUES (:id, :title, :description, :color, :created_at, :updated_at);`
	_, err := r.db.NamedExecContext(ctx, q, map[string]any{"id": contextObj.ID, "title": contextObj.Title, "description": contextObj.Description, "color": contextObj.Color, "created_at": contextObj.CreatedAt, "updated_at": contextObj.UpdatedAt})
	return err
}

func (r *ContextRepository) Get(ctx context.Context, id string) (*domain.Context, error) {
	const q = `SELECT id, title, description, color, created_at, updated_at FROM contexts WHERE id = ?;`
	var row contextRow
	if err := r.db.GetContext(ctx, &row, q, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return toContextDomain(row), nil
}

func (r *ContextRepository) List(ctx context.Context, q domain.ContextListQuery) ([]*domain.Context, error) {
	if err := (&q).Validate(); err != nil {
		return nil, err
	}
	base := `SELECT id, title, description, color, created_at, updated_at FROM contexts`
	args := []any{}
	base, args = applyContextFilters(base, q, args)
	ob, err := contextOrderBy(q)
	if err != nil {
		return nil, err
	}
	base += ob
	base, args = applyContextLimitOffset(base, q, args)
	rows := make([]contextRow, 0)
	if err := r.db.SelectContext(ctx, &rows, base, args...); err != nil {
		return nil, err
	}
	out := make([]*domain.Context, 0, len(rows))
	for _, row := range rows {
		out = append(out, toContextDomain(row))
	}
	return out, nil
}

func (r *ContextRepository) Update(ctx context.Context, contextObj *domain.Context) error {
	if err := contextObj.Validate(); err != nil {
		return err
	}
	const q = `UPDATE contexts SET title = :title, description = :description, color = :color, updated_at = :updated_at WHERE id = :id;`
	res, err := r.db.NamedExecContext(ctx, q, map[string]any{"id": contextObj.ID, "title": contextObj.Title, "description": contextObj.Description, "color": contextObj.Color, "updated_at": contextObj.UpdatedAt})
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

func (r *ContextRepository) Delete(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM contexts WHERE id = ?;`, id)
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
