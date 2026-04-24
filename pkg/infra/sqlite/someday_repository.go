package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"

	"github.com/anyvoxel/multivac/pkg/domain"
)

type SomedayRepository struct{ db *sqlx.DB }

func NewSomedayRepository(db *sqlx.DB) *SomedayRepository { return &SomedayRepository{db: db} }

func (r *SomedayRepository) Migrate(ctx context.Context) error {
	const ddl = `
		CREATE TABLE IF NOT EXISTS somedays (
		  id TEXT PRIMARY KEY,
		  title TEXT NOT NULL,
		  description TEXT NOT NULL,
		  created_at DATETIME NOT NULL,
		  updated_at DATETIME NOT NULL
		);
		CREATE INDEX IF NOT EXISTS idx_somedays_created_at ON somedays(created_at);
		CREATE INDEX IF NOT EXISTS idx_somedays_updated_at ON somedays(updated_at);
		`
	_, err := r.db.ExecContext(ctx, ddl)
	return err
}

type somedayRow struct {
	ID          string       `db:"id"`
	Title       string       `db:"title"`
	Description string       `db:"description"`
	CreatedAt   sql.NullTime `db:"created_at"`
	UpdatedAt   sql.NullTime `db:"updated_at"`
}

type somedayInboxRow struct {
	ID          string `db:"id"`
	Title       string `db:"title"`
	Description string `db:"description"`
}

func toSomedayDomain(row somedayRow) *domain.Someday {
	someday := &domain.Someday{ID: row.ID, Title: row.Title, Description: row.Description}
	if row.CreatedAt.Valid {
		someday.CreatedAt = row.CreatedAt.Time
	}
	if row.UpdatedAt.Valid {
		someday.UpdatedAt = row.UpdatedAt.Time
	}
	return someday
}

func applySomedayFilters(base string, q domain.SomedayListQuery, args []any) (string, []any) {
	base += " WHERE 1=1"
	if q.Search != "" {
		like := "%" + strings.ToLower(q.Search) + "%"
		base += " AND (LOWER(title) LIKE ? OR LOWER(description) LIKE ?)"
		args = append(args, like, like)
	}
	return base, args
}

func somedayOrderBy(q domain.SomedayListQuery) (string, error) {
	orderByMap := map[domain.SortBy]string{
		domain.SomedaySortByCreatedAt: "created_at",
		domain.SomedaySortByUpdatedAt: "updated_at",
		domain.SomedaySortByTitle:     "title",
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

func applySomedayLimitOffset(base string, q domain.SomedayListQuery, args []any) (string, []any) {
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

func (r *SomedayRepository) Create(ctx context.Context, someday *domain.Someday) error {
	if err := someday.Validate(); err != nil {
		return err
	}
	const q = `INSERT INTO somedays (id, title, description, created_at, updated_at) VALUES (:id, :title, :description, :created_at, :updated_at);`
	_, err := r.db.NamedExecContext(ctx, q, map[string]any{
		"id":          someday.ID,
		"title":       someday.Title,
		"description": someday.Description,
		"created_at":  someday.CreatedAt,
		"updated_at":  someday.UpdatedAt,
	})
	return err
}

func (r *SomedayRepository) Get(ctx context.Context, id string) (*domain.Someday, error) {
	const q = `SELECT id, title, description, created_at, updated_at FROM somedays WHERE id = ?;`
	var row somedayRow
	if err := r.db.GetContext(ctx, &row, q, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return toSomedayDomain(row), nil
}

func (r *SomedayRepository) List(ctx context.Context, q domain.SomedayListQuery) ([]*domain.Someday, error) {
	if err := (&q).Validate(); err != nil {
		return nil, err
	}
	base := `SELECT id, title, description, created_at, updated_at FROM somedays`
	args := []any{}
	base, args = applySomedayFilters(base, q, args)
	ob, err := somedayOrderBy(q)
	if err != nil {
		return nil, err
	}
	base += ob
	base, args = applySomedayLimitOffset(base, q, args)
	rows := make([]somedayRow, 0)
	if err := r.db.SelectContext(ctx, &rows, base, args...); err != nil {
		return nil, err
	}
	out := make([]*domain.Someday, 0, len(rows))
	for _, row := range rows {
		out = append(out, toSomedayDomain(row))
	}
	return out, nil
}

func (r *SomedayRepository) Update(ctx context.Context, someday *domain.Someday) error {
	if err := someday.Validate(); err != nil {
		return err
	}
	const q = `UPDATE somedays SET title = :title, description = :description, updated_at = :updated_at WHERE id = :id;`
	res, err := r.db.NamedExecContext(ctx, q, map[string]any{
		"id":          someday.ID,
		"title":       someday.Title,
		"description": someday.Description,
		"updated_at":  someday.UpdatedAt,
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

func (r *SomedayRepository) Delete(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM somedays WHERE id = ?;`, id)
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

func (r *SomedayRepository) ConvertFromInbox(ctx context.Context, inboxID string, title, description *string, now time.Time) (*domain.Someday, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	var inbox somedayInboxRow
	if err := tx.GetContext(ctx, &inbox, `SELECT id, title, description FROM inboxes WHERE id = ?;`, inboxID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	finalTitle := inbox.Title
	if title != nil {
		finalTitle = *title
	}
	finalDescription := inbox.Description
	if description != nil {
		finalDescription = *description
	}

	someday, err := domain.NewSomeday(inbox.ID, finalTitle, finalDescription, now.UTC())
	if err != nil {
		return nil, err
	}

	const insertSomeday = `INSERT INTO somedays (id, title, description, created_at, updated_at) VALUES (:id, :title, :description, :created_at, :updated_at);`
	if _, err := tx.NamedExecContext(ctx, insertSomeday, map[string]any{
		"id":          someday.ID,
		"title":       someday.Title,
		"description": someday.Description,
		"created_at":  someday.CreatedAt,
		"updated_at":  someday.UpdatedAt,
	}); err != nil {
		return nil, err
	}

	res, err := tx.ExecContext(ctx, `DELETE FROM inboxes WHERE id = ?;`, inbox.ID)
	if err != nil {
		return nil, err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return nil, err
	}
	if affected == 0 {
		return nil, domain.ErrNotFound
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return someday, nil
}
