# ADR-002: SQLite WAL

## Status: **OBSOLETE** (Superseded by [ADR-016: Use pgx with PostgreSQL](ADR-016-pgx-with-postgres.md))

This ADR described the original SQLite-based architecture. The project has since migrated to PostgreSQL. See ADR-016 for the current database architecture.

---

## Context

A lightweight, serverless database is needed for the skeleton project that:
- Doesn't require a separate server
- Supports transactions and foreign keys
- Has good performance for read-heavy workloads
- Is easily configured for dev/test/prod

## Decision

Use SQLite in WAL (Write-Ahead Logging) mode with the `modernc.org/sqlite` driver (pure Go, no CGO).

PRAGMA settings:
- `journal_mode=WAL` — concurrent reads during writes
- `synchronous=NORMAL` — balance between safety and speed
- `foreign_keys=ON` — data integrity
- `busy_timeout=5000` — wait when blocked

`SetMaxOpenConns(1)` — SQLite doesn't support concurrent writes.

## Consequences

### Positive
- Zero infrastructure — no separate database server
- Pure Go driver — easy compilation without CGO
- WAL allows concurrent reads

### Negative
- Limited write concurrency (1 connection)
- Doesn't scale horizontally
- For high-load, migration to PostgreSQL is needed