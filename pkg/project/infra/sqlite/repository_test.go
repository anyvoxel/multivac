package sqlite

import (
	"context"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/onsi/gomega"

	"github.com/anyvoxel/multivac/pkg/project/domain"
)

func TestRepositoryCRUD(t *testing.T) {
	g := gomega.NewWithT(t)

	db, err := sqlx.Open("sqlite3", ":memory:")
	g.Expect(err).ToNot(gomega.HaveOccurred())
	defer func() { _ = db.Close() }()

	repo := NewRepository(db)
	ctx := context.Background()
	g.Expect(repo.Migrate(ctx)).To(gomega.Succeed())

	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	p, err := domain.NewProject("p1", "n", "g", "pr", "vr", "d", nil, now)
	g.Expect(err).ToNot(gomega.HaveOccurred())

	g.Expect(repo.Create(ctx, p)).To(gomega.Succeed())

	got, err := repo.Get(ctx, "p1")
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(got.ID).To(gomega.Equal("p1"))
	g.Expect(got.Name).To(gomega.Equal("n"))

	status := domain.StatusDraft
	list, err := repo.List(ctx, domain.ListQuery{Status: &status})
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(list).To(gomega.HaveLen(1))

	now2 := now.Add(time.Hour)
	g.Expect(p.UpdateDetails("n2", "g2", "pr2", "vr2", "d2", nil, now2)).To(gomega.Succeed())
	g.Expect(p.SetStatus(domain.StatusActive, now2)).To(gomega.Succeed())
	g.Expect(repo.Update(ctx, p)).To(gomega.Succeed())

	got2, err := repo.Get(ctx, "p1")
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(got2.Name).To(gomega.Equal("n2"))
	g.Expect(got2.Status).To(gomega.Equal(domain.StatusActive))
	g.Expect(got2.StartedAt).ToNot(gomega.BeNil())

	g.Expect(repo.Delete(ctx, "p1")).To(gomega.Succeed())
	_, err = repo.Get(ctx, "p1")
	g.Expect(err).To(gomega.Equal(domain.ErrNotFound))
}

func TestRepositoryListSupportsSearchPaginationAndSort(t *testing.T) {
	g := gomega.NewWithT(t)

	db, err := sqlx.Open("sqlite3", ":memory:")
	g.Expect(err).ToNot(gomega.HaveOccurred())
	defer func() { _ = db.Close() }()

	repo := NewRepository(db)
	ctx := context.Background()
	g.Expect(repo.Migrate(ctx)).To(gomega.Succeed())

	now := time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)
	alpha, err := domain.NewProject("p1", "Alpha Search", "Goal One", "Principle", "Vision", "Desc", nil, now)
	g.Expect(err).ToNot(gomega.HaveOccurred())
	beta, err := domain.NewProject("p2", "Beta", "Search Goal", "Principle", "Vision", "Desc", nil, now.Add(time.Minute))
	g.Expect(err).ToNot(gomega.HaveOccurred())
	gamma, err := domain.NewProject("p3", "Gamma", "Other", "Principle", "Vision Search", "Desc", nil, now.Add(2*time.Minute))
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(gamma.SetStatus(domain.StatusActive, now.Add(3*time.Minute))).To(gomega.Succeed())

	g.Expect(repo.Create(ctx, alpha)).To(gomega.Succeed())
	g.Expect(repo.Create(ctx, beta)).To(gomega.Succeed())
	g.Expect(repo.Create(ctx, gamma)).To(gomega.Succeed())

	list, err := repo.List(ctx, domain.ListQuery{Search: "search"})
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(list).To(gomega.HaveLen(3))

	paged, err := repo.List(ctx, domain.ListQuery{
		Search: "search",
		Sorts:  []domain.Sort{{By: domain.ProjectSortByName, Dir: domain.SortAsc}},
		Limit:  1,
		Offset: 1,
	})
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(paged).To(gomega.HaveLen(1))
	g.Expect(paged[0].ID).To(gomega.Equal("p2"))

	active := domain.StatusActive
	filtered, err := repo.List(ctx, domain.ListQuery{
		Status: &active,
		Search: "vision",
	})
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(filtered).To(gomega.HaveLen(1))
	g.Expect(filtered[0].ID).To(gomega.Equal("p3"))

	caseInsensitive, err := repo.List(ctx, domain.ListQuery{Search: "ALPHA"})
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(caseInsensitive).To(gomega.HaveLen(1))
	g.Expect(caseInsensitive[0].ID).To(gomega.Equal("p1"))
}
