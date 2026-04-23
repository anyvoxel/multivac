package http

import (
	"errors"
	"testing"
	"time"

	"github.com/onsi/gomega"

	"github.com/anyvoxel/multivac/pkg/action/domain"
)

func TestParseKindSupportsScheduled(t *testing.T) {
	g := gomega.NewWithT(t)

	kind, err := parseKind("Scheduled")
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(kind).To(gomega.Equal(domain.KindScheduled))

	kind, err = parseKind("scheduled")
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(kind).To(gomega.Equal(domain.KindScheduled))

	kind, err = parseKind(" ")
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(kind).To(gomega.Equal(domain.Kind("")))
}

func TestParseKindRejectsInvalidValue(t *testing.T) {
	g := gomega.NewWithT(t)

	_, err := parseKind("unknown")
	g.Expect(err).To(gomega.HaveOccurred())
	g.Expect(errors.Is(err, domain.ErrInvalidArg)).To(gomega.BeTrue())
}

func TestToRespIncludesScheduledAttributes(t *testing.T) {
	g := gomega.NewWithT(t)

	now := time.Date(2026, 2, 2, 0, 0, 0, 0, time.UTC)
	startAt := now.Add(2 * time.Hour)
	endAt := startAt.Add(time.Hour)
	action := &domain.Action{
		ID:          "a1",
		Title:       "会议",
		Description: "周会",
		Kind:        domain.KindScheduled,
		ContextIDs:  []string{},
		Labels:      []domain.Label{},
		Attributes: domain.Attributes{
			Scheduled: &domain.ScheduledAttributes{StartAt: &startAt, EndAt: &endAt},
		},
		CreatedAt: now,
		UpdatedAt: now,
	}

	resp := toResp(action)
	g.Expect(resp.Kind).To(gomega.Equal("Scheduled"))
	g.Expect(resp.Attributes.Scheduled).ToNot(gomega.BeNil())
	g.Expect(resp.Attributes.Scheduled.StartAt).ToNot(gomega.BeNil())
	g.Expect(resp.Attributes.Scheduled.EndAt).ToNot(gomega.BeNil())
	g.Expect(resp.Attributes.Scheduled.StartAt.UTC()).To(gomega.Equal(startAt.UTC()))
	g.Expect(resp.Attributes.Scheduled.EndAt.UTC()).To(gomega.Equal(endAt.UTC()))
}
