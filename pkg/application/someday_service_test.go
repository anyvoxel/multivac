package application

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/onsi/gomega"

	inboxsqlite "github.com/anyvoxel/multivac/pkg/infra/sqlite"
	somedaysqlite "github.com/anyvoxel/multivac/pkg/infra/sqlite"
)

func TestSomedayServiceCreateGeneratesULIDAndUTCFields(t *testing.T) {
	g := gomega.NewWithT(t)
	db, err := sqlx.Open("sqlite3", ":memory:")
	g.Expect(err).ToNot(gomega.HaveOccurred())
	defer func() { _ = db.Close() }()
	ctx := context.Background()
	repo := somedaysqlite.NewSomedayRepository(db)
	svc := NewSomedayService(repo, WithSomedayNow(func() time.Time {
		return time.Date(2026, 1, 11, 8, 30, 0, 0, time.FixedZone("UTC+8", 8*3600))
	}))
	g.Expect(svc.Migrate(ctx)).To(gomega.Succeed())

	someday, err := svc.Create(ctx, CreateSomedayCmd{Title: "Someday", Description: "body"})
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(len(someday.ID)).To(gomega.Equal(26))
	g.Expect(strings.ToUpper(someday.ID)).To(gomega.Equal(someday.ID))
	g.Expect(someday.CreatedAt.Location()).To(gomega.Equal(time.UTC))
	g.Expect(someday.UpdatedAt.Location()).To(gomega.Equal(time.UTC))
	g.Expect(someday.CreatedAt).To(gomega.Equal(time.Date(2026, 1, 11, 0, 30, 0, 0, time.UTC)))
	loaded, err := svc.Get(ctx, someday.ID)
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(loaded.Title).To(gomega.Equal("Someday"))
	g.Expect(loaded.Description).To(gomega.Equal("body"))
}

func TestSomedayServiceConvertFromInboxKeepsIDAndRemovesInbox(t *testing.T) {
	g := gomega.NewWithT(t)
	db, err := sqlx.Open("sqlite3", ":memory:")
	g.Expect(err).ToNot(gomega.HaveOccurred())
	defer func() { _ = db.Close() }()
	ctx := context.Background()

	inboxRepo := inboxsqlite.NewInboxRepository(db)
	repo := somedaysqlite.NewSomedayRepository(db)
	svc := NewSomedayService(repo, WithSomedayNow(func() time.Time {
		return time.Date(2026, 1, 15, 8, 0, 0, 0, time.UTC)
	}))
	g.Expect(inboxRepo.Migrate(ctx)).To(gomega.Succeed())
	g.Expect(svc.Migrate(ctx)).To(gomega.Succeed())

	now := time.Date(2026, 1, 14, 8, 0, 0, 0, time.UTC)
	_, err = db.ExecContext(ctx, `INSERT INTO inboxes (id, title, description, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`, "01J0000000000000000000301", "Inbox title", "Inbox body", now, now)
	g.Expect(err).ToNot(gomega.HaveOccurred())

	overrideTitle := "Someday title"
	overrideDesc := "Someday body"
	someday, err := svc.ConvertFromInbox(ctx, "01J0000000000000000000301", ConvertSomedayFromInboxCmd{Title: &overrideTitle, Description: &overrideDesc})
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(someday.ID).To(gomega.Equal("01J0000000000000000000301"))
	g.Expect(someday.Title).To(gomega.Equal("Someday title"))
	g.Expect(someday.Description).To(gomega.Equal("Someday body"))

	var inboxCount int
	err = db.GetContext(ctx, &inboxCount, `SELECT COUNT(1) FROM inboxes WHERE id = ?`, "01J0000000000000000000301")
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(inboxCount).To(gomega.Equal(0))
}
