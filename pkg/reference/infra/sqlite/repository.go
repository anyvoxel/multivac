package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"

	"github.com/anyvoxel/multivac/pkg/reference/domain"
)

type Repository struct{ db *sqlx.DB }

func NewRepository(db *sqlx.DB) *Repository { return &Repository{db: db} }

func (r *Repository) Migrate(ctx context.Context) error {
	const ddl = `
		PRAGMA foreign_keys = ON;

		CREATE TABLE IF NOT EXISTS "references" (
		  id TEXT PRIMARY KEY,
		  title TEXT NOT NULL,
		  description TEXT NOT NULL,
		  "references" JSONB NOT NULL DEFAULT (jsonb('[]')),
		  created_at DATETIME NOT NULL,
		  updated_at DATETIME NOT NULL
		);
		CREATE INDEX IF NOT EXISTS idx_references_created_at ON "references"(created_at);
		CREATE INDEX IF NOT EXISTS idx_references_updated_at ON "references"(updated_at);
		`
	_, err := r.db.ExecContext(ctx, ddl)
	return err
}

type referenceRow struct {
	ID          string       `db:"id"`
	Title       string       `db:"title"`
	Description string       `db:"description"`
	References  []byte       `db:"reference_links"`
	CreatedAt   sql.NullTime `db:"created_at"`
	UpdatedAt   sql.NullTime `db:"updated_at"`
}

func toDomain(row referenceRow) (*domain.Reference, error) {
	references, err := unmarshalReferences(row.References)
	if err != nil {
		return nil, err
	}
	reference := &domain.Reference{
		ID:          row.ID,
		Title:       row.Title,
		Description: row.Description,
		References:  references,
	}
	if row.CreatedAt.Valid {
		reference.CreatedAt = row.CreatedAt.Time
	}
	if row.UpdatedAt.Valid {
		reference.UpdatedAt = row.UpdatedAt.Time
	}
	return reference, nil
}

func marshalReferences(references []domain.ReferenceLink) (string, error) {
	if len(references) == 0 {
		return "[]", nil
	}
	data, err := json.Marshal(references)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func unmarshalReferences(raw []byte) ([]domain.ReferenceLink, error) {
	raw = []byte(strings.TrimSpace(string(raw)))
	if len(raw) == 0 {
		return nil, nil
	}
	var references []domain.ReferenceLink
	if err := json.Unmarshal(raw, &references); err != nil {
		return nil, err
	}
	return references, nil
}

func applyFilters(base string, q domain.ListQuery, args []any) (string, []any) {
	base += " WHERE 1=1"
	if q.Search != "" {
		like := "%" + strings.ToLower(q.Search) + "%"
		base += " AND (LOWER(title) LIKE ? OR LOWER(description) LIKE ?)"
		args = append(args, like, like)
	}
	return base, args
}

func orderBy(q domain.ListQuery) (string, error) {
	orderByMap := map[domain.SortBy]string{
		domain.ReferenceSortByCreatedAt: "created_at",
		domain.ReferenceSortByUpdatedAt: "updated_at",
		domain.ReferenceSortByTitle:     "title",
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

func (r *Repository) Create(ctx context.Context, reference *domain.Reference) error {
	if err := reference.Validate(); err != nil {
		return err
	}
	references, err := marshalReferences(reference.References)
	if err != nil {
		return err
	}
	const q = `
		INSERT INTO "references" (id, title, description, "references", created_at, updated_at)
		VALUES (:id, :title, :description, jsonb(:references), :created_at, :updated_at);
	`
	_, err = r.db.NamedExecContext(ctx, q, map[string]any{
		"id":          reference.ID,
		"title":       reference.Title,
		"description": reference.Description,
		"references":  references,
		"created_at":  reference.CreatedAt,
		"updated_at":  reference.UpdatedAt,
	})
	return err
}

func (r *Repository) Get(ctx context.Context, id string) (*domain.Reference, error) {
	const q = `
		SELECT id, title, description, json("references") AS reference_links, created_at, updated_at
		FROM "references"
		WHERE id = ?;
	`
	var row referenceRow
	if err := r.db.GetContext(ctx, &row, q, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return toDomain(row)
}

func (r *Repository) List(ctx context.Context, q domain.ListQuery) ([]*domain.Reference, error) {
	if err := (&q).Validate(); err != nil {
		return nil, err
	}
	base := `
		SELECT id, title, description, json("references") AS reference_links, created_at, updated_at
		FROM "references"
	`
	args := []any{}
	base, args = applyFilters(base, q, args)
	ob, err := orderBy(q)
	if err != nil {
		return nil, err
	}
	base += ob
	base, args = applyLimitOffset(base, q, args)
	rows := make([]referenceRow, 0)
	if err := r.db.SelectContext(ctx, &rows, base, args...); err != nil {
		return nil, err
	}
	out := make([]*domain.Reference, 0, len(rows))
	for _, row := range rows {
		reference, err := toDomain(row)
		if err != nil {
			return nil, err
		}
		out = append(out, reference)
	}
	return out, nil
}

func (r *Repository) Update(ctx context.Context, reference *domain.Reference) error {
	if err := reference.Validate(); err != nil {
		return err
	}
	references, err := marshalReferences(reference.References)
	if err != nil {
		return err
	}
	const q = `
		UPDATE "references"
		SET title = :title, description = :description, "references" = jsonb(:references), updated_at = :updated_at
		WHERE id = :id;
	`
	res, err := r.db.NamedExecContext(ctx, q, map[string]any{
		"id":          reference.ID,
		"title":       reference.Title,
		"description": reference.Description,
		"references":  references,
		"updated_at":  reference.UpdatedAt,
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

func (r *Repository) Delete(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM "references" WHERE id = ?;`, id)
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
