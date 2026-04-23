package domain

import (
	"testing"
	"time"

	"github.com/onsi/gomega"
)

func TestProjectLifecycle(t *testing.T) {
	g := gomega.NewWithT(t)

	now := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	p, err := NewProject("id", "title", []ProjectGoal{{Title: "goal"}}, "desc", nil, now)
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(p.Status).To(gomega.Equal(StatusDraft))
	g.Expect(p.CreatedAt).To(gomega.Equal(now))
	g.Expect(p.UpdatedAt).To(gomega.Equal(now))
	g.Expect(p.Goals).To(gomega.HaveLen(1))
	g.Expect(p.Goals[0].CreatedAt).To(gomega.Equal(now))
	g.Expect(p.Goals[0].CompletedAt).To(gomega.BeNil())

	now2 := now.Add(time.Hour)
	g.Expect(p.SetStatus(StatusActive, now2)).To(gomega.Succeed())
	g.Expect(p.Status).To(gomega.Equal(StatusActive))
	g.Expect(p.StartAt).ToNot(gomega.BeNil())
	g.Expect(*p.StartAt).To(gomega.Equal(now2))

	now3 := now2.Add(time.Hour)
	g.Expect(p.SetStatus(StatusCompleted, now3)).To(gomega.Succeed())
	g.Expect(p.CompletedAt).ToNot(gomega.BeNil())
	g.Expect(*p.CompletedAt).To(gomega.Equal(now3))
}

func TestNewProjectValidationErrorIncludesMissingFields(t *testing.T) {
	g := gomega.NewWithT(t)

	_, err := NewProject("id", "", nil, "", nil, time.Now())
	g.Expect(err).To(gomega.MatchError("invalid argument: [title: Required value, description: Required value]"))
	g.Expect(err).To(gomega.MatchError(gomega.HavePrefix("invalid argument:")))
	g.Expect(err).To(gomega.MatchError(gomega.ContainSubstring("title: Required value")))
}

func TestProjectUpdateValidationErrorIncludesMissingFields(t *testing.T) {
	g := gomega.NewWithT(t)

	p, err := NewProject("id", "title", []ProjectGoal{{Title: "goal"}}, "desc", nil, time.Now())
	g.Expect(err).ToNot(gomega.HaveOccurred())

	err = p.UpdateDetails("", []ProjectGoal{{Title: "goal"}}, "", nil, time.Now())
	g.Expect(err).To(gomega.MatchError("invalid argument: [title: Required value, description: Required value]"))
}

func TestProjectSetStatusValidationErrorIncludesField(t *testing.T) {
	g := gomega.NewWithT(t)

	p, err := NewProject("id", "title", []ProjectGoal{{Title: "goal"}}, "desc", nil, time.Now())
	g.Expect(err).ToNot(gomega.HaveOccurred())

	err = p.SetStatus(Status("wat"), time.Now())
	g.Expect(err).To(gomega.MatchError("invalid argument: [status: Unsupported value: \"wat\"]"))
}

func TestProjectGoalTimestampNormalization(t *testing.T) {
	g := gomega.NewWithT(t)
	now := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	completedAt := now.Add(time.Hour)

	p, err := NewProject("id", "title", []ProjectGoal{{Title: "  my goal  ", CompletedAt: &completedAt}}, "desc", nil, now)
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(p.Goals).To(gomega.HaveLen(1))
	g.Expect(p.Goals[0].Title).To(gomega.Equal("my goal"))
	g.Expect(p.Goals[0].CompletedAt).ToNot(gomega.BeNil())
	g.Expect(*p.Goals[0].CompletedAt).To(gomega.Equal(completedAt))
	g.Expect(p.Goals[0].CreatedAt).To(gomega.Equal(now))

	now2 := now.Add(2 * time.Hour)
	err = p.UpdateDetails("title", []ProjectGoal{{Title: "goal"}}, "desc", nil, now2)
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(p.Goals[0].CompletedAt).To(gomega.BeNil())
}

func TestProjectGoalRejectsEmptyOrMultilineTitle(t *testing.T) {
	g := gomega.NewWithT(t)

	_, err := NewProject("id", "title", []ProjectGoal{{Title: "  "}}, "desc", nil, time.Now())
	g.Expect(err).To(gomega.MatchError(gomega.ContainSubstring("goals[0].title: Required value")))

	_, err = NewProject("id", "title", []ProjectGoal{{Title: "line1\nline2"}}, "desc", nil, time.Now())
	g.Expect(err).To(gomega.MatchError(gomega.ContainSubstring("goals[0].title: Unsupported value")))
}

func TestProjectReferenceValidation(t *testing.T) {
	g := gomega.NewWithT(t)

	_, err := NewProject("id", "title", []ProjectGoal{{Title: "goal"}}, "desc", []ProjectReference{{Title: "", URL: "https://example.com"}}, time.Now())
	g.Expect(err).To(gomega.MatchError(gomega.ContainSubstring("references[0].title: Required value")))

	_, err = NewProject("id", "title", []ProjectGoal{{Title: "goal"}}, "desc", []ProjectReference{{Title: "ref", URL: "notaurl"}}, time.Now())
	g.Expect(err).To(gomega.MatchError(gomega.ContainSubstring("references[0].url: Unsupported value")))
}

func TestProjectStatusTransitions(t *testing.T) {
	g := gomega.NewWithT(t)
	now := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)

	p, err := NewProject("id", "title", []ProjectGoal{{Title: "goal"}}, "desc", nil, now)
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(p.SetStatus(StatusHold, now.Add(time.Hour))).To(gomega.Succeed())
	g.Expect(p.SetStatus(StatusActive, now.Add(2*time.Hour))).To(gomega.Succeed())
	g.Expect(p.SetStatus(StatusCompleted, now.Add(3*time.Hour))).To(gomega.Succeed())
	g.Expect(p.SetStatus(StatusActive, now.Add(4*time.Hour))).To(gomega.MatchError(gomega.ContainSubstring("Unsupported transition")))
}
