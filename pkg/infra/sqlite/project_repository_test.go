package sqlite

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/onsi/gomega"

	"github.com/anyvoxel/multivac/pkg/domain"
)

func TestProjectRepositoryCRUD(t *testing.T) {
	g := gomega.NewWithT(t)

	db, err := sqlx.Open("sqlite3", ":memory:")
	g.Expect(err).ToNot(gomega.HaveOccurred())
	defer func() { _ = db.Close() }()

	repo := NewProjectRepository(db)
	ctx := context.Background()
	g.Expect(repo.Migrate(ctx)).To(gomega.Succeed())

	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	p, err := domain.NewProject("p1", "n", []domain.ProjectGoal{{Title: "g"}}, "d", []domain.ProjectReference{{Title: "ref", URL: "https://example.com"}}, now)
	g.Expect(err).ToNot(gomega.HaveOccurred())

	g.Expect(repo.Create(ctx, p)).To(gomega.Succeed())

	got, err := repo.Get(ctx, "p1")
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(got.ID).To(gomega.Equal("p1"))
	g.Expect(got.Title).To(gomega.Equal("n"))
	g.Expect(got.Description).To(gomega.Equal("d"))
	g.Expect(got.References).To(gomega.HaveLen(1))

	status := domain.StatusDraft
	list, err := repo.List(ctx, domain.ListQuery{Status: &status})
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(list).To(gomega.HaveLen(1))

	now2 := now.Add(time.Hour)
	g.Expect(p.UpdateDetails("n2", []domain.ProjectGoal{{Title: "g2"}}, "d2", []domain.ProjectReference{{Title: "docs", URL: "https://example.com/docs"}}, now2)).To(gomega.Succeed())
	g.Expect(p.SetStatus(domain.StatusActive, now2)).To(gomega.Succeed())
	g.Expect(repo.Update(ctx, p)).To(gomega.Succeed())

	got2, err := repo.Get(ctx, "p1")
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(got2.Title).To(gomega.Equal("n2"))
	g.Expect(got2.Status).To(gomega.Equal(domain.StatusActive))
	g.Expect(got2.StartAt).ToNot(gomega.BeNil())

	g.Expect(repo.Delete(ctx, "p1")).To(gomega.Succeed())
	_, err = repo.Get(ctx, "p1")
	g.Expect(err).To(gomega.Equal(domain.ErrNotFound))
}

func TestProjectRepositoryListSupportsSearchPaginationAndSort(t *testing.T) {
	g := gomega.NewWithT(t)

	db, err := sqlx.Open("sqlite3", ":memory:")
	g.Expect(err).ToNot(gomega.HaveOccurred())
	defer func() { _ = db.Close() }()

	repo := NewProjectRepository(db)
	ctx := context.Background()
	g.Expect(repo.Migrate(ctx)).To(gomega.Succeed())

	now := time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)
	alpha, err := domain.NewProject("p1", "Alpha Search", []domain.ProjectGoal{{Title: "Goal One"}}, "Desc", nil, now)
	g.Expect(err).ToNot(gomega.HaveOccurred())
	beta, err := domain.NewProject("p2", "Beta", []domain.ProjectGoal{{Title: "Search Goal"}}, "Search Desc", nil, now.Add(time.Minute))
	g.Expect(err).ToNot(gomega.HaveOccurred())
	gamma, err := domain.NewProject("p3", "Gamma", []domain.ProjectGoal{{Title: "Other"}}, "Vision Search", nil, now.Add(2*time.Minute))
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
		Sorts:  []domain.Sort{{By: domain.ProjectSortByTitle, Dir: domain.SortAsc}},
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

func TestProjectRepositoryStoresJSONBColumns(t *testing.T) {
	g := gomega.NewWithT(t)

	db, err := sqlx.Open("sqlite3", ":memory:")
	g.Expect(err).ToNot(gomega.HaveOccurred())
	defer func() { _ = db.Close() }()

	repo := NewProjectRepository(db)
	ctx := context.Background()
	g.Expect(repo.Migrate(ctx)).To(gomega.Succeed())

	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	p, err := domain.NewProject("p1", "n", []domain.ProjectGoal{{Title: "g"}}, "d", []domain.ProjectReference{{Title: "ref", URL: "https://example.com"}}, now)
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(repo.Create(ctx, p)).To(gomega.Succeed())

	var goalsType string
	err = db.GetContext(ctx, &goalsType, `SELECT typeof(goals) FROM projects WHERE id = ?`, "p1")
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(goalsType).To(gomega.Equal("blob"))

	var referencesType string
	err = db.GetContext(ctx, &referencesType, `SELECT typeof("references") FROM projects WHERE id = ?`, "p1")
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(referencesType).To(gomega.Equal("blob"))

	var goalJSON sql.NullString
	err = db.GetContext(ctx, &goalJSON, `SELECT json(goals) FROM projects WHERE id = ?`, "p1")
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(goalJSON.Valid).To(gomega.BeTrue())
	g.Expect(goalJSON.String).To(gomega.ContainSubstring(`"title":"g"`))
}
