package domain

import (
	"testing"
	"time"

	"github.com/onsi/gomega"
)

func TestNewActionTaskAndNormalization(t *testing.T) {
	g := gomega.NewWithT(t)
	now := time.Date(2026, 1, 2, 3, 4, 5, 0, time.FixedZone("UTC+8", 8*3600))
	projectID := "  p1  "
	action, err := NewAction(
		" id1 ",
		"  title  ",
		"desc",
		&projectID,
		KindTask,
		[]string{" c1 ", "c1", "", "c2"},
		[]Label{{Name: " l1 "}, {Name: "l2"}},
		Attributes{},
		now,
	)
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(action.ID).To(gomega.Equal("id1"))
	g.Expect(action.Title).To(gomega.Equal("title"))
	g.Expect(*action.ProjectID).To(gomega.Equal("p1"))
	g.Expect(action.ContextIDs).To(gomega.Equal([]string{"c1", "c2"}))
	g.Expect(action.Labels).To(gomega.Equal([]Label{{Name: "l1"}, {Name: "l2"}}))
	g.Expect(action.Attributes.Task).ToNot(gomega.BeNil())
	g.Expect(action.CreatedAt.Location()).To(gomega.Equal(time.UTC))
	g.Expect(action.UpdatedAt.Location()).To(gomega.Equal(time.UTC))
}

func TestActionValidateAttributesByKind(t *testing.T) {
	g := gomega.NewWithT(t)
	now := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)

	_, err := NewAction("id", "title", "desc", nil, KindWaiting, nil, nil, Attributes{Waiting: &WaitingAttributes{}}, now)
	g.Expect(err).To(gomega.MatchError(gomega.ContainSubstring("attributes.waiting.delegatee: Required value")))

	dueAt := now.Add(time.Hour)
	_, err = NewAction("id", "title", "desc", nil, KindWaiting, nil, nil, Attributes{Waiting: &WaitingAttributes{Delegatee: "alice", DueAt: &dueAt}}, now)
	g.Expect(err).ToNot(gomega.HaveOccurred())

	startAt := now.Add(time.Hour)
	endAt := now
	_, err = NewAction("id", "title", "desc", nil, KindScheduled, nil, nil, Attributes{Scheduled: &ScheduledAttributes{StartAt: &startAt, EndAt: &endAt}}, now)
	g.Expect(err).To(gomega.MatchError(gomega.ContainSubstring("attributes.scheduled.end_at: Unsupported value")))
}

func TestActionUpdateDetails(t *testing.T) {
	g := gomega.NewWithT(t)
	now := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	action, err := NewAction("id", "title", "desc", nil, KindTask, nil, nil, Attributes{Task: &TaskAttributes{}}, now)
	g.Expect(err).ToNot(gomega.HaveOccurred())

	next := now.Add(2 * time.Hour)
	projectID := "p1"
	dueAt := now.Add(24 * time.Hour)
	err = action.UpdateDetails("new title", "new desc", &projectID, KindWaiting, []string{"c1"}, []Label{{Name: "L"}}, Attributes{Waiting: &WaitingAttributes{Delegatee: "bob", DueAt: &dueAt}}, next)
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(action.Kind).To(gomega.Equal(KindWaiting))
	g.Expect(action.Attributes.Waiting).ToNot(gomega.BeNil())
	g.Expect(action.Attributes.Waiting.Delegatee).To(gomega.Equal("bob"))
	g.Expect(action.UpdatedAt).To(gomega.Equal(next.UTC()))
}
