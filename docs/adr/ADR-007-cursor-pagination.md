# ADR-007: Cursor-based Pagination

## Status: Accepted

## Context

Web interface needs efficient pagination for lists (users, roles, etc.).

Offset-based pagination (`LIMIT/OFFSET`) has critical disadvantages:
1. **Performance degradation** — OFFSET scans and discards rows, performance drops linearly
2. **Inconsistent results** — when adding/deleting records between requests data "shifts"
3. **No bookmarking** — impossible to return to the same position after changes

## Decision

Cursor-based pagination with UUID v7 as cursor.

### Why UUID v7 works perfectly

UUID v7 contains timestamp in high bits → lexicographically ordered → ideal cursor.

### API Design

**Request:**
```
GET /api/v1/users?limit=20&cursor=019d65d6-de90-7200-b1cf-4f8745597e0a
```

**Response:**
```json
{
  "items": [...],
  "pagination": {
    "next_cursor": "019d65d6-de98-7e00-b590-2d70f5506278",
    "has_more": true,
    "limit": 20
  }
}
```

### Conventions

| Parameter | Default | Max | Description |
|----------|---------|-----|-------------|
| `limit` | 20 | 100 | Number of records |
| `cursor` | — | — | UUID v7 of last element |
| `sort` | `desc` | — | Direction (desc only) |

### Repository Pattern

```go
type PageResult[T any] struct {
    Items      []T    `json:"items"`
    NextCursor string `json:"next_cursor"`
    HasMore    bool   `json:"has_more"`
    Limit      int    `json:"limit"`
}

type PageQuery struct {
    Cursor string
    Limit  int
}
```

SQL pattern:
```sql
SELECT * FROM users
WHERE id < ?  -- cursor-based seek
ORDER BY id DESC
LIMIT ? + 1   -- fetch one extra to detect has_more
```

### Benefits

| Characteristic | Offset | Cursor (UUID v7) |
|---------------|--------|------------------|
| Performance | O(n) degradation | O(log n) — index seek |
| Consistency | Shifts on changes | Stable |
| Bookmarking | Impossible | Possible (cursor = URL) |
| Deep pages | Slow | Same speed |

## Consequences

**Positive:**
- Stable performance at any depth
- Impossible to "skip" a record during concurrent changes
- Cursor = UUID v7 = timestamp → can filter by time
- Stateless — doesn't require server state

**Negative:**
- Cannot jump to a specific page (only forward/backward)
- Need to store cursor on client
- Total count unavailable (no COUNT)