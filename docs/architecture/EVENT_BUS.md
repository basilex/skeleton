# Event Bus

## Коли використовувати події

**Події** — коли:
- Інший bounded context має реагувати на дію
- Потрібна слабка зв'язаність між компонентами
- Операція не вимагає синхронної відповіді

**Прямий виклик** — коли:
- Потрібен результат тут і зараз
- Обидва компоненти в одному bounded context
- Операція критична і не може бути відкладена

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
- Приклади: `identity.user_registered`, `identity.role_assigned`, `identity.role_revoked`
- Snake_case для event names

## Delivery Guarantees

| Implementation | Guarantee | Order | Use Case |
|---------------|-----------|-------|----------|
| In-Memory | At-least-once (sync) | Sequential | dev, test |
| Redis Pub/Sub | At-most-once (fire-and-forget) | Per-channel | prod |

## Приклад реєстрації хендлера

```go
// cmd/api/main.go
bus.Subscribe("identity.user_registered", func(ctx context.Context, e eventbus.Event) error {
    evt := e.(identity.UserRegistered)
    // handle event
    return nil
})
```

## Panic Recovery

Кожен handler захищений recover — panic не падає весь bus, тільки логується.
