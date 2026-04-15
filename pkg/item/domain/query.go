package domain

import "strings"

type SortDir string

const (
	SortAsc  SortDir = "Asc"
	SortDesc SortDir = "Desc"
)

func (d SortDir) valid() bool { return d == SortAsc || d == SortDesc }

type SortBy string

const (
	ItemSortByCreatedAt  SortBy = "CreatedAt"
	ItemSortByUpdatedAt  SortBy = "UpdatedAt"
	ItemSortByTitle      SortBy = "Title"
	ItemSortByDueAt      SortBy = "DueAt"
	ItemSortByExpectedAt SortBy = "ExpectedAt"
	ItemSortByPriority   SortBy = "Priority"
)

func (s SortBy) valid() bool {
	switch s {
	case ItemSortByCreatedAt, ItemSortByUpdatedAt, ItemSortByTitle, ItemSortByDueAt, ItemSortByExpectedAt, ItemSortByPriority:
		return true
	default:
		return false
	}
}

type Sort struct {
	By  SortBy
	Dir SortDir
}

type ListQuery struct {
	Bucket     *Bucket
	Kind       *Kind
	ProjectID  string
	TaskStatus string
	Search     string
	Sorts      []Sort
	Limit      int
	Offset     int
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
		q.Sorts = []Sort{{By: ItemSortByCreatedAt, Dir: SortDesc}}
	}
}

func (q *ListQuery) Validate() error {
	q.normalize()
	if q == nil {
		return ErrInvalidArg
	}
	if q.Bucket != nil && !q.Bucket.Valid() {
		return InvalidBucket(string(*q.Bucket))
	}
	if q.Kind != nil && !q.Kind.Valid() {
		return InvalidKind(string(*q.Kind))
	}
	for _, s := range q.Sorts {
		if !s.By.valid() {
			return InvalidSortBy(string(s.By))
		}
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
