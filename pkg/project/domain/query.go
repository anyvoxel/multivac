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

// SortBy represents sortable fields for Project.
type SortBy string

const (
	// ProjectSortByCreatedAt sorts by created time.
	ProjectSortByCreatedAt SortBy = "CreatedAt"
	// ProjectSortByUpdatedAt sorts by updated time.
	ProjectSortByUpdatedAt SortBy = "UpdatedAt"
	// ProjectSortByTitle sorts by project title.
	ProjectSortByTitle SortBy = "Title"
)

func (s SortBy) valid() bool {
	switch s {
	case ProjectSortByCreatedAt, ProjectSortByUpdatedAt, ProjectSortByTitle:
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

// ListQuery is used to query projects.
type ListQuery struct {
	Status *Status
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
		q.Sorts = []Sort{{By: ProjectSortByCreatedAt, Dir: SortDesc}}
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
