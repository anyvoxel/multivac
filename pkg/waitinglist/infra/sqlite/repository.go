// Package sqlite provides sqlite implementation for waiting list repository.
package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"

	"github.com/anyvoxel/multivac/pkg/waitinglist/domain"
)

func applyWaitingListFilters(base string, q domain.ListQuery, args []any) (string, []any) {
	base += " WHERE 1=1"
	if q.Search != "" {
		like := "%" + strings.ToLower(q.Search) + "%"
		base += " AND (LOWER(name) LIKE ? OR LOWER(details) LIKE ? OR LOWER(owner) LIKE ?)"
		args = append(args, like, like, like)
	}
	return base, args
}

func waitingListOrderBy(q domain.ListQuery) (string, error) {
	orderParts := make([]string, 0, len(q.Sorts))
	for _, s := range q.Sorts {
		switch s.By {
		case domain.WaitingListSortByCreatedAt:
			dir := "ASC"
			if s.Dir == domain.SortDesc {
				dir = "DESC"
			}
			orderParts = append(orderParts, "created_at "+dir)
		case domain.WaitingListSortByUpdatedAt:
			dir := "ASC"
			if s.Dir == domain.SortDesc {
				dir = "DESC"
			}
			orderParts = append(orderParts, "updated_at "+dir)
		case domain.WaitingListSortByName:
			dir := "ASC"
			if s.Dir == domain.SortDesc {
				dir = "DESC"
			}
			orderParts = append(orderParts, "name "+dir)
		case domain.WaitingListSortByExpectedAt:
			dir := "ASC"
			if s.Dir == domain.SortDesc {
				dir = "DESC"
			}
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

// Repository implements domain.Repository using sqlite.
type Repository struct {
	db *sqlx.DB
}

// NewRepository creates a sqlite-backed waiting list repository.
func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

// Migrate creates required tables and indices.
func (r *Repository) Migrate(ctx context.Context) error {
	const ddl = `
	PRAGMA foreign_keys = ON;

	CREATE TABLE IF NOT EXISTS waiting_lists (
	  id TEXT PRIMARY KEY,
	  name TEXT NOT NULL,
	  details TEXT NOT NULL,
	  owner TEXT NOT NULL,
	  expected_at DATETIME NULL,
	  created_at DATETIME NOT NULL,
	  updated_at DATETIME NOT NULL
	);
	`
	_, err := r.db.ExecContext(ctx, ddl)
	return err
}

type waitingListRow struct {
	ID         string       `db:"id"`
	Name       string       `db:"name"`
	Details    string       `db:"details"`
	Owner      string       `db:"owner"`
	ExpectedAt sql.NullTime `db:"expected_at"`
	CreatedAt  sql.NullTime `db:"created_at"`
	UpdatedAt  sql.NullTime `db:"updated_at"`
}

func toDomain(row waitingListRow) *domain.WaitingList {
	item := &domain.WaitingList{
		ID:      row.ID,
		Name:    row.Name,
		Details: row.Details,
		Owner:   row.Owner,
	}
	if row.ExpectedAt.Valid {
		expectedAt := row.ExpectedAt.Time
		item.ExpectedAt = &expectedAt
	}
	if row.CreatedAt.Valid {
		item.CreatedAt = row.CreatedAt.Time
	}
	if row.UpdatedAt.Valid {
		item.UpdatedAt = row.UpdatedAt.Time
	}
	return item
}

// Create inserts a new waiting list item.
func (r *Repository) Create(ctx context.Context, item *domain.WaitingList) error {
	const q = `
	INSERT INTO waiting_lists (
	  id, name, details, owner, expected_at, created_at, updated_at
	) VALUES (
	  :id, :name, :details, :owner, :expected_at, :created_at, :updated_at
	);
	`
	params := map[string]any{
		"id":          item.ID,
		"name":        item.Name,
		"details":     item.Details,
		"owner":       item.Owner,
		"expected_at": timePtrOrNil(item.ExpectedAt),
		"created_at":  item.CreatedAt,
		"updated_at":  item.UpdatedAt,
	}
	_, err := r.db.NamedExecContext(ctx, q, params)
	return err
}

// Get loads a waiting list item by id.
func (r *Repository) Get(ctx context.Context, id string) (*domain.WaitingList, error) {
	const q = `
	SELECT id, name, details, owner, expected_at, created_at, updated_at
	FROM waiting_lists
	WHERE id = ?;
	`
	var row waitingListRow
	if err := r.db.GetContext(ctx, &row, q, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return toDomain(row), nil
}

// List returns waiting list items by query.
func (r *Repository) List(ctx context.Context, q domain.ListQuery) ([]*domain.WaitingList, error) {
	if err := (&q).Validate(); err != nil {
		return nil, err
	}

	base := `
	SELECT id, name, details, owner, expected_at, created_at, updated_at
	FROM waiting_lists
	`
	args := []any{}
	base, args = applyWaitingListFilters(base, q, args)
	orderBy, err := waitingListOrderBy(q)
	if err != nil {
		return nil, err
	}
	base += orderBy
	base, args = applyLimitOffset(base, q, args)

	rows := make([]waitingListRow, 0)
	if err := r.db.SelectContext(ctx, &rows, base, args...); err != nil {
		return nil, err
	}
	out := make([]*domain.WaitingList, 0, len(rows))
	for _, row := range rows {
		out = append(out, toDomain(row))
	}
	return out, nil
}

// Update updates a waiting list item by id.
func (r *Repository) Update(ctx context.Context, item *domain.WaitingList) error {
	const q = `
	UPDATE waiting_lists SET
	  name = :name,
	  details = :details,
	  owner = :owner,
	  expected_at = :expected_at,
	  updated_at = :updated_at
	WHERE id = :id;
	`
	params := map[string]any{
		"id":          item.ID,
		"name":        item.Name,
		"details":     item.Details,
		"owner":       item.Owner,
		"expected_at": timePtrOrNil(item.ExpectedAt),
		"updated_at":  item.UpdatedAt,
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

// Delete removes a waiting list item by id.
func (r *Repository) Delete(ctx context.Context, id string) error {
	const q = `DELETE FROM waiting_lists WHERE id = ?;`
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

func timePtrOrNil(v *time.Time) any {
	if v == nil {
		return nil
	}
	return *v
}
