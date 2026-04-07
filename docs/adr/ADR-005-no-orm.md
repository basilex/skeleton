# ADR-005: No ORM

## Status: Accepted

## Context

Need a way to interact with DB:
- Control over SQL queries
- Avoid N+1 problem
- Performance transparency
- Idiomatic Go

## Decision

Use `sqlx` instead of ORM (GORM etc.):
- Named parameters for readability
- Struct scanning for convenience
- Direct SQL — full control
- No string concatenation in SQL

## Consequences

### Positive
- Full control over queries
- No hidden queries / N+1
- Easier to optimize performance
- Fewer dependencies

### Negative
- More boilerplate for CRUD
- Need to manually write migrations
- More complex schema changes