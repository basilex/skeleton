# ADR-006: UUID v7 for Primary Keys

## Статус: Accepted

## Контекст

UUID v4 (random) має критичні недоліки як primary key у БД:

1. **Fragmentation** — випадковий порядок вставки розбиває B-tree індекси, викликаючи page splits
2. **Poor locality** — сусідні за часом записи розкидані по всьому диску
3. **Cache inefficiency** — кожна нова вставка ймовірно потрапляє на нову сторінку
4. **Write amplification** — page splits генерують зайві I/O операції

Для SQLite WAL це особливо критично при зростанні таблиць.

## Рішення

Власна реалізація UUID v7 (`pkg/uuid/uuid.go`) — zero external dependencies.

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

### Переваги UUID v7

| Характеристика | UUID v4 | UUID v7 |
|---------------|---------|---------|
| Порядок | Випадковий | Часовий |
| B-tree fragmentation | Висока | Мінімальна |
| Write amplification | Високий | Низький |
| Cache locality | Погана | Хороша |
| Sortable | Ні | Так |
| Timestamp extraction | Ні | Так |

### Реалізація

- `NewV7()` — генерація нового UUID
- `Parse(s)` / `MustParse(s)` — парсинг з рядка
- `String()` — формат `xxxxxxxx-xxxx-7xxx-8xxx-xxxxxxxxxxxx`
- `Timestamp()` — витягнення часу з UUID
- `Version()` — повертає 7
- `MarshalText()` / `UnmarshalText()` — серіалізація для JSON/DB

### Використання в проекті

Всі ID типи використовують UUID v7:
```go
func NewUserID() UserID {
    return UserID(uuid.NewV7().String())
}
```

## Наслідки

### Позитивні
- Часово-впорядковані PK — оптимальна продуктивність B-tree індексів
- Вбудований timestamp — можна сортувати/фільтрувати без JOIN
- Zero dependencies — немає `google/uuid`
- Thread-safe — `crypto/rand` з mutex для rand_a
- RFC 9562 compliant — сумісний з іншими UUID v7 реалізаціями

### Негативні
- Власна реалізація — потрібно підтримувати та тестувати
- Менше battle-tested ніж `google/uuid`
- 48-bit timestamp — працює до ~10889 року

### Міграція
Якщо проект починався з UUID v4 — міграція не потрібна, нові записи автоматично отримують v7.
