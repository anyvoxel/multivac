package domain

import (
	"testing"
	"time"

	"github.com/onsi/gomega"
)

func TestProjectLifecycle(t *testing.T) {
	g := gomega.NewWithT(t)

	now := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	p, err := NewProject("id", "name", []Goal{{Text: "goal"}}, "desc", nil, now)
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(p.Status).To(gomega.Equal(StatusDraft))
	g.Expect(p.CreatedAt).To(gomega.Equal(now))
	g.Expect(p.UpdatedAt).To(gomega.Equal(now))
	g.Expect(p.Goals).To(gomega.HaveLen(1))
	g.Expect(p.Goals[0].Completed).To(gomega.BeFalse())
	g.Expect(p.Goals[0].CreatedAt).To(gomega.Equal(now))
	g.Expect(p.Goals[0].CompletedAt).To(gomega.BeNil())

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

	_, err := NewProject("id", "", nil, "", nil, time.Now())
	g.Expect(err).To(gomega.MatchError("invalid argument: [name: Required value, description: Required value]"))
	g.Expect(err).To(gomega.MatchError(gomega.HavePrefix("invalid argument:")))
	g.Expect(err).To(gomega.MatchError(gomega.ContainSubstring("name: Required value")))
}

func TestProjectUpdateValidationErrorIncludesMissingFields(t *testing.T) {
	g := gomega.NewWithT(t)

	p, err := NewProject("id", "name", []Goal{{Text: "goal"}}, "desc", nil, time.Now())
	g.Expect(err).ToNot(gomega.HaveOccurred())

	err = p.UpdateDetails("", []Goal{{Text: "goal"}}, "", nil, time.Now())
	g.Expect(err).To(gomega.MatchError("invalid argument: [name: Required value, description: Required value]"))
}

func TestProjectSetStatusValidationErrorIncludesField(t *testing.T) {
	g := gomega.NewWithT(t)

	p, err := NewProject("id", "name", []Goal{{Text: "goal"}}, "desc", nil, time.Now())
	g.Expect(err).ToNot(gomega.HaveOccurred())

	err = p.SetStatus(Status("wat"), time.Now())
	g.Expect(err).To(gomega.MatchError("invalid argument: [status: Unsupported value: \"wat\"]"))
}

func TestProjectGoalCompletionAndTimestampNormalization(t *testing.T) {
	g := gomega.NewWithT(t)
	now := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)

	p, err := NewProject("id", "name", []Goal{{Text: "  my goal  ", Completed: true}}, "desc", nil, now)
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(p.Goals).To(gomega.HaveLen(1))
	g.Expect(p.Goals[0].Text).To(gomega.Equal("my goal"))
	g.Expect(p.Goals[0].Completed).To(gomega.BeTrue())
	g.Expect(p.Goals[0].CompletedAt).ToNot(gomega.BeNil())
	g.Expect(*p.Goals[0].CompletedAt).To(gomega.Equal(now))
	g.Expect(p.Goals[0].CreatedAt).To(gomega.Equal(now))

	now2 := now.Add(time.Hour)
	err = p.UpdateDetails("name", []Goal{{Text: "goal", Completed: false, CompletedAt: &now2}}, "desc", nil, now2)
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(p.Goals[0].Completed).To(gomega.BeFalse())
	g.Expect(p.Goals[0].CompletedAt).To(gomega.BeNil())
}

func TestProjectGoalRejectsEmptyOrMultilineText(t *testing.T) {
	g := gomega.NewWithT(t)

	_, err := NewProject("id", "name", []Goal{{Text: "  "}}, "desc", nil, time.Now())
	g.Expect(err).To(gomega.MatchError(gomega.ContainSubstring("goals[0].text: Required value")))

	_, err = NewProject("id", "name", []Goal{{Text: "line1\nline2"}}, "desc", nil, time.Now())
	g.Expect(err).To(gomega.MatchError(gomega.ContainSubstring("goals[0].text: Unsupported value")))
}
