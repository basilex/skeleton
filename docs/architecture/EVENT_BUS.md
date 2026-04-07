# Event Bus

## When to Use Events

**Events** — when:
- Another bounded context needs to react to an action
- Loose coupling between components is needed
- The operation doesn't require a synchronous response

**Direct Call** — when:
- A result is needed here and now
- Both components are in the same bounded context
- The operation is critical and cannot be deferred

## Interface

```go
type Bus interface {
    Publish(ctx context.Context, event Event) error
    Subscribe(eventName string, handler Handler)
}
```

## Envelope Format (Redis)

```json
{
  "event_name": "identity.user_registered",
  "occurred_at": "2025-04-07T10:00:00Z",
  "payload": { "user_id": "...", "email": "..." }
}
```

## Naming Conventions

- Format: `{context}.{entity}_{action}`
- Examples: 
  - `identity.user_registered` — new user registered
  - `identity.role_assigned` — role assigned to user
  - `identity.role_revoked` — role revoked from user
  - `identity.login` — user logged in
  - `identity.logout` — user logged out
- Snake_case for event names

## Delivery Guarantees

| Implementation | Guarantee | Order | Use Case |
|---------------|-----------|-------|----------|
| In-Memory | At-least-once (sync) | Sequential | dev, test |
| Redis Pub/Sub | At-most-once (fire-and-forget) | Per-channel | prod |

## Handler Registration Example

```go
// cmd/api/main.go
bus.Subscribe("identity.user_registered", func(ctx context.Context, e eventbus.Event) error {
    evt := e.(identity.UserRegistered)
    // handle event
    return nil
})
```

## Panic Recovery

Each handler is protected by recover — panic doesn't bring down the entire bus, it's only logged.