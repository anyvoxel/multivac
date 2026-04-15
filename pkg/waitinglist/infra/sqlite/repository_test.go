package sqlite

import (
	"context"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/onsi/gomega"

	"github.com/anyvoxel/multivac/pkg/waitinglist/domain"
)

func TestRepositoryCRUD(t *testing.T) {
	g := gomega.NewWithT(t)

	db, err := sqlx.Open("sqlite3", ":memory:")
	g.Expect(err).ToNot(gomega.HaveOccurred())
	defer func() { _ = db.Close() }()

	repo := NewRepository(db)
	ctx := context.Background()
	g.Expect(repo.Migrate(ctx)).To(gomega.Succeed())

	now := time.Date(2026, 1, 5, 9, 0, 0, 0, time.UTC)
	expectedAt := now.Add(24 * time.Hour)
	item, err := domain.NewWaitingList("w1", "name", "details", "alice", &expectedAt, now)
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(repo.Create(ctx, item)).To(gomega.Succeed())

	got, err := repo.Get(ctx, "w1")
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(got.ID).To(gomega.Equal("w1"))
	g.Expect(got.Name).To(gomega.Equal("name"))
	g.Expect(got.Owner).To(gomega.Equal("alice"))
	g.Expect(got.ExpectedAt).ToNot(gomega.BeNil())
	g.Expect(got.ExpectedAt.UTC()).To(gomega.Equal(expectedAt))

	list, err := repo.List(ctx, domain.ListQuery{})
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(list).To(gomega.HaveLen(1))

	now2 := now.Add(time.Hour)
	g.Expect(item.UpdateDetails("name2", "details2", "bob", nil, now2)).To(gomega.Succeed())
	g.Expect(repo.Update(ctx, item)).To(gomega.Succeed())

	updated, err := repo.Get(ctx, "w1")
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(updated.Name).To(gomega.Equal("name2"))
	g.Expect(updated.Details).To(gomega.Equal("details2"))
	g.Expect(updated.Owner).To(gomega.Equal("bob"))
	g.Expect(updated.ExpectedAt).To(gomega.BeNil())

	g.Expect(repo.Delete(ctx, "w1")).To(gomega.Succeed())
	_, err = repo.Get(ctx, "w1")
	g.Expect(err).To(gomega.Equal(domain.ErrNotFound))
}

func TestRepositoryListSupportsSearchPaginationAndExpectedAtSort(t *testing.T) {
	g := gomega.NewWithT(t)

	db, err := sqlx.Open("sqlite3", ":memory:")
	g.Expect(err).ToNot(gomega.HaveOccurred())
	defer func() { _ = db.Close() }()

	repo := NewRepository(db)
	ctx := context.Background()
	g.Expect(repo.Migrate(ctx)).To(gomega.Succeed())

	now := time.Date(2026, 1, 6, 9, 0, 0, 0, time.UTC)
	a1 := now.Add(48 * time.Hour)
	a2 := now.Add(24 * time.Hour)
	alpha, err := domain.NewWaitingList("w1", "Alpha", "Search item", "Alice", &a1, now)
	g.Expect(err).ToNot(gomega.HaveOccurred())
	beta, err := domain.NewWaitingList("w2", "Beta", "desc", "SearchOwner", &a2, now.Add(time.Minute))
	g.Expect(err).ToNot(gomega.HaveOccurred())
	gamma, err := domain.NewWaitingList("w3", "Gamma", "other", "Carol", nil, now.Add(2*time.Minute))
	g.Expect(err).ToNot(gomega.HaveOccurred())

	g.Expect(repo.Create(ctx, alpha)).To(gomega.Succeed())
	g.Expect(repo.Create(ctx, beta)).To(gomega.Succeed())
	g.Expect(repo.Create(ctx, gamma)).To(gomega.Succeed())

	list, err := repo.List(ctx, domain.ListQuery{Search: "search"})
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(list).To(gomega.HaveLen(2))

	ownerHit, err := repo.List(ctx, domain.ListQuery{Search: "searchowner"})
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(ownerHit).To(gomega.HaveLen(1))
	g.Expect(ownerHit[0].ID).To(gomega.Equal("w2"))

	paged, err := repo.List(ctx, domain.ListQuery{
		Sorts:  []domain.Sort{{By: domain.WaitingListSortByName, Dir: domain.SortAsc}},
		Limit:  1,
		Offset: 1,
	})
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(paged).To(gomega.HaveLen(1))
	g.Expect(paged[0].ID).To(gomega.Equal("w2"))

	sorted, err := repo.List(ctx, domain.ListQuery{
		Sorts: []domain.Sort{{By: domain.WaitingListSortByExpectedAt, Dir: domain.SortAsc}},
	})
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(sorted).To(gomega.HaveLen(3))
	g.Expect(sorted[0].ID).To(gomega.Equal("w2"))
	g.Expect(sorted[1].ID).To(gomega.Equal("w1"))
	g.Expect(sorted[2].ID).To(gomega.Equal("w3"))
}
