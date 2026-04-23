package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"

	"github.com/anyvoxel/multivac/pkg/domain"
)

type ActionRepository struct{ db *sqlx.DB }

func NewActionRepository(db *sqlx.DB) *ActionRepository { return &ActionRepository{db: db} }

func (r *ActionRepository) Migrate(ctx context.Context) error {
	const ddl = `
		PRAGMA foreign_keys = ON;

		CREATE TABLE IF NOT EXISTS actions (
		  id TEXT PRIMARY KEY,
		  title TEXT NOT NULL,
		  description TEXT NOT NULL,
		  project_id TEXT NULL REFERENCES projects(id) ON DELETE SET NULL,
		  kind TEXT NOT NULL,
		  "context" JSONB NOT NULL DEFAULT (jsonb('[]')),
		  labels JSONB NOT NULL DEFAULT (jsonb('[]')),
		  attributes JSONB NOT NULL DEFAULT (jsonb('{}')),
		  created_at DATETIME NOT NULL,
		  updated_at DATETIME NOT NULL
		);
		CREATE INDEX IF NOT EXISTS idx_actions_project_id ON actions(project_id);
		CREATE INDEX IF NOT EXISTS idx_actions_kind ON actions(kind);
		CREATE INDEX IF NOT EXISTS idx_actions_created_at ON actions(created_at);
		CREATE INDEX IF NOT EXISTS idx_actions_updated_at ON actions(updated_at);
	`
	_, err := r.db.ExecContext(ctx, ddl)
	return err
}

type actionRow struct {
	ID          string         `db:"id"`
	Title       string         `db:"title"`
	Description string         `db:"description"`
	ProjectID   sql.NullString `db:"project_id"`
	Kind        string         `db:"kind"`
	Context     []byte         `db:"context_ids"`
	Labels      []byte         `db:"labels"`
	Attributes  []byte         `db:"attributes"`
	CreatedAt   sql.NullTime   `db:"created_at"`
	UpdatedAt   sql.NullTime   `db:"updated_at"`
}

type actionInboxRow struct {
	ID          string `db:"id"`
	Title       string `db:"title"`
	Description string `db:"description"`
}

func toActionDomain(row actionRow) (*domain.Action, error) {
	kind, ok := domain.ParseKind(row.Kind)
	if !ok {
		return nil, domain.ErrInvalidArg
	}
	contextIDs, err := unmarshalContextIDs(row.Context)
	if err != nil {
		return nil, err
	}
	labels, err := unmarshalLabels(row.Labels)
	if err != nil {
		return nil, err
	}
	attributes, err := unmarshalAttributes(row.Attributes)
	if err != nil {
		return nil, err
	}
	action := &domain.Action{
		ID:          row.ID,
		Title:       row.Title,
		Description: row.Description,
		Kind:        kind,
		ContextIDs:  contextIDs,
		Labels:      labels,
		Attributes:  attributes,
	}
	if row.ProjectID.Valid {
		v := row.ProjectID.String
		action.ProjectID = &v
	}
	if row.CreatedAt.Valid {
		action.CreatedAt = row.CreatedAt.Time
	}
	if row.UpdatedAt.Valid {
		action.UpdatedAt = row.UpdatedAt.Time
	}
	return action, nil
}

