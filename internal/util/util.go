package util

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/chisty/gopherhub/internal/store"
)

func ParsePagination(r *http.Request) (store.PaginatedFeedQuery, error) {
	qs := r.URL.Query()

	limit, err := strconv.Atoi(qs.Get("limit"))
	if err != nil {
		return store.PaginatedFeedQuery{}, fmt.Errorf("failed to parse limit: %w", err)
	}

	offset, err := strconv.Atoi(qs.Get("offset"))
	if err != nil {
		return store.PaginatedFeedQuery{}, fmt.Errorf("failed to parse offset: %w", err)
	}

	sort := qs.Get("sort")
	if sort == "" {
		sort = "desc"
	}

	userID, err := strconv.Atoi(qs.Get("userID"))
	if err != nil {
		return store.PaginatedFeedQuery{}, fmt.Errorf("failed to parse userID: %w", err)
	}

	pq := store.PaginatedFeedQuery{
		Limit:  limit,
		Offset: offset,
		Sort:   sort,
		UserID: int64(userID),
	}

	tags := qs.Get("tags")
	if tags != "" {
		pq.Tags = strings.Split(tags, ",")
	}

	search := qs.Get("search")
	if search != "" {
		pq.Search = search
	}

	since := qs.Get("since")
	if since != "" {
		pq.Since = parseTime(since)
	}

	until := qs.Get("until")
	if until != "" {
		pq.Until = parseTime(until)
	}

	return pq, nil
}

func parseTime(s string) string {
	t, err := time.Parse(time.DateTime, s)
	if err != nil {
		return ""
	}

	return t.Format(time.DateTime)
}
