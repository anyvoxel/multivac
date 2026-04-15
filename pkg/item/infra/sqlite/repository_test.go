package sqlite

import (
	"context"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/onsi/gomega"

	"github.com/anyvoxel/multivac/pkg/item/domain"
	projectsqlite "github.com/anyvoxel/multivac/pkg/project/infra/sqlite"
)

func TestRepositoryMigratesLegacyDataIdempotently(t *testing.T) {
	g := gomega.NewWithT(t)
	db, err := sqlx.Open("sqlite3", ":memory:")
	g.Expect(err).ToNot(gomega.HaveOccurred())
	defer func() { _ = db.Close() }()
	ctx := context.Background()
	projRepo := projectsqlite.NewRepository(db)
	g.Expect(projRepo.Migrate(ctx)).To(gomega.Succeed())
	now := time.Date(2026, 1, 8, 9, 0, 0, 0, time.UTC)
	_, err = db.ExecContext(ctx, `INSERT INTO projects (id, name, goal, principles, vision_result, description, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?);`, "p1", "Project", "g", "pr", "vr", "desc", "Draft", now, now)
	g.Expect(err).ToNot(gomega.HaveOccurred())
	_, err = db.ExecContext(ctx, `CREATE TABLE inboxes (id TEXT PRIMARY KEY, name TEXT NOT NULL, description TEXT NOT NULL, created_at DATETIME NOT NULL, updated_at DATETIME NOT NULL);`)
	g.Expect(err).ToNot(gomega.HaveOccurred())
	_, err = db.ExecContext(ctx, `CREATE TABLE somedays (id TEXT PRIMARY KEY, name TEXT NOT NULL, description TEXT NOT NULL, created_at DATETIME NOT NULL, updated_at DATETIME NOT NULL);`)
	g.Expect(err).ToNot(gomega.HaveOccurred())
	_, err = db.ExecContext(ctx, `CREATE TABLE waiting_lists (id TEXT PRIMARY KEY, name TEXT NOT NULL, details TEXT NOT NULL, owner TEXT NOT NULL, expected_at DATETIME NULL, created_at DATETIME NOT NULL, updated_at DATETIME NOT NULL);`)
	g.Expect(err).ToNot(gomega.HaveOccurred())
	_, err = db.ExecContext(ctx, `CREATE TABLE tasks (id TEXT PRIMARY KEY, project_id TEXT NULL, name TEXT NOT NULL, description TEXT NOT NULL, context TEXT NOT NULL, details TEXT NOT NULL, status TEXT NOT NULL, priority TEXT NOT NULL, due_at DATETIME NULL, created_at DATETIME NOT NULL, updated_at DATETIME NOT NULL);`)
	g.Expect(err).ToNot(gomega.HaveOccurred())
	expectedAt := now.Add(24 * time.Hour)
	dueAt := now.Add(48 * time.Hour)
	_, err = db.ExecContext(ctx, `INSERT INTO inboxes VALUES (?, ?, ?, ?, ?);`, "i1", "Inbox", "inbox-desc", now, now)
	g.Expect(err).ToNot(gomega.HaveOccurred())
	_, err = db.ExecContext(ctx, `INSERT INTO somedays VALUES (?, ?, ?, ?, ?);`, "s1", "Someday", "someday-desc", now, now)
	g.Expect(err).ToNot(gomega.HaveOccurred())
	_, err = db.ExecContext(ctx, `INSERT INTO waiting_lists VALUES (?, ?, ?, ?, ?, ?, ?);`, "w1", "Wait", "wait-details", "alice", expectedAt, now, now)
	g.Expect(err).ToNot(gomega.HaveOccurred())
	_, err = db.ExecContext(ctx, `INSERT INTO tasks VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);`, "t1", "p1", "Task", "task-desc", "ctx", "task-details", "Done", "High", dueAt, now, now)
	g.Expect(err).ToNot(gomega.HaveOccurred())
	repo := NewRepository(db)
	g.Expect(repo.Migrate(ctx)).To(gomega.Succeed())
	g.Expect(repo.Migrate(ctx)).To(gomega.Succeed())
	items, err := repo.List(ctx, domain.ListQuery{Sorts: []domain.Sort{{By: domain.ItemSortByTitle, Dir: domain.SortAsc}}})
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(items).To(gomega.HaveLen(4))
	taskItem, err := repo.Get(ctx, "t1")
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(taskItem.Kind).To(gomega.Equal(domain.KindTask))
	g.Expect(taskItem.Bucket).To(gomega.Equal(domain.BucketCompleted))
	g.Expect(taskItem.TaskStatus).To(gomega.Equal("Done"))
	g.Expect(taskItem.ProjectID).To(gomega.Equal("p1"))
	g.Expect(taskItem.DueAt).ToNot(gomega.BeNil())
	waitingItem, err := repo.Get(ctx, "w1")
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(waitingItem.Kind).To(gomega.Equal(domain.KindWaitingFor))
	g.Expect(waitingItem.Bucket).To(gomega.Equal(domain.BucketWaitingFor))
	g.Expect(waitingItem.WaitingFor).To(gomega.Equal("alice"))
	g.Expect(waitingItem.ExpectedAt).ToNot(gomega.BeNil())
}

