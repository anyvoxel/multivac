package sqlite

import (
	"context"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/onsi/gomega"

	"github.com/anyvoxel/multivac/pkg/inbox/domain"
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
	inbox, err := domain.NewInbox("i1", "name", "desc", now)
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(repo.Create(ctx, inbox)).To(gomega.Succeed())

	got, err := repo.Get(ctx, "i1")
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(got.ID).To(gomega.Equal("i1"))
	g.Expect(got.Name).To(gomega.Equal("name"))

	list, err := repo.List(ctx, domain.ListQuery{})
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(list).To(gomega.HaveLen(1))

	now2 := now.Add(time.Hour)
	g.Expect(inbox.UpdateDetails("name2", "desc2", now2)).To(gomega.Succeed())
	g.Expect(repo.Update(ctx, inbox)).To(gomega.Succeed())

	updated, err := repo.Get(ctx, "i1")
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(updated.Name).To(gomega.Equal("name2"))
	g.Expect(updated.Description).To(gomega.Equal("desc2"))

	g.Expect(repo.Delete(ctx, "i1")).To(gomega.Succeed())
	_, err = repo.Get(ctx, "i1")
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
	alpha, err := domain.NewInbox("i1", "Alpha", "Search item", now)
	g.Expect(err).ToNot(gomega.HaveOccurred())
	beta, err := domain.NewInbox("i2", "Beta Search", "desc", now.Add(time.Minute))
	g.Expect(err).ToNot(gomega.HaveOccurred())
	gamma, err := domain.NewInbox("i3", "Gamma", "other", now.Add(2*time.Minute))
	g.Expect(err).ToNot(gomega.HaveOccurred())

	g.Expect(repo.Create(ctx, alpha)).To(gomega.Succeed())
	g.Expect(repo.Create(ctx, beta)).To(gomega.Succeed())
	g.Expect(repo.Create(ctx, gamma)).To(gomega.Succeed())

	list, err := repo.List(ctx, domain.ListQuery{Search: "search"})
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(list).To(gomega.HaveLen(2))

	paged, err := repo.List(ctx, domain.ListQuery{
		Sorts:  []domain.Sort{{By: domain.InboxSortByName, Dir: domain.SortAsc}},
		Limit:  1,
		Offset: 1,
	})
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(paged).To(gomega.HaveLen(1))
	g.Expect(paged[0].ID).To(gomega.Equal("i2"))

	caseInsensitive, err := repo.List(ctx, domain.ListQuery{Search: "ALPHA"})
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(caseInsensitive).To(gomega.HaveLen(1))
	g.Expect(caseInsensitive[0].ID).To(gomega.Equal("i1"))
}
