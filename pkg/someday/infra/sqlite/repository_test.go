package sqlite

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/onsi/gomega"

	inboxsqlite "github.com/anyvoxel/multivac/pkg/inbox/infra/sqlite"
	"github.com/anyvoxel/multivac/pkg/someday/domain"
)

func TestRepositoryCRUDAndFilters(t *testing.T) {
	g := gomega.NewWithT(t)
	db, err := sqlx.Open("sqlite3", ":memory:")
	g.Expect(err).ToNot(gomega.HaveOccurred())
	defer func() { _ = db.Close() }()
	ctx := context.Background()
	repo := NewRepository(db)
	g.Expect(repo.Migrate(ctx)).To(gomega.Succeed())
	g.Expect(repo.Migrate(ctx)).To(gomega.Succeed())
	now := time.Date(2026, 1, 9, 9, 0, 0, 0, time.UTC)
	first := &domain.Someday{ID: "01J0000000000000000000101", Title: "Alpha", Description: "plain text", CreatedAt: now, UpdatedAt: now}
	second := &domain.Someday{ID: "01J0000000000000000000102", Title: "Beta", Description: "contains markdown **search**", CreatedAt: now.Add(time.Minute), UpdatedAt: now.Add(2 * time.Minute)}
	g.Expect(repo.Create(ctx, first)).To(gomega.Succeed())
	g.Expect(repo.Create(ctx, second)).To(gomega.Succeed())

	listed, err := repo.List(ctx, domain.ListQuery{})
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(listed).To(gomega.HaveLen(2))
	g.Expect(listed[0].ID).To(gomega.Equal(second.ID))

	search, err := repo.List(ctx, domain.ListQuery{Search: "search"})
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(search).To(gomega.HaveLen(1))
	g.Expect(search[0].ID).To(gomega.Equal(second.ID))

	sorted, err := repo.List(ctx, domain.ListQuery{Sorts: []domain.Sort{{By: domain.SomedaySortByTitle, Dir: domain.SortAsc}}, Limit: 1, Offset: 1})
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(sorted).To(gomega.HaveLen(1))
	g.Expect(sorted[0].ID).To(gomega.Equal(second.ID))

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
	g.Expect(errors.Is(err, domain.ErrNotFound)).To(gomega.BeTrue())
}

func TestRepositoryValidationAndNotFound(t *testing.T) {
	g := gomega.NewWithT(t)
	db, err := sqlx.Open("sqlite3", ":memory:")
	g.Expect(err).ToNot(gomega.HaveOccurred())
	defer func() { _ = db.Close() }()
	ctx := context.Background()
	repo := NewRepository(db)
	g.Expect(repo.Migrate(ctx)).To(gomega.Succeed())
	now := time.Date(2026, 1, 10, 9, 0, 0, 0, time.UTC)

	err = repo.Create(ctx, &domain.Someday{ID: "01J0000000000000000000103", Description: "missing title", CreatedAt: now, UpdatedAt: now})
	g.Expect(errors.Is(err, domain.ErrInvalidArg)).To(gomega.BeTrue())

	_, err = repo.List(ctx, domain.ListQuery{Sorts: []domain.Sort{{By: domain.SortBy("BadField"), Dir: domain.SortAsc}}})
	g.Expect(errors.Is(err, domain.ErrInvalidArg)).To(gomega.BeTrue())

	err = repo.Delete(ctx, "missing")
	g.Expect(errors.Is(err, domain.ErrNotFound)).To(gomega.BeTrue())
}

func TestRepositoryConvertFromInbox(t *testing.T) {
	g := gomega.NewWithT(t)
	db, err := sqlx.Open("sqlite3", ":memory:")
	g.Expect(err).ToNot(gomega.HaveOccurred())
	defer func() { _ = db.Close() }()
	ctx := context.Background()

	inboxRepo := inboxsqlite.NewRepository(db)
	repo := NewRepository(db)
	g.Expect(inboxRepo.Migrate(ctx)).To(gomega.Succeed())
	g.Expect(repo.Migrate(ctx)).To(gomega.Succeed())

	now := time.Date(2026, 1, 10, 12, 0, 0, 0, time.UTC)
	_, err = db.ExecContext(ctx, `INSERT INTO inboxes (id, title, description, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`, "01J0000000000000000000201", "Inbox title", "Inbox body", now, now)
	g.Expect(err).ToNot(gomega.HaveOccurred())

	overrideTitle := "Clarified someday"
	someday, err := repo.ConvertFromInbox(ctx, "01J0000000000000000000201", &overrideTitle, nil, now.Add(time.Minute))
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(someday.ID).To(gomega.Equal("01J0000000000000000000201"))
	g.Expect(someday.Title).To(gomega.Equal("Clarified someday"))
	g.Expect(someday.Description).To(gomega.Equal("Inbox body"))

	var inboxCount int
	err = db.GetContext(ctx, &inboxCount, `SELECT COUNT(1) FROM inboxes WHERE id = ?`, "01J0000000000000000000201")
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(inboxCount).To(gomega.Equal(0))

	loaded, err := repo.Get(ctx, "01J0000000000000000000201")
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(loaded.Title).To(gomega.Equal("Clarified someday"))
	g.Expect(loaded.Description).To(gomega.Equal("Inbox body"))
}
