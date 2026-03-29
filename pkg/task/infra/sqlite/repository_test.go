package sqlite

import (
	"context"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/onsi/gomega"

	projectsqlite "github.com/anyvoxel/multivac/pkg/project/infra/sqlite"
	"github.com/anyvoxel/multivac/pkg/task/domain"
)

func TestRepositoryCRUD(t *testing.T) {
	g := gomega.NewWithT(t)

	db, err := sqlx.Open("sqlite3", ":memory:")
	g.Expect(err).ToNot(gomega.HaveOccurred())
	defer func() { _ = db.Close() }()

	// Prepare parent table.
	projRepo := projectsqlite.NewRepository(db)
	ctx := context.Background()
	g.Expect(projRepo.Migrate(ctx)).To(gomega.Succeed())

	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	insertProject := `
INSERT INTO projects (id, name, goal, principles, vision_result, description, status, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?);
`
	_, err = db.ExecContext(ctx, insertProject, "p1", "pname", "g", "pr", "vr", "d", "Draft", now, now)
	g.Expect(err).ToNot(gomega.HaveOccurred())

	repo := NewRepository(db)
	g.Expect(repo.Migrate(ctx)).To(gomega.Succeed())

	due := now.Add(48 * time.Hour)
	tk, err := domain.NewTask("t1", "p1", "n", "desc", "ctx", "details", domain.PriorityMedium, &due, now)
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(repo.Create(ctx, tk)).To(gomega.Succeed())

	got, err := repo.Get(ctx, "t1")
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(got.ProjectID).To(gomega.Equal("p1"))

	list, err := repo.List(ctx, domain.ListQuery{ProjectID: "p1"})
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(list).To(gomega.HaveLen(1))

	now2 := now.Add(time.Hour)
	g.Expect(tk.SetStatus(domain.StatusDone, now2)).To(gomega.Succeed())
	g.Expect(repo.Update(ctx, tk)).To(gomega.Succeed())

	st := domain.StatusDone
	list2, err := repo.List(ctx, domain.ListQuery{ProjectID: "p1", Status: &st})
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(list2).To(gomega.HaveLen(1))

	g.Expect(repo.Delete(ctx, "t1")).To(gomega.Succeed())
	_, err = repo.Get(ctx, "t1")
	g.Expect(err).To(gomega.Equal(domain.ErrNotFound))
}
