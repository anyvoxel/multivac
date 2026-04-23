package domain

import (
	"strings"
	"time"
)

type Kind string

const (
	KindTask      Kind = "Task"
	KindWaiting   Kind = "Waiting"
	KindScheduled Kind = "Scheduled"
)

func (k Kind) Valid() bool {
	switch k {
	case KindTask, KindWaiting, KindScheduled:
		return true
	default:
		return false
	}
}

func ParseKind(v string) (Kind, bool) {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "task":
		return KindTask, true
	case "waiting":
		return KindWaiting, true
	case "scheduled":
		return KindScheduled, true
	default:
		return "", false
	}
}

type Label struct {
	Name string `json:"name"`
}

type TaskAttributes struct {
	ExpectedAt *time.Time `json:"expected_at,omitempty"`
}

type WaitingAttributes struct {
	Delegatee string     `json:"delegatee"`
	DueAt     *time.Time `json:"due_at,omitempty"`
}

type ScheduledAttributes struct {
	StartAt *time.Time `json:"start_at,omitempty"`
	EndAt   *time.Time `json:"end_at,omitempty"`
}

type Attributes struct {
	Task      *TaskAttributes      `json:"task,omitempty"`
	Waiting   *WaitingAttributes   `json:"waiting,omitempty"`
	Scheduled *ScheduledAttributes `json:"scheduled,omitempty"`
}

