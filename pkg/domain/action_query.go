package domain

import "strings"

type SortDir string

const (
	SortAsc  SortDir = "Asc"
	SortDesc SortDir = "Desc"
)

func (d SortDir) valid() bool { return d == SortAsc || d == SortDesc }

type SortBy string

type Sort struct {
	By  SortBy
	Dir SortDir
}

type ListQuery struct {
	Search     string
	Kind       *Kind
	ProjectID  *string
	Status     *Status
	ContextIDs []string
	Tags       []string
	Sorts      []Sort
	Limit      int
	Offset     int
}

type ActionListQuery = ListQuery

const (
	ActionSortByCreatedAt SortBy = "CreatedAt"
	ActionSortByUpdatedAt SortBy = "UpdatedAt"
	ActionSortByTitle     SortBy = "Title"
)

func (q *ListQuery) normalize() {
	if q == nil {
		return
	}
	if q.Limit < 0 {
		q.Limit = 0
	}
	if q.Offset < 0 {
		q.Offset = 0
	}
	if len(q.Sorts) == 0 {
		q.Sorts = []Sort{{By: ActionSortByCreatedAt, Dir: SortDesc}}
	}
}

func (q *ListQuery) Validate() error {
	q.normalize()
	if q == nil {
		return ErrInvalidArg
	}
	if q.Kind != nil && !q.Kind.Valid() {
		return InvalidKind(string(*q.Kind))
	}
	if q.Status != nil && !q.Status.Valid() {
		return InvalidStatus(string(*q.Status))
	}
	for _, s := range q.Sorts {
		if !s.Dir.valid() {
			return InvalidSortDir(string(s.Dir))
		}
	}
	return nil
}

func ParseSortDir(v string) (SortDir, bool) {
	switch strings.ToLower(v) {
	case "asc":
		return SortAsc, true
	case "desc":
		return SortDesc, true
	default:
		return "", false
	}
}
