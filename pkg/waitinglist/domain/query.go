package domain

import "strings"

// SortDir represents the sort direction.
type SortDir string

const (
	// SortAsc sorts in ascending order.
	SortAsc SortDir = "Asc"
	// SortDesc sorts in descending order.
	SortDesc SortDir = "Desc"
)

func (d SortDir) valid() bool {
	return d == SortAsc || d == SortDesc
}

// SortBy represents sortable fields for Waiting List.
type SortBy string

const (
	WaitingListSortByCreatedAt  SortBy = "CreatedAt"
	WaitingListSortByUpdatedAt  SortBy = "UpdatedAt"
	WaitingListSortByName       SortBy = "Name"
	WaitingListSortByExpectedAt SortBy = "ExpectedAt"
)

func (s SortBy) valid() bool {
	switch s {
	case WaitingListSortByCreatedAt, WaitingListSortByUpdatedAt, WaitingListSortByName, WaitingListSortByExpectedAt:
		return true
	default:
		return false
	}
}

// Sort defines one sorting rule.
type Sort struct {
	By  SortBy
	Dir SortDir
}

// ListQuery is used to query waiting list items.
type ListQuery struct {
	Search string
	Sorts  []Sort
	Limit  int
	Offset int
}

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
		q.Sorts = []Sort{{By: WaitingListSortByCreatedAt, Dir: SortDesc}}
	}
}

// Validate validates and normalizes the query.
func (q *ListQuery) Validate() error {
	q.normalize()
	if q == nil {
		return ErrInvalidArg
	}
	for _, s := range q.Sorts {
		if !s.By.valid() {
			return invalidFieldValueError("sortBy", string(s.By))
		}
		if !s.Dir.valid() {
			return invalidFieldValueError("sortDir", string(s.Dir))
		}
	}
	return nil
}

// ParseSortDir parses sort direction.
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
