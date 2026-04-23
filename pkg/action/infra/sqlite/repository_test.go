package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/onsi/gomega"

	actiondomain "github.com/anyvoxel/multivac/pkg/action/domain"
	inboxsqlite "github.com/anyvoxel/multivac/pkg/inbox/infra/sqlite"
	projectdomain "github.com/anyvoxel/multivac/pkg/project/domain"
	projectsqlite "github.com/anyvoxel/multivac/pkg/project/infra/sqlite"
)

func TestRepositoryCRUDAndFilters(t *testing.T) {
	g := gomega.NewWithT(t)
	db, err := sqlx.Open("sqlite3", ":memory:")
	g.Expect(err).ToNot(gomega.HaveOccurred())
	defer func() { _ = db.Close() }()
	ctx := context.Background()

	projectRepo := projectsqlite.NewRepository(db)
	repo := NewRepository(db)
	g.Expect(projectRepo.Migrate(ctx)).To(gomega.Succeed())
	g.Expect(repo.Migrate(ctx)).To(gomega.Succeed())

	now := time.Date(2026, 1, 9, 9, 0, 0, 0, time.UTC)
	project, err := projectdomain.NewProject("p1", "Project", []projectdomain.Goal{{Title: "Goal"}}, "Desc", []projectdomain.Reference{{Title: "ref", URL: "https://example.com"}}, now)
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(projectRepo.Create(ctx, project)).To(gomega.Succeed())

	projectID := "p1"
	first, err := actiondomain.NewAction("01J0000000000000000000101", "Alpha", "plain text", &projectID, actiondomain.KindTask, []string{"ctx-1"}, []actiondomain.Label{{Name: "today"}}, actiondomain.Attributes{Task: &actiondomain.TaskAttributes{}}, now)
	g.Expect(err).ToNot(gomega.HaveOccurred())
	dueAt := now.Add(24 * time.Hour)
	second, err := actiondomain.NewAction("01J0000000000000000000102", "Beta", "contains search", nil, actiondomain.KindWaiting, nil, nil, actiondomain.Attributes{Waiting: &actiondomain.WaitingAttributes{Delegatee: "alice", DueAt: &dueAt}}, now.Add(time.Minute))
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(repo.Create(ctx, first)).To(gomega.Succeed())
	g.Expect(repo.Create(ctx, second)).To(gomega.Succeed())

	listed, err := repo.List(ctx, actiondomain.ListQuery{})
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(listed).To(gomega.HaveLen(2))
	g.Expect(listed[0].ID).To(gomega.Equal(second.ID))

	search, err := repo.List(ctx, actiondomain.ListQuery{Search: "search"})
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(search).To(gomega.HaveLen(1))
	g.Expect(search[0].ID).To(gomega.Equal(second.ID))

	kind := actiondomain.KindWaiting
	byKind, err := repo.List(ctx, actiondomain.ListQuery{Kind: &kind})
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(byKind).To(gomega.HaveLen(1))
	g.Expect(byKind[0].ID).To(gomega.Equal(second.ID))

	byProject, err := repo.List(ctx, actiondomain.ListQuery{ProjectID: &projectID})
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(byProject).To(gomega.HaveLen(1))
	g.Expect(byProject[0].ID).To(gomega.Equal(first.ID))

	loaded, err := repo.Get(ctx, first.ID)
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(loaded.Title).To(gomega.Equal("Alpha"))

	loaded.Title = "Alpha Updated"
	loaded.Description = "updated"
	loaded.UpdatedAt = now.Add(3 * time.Hour)
	g.Expect(repo.Update(ctx, loaded)).To(gomega.Succeed())

	updated, err := repo.Get(ctx, first.ID)
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(updated.Title).To(gomega.Equal("Alpha Updated"))
	g.Expect(updated.Description).To(gomega.Equal("updated"))

	g.Expect(repo.Delete(ctx, second.ID)).To(gomega.Succeed())
	_, err = repo.Get(ctx, second.ID)
	g.Expect(errors.Is(err, actiondomain.ErrNotFound)).To(gomega.BeTrue())
}