func marshalContextIDs(contextIDs []string) (string, error) {
	if len(contextIDs) == 0 {
		return "[]", nil
	}
	data, err := json.Marshal(contextIDs)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func unmarshalContextIDs(raw []byte) ([]string, error) {
	raw = []byte(strings.TrimSpace(string(raw)))
	if len(raw) == 0 {
		return []string{}, nil
	}
	var contextIDs []string
	if err := json.Unmarshal(raw, &contextIDs); err != nil {
		return nil, err
	}
	return contextIDs, nil
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

func unmarshalLabels(raw []byte) ([]domain.Label, error) {
	raw = []byte(strings.TrimSpace(string(raw)))
	if len(raw) == 0 {
		return []domain.Label{}, nil
	}
	var labels []domain.Label
	if err := json.Unmarshal(raw, &labels); err != nil {
		return nil, err
	}
	return labels, nil
}

func marshalAttributes(attributes domain.Attributes) (string, error) {
	data, err := json.Marshal(attributes)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func unmarshalAttributes(raw []byte) (domain.Attributes, error) {
	raw = []byte(strings.TrimSpace(string(raw)))
	if len(raw) == 0 {
		return domain.Attributes{}, nil
	}
	var attributes domain.Attributes
	if err := json.Unmarshal(raw, &attributes); err != nil {
		return domain.Attributes{}, err
	}
	return attributes, nil
}

func applyActionFilters(base string, q domain.ActionListQuery, args []any) (string, []any) {
	base += " WHERE 1=1"
	if q.Search != "" {
		like := "%" + strings.ToLower(q.Search) + "%"
		base += " AND (LOWER(title) LIKE ? OR LOWER(description) LIKE ?)"
		args = append(args, like, like)
	}
	if q.Kind != nil {
		base += " AND kind = ?"
		args = append(args, string(*q.Kind))
	}
	if q.ProjectID != nil {
		base += " AND project_id = ?"
		args = append(args, strings.TrimSpace(*q.ProjectID))
	}
	return base, args
}

func actionOrderBy(q domain.ActionListQuery) (string, error) {
	orderByMap := map[domain.SortBy]string{
		domain.ActionSortByCreatedAt: "created_at",
		domain.ActionSortByUpdatedAt: "updated_at",
		domain.ActionSortByTitle:     "title",
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

func applyActionLimitOffset(base string, q domain.ActionListQuery, args []any) (string, []any) {
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

func (r *ActionRepository) Create(ctx context.Context, action *domain.Action) error {
	if err := action.Validate(); err != nil {
		return err
	}
	contextIDs, err := marshalContextIDs(action.ContextIDs)
	if err != nil {
		return err
	}
	labels, err := marshalLabels(action.Labels)
	if err != nil {
		return err
	}
	attributes, err := marshalAttributes(action.Attributes)
	if err != nil {
		return err
	}
	const q = `
		INSERT INTO actions (id, title, description, project_id, kind, "context", labels, attributes, created_at, updated_at)
		VALUES (:id, :title, :description, :project_id, :kind, jsonb(:context), jsonb(:labels), jsonb(:attributes), :created_at, :updated_at);
	`
	_, err = r.db.NamedExecContext(ctx, q, map[string]any{
		"id":          action.ID,
		"title":       action.Title,
		"description": action.Description,
		"project_id":  action.ProjectID,
		"kind":        string(action.Kind),
		"context":     contextIDs,
		"labels":      labels,
		"attributes":  attributes,
		"created_at":  action.CreatedAt,
		"updated_at":  action.UpdatedAt,
	})
	return err
}

func (r *ActionRepository) Get(ctx context.Context, id string) (*domain.Action, error) {
	const q = `
		SELECT id, title, description, project_id, kind, json("context") AS context_ids, json(labels) AS labels, json(attributes) AS attributes, created_at, updated_at
		FROM actions
		WHERE id = ?;
	`
	var row actionRow
	if err := r.db.GetContext(ctx, &row, q, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return toActionDomain(row)
}

func (r *ActionRepository) List(ctx context.Context, q domain.ActionListQuery) ([]*domain.Action, error) {
	if err := (&q).Validate(); err != nil {
		return nil, err
	}
	base := `
		SELECT id, title, description, project_id, kind, json("context") AS context_ids, json(labels) AS labels, json(attributes) AS attributes, created_at, updated_at
		FROM actions
	`
	args := []any{}
	base, args = applyActionFilters(base, q, args)
	ob, err := actionOrderBy(q)
	if err != nil {
		return nil, err
	}
	base += ob
	base, args = applyActionLimitOffset(base, q, args)
	rows := make([]actionRow, 0)
	if err := r.db.SelectContext(ctx, &rows, base, args...); err != nil {
		return nil, err
	}
	out := make([]*domain.Action, 0, len(rows))
	for _, row := range rows {
		action, err := toActionDomain(row)
		if err != nil {
			return nil, err
		}
		out = append(out, action)
	}
	return out, nil
}

func (r *ActionRepository) Update(ctx context.Context, action *domain.Action) error {
	if err := action.Validate(); err != nil {
		return err
	}
	contextIDs, err := marshalContextIDs(action.ContextIDs)
	if err != nil {
		return err
	}
	labels, err := marshalLabels(action.Labels)
	if err != nil {
		return err
	}
	attributes, err := marshalAttributes(action.Attributes)
	if err != nil {
		return err
	}
	const q = `
		UPDATE actions
		SET title = :title, description = :description, project_id = :project_id, kind = :kind, "context" = jsonb(:context), labels = jsonb(:labels), attributes = jsonb(:attributes), updated_at = :updated_at
		WHERE id = :id;
	`
	res, err := r.db.NamedExecContext(ctx, q, map[string]any{
		"id":          action.ID,
		"title":       action.Title,
		"description": action.Description,
		"project_id":  action.ProjectID,
		"kind":        string(action.Kind),
		"context":     contextIDs,
		"labels":      labels,
		"attributes":  attributes,
		"updated_at":  action.UpdatedAt,
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

func (r *ActionRepository) Delete(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM actions WHERE id = ?;`, id)
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

func (r *ActionRepository) ConvertFromInbox(ctx context.Context, inboxID string, title, description *string, kind *domain.Kind, projectID *string, contextIDs []string, labels []domain.Label, attributes *domain.Attributes, now time.Time) (*domain.Action, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	var inbox actionInboxRow
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
	finalKind := domain.KindTask
	if kind != nil {
		finalKind = *kind
	}
	finalAttributes := domain.Attributes{Task: &domain.TaskAttributes{}}
	if attributes != nil {
		finalAttributes = *attributes
	}
	action, err := domain.NewAction(inbox.ID, finalTitle, finalDescription, projectID, finalKind, contextIDs, labels, finalAttributes, now.UTC())
	if err != nil {
		return nil, err
	}

	contextJSON, err := marshalContextIDs(action.ContextIDs)
	if err != nil {
		return nil, err
	}
	labelsJSON, err := marshalLabels(action.Labels)
	if err != nil {
		return nil, err
	}
	attributesJSON, err := marshalAttributes(action.Attributes)
	if err != nil {
		return nil, err
	}

	const insertAction = `
		INSERT INTO actions (id, title, description, project_id, kind, "context", labels, attributes, created_at, updated_at)
		VALUES (:id, :title, :description, :project_id, :kind, jsonb(:context), jsonb(:labels), jsonb(:attributes), :created_at, :updated_at);
	`
	if _, err := tx.NamedExecContext(ctx, insertAction, map[string]any{
		"id":          action.ID,
		"title":       action.Title,
		"description": action.Description,
		"project_id":  action.ProjectID,
		"kind":        string(action.Kind),
		"context":     contextJSON,
		"labels":      labelsJSON,
		"attributes":  attributesJSON,
		"created_at":  action.CreatedAt,
		"updated_at":  action.UpdatedAt,
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
	return action, nil
}
