package domain

type ProjectListQuery = ListQuery

const (
	ProjectSortByCreatedAt SortBy = "CreatedAt"
	ProjectSortByUpdatedAt SortBy = "UpdatedAt"
	ProjectSortByTitle     SortBy = "Title"
)
