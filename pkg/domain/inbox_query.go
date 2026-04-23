package domain

type InboxListQuery = ListQuery

const (
	InboxSortByCreatedAt SortBy = "CreatedAt"
	InboxSortByUpdatedAt SortBy = "UpdatedAt"
	InboxSortByTitle     SortBy = "Title"
)
