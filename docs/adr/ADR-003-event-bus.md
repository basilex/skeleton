# ADR-003: Event Bus Strategy

## Status: Accepted

## Context

A communication mechanism between bounded contexts is needed:
- Loose coupling between contexts
- Different requirements for dev/test vs prod
- Simple development and testing

## Decision

Two implementations of the same `eventbus.Bus` interface:
- **In-Memory** — synchronous, for dev/test
- **Redis Pub/Sub** — asynchronous, for prod

Selection via env: `APP_ENV=prod` → Redis, otherwise → In-Memory.

## Consequences

### Positive
- Simple testing without Redis
- One interface — easy replacement
- Panic recovery in each handler

### Negative
- Redis Pub/Sub — at-most-once delivery (messages can be lost)
- In-memory — doesn't work with multiple instances