package domain

import (
	"testing"
	"time"

	"github.com/onsi/gomega"
)

func TestProjectLifecycle(t *testing.T) {
	g := gomega.NewWithT(t)

	now := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	p, err := NewProject("id", "name", "goal", "principles", "vision", "desc", now)
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
