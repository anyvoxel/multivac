package domain

import (
	"testing"
	"time"

	"github.com/onsi/gomega"
)

func TestProjectLifecycle(t *testing.T) {
	g := gomega.NewWithT(t)

	now := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	p, err := NewProject("id", "name", "goal", "principles", "vision", "desc", nil, now)
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(p.Status).To(gomega.Equal(StatusDraft))
	g.Expect(p.CreatedAt).To(gomega.Equal(now))
	g.Expect(p.UpdatedAt).To(gomega.Equal(now))

	// Activate
	now2 := now.Add(time.Hour)
	g.Expect(p.SetStatus(StatusActive, now2)).To(gomega.Succeed())
	g.Expect(p.Status).To(gomega.Equal(StatusActive))
	g.Expect(p.StartedAt).ToNot(gomega.BeNil())
	g.Expect(*p.StartedAt).To(gomega.Equal(now2))

	// Complete
	now3 := now2.Add(time.Hour)
	g.Expect(p.SetStatus(StatusCompleted, now3)).To(gomega.Succeed())
	g.Expect(p.CompletedAt).ToNot(gomega.BeNil())
	g.Expect(*p.CompletedAt).To(gomega.Equal(now3))
}

func TestNewProjectValidationErrorIncludesMissingFields(t *testing.T) {
	g := gomega.NewWithT(t)

	_, err := NewProject("id", "", "", "principles", "", "", nil, time.Now())
	g.Expect(err).To(gomega.MatchError("invalid argument: [name: Required value, goal: Required value, visionResult: Required value, description: Required value]"))
	g.Expect(err).To(gomega.MatchError(gomega.HavePrefix("invalid argument:")))
	g.Expect(err).To(gomega.MatchError(gomega.ContainSubstring("name: Required value")))
	g.Expect(err).To(gomega.MatchError(gomega.ContainSubstring("goal: Required value")))
}

func TestProjectUpdateValidationErrorIncludesMissingFields(t *testing.T) {
	g := gomega.NewWithT(t)

	p, err := NewProject("id", "name", "goal", "principles", "vision", "desc", nil, time.Now())
	g.Expect(err).ToNot(gomega.HaveOccurred())

	err = p.UpdateDetails("", "goal", "", "vision", "", nil, time.Now())
	g.Expect(err).To(gomega.MatchError("invalid argument: [name: Required value, principles: Required value, description: Required value]"))
}

func TestProjectSetStatusValidationErrorIncludesField(t *testing.T) {
	g := gomega.NewWithT(t)

	p, err := NewProject("id", "name", "goal", "principles", "vision", "desc", nil, time.Now())
	g.Expect(err).ToNot(gomega.HaveOccurred())

	err = p.SetStatus(Status("wat"), time.Now())
	g.Expect(err).To(gomega.MatchError("invalid argument: [status: Unsupported value: \"wat\"]"))
}
