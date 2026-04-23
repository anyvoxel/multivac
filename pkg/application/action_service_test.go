package application

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/onsi/gomega"

	actiondomain "github.com/anyvoxel/multivac/pkg/domain"
	actionsqlite "github.com/anyvoxel/multivac/pkg/infra/sqlite"
	inboxsqlite "github.com/anyvoxel/multivac/pkg/infra/sqlite"
	projectsqlite "github.com/anyvoxel/multivac/pkg/infra/sqlite"
)

func TestActionServiceCreateGeneratesULIDAndUTCFields(t *testing.T) {
	g := gomega.NewWithT(t)
	db, err := sqlx.Open("sqlite3", ":memory:")
	g.Expect(err).ToNot(gomega.HaveOccurred())
	defer func() { _ = db.Close() }()
	ctx := context.Background()
	projectRepo := projectsqlite.NewProjectRepository(db)
	repo := actionsqlite.NewActionRepository(db)
	svc := NewActionService(repo, WithActionNow(func() time.Time {
		return time.Date(2026, 1, 11, 8, 30, 0, 0, time.FixedZone("UTC+8", 8*3600))
	}))
	g.Expect(projectRepo.Migrate(ctx)).To(gomega.Succeed())
	g.Expect(svc.Migrate(ctx)).To(gomega.Succeed())

	action, err := svc.Create(ctx, CreateActionCmd{Title: "Action", Description: "body", Kind: actiondomain.KindTask, Attributes: actiondomain.Attributes{Task: &actiondomain.TaskAttributes{}}})
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(len(action.ID)).To(gomega.Equal(26))
	g.Expect(strings.ToUpper(action.ID)).To(gomega.Equal(action.ID))
	g.Expect(action.CreatedAt.Location()).To(gomega.Equal(time.UTC))
	g.Expect(action.UpdatedAt.Location()).To(gomega.Equal(time.UTC))
	g.Expect(action.CreatedAt).To(gomega.Equal(time.Date(2026, 1, 11, 0, 30, 0, 0, time.UTC)))
	loaded, err := svc.Get(ctx, action.ID)
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(loaded.Title).To(gomega.Equal("Action"))
	g.Expect(loaded.Description).To(gomega.Equal("body"))
	g.Expect(loaded.Attributes.Task).ToNot(gomega.BeNil())
	g.Expect(loaded.Attributes.Task.Status).To(gomega.Equal(actiondomain.TaskStatusPending))
}

func TestActionServiceConvertFromInboxKeepsIDAndRemovesInbox(t *testing.T) {
	g := gomega.NewWithT(t)
	db, err := sqlx.Open("sqlite3", ":memory:")
	g.Expect(err).ToNot(gomega.HaveOccurred())
	defer func() { _ = db.Close() }()
	ctx := context.Background()

	inboxRepo := inboxsqlite.NewInboxRepository(db)
	projectRepo := projectsqlite.NewProjectRepository(db)
	repo := actionsqlite.NewActionRepository(db)
	svc := NewActionService(repo, WithActionNow(func() time.Time {
		return time.Date(2026, 1, 15, 8, 0, 0, 0, time.UTC)
	}))
	g.Expect(inboxRepo.Migrate(ctx)).To(gomega.Succeed())
	g.Expect(projectRepo.Migrate(ctx)).To(gomega.Succeed())
	g.Expect(svc.Migrate(ctx)).To(gomega.Succeed())

	now := time.Date(2026, 1, 14, 8, 0, 0, 0, time.UTC)
	_, err = db.ExecContext(ctx, `INSERT INTO inboxes (id, title, description, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`, "01J0000000000000000000301", "Inbox title", "Inbox body", now, now)
	g.Expect(err).ToNot(gomega.HaveOccurred())

	overrideTitle := "Action title"
	action, err := svc.ConvertFromInbox(ctx, "01J0000000000000000000301", ConvertActionFromInboxCmd{Title: &overrideTitle})
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(action.ID).To(gomega.Equal("01J0000000000000000000301"))
	g.Expect(action.Title).To(gomega.Equal("Action title"))
	g.Expect(action.Description).To(gomega.Equal("Inbox body"))

	var inboxCount int
	err = db.GetContext(ctx, &inboxCount, `SELECT COUNT(1) FROM inboxes WHERE id = ?`, "01J0000000000000000000301")
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(inboxCount).To(gomega.Equal(0))
}

func TestActionServiceCreateScheduledAndListByKind(t *testing.T) {
	g := gomega.NewWithT(t)
	db, err := sqlx.Open("sqlite3", ":memory:")
	g.Expect(err).ToNot(gomega.HaveOccurred())
	defer func() { _ = db.Close() }()
	ctx := context.Background()

	projectRepo := projectsqlite.NewProjectRepository(db)
	repo := actionsqlite.NewActionRepository(db)
	svc := NewActionService(repo)
	g.Expect(projectRepo.Migrate(ctx)).To(gomega.Succeed())
	g.Expect(svc.Migrate(ctx)).To(gomega.Succeed())

	startAt := time.Date(2026, 2, 1, 9, 0, 0, 0, time.FixedZone("UTC+8", 8*3600))
	endAt := startAt.Add(90 * time.Minute)
	action, err := svc.Create(ctx, CreateActionCmd{
		Title:       "会议",
		Description: "周会",
		Kind:        actiondomain.KindScheduled,
		Attributes: actiondomain.Attributes{
			Scheduled: &actiondomain.ScheduledAttributes{StartAt: &startAt, EndAt: &endAt},
		},
	})
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(action.Kind).To(gomega.Equal(actiondomain.KindScheduled))
	g.Expect(action.Attributes.Scheduled).ToNot(gomega.BeNil())
	g.Expect(action.Attributes.Scheduled.StartAt.Location()).To(gomega.Equal(time.UTC))
	g.Expect(action.Attributes.Scheduled.EndAt.Location()).To(gomega.Equal(time.UTC))

	kind := actiondomain.KindScheduled
	listed, err := svc.List(ctx, actiondomain.ListQuery{Kind: &kind})
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(listed).To(gomega.HaveLen(1))
	g.Expect(listed[0].ID).To(gomega.Equal(action.ID))
}
