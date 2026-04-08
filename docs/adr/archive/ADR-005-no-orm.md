# ADR-005: No ORM

## Status: **SUPERSEDED by ADR-016 and ADR-017**

## Context

Need a way to interact with DB:
- Control over SQL queries
- Avoid N+1 problem
- Performance transparency
- Idiomatic Go

## Decision

**SUPERSEDED**: This ADR has been replaced by:
- **ADR-016**: Use PostgreSQL with Pure pgx Driver (defines pgx v5 + pgxpool as standard)
- **ADR-017**: scany v2 + squirrel Repository Standard (eliminates boilerplate while maintaining pgx control)

The original decision to use `sqlx` has been replaced with **pure pgx/scany v2/squirrel** stack:

**New Stack (ADR-016 + ADR-017):**
- ✅ pgx/v5 + pgxpool - Pure PostgreSQL driver, zero reflection
- ✅ scany v2 - Minimal reflection for struct scanning
- ✅ squirrel - Type-safe dynamic query builder
- ❌ sqlx - Replaced (reflection overhead, not pgx-native)

## Migration Path

1. **From sqlx to pgx** (ADR-016): Replace all `sqlx.DB` with `pgxpool.Pool`
2. **From manual Scan to scany** (ADR-017): Replace manual variable scanning with DTO structs
3. **From stringconcat to squirrel** (ADR-017): Replace `fmt.Sprintf` queries with type-safe builders

## References

- ADR-016: Use PostgreSQL with Pure pgx Driver
- ADR-017: scany v2 + squirrel Repository Standard

## Date

2026-04-08 (Original: 2025-03-15)