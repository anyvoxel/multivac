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
	g.Expect(tk.UpdateDetails("n2", "d2", "ctx2", "details2", PriorityLow, nil, now3)).To(gomega.Succeed())
	g.Expect(tk.Name).To(gomega.Equal("n2"))
	g.Expect(tk.DueAt).To(gomega.BeNil())
}
