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
	g.Expect(tk.UpdateDetails("", tk.Name, tk.Description, tk.Context, tk.Details, tk.Priority, tk.DueAt, now2)).To(gomega.Succeed())
	g.Expect(tk.SetStatus(domain.StatusDone, now2)).To(gomega.Succeed())
	g.Expect(repo.Update(ctx, tk)).To(gomega.Succeed())

	updated, err := repo.Get(ctx, "t1")
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(updated.ProjectID).To(gomega.Equal(""))

	st := domain.StatusDone
	list2, err := repo.List(ctx, domain.ListQuery{ProjectID: "p1", Status: &st})
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(list2).To(gomega.HaveLen(0))

	list3, err := repo.List(ctx, domain.ListQuery{Status: &st})
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(list3).To(gomega.HaveLen(1))
	g.Expect(list3[0].ProjectID).To(gomega.Equal(""))

	noProject, err := domain.NewTask("t2", "", "n2", "desc2", "ctx", "details2", domain.PriorityLow, nil, now)
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(repo.Create(ctx, noProject)).To(gomega.Succeed())

	got2, err := repo.Get(ctx, "t2")
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(got2.ProjectID).To(gomega.Equal(""))

	g.Expect(repo.Delete(ctx, "t1")).To(gomega.Succeed())
	_, err = repo.Get(ctx, "t1")
	g.Expect(err).To(gomega.Equal(domain.ErrNotFound))
}

func TestRepositoryListSupportsSearchPaginationAndSort(t *testing.T) {
	g := gomega.NewWithT(t)

	db, err := sqlx.Open("sqlite3", ":memory:")
	g.Expect(err).ToNot(gomega.HaveOccurred())
	defer func() { _ = db.Close() }()

	projRepo := projectsqlite.NewRepository(db)
	ctx := context.Background()
	g.Expect(projRepo.Migrate(ctx)).To(gomega.Succeed())

	now := time.Date(2026, 1, 3, 0, 0, 0, 0, time.UTC)
	insertProject := `
INSERT INTO projects (id, name, goal, principles, vision_result, description, status, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?);
`
	_, err = db.ExecContext(ctx, insertProject, "p1", "Project One", "g", "pr", "vr", "d", "Draft", now, now)
	g.Expect(err).ToNot(gomega.HaveOccurred())
	_, err = db.ExecContext(ctx, insertProject, "p2", "Project Two", "g", "pr", "vr", "d", "Draft", now, now)
	g.Expect(err).ToNot(gomega.HaveOccurred())

	repo := NewRepository(db)
	g.Expect(repo.Migrate(ctx)).To(gomega.Succeed())

	t1, err := domain.NewTask("t1", "p1", "Alpha Search", "desc", "ctx", "details", domain.PriorityHigh, nil, now)
	g.Expect(err).ToNot(gomega.HaveOccurred())
	t2, err := domain.NewTask("t2", "p1", "Beta", "desc", "Search Context", "details", domain.PriorityLow, nil, now.Add(time.Minute))
	g.Expect(err).ToNot(gomega.HaveOccurred())
	t3, err := domain.NewTask("t3", "p2", "Gamma", "desc", "ctx", "Search Details", domain.PriorityMedium, nil, now.Add(2*time.Minute))
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(t3.SetStatus(domain.StatusDone, now.Add(3*time.Minute))).To(gomega.Succeed())

	g.Expect(repo.Create(ctx, t1)).To(gomega.Succeed())
	g.Expect(repo.Create(ctx, t2)).To(gomega.Succeed())
	g.Expect(repo.Create(ctx, t3)).To(gomega.Succeed())

	list, err := repo.List(ctx, domain.ListQuery{Search: "search"})
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(list).To(gomega.HaveLen(3))

	paged, err := repo.List(ctx, domain.ListQuery{
		Search: "search",
		Sorts:  []domain.Sort{{By: domain.TaskSortByPriority, Dir: domain.SortAsc}},
		Limit:  1,
		Offset: 1,
	})
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(paged).To(gomega.HaveLen(1))
	g.Expect(paged[0].ID).To(gomega.Equal("t3"))

	done := domain.StatusDone
	filtered, err := repo.List(ctx, domain.ListQuery{
		ProjectID: "p2",
		Status:    &done,
		Search:    "details",
	})
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(filtered).To(gomega.HaveLen(1))
	g.Expect(filtered[0].ID).To(gomega.Equal("t3"))

	caseInsensitive, err := repo.List(ctx, domain.ListQuery{Search: "ALPHA"})
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(caseInsensitive).To(gomega.HaveLen(1))
	g.Expect(caseInsensitive[0].ID).To(gomega.Equal("t1"))

	dueSoon := now.Add(4 * time.Hour)
	dueLater := now.Add(24 * time.Hour)
	t4, err := domain.NewTask("t4", "p1", "Soon", "desc", "ctx", "details", domain.PriorityLow, &dueSoon, now.Add(4*time.Minute))
	g.Expect(err).ToNot(gomega.HaveOccurred())
	t5, err := domain.NewTask("t5", "p1", "Later", "desc", "ctx", "details", domain.PriorityLow, &dueLater, now.Add(5*time.Minute))
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(repo.Create(ctx, t4)).To(gomega.Succeed())
	g.Expect(repo.Create(ctx, t5)).To(gomega.Succeed())

	dueAsc, err := repo.List(ctx, domain.ListQuery{Sorts: []domain.Sort{{By: domain.TaskSortByDueAt, Dir: domain.SortAsc}}})
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(dueAsc[0].ID).To(gomega.Equal("t4"))
	g.Expect(dueAsc[1].ID).To(gomega.Equal("t5"))
	g.Expect(dueAsc[len(dueAsc)-1].DueAt).To(gomega.BeNil())

	dueDesc, err := repo.List(ctx, domain.ListQuery{Sorts: []domain.Sort{{By: domain.TaskSortByDueAt, Dir: domain.SortDesc}}})
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(dueDesc[0].ID).To(gomega.Equal("t5"))
	g.Expect(dueDesc[1].ID).To(gomega.Equal("t4"))
	g.Expect(dueDesc[len(dueDesc)-1].DueAt).To(gomega.BeNil())
}
