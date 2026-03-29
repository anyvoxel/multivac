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
	p, err := domain.NewProject("p1", "n", "g", "pr", "vr", "d", now)
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
	g.Expect(p.UpdateDetails("n2", "g2", "pr2", "vr2", "d2", now2)).To(gomega.Succeed())
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
