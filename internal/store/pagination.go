package store

type PaginatedFeedQuery struct {
	Limit  int
	Offset int
	Sort   string
	Tags   []string
	Search string
	UserID int64
	Since  string
	Until  string
}
