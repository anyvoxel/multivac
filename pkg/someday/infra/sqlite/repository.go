// Package sqlite provides sqlite implementation for someday repository.
package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"

	"github.com/anyvoxel/multivac/pkg/someday/domain"
)

func applySomedayFilters(base string, q domain.ListQuery, args []any) (string, []any) {
	base += " WHERE 1=1"
	if q.Search != "" {
		like := "%" + strings.ToLower(q.Search) + "%"
		base += " AND (LOWER(name) LIKE ? OR LOWER(description) LIKE ?)"
		args = append(args, like, like)
	}
	return base, args
}

func somedayOrderBy(q domain.ListQuery) (string, error) {
	orderByMap := map[domain.SortBy]string{
		domain.SomedaySortByCreatedAt: "created_at",
		domain.SomedaySortByUpdatedAt: "updated_at",
		domain.SomedaySortByName:      "name",
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

// NewRepository creates a sqlite-backed someday repository.
func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

// Migrate creates required tables and indices.
func (r *Repository) Migrate(ctx context.Context) error {
	const ddl = `
	PRAGMA foreign_keys = ON;

	CREATE TABLE IF NOT EXISTS somedays (
	  id TEXT PRIMARY KEY,
	  name TEXT NOT NULL,
	  description TEXT NOT NULL,
	  created_at DATETIME NOT NULL,
	  updated_at DATETIME NOT NULL
	);
	`
	_, err := r.db.ExecContext(ctx, ddl)
	return err
}

type somedayRow struct {
	ID          string       `db:"id"`
	Name        string       `db:"name"`
	Description string       `db:"description"`
	CreatedAt   sql.NullTime `db:"created_at"`
	UpdatedAt   sql.NullTime `db:"updated_at"`
}

func toDomain(row somedayRow) *domain.Someday {
	someday := &domain.Someday{
		ID:          row.ID,
		Name:        row.Name,
		Description: row.Description,
	}
	if row.CreatedAt.Valid {
		someday.CreatedAt = row.CreatedAt.Time
	}
	if row.UpdatedAt.Valid {
		someday.UpdatedAt = row.UpdatedAt.Time
	}
	return someday
}

// Create inserts a new someday item.
func (r *Repository) Create(ctx context.Context, someday *domain.Someday) error {
	const q = `
	INSERT INTO somedays (
	  id, name, description, created_at, updated_at
	) VALUES (
	  :id, :name, :description, :created_at, :updated_at
	);
	`
	params := map[string]any{
		"id":          someday.ID,
		"name":        someday.Name,
		"description": someday.Description,
		"created_at":  someday.CreatedAt,
		"updated_at":  someday.UpdatedAt,
	}
	_, err := r.db.NamedExecContext(ctx, q, params)
	return err
}

// Get loads a someday item by id.
func (r *Repository) Get(ctx context.Context, id string) (*domain.Someday, error) {
	const q = `
	SELECT id, name, description, created_at, updated_at
	FROM somedays
	WHERE id = ?;
	`
	var row somedayRow
	if err := r.db.GetContext(ctx, &row, q, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return toDomain(row), nil
}

// List returns someday items by query.
func (r *Repository) List(ctx context.Context, q domain.ListQuery) ([]*domain.Someday, error) {
	if err := (&q).Validate(); err != nil {
		return nil, err
	}

	base := `
	SELECT id, name, description, created_at, updated_at
	FROM somedays
	`
	args := []any{}
	base, args = applySomedayFilters(base, q, args)
	orderBy, err := somedayOrderBy(q)
	if err != nil {
		return nil, err
	}
	base += orderBy
	base, args = applyLimitOffset(base, q, args)

	rows := make([]somedayRow, 0)
	if err := r.db.SelectContext(ctx, &rows, base, args...); err != nil {
		return nil, err
	}
	out := make([]*domain.Someday, 0, len(rows))
	for _, row := range rows {
		out = append(out, toDomain(row))
	}
	return out, nil
}

// Update updates a someday item by id.
func (r *Repository) Update(ctx context.Context, someday *domain.Someday) error {
	const q = `
	UPDATE somedays SET
	  name = :name,
	  description = :description,
	  updated_at = :updated_at
	WHERE id = :id;
	`
	params := map[string]any{
		"id":          someday.ID,
		"name":        someday.Name,
		"description": someday.Description,
		"updated_at":  someday.UpdatedAt,
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

// Delete removes a someday item by id.
func (r *Repository) Delete(ctx context.Context, id string) error {
	const q = `DELETE FROM somedays WHERE id = ?;`
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
