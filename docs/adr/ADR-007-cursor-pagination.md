# ADR-007: Cursor-based Pagination

## Статус: Accepted

## Контекст

Web-інтерфейс потребує ефективної пагінації для списків (users, roles тощо).

Offset-based пагінація (`LIMIT/OFFSET`) має критичні недоліки:
1. **Performance degradation** — OFFSET сканує і відкидає рядки, продуктивність падає лінійно
2. **Inconsistent results** — при додаванні/видаленні записів між запитами дані "зсуваються"
3. **No bookmarking** — неможливо повернутись до тієї ж позиції після змін

## Рішення

Cursor-based пагінація з UUID v7 як курсор.

### Чому UUID v7 працює ідеально

UUID v7 містить timestamp у старших бітах → лексикографічно впорядкований → ідеальний курсор.

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

### Конвенції

| Параметр | Default | Max | Description |
|----------|---------|-----|-------------|
| `limit` | 20 | 100 | Кількість записів |
| `cursor` | — | — | UUID v7 останнього елемента |
| `sort` | `desc` | — | Напрямок (тільки desc) |

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

### Переваги

| Характеристика | Offset | Cursor (UUID v7) |
|---------------|--------|------------------|
| Performance | O(n) деградація | O(log n) — index seek |
| Consistency | Зсуви при змінах | Стабільний |
| Bookmarking | Неможливо | Можливо (cursor = URL) |
| Deep pages | Повільні | Однакова швидкість |

### Наслідки

**Позитивні:**
- Стабільна продуктивність на будь-якій глибині
- Неможливо "пропустити" запис при одночасних змінах
- Cursor = UUID v7 = timestamp → можна фільтрувати за часом
- Stateless — не потребує серверного стану

**Негативні:**
- Не можна стрибнути на конкретну сторінку (тільки forward/backward)
- Потрібно зберігати cursor на клієнті
- Total count недоступний (не робимо COUNT)
