package sqlite

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/onsi/gomega"

	"github.com/anyvoxel/multivac/pkg/reference/domain"
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
	reference, err := domain.NewReference("r1", "n", "d", []domain.ReferenceLink{{Title: "ref", URL: "https://example.com"}}, now)
	g.Expect(err).ToNot(gomega.HaveOccurred())

	g.Expect(repo.Create(ctx, reference)).To(gomega.Succeed())

	got, err := repo.Get(ctx, "r1")
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(got.ID).To(gomega.Equal("r1"))
	g.Expect(got.Title).To(gomega.Equal("n"))
	g.Expect(got.Description).To(gomega.Equal("d"))
	g.Expect(got.References).To(gomega.HaveLen(1))
	g.Expect(got.References[0].URL).To(gomega.Equal("https://example.com"))

	list, err := repo.List(ctx, domain.ListQuery{})
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(list).To(gomega.HaveLen(1))

	now2 := now.Add(time.Hour)
	g.Expect(reference.UpdateDetails("n2", "d2", []domain.ReferenceLink{{Title: "docs", URL: "https://example.com/docs"}}, now2)).To(gomega.Succeed())
	g.Expect(repo.Update(ctx, reference)).To(gomega.Succeed())

	got2, err := repo.Get(ctx, "r1")
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(got2.Title).To(gomega.Equal("n2"))
	g.Expect(got2.References).To(gomega.Equal([]domain.ReferenceLink{{Title: "docs", URL: "https://example.com/docs"}}))

	g.Expect(repo.Delete(ctx, "r1")).To(gomega.Succeed())
	_, err = repo.Get(ctx, "r1")
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
	alpha, err := domain.NewReference("r1", "Alpha Search", "Desc", nil, now)
	g.Expect(err).ToNot(gomega.HaveOccurred())
	beta, err := domain.NewReference("r2", "Beta", "Search Desc", nil, now.Add(time.Minute))
	g.Expect(err).ToNot(gomega.HaveOccurred())
	gamma, err := domain.NewReference("r3", "Gamma", "Vision Search", nil, now.Add(2*time.Minute))
	g.Expect(err).ToNot(gomega.HaveOccurred())

	g.Expect(repo.Create(ctx, alpha)).To(gomega.Succeed())
	g.Expect(repo.Create(ctx, beta)).To(gomega.Succeed())
	g.Expect(repo.Create(ctx, gamma)).To(gomega.Succeed())

	list, err := repo.List(ctx, domain.ListQuery{Search: "search"})
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(list).To(gomega.HaveLen(3))

	paged, err := repo.List(ctx, domain.ListQuery{
		Search: "search",
		Sorts:  []domain.Sort{{By: domain.ReferenceSortByTitle, Dir: domain.SortAsc}},
		Limit:  1,
		Offset: 1,
	})
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(paged).To(gomega.HaveLen(1))
	g.Expect(paged[0].ID).To(gomega.Equal("r2"))

	caseInsensitive, err := repo.List(ctx, domain.ListQuery{Search: "ALPHA"})
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(caseInsensitive).To(gomega.HaveLen(1))
	g.Expect(caseInsensitive[0].ID).To(gomega.Equal("r1"))
}

func TestRepositoryStoresJSONBColumns(t *testing.T) {
	g := gomega.NewWithT(t)

	db, err := sqlx.Open("sqlite3", ":memory:")
	g.Expect(err).ToNot(gomega.HaveOccurred())
	defer func() { _ = db.Close() }()

	repo := NewRepository(db)
	ctx := context.Background()
	g.Expect(repo.Migrate(ctx)).To(gomega.Succeed())

	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	reference, err := domain.NewReference("r1", "n", "d", []domain.ReferenceLink{{Title: "ref", URL: "https://example.com"}}, now)
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(repo.Create(ctx, reference)).To(gomega.Succeed())

	var referencesType string
	err = db.GetContext(ctx, &referencesType, `SELECT typeof("references") FROM "references" WHERE id = ?`, "r1")
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(referencesType).To(gomega.Equal("blob"))

	var referencesJSON sql.NullString
	err = db.GetContext(ctx, &referencesJSON, `SELECT json("references") FROM "references" WHERE id = ?`, "r1")
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(referencesJSON.Valid).To(gomega.BeTrue())
	g.Expect(referencesJSON.String).To(gomega.ContainSubstring(`"title":"ref"`))
	g.Expect(referencesJSON.String).To(gomega.ContainSubstring(`"url":"https://example.com"`))
}
