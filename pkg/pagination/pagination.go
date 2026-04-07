package pagination

import (
	"fmt"
)

const (
	DefaultLimit = 20
	MaxLimit     = 100
)

type PageQuery struct {
	Cursor string
	Limit  int
}

func (q *PageQuery) Normalize() {
	if q.Limit <= 0 {
		q.Limit = DefaultLimit
	}
	if q.Limit > MaxLimit {
		q.Limit = MaxLimit
	}
}

type PageResult[T any] struct {
	Items      []T    `json:"items"`
	NextCursor string `json:"next_cursor"`
	HasMore    bool   `json:"has_more"`
	Limit      int    `json:"limit"`
}

func NewPageResult[T any](items []T, limit int) PageResult[T] {
	hasMore := false
	nextCursor := ""

	if len(items) > limit {
		hasMore = true
		items = items[:limit]
	}

	if len(items) > 0 {
		if lastItem, ok := any(items[len(items)-1]).(interface{ ID() string }); ok {
			nextCursor = lastItem.ID()
		}
	}

	return PageResult[T]{
		Items:      items,
		NextCursor: nextCursor,
		HasMore:    hasMore,
		Limit:      limit,
	}
}

func NewPageResultWithCursor[T any](items []T, nextCursor string, hasMore bool, limit int) PageResult[T] {
	return PageResult[T]{
		Items:      items,
		NextCursor: nextCursor,
		HasMore:    hasMore,
		Limit:      limit,
	}
}

func ParseCursor(cursor string) (string, error) {
	if cursor == "" {
		return "", nil
	}
	if len(cursor) != 36 {
		return "", fmt.Errorf("invalid cursor format: must be UUID v7")
	}
	return cursor, nil
}
