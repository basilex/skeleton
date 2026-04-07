# ADR-003: Event Bus Strategy

## Статус: Accepted

## Контекст

Потрібен механізм комунікації між bounded contexts:
- Слабка зв'язаність між контекстами
- Різні вимоги для dev/test vs prod
- Проста розробка та тестування

## Рішення

Два implementations одного інтерфейсу `eventbus.Bus`:
- **In-Memory** — синхронний, для dev/test
- **Redis Pub/Sub** — асинхронний, для prod

Вибір через env: `APP_ENV=prod` → Redis, інакше → In-Memory.

## Наслідки

### Позитивні
- Просте тестування без Redis
- Один інтерфейс — легка заміна
- Panic recovery у кожному handler

### Негативні
- Redis Pub/Sub — at-most-once delivery (повідомлення можуть губитись)
- In-memory — не працює з кількома інстансами
