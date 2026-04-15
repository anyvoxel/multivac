package domain

import (
	"testing"
	"time"

	"github.com/onsi/gomega"
)

func TestTaskLifecycle(t *testing.T) {
	g := gomega.NewWithT(t)

	now := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	due := now.Add(24 * time.Hour)
	tk, err := NewTask("id", "pid", "name", "desc", "ctx", "details", PriorityHigh, &due, now)
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(tk.Status).To(gomega.Equal(StatusTodo))
	g.Expect(tk.Priority).To(gomega.Equal(PriorityHigh))
	g.Expect(tk.DueAt).ToNot(gomega.BeNil())

	now2 := now.Add(time.Hour)
	g.Expect(tk.SetStatus(StatusInProgress, now2)).To(gomega.Succeed())
	g.Expect(tk.Status).To(gomega.Equal(StatusInProgress))

	now3 := now2.Add(time.Hour)
	g.Expect(tk.UpdateDetails("", "n2", "d2", "ctx2", "details2", PriorityLow, nil, now3)).To(gomega.Succeed())
	g.Expect(tk.ProjectID).To(gomega.Equal(""))
	g.Expect(tk.Name).To(gomega.Equal("n2"))
	g.Expect(tk.DueAt).To(gomega.BeNil())
}

func TestNewTaskValidationErrorIncludesMissingFields(t *testing.T) {
	g := gomega.NewWithT(t)

	_, err := NewTask("id", "", "", "", "ctx", "", PriorityHigh, nil, time.Now())
	g.Expect(err).To(gomega.MatchError("invalid argument: [name: Required value]"))
}

func TestTaskUpdateValidationErrorIncludesMissingFields(t *testing.T) {
	g := gomega.NewWithT(t)

	tk, err := NewTask("id", "pid", "name", "desc", "ctx", "details", PriorityHigh, nil, time.Now())
	g.Expect(err).ToNot(gomega.HaveOccurred())

	err = tk.UpdateDetails("pid", "", "", "ctx", "", PriorityLow, nil, time.Now())
	g.Expect(err).To(gomega.MatchError("invalid argument: [name: Required value]"))
}

func TestTaskAllowsEmptyDescriptionAndDetails(t *testing.T) {
	g := gomega.NewWithT(t)

	tk, err := NewTask("id", "pid", "name", "", "ctx", "", PriorityHigh, nil, time.Now())
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(tk.Description).To(gomega.Equal(""))
	g.Expect(tk.Details).To(gomega.Equal(""))

	err = tk.UpdateDetails("", "name2", "", "ctx2", "", PriorityLow, nil, time.Now())
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(tk.ProjectID).To(gomega.Equal(""))
	g.Expect(tk.Description).To(gomega.Equal(""))
	g.Expect(tk.Details).To(gomega.Equal(""))
}

func TestTaskAllowsEmptyProjectID(t *testing.T) {
	g := gomega.NewWithT(t)

	tk, err := NewTask("id", "", "name", "desc", "ctx", "details", PriorityHigh, nil, time.Now())
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(tk.ProjectID).To(gomega.Equal(""))
}

func TestTaskValidationErrorIncludesUnsupportedValues(t *testing.T) {
	g := gomega.NewWithT(t)

	_, err := NewTask("id", "pid", "name", "desc", "ctx", "details", Priority("wat"), nil, time.Now())
	g.Expect(err).To(gomega.MatchError("invalid argument: [priority: Unsupported value: \"wat\"]"))

	tk, err := NewTask("id", "pid", "name", "desc", "ctx", "details", PriorityHigh, nil, time.Now())
	g.Expect(err).ToNot(gomega.HaveOccurred())

	err = tk.SetStatus(Status("wat"), time.Now())
	g.Expect(err).To(gomega.MatchError("invalid argument: [status: Unsupported value: \"wat\"]"))
}