type Action struct {
	ID          string
	Title       string
	Description string
	ProjectID   *string
	Kind        Kind
	ContextIDs  []string
	Labels      []Label
	Attributes  Attributes
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func NewAction(id, title, description string, projectID *string, kind Kind, contextIDs []string, labels []Label, attributes Attributes, now time.Time) (*Action, error) {
	now = now.UTC()
	action := &Action{
		ID:          strings.TrimSpace(id),
		Title:       strings.TrimSpace(title),
		Description: description,
		ProjectID:   normalizeProjectID(projectID),
		Kind:        kind,
		ContextIDs:  normalizeContextIDs(contextIDs),
		Labels:      normalizeLabels(labels),
		Attributes:  normalizeAttributes(kind, attributes),
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := action.Validate(); err != nil {
		return nil, err
	}
	return action, nil
}

func (a *Action) Validate() error {
	if a == nil {
		return ErrInvalidArg
	}
	if err := requiredFieldsError(requiredActionField(a.ID, "id"), requiredActionField(a.Title, "title")); err != nil {
		return err
	}
	if !a.Kind.Valid() {
		return InvalidKind(string(a.Kind))
	}
	if err := validateContextIDs(a.ContextIDs); err != nil {
		return err
	}
	if err := validateLabels(a.Labels); err != nil {
		return err
	}
	if err := validateAttributes(a.Kind, a.Attributes); err != nil {
		return err
	}
	return nil
}

func (a *Action) UpdateDetails(title, description string, projectID *string, kind Kind, contextIDs []string, labels []Label, attributes Attributes, now time.Time) error {
	if a == nil {
		return ErrInvalidArg
	}
	a.Title = strings.TrimSpace(title)
	a.Description = description
	a.ProjectID = normalizeProjectID(projectID)
	a.Kind = kind
	a.ContextIDs = normalizeContextIDs(contextIDs)
	a.Labels = normalizeLabels(labels)
	a.Attributes = normalizeAttributes(kind, attributes)
	a.UpdatedAt = now.UTC()
	return a.Validate()
}

func normalizeProjectID(projectID *string) *string {
	if projectID == nil {
		return nil
	}
	v := strings.TrimSpace(*projectID)
	if v == "" {
		return nil
	}
	return &v
}

func normalizeContextIDs(contextIDs []string) []string {
	if len(contextIDs) == 0 {
		return []string{}
	}
	out := make([]string, 0, len(contextIDs))
	seen := make(map[string]struct{}, len(contextIDs))
	for _, id := range contextIDs {
		v := strings.TrimSpace(id)
		if v == "" {
			continue
		}
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}
	return out
}

func normalizeLabels(labels []Label) []Label {
	if len(labels) == 0 {
		return []Label{}
	}
	out := make([]Label, 0, len(labels))
	for _, label := range labels {
		out = append(out, Label{Name: strings.TrimSpace(label.Name)})
	}
	return out
}

func normalizeAttributes(kind Kind, attributes Attributes) Attributes {
	var out Attributes
	switch kind {
	case KindTask:
		if attributes.Task == nil {
			out.Task = &TaskAttributes{}
		} else {
			task := *attributes.Task
			if task.ExpectedAt != nil {
				t := task.ExpectedAt.UTC()
				task.ExpectedAt = &t
			}
			out.Task = &task
		}
	case KindWaiting:
		if attributes.Waiting == nil {
			out.Waiting = &WaitingAttributes{}
		} else {
			waiting := *attributes.Waiting
			waiting.Delegatee = strings.TrimSpace(waiting.Delegatee)
			if waiting.DueAt != nil {
				t := waiting.DueAt.UTC()
				waiting.DueAt = &t
			}
			out.Waiting = &waiting
		}
	case KindScheduled:
		if attributes.Scheduled == nil {
			out.Scheduled = &ScheduledAttributes{}
		} else {
			scheduled := *attributes.Scheduled
			if scheduled.StartAt != nil {
				t := scheduled.StartAt.UTC()
				scheduled.StartAt = &t
			}
			if scheduled.EndAt != nil {
				t := scheduled.EndAt.UTC()
				scheduled.EndAt = &t
			}
			out.Scheduled = &scheduled
		}
	}
	return out
}

func validateContextIDs(contextIDs []string) error {
	for i, id := range contextIDs {
		if strings.TrimSpace(id) == "" {
			return &ValidationError{Problems: []string{"context[" + itoaAction(i) + "]: Required value"}}
		}
	}
	return nil
}

func validateLabels(labels []Label) error {
	for i, label := range labels {
		if strings.TrimSpace(label.Name) == "" {
			return &ValidationError{Problems: []string{"labels[" + itoaAction(i) + "].name: Required value"}}
		}
	}
	return nil
}

func validateAttributes(kind Kind, attributes Attributes) error {
	switch kind {
	case KindTask:
		if attributes.Task == nil {
			return &ValidationError{Problems: []string{"attributes.task: Required value"}}
		}
	case KindWaiting:
		if attributes.Waiting == nil {
			return &ValidationError{Problems: []string{"attributes.waiting: Required value"}}
		}
		if strings.TrimSpace(attributes.Waiting.Delegatee) == "" {
			return &ValidationError{Problems: []string{"attributes.waiting.delegatee: Required value"}}
		}
		if attributes.Waiting.DueAt == nil || attributes.Waiting.DueAt.IsZero() {
			return &ValidationError{Problems: []string{"attributes.waiting.due_at: Required value"}}
		}
	case KindScheduled:
		if attributes.Scheduled == nil {
			return &ValidationError{Problems: []string{"attributes.scheduled: Required value"}}
		}
		if attributes.Scheduled.StartAt == nil || attributes.Scheduled.StartAt.IsZero() {
			return &ValidationError{Problems: []string{"attributes.scheduled.start_at: Required value"}}
		}
		if attributes.Scheduled.EndAt == nil || attributes.Scheduled.EndAt.IsZero() {
			return &ValidationError{Problems: []string{"attributes.scheduled.end_at: Required value"}}
		}
		if attributes.Scheduled.EndAt.Before(*attributes.Scheduled.StartAt) {
			return &ValidationError{Problems: []string{"attributes.scheduled.end_at: Unsupported value"}}
		}
	default:
		return InvalidKind(string(kind))
	}
	return nil
}

func requiredActionField(value, field string) string {
	if strings.TrimSpace(value) == "" {
		return field
	}
	return ""
}

func itoaAction(v int) string {
	if v == 0 {
		return "0"
	}
	buf := [20]byte{}
	pos := len(buf)
	for v > 0 {
		pos--
		buf[pos] = byte('0' + v%10)
		v /= 10
	}
	return string(buf[pos:])
}