func TestRepositoryCRUDAndFilters(t *testing.T) {
	g := gomega.NewWithT(t)
	db, err := sqlx.Open("sqlite3", ":memory:")
	g.Expect(err).ToNot(gomega.HaveOccurred())
	defer func() { _ = db.Close() }()
	ctx := context.Background()
	projRepo := projectsqlite.NewRepository(db)
	g.Expect(projRepo.Migrate(ctx)).To(gomega.Succeed())
	now := time.Date(2026, 1, 9, 9, 0, 0, 0, time.UTC)
	_, err = db.ExecContext(ctx, `INSERT INTO projects (id, name, goal, principles, vision_result, description, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?);`, "p1", "Project", "g", "pr", "vr", "desc", "Draft", now, now)
	g.Expect(err).ToNot(gomega.HaveOccurred())
	repo := NewRepository(db)
	g.Expect(repo.Migrate(ctx)).To(gomega.Succeed())
	dueSoon := now.Add(3 * time.Hour)
	dueLater := now.Add(24 * time.Hour)
	first := &domain.Item{ID: "it1", Kind: domain.KindTask, Bucket: domain.BucketNextAction, ProjectID: "p1", Title: "Alpha", Description: "desc", Context: "ctx", Details: "details", TaskStatus: "Todo", Priority: "High", CreatedAt: now, UpdatedAt: now, DueAt: &dueLater}
	second := &domain.Item{ID: "it2", Kind: domain.KindTask, Bucket: domain.BucketNextAction, Title: "Beta Search", Description: "desc", Context: "ctx", Details: "details", TaskStatus: "InProgress", Priority: "Low", CreatedAt: now.Add(time.Minute), UpdatedAt: now.Add(time.Minute), DueAt: &dueSoon}
	g.Expect(repo.Create(ctx, first)).To(gomega.Succeed())
	g.Expect(repo.Create(ctx, second)).To(gomega.Succeed())
	list, err := repo.List(ctx, domain.ListQuery{ProjectID: "p1"})
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(list).To(gomega.HaveLen(1))
	search, err := repo.List(ctx, domain.ListQuery{Search: "search"})
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(search).To(gomega.HaveLen(1))
	bucket := domain.BucketNextAction
	kind := domain.KindTask
	sorted, err := repo.List(ctx, domain.ListQuery{Bucket: &bucket, Kind: &kind, Sorts: []domain.Sort{{By: domain.ItemSortByDueAt, Dir: domain.SortAsc}}})
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(sorted[0].ID).To(gomega.Equal("it2"))
	loaded, err := repo.Get(ctx, "it1")
	g.Expect(err).ToNot(gomega.HaveOccurred())
	loaded.Title = "Alpha Updated"
	loaded.UpdatedAt = now.Add(2 * time.Hour)
	g.Expect(repo.Update(ctx, loaded)).To(gomega.Succeed())
	updated, err := repo.Get(ctx, "it1")
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(updated.Title).To(gomega.Equal("Alpha Updated"))
	g.Expect(repo.Delete(ctx, "it2")).To(gomega.Succeed())
}
