# ADR-001: Hexagonal Architecture

## Status: Accepted

## Context

An architecture is needed that provides:
- Clear separation of business logic and infrastructure
- Independence from frameworks, databases, and external services
- Easy testing of the domain layer without infrastructure
- Ability to replace technologies without changing business logic

## Decision

Use Hexagonal Architecture (Ports & Adapters) with DDD:
- **Domain** — aggregates, value objects, domain events, repository interfaces
- **Application** — command/query handlers (use cases)
- **Infrastructure** — repository implementations, services
- **Ports** — HTTP handlers as entry points

Dependency rule: dependencies go inward, domain knows nothing external.

## Consequences

### Positive
- Domain layer is completely isolated and easy to test
- Can replace database/framework without changing business logic
- Clear boundaries between layers

### Negative
- More boilerplate code
- Discipline required to follow the rules
- May be overkill for simple CRUD applications