# ADR-006: UUID v7 for Primary Keys

## Status: Accepted

## Context

UUID v4 (random) has critical disadvantages as a primary key in DB:

1. **Fragmentation** — random insert order breaks B-tree indexes, causing page splits
2. **Poor locality** — records adjacent in time are scattered across the entire disk
3. **Cache inefficiency** — each new insert likely hits a new page
4. **Write amplification** — page splits generate unnecessary I/O operations

For SQLite WAL this is especially critical as tables grow.

## Decision

Custom UUID v7 implementation (`pkg/uuid/uuid.go`) — zero external dependencies.

UUID v7 structure:
```
 0                   1                   2                   3
 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|                           unix_ts_ms                          |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|          unix_ts_ms           |  ver  |       rand_a          |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|var|                        rand_b                             |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|                            rand_b                             |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
```

- **48 bits** — Unix timestamp in milliseconds (time-ordered)
- **4 bits** — version (0x7)
- **12 bits** — random (uniqueness within same ms)
- **2 bits** — variant (0b10)
- **62 bits** — random (collision resistance)

### UUID v7 Benefits

| Characteristic | UUID v4 | UUID v7 |
|---------------|---------|---------|
| Order | Random | Time-based |
| B-tree fragmentation | High | Minimal |
| Write amplification | High | Low |
| Cache locality | Poor | Good |
| Sortable | No | Yes |
| Timestamp extraction | No | Yes |

### Implementation

- `NewV7()` — generate new UUID
- `Parse(s)` / `MustParse(s)` — parse from string
- `String()` — format `xxxxxxxx-xxxx-7xxx-8xxx-xxxxxxxxxxxx`
- `Timestamp()` — extract time from UUID
- `Version()` — returns 7
- `MarshalText()` / `UnmarshalText()` — serialization for JSON/DB

### Usage in Project

All ID types use UUID v7:
```go
func NewUserID() UserID {
    return UserID(uuid.NewV7().String())
}
```

## Consequences

### Positive
- Time-ordered PK — optimal B-tree index performance
- Built-in timestamp — can sort/filter without JOIN
- Zero dependencies — no `google/uuid`
- Thread-safe — `crypto/rand` with mutex for rand_a
- RFC 9562 compliant — compatible with other UUID v7 implementations

### Negative
- Custom implementation — need to maintain and test
- Less battle-tested than `google/uuid`
- 48-bit timestamp — works until ~10889 year

### Migration
If project started with UUID v4 — migration not needed, new records automatically get v7.