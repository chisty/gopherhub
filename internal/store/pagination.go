package store

type PaginatedFeedQuery struct {
	Limit  int
	Offset int
	Sort   string
	Tags   []string
	Search string
	Since  string
	Until  string
}
