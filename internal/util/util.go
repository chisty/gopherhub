package util

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/chisty/gopherhub/internal/store"
)

func ParsePagination(r *http.Request) (store.PaginatedQuery, error) {
	qs := r.URL.Query()

	limit, err := strconv.Atoi(qs.Get("limit"))
	if err != nil {
		return store.PaginatedQuery{}, fmt.Errorf("failed to parse limit: %w", err)
	}

	offset, err := strconv.Atoi(qs.Get("offset"))
	if err != nil {
		return store.PaginatedQuery{}, fmt.Errorf("failed to parse offset: %w", err)
	}

	sort := qs.Get("sort")
	if sort == "" {
		sort = "desc"
	}

	pq := store.PaginatedQuery{
		Limit:  limit,
		Offset: offset,
		Sort:   sort,
	}

	return pq, nil
}
