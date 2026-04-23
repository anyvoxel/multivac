package application

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/onsi/gomega"

	inboxsqlite "github.com/anyvoxel/multivac/pkg/infra/sqlite"
)

func TestInboxServiceCreateGeneratesULIDAndUTCFields(t *testing.T) {
	g := gomega.NewWithT(t)
	db, err := sqlx.Open("sqlite3", ":memory:")
	g.Expect(err).ToNot(gomega.HaveOccurred())
	defer func() { _ = db.Close() }()
	ctx := context.Background()
	repo := inboxsqlite.NewInboxRepository(db)
	svc := NewInboxService(repo, WithInboxNow(func() time.Time {
		return time.Date(2026, 1, 11, 8, 30, 0, 0, time.FixedZone("UTC+8", 8*3600))
	}))
	g.Expect(svc.Migrate(ctx)).To(gomega.Succeed())

	inbox, err := svc.Create(ctx, CreateInboxCmd{Title: "Inbox", Description: "body"})
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(len(inbox.ID)).To(gomega.Equal(26))
	g.Expect(strings.ToUpper(inbox.ID)).To(gomega.Equal(inbox.ID))
	g.Expect(inbox.CreatedAt.Location()).To(gomega.Equal(time.UTC))
	g.Expect(inbox.UpdatedAt.Location()).To(gomega.Equal(time.UTC))
	g.Expect(inbox.CreatedAt).To(gomega.Equal(time.Date(2026, 1, 11, 0, 30, 0, 0, time.UTC)))
	loaded, err := svc.Get(ctx, inbox.ID)
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(loaded.Title).To(gomega.Equal("Inbox"))
	g.Expect(loaded.Description).To(gomega.Equal("body"))
}