func TestRepositoryStoresJSONBColumns(t *testing.T) {
	g := gomega.NewWithT(t)
	db, err := sqlx.Open("sqlite3", ":memory:")
	g.Expect(err).ToNot(gomega.HaveOccurred())
	defer func() { _ = db.Close() }()
	ctx := context.Background()

	projectRepo := projectsqlite.NewRepository(db)
	repo := NewRepository(db)
	g.Expect(projectRepo.Migrate(ctx)).To(gomega.Succeed())
	g.Expect(repo.Migrate(ctx)).To(gomega.Succeed())

	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	action, err := actiondomain.NewAction("a1", "n", "d", nil, actiondomain.KindTask, []string{"ctx-1"}, []actiondomain.Label{{Name: "L"}}, actiondomain.Attributes{Task: &actiondomain.TaskAttributes{}}, now)
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(repo.Create(ctx, action)).To(gomega.Succeed())

	var contextType string
	err = db.GetContext(ctx, &contextType, `SELECT typeof("context") FROM actions WHERE id = ?`, "a1")
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(contextType).To(gomega.Equal("blob"))

	var labelsType string
	err = db.GetContext(ctx, &labelsType, `SELECT typeof(labels) FROM actions WHERE id = ?`, "a1")
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(labelsType).To(gomega.Equal("blob"))

	var attributesType string
	err = db.GetContext(ctx, &attributesType, `SELECT typeof(attributes) FROM actions WHERE id = ?`, "a1")
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(attributesType).To(gomega.Equal("blob"))

	var contextJSON sql.NullString
	err = db.GetContext(ctx, &contextJSON, `SELECT json("context") FROM actions WHERE id = ?`, "a1")
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(contextJSON.Valid).To(gomega.BeTrue())
	g.Expect(contextJSON.String).To(gomega.ContainSubstring("ctx-1"))
}

func TestRepositoryConvertFromInbox(t *testing.T) {
	g := gomega.NewWithT(t)
	db, err := sqlx.Open("sqlite3", ":memory:")
	g.Expect(err).ToNot(gomega.HaveOccurred())
	defer func() { _ = db.Close() }()
	ctx := context.Background()

	inboxRepo := inboxsqlite.NewRepository(db)
	projectRepo := projectsqlite.NewRepository(db)
	repo := NewRepository(db)
	g.Expect(inboxRepo.Migrate(ctx)).To(gomega.Succeed())
	g.Expect(projectRepo.Migrate(ctx)).To(gomega.Succeed())
	g.Expect(repo.Migrate(ctx)).To(gomega.Succeed())

	now := time.Date(2026, 1, 10, 12, 0, 0, 0, time.UTC)
	_, err = db.ExecContext(ctx, `INSERT INTO inboxes (id, title, description, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`, "01J0000000000000000000201", "Inbox title", "Inbox body", now, now)
	g.Expect(err).ToNot(gomega.HaveOccurred())

	overrideTitle := "Clarified action"
	action, err := repo.ConvertFromInbox(ctx, "01J0000000000000000000201", &overrideTitle, nil, nil, nil, nil, nil, nil, now.Add(time.Minute))
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(action.ID).To(gomega.Equal("01J0000000000000000000201"))
	g.Expect(action.Title).To(gomega.Equal("Clarified action"))
	g.Expect(action.Description).To(gomega.Equal("Inbox body"))

	var inboxCount int
	err = db.GetContext(ctx, &inboxCount, `SELECT COUNT(1) FROM inboxes WHERE id = ?`, "01J0000000000000000000201")
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(inboxCount).To(gomega.Equal(0))

	loaded, err := repo.Get(ctx, "01J0000000000000000000201")
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(loaded.Title).To(gomega.Equal("Clarified action"))
}
