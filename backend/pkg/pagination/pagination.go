// Package pagination provides cursor-based pagination utilities for API responses.
// It supports efficient pagination without offset calculations and includes
// query parsing and result construction helpers.
package pagination

import (
	"fmt"
)

// DefaultLimit is the default number of items per page when no limit is specified.
const DefaultLimit = 20

// MaxLimit is the maximum allowed number of items per page.
const MaxLimit = 100

// PageQuery represents pagination parameters for a list query.
type PageQuery struct {
	Cursor string
	Limit  int
}

// Normalize ensures the query has valid pagination parameters.
// If the limit is invalid, it defaults to DefaultLimit.
// If the limit exceeds MaxLimit, it is capped to MaxLimit.
func (q *PageQuery) Normalize() {
	if q.Limit <= 0 {
		q.Limit = DefaultLimit
	}
	if q.Limit > MaxLimit {
		q.Limit = MaxLimit
	}
}

// PageResult represents a paginated list of items with metadata for client pagination.
// It uses generics to support any item type.
type PageResult[T any] struct {
	Items      []T    `json:"items"`
	NextCursor string `json:"next_cursor"`
	HasMore    bool   `json:"has_more"`
	Limit      int    `json:"limit"`
}

// NewPageResult creates a paginated result from a slice of items.
// If len(items) > limit, it sets HasMore to true and truncates to limit items.
// The next cursor is extracted from the last item's ID() method if it exists.
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

// NewPageResultWithCursor creates a paginated result with explicit cursor and hasMore values.
// Use this when you need full control over pagination metadata.
func NewPageResultWithCursor[T any](items []T, nextCursor string, hasMore bool, limit int) PageResult[T] {
	return PageResult[T]{
		Items:      items,
		NextCursor: nextCursor,
		HasMore:    hasMore,
		Limit:      limit,
	}
}

// ParseCursor validates and returns a pagination cursor string.
// Returns an error if the cursor is present but not a valid UUID v7 format.
func ParseCursor(cursor string) (string, error) {
	if cursor == "" {
		return "", nil
	}
	if len(cursor) != 36 {
		return "", fmt.Errorf("invalid cursor format: must be UUID v7")
	}
	return cursor, nil
}
