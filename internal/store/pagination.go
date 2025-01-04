package store

type PaginatedQuery struct {
	Limit  int
	Offset int
	Sort   string
}
