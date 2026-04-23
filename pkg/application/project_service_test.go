package application

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/onsi/gomega"

	projectdomain "github.com/anyvoxel/multivac/pkg/domain"
	projectsqlite "github.com/anyvoxel/multivac/pkg/infra/sqlite"
)

func TestProjectServiceCreateGeneratesULIDAndUTCFields(t *testing.T) {
	g := gomega.NewWithT(t)
	db, err := sqlx.Open("sqlite3", ":memory:")
	g.Expect(err).ToNot(gomega.HaveOccurred())
	defer func() { _ = db.Close() }()
	ctx := context.Background()
	repo := projectsqlite.NewProjectRepository(db)
	svc := NewProjectService(repo, WithProjectNow(func() time.Time {
		return time.Date(2026, 1, 11, 8, 30, 0, 0, time.FixedZone("UTC+8", 8*3600))
	}))
	g.Expect(svc.Migrate(ctx)).To(gomega.Succeed())

	project, err := svc.Create(ctx, CreateProjectCmd{
		Title:       "Project",
		Goals:       []projectdomain.ProjectGoal{{Title: "Goal"}},
		Description: "body",
		References:  []projectdomain.ProjectReference{{Title: "ref", URL: "https://example.com"}},
	})
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(len(project.ID)).To(gomega.Equal(26))
	g.Expect(strings.ToUpper(project.ID)).To(gomega.Equal(project.ID))
	g.Expect(project.CreatedAt.Location()).To(gomega.Equal(time.UTC))
	g.Expect(project.UpdatedAt.Location()).To(gomega.Equal(time.UTC))
	g.Expect(project.CreatedAt).To(gomega.Equal(time.Date(2026, 1, 11, 0, 30, 0, 0, time.UTC)))

	loaded, err := svc.Get(ctx, project.ID)
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(loaded.Title).To(gomega.Equal("Project"))
	g.Expect(loaded.Description).To(gomega.Equal("body"))
}
