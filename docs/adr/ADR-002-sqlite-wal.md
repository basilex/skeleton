# ADR-002: SQLite WAL

## Статус: Accepted

## Контекст

Потрібна легка, serverless БД для skeleton проекту, яка:
- Не вимагає окремого сервера
- Підтримує транзакції та foreign keys
- Має хорошу продуктивність для read-heavy навантажень
- Легко налаштовується для dev/test/prod

## Рішення

Використовувати SQLite у WAL (Write-Ahead Logging) режимі з драйвером `modernc.org/sqlite` (pure Go, без CGO).

PRAGMA налаштування:
- `journal_mode=WAL` — конкурентні читання під час запису
- `synchronous=NORMAL` — баланс між безпекою та швидкістю
- `foreign_keys=ON` — цілісність даних
- `busy_timeout=5000` — чекати при блокуванні

`SetMaxOpenConns(1)` — SQLite не підтримує конкурентний запис.

## Наслідки

### Позитивні
- Zero infrastructure — немає окремого БД сервера
- Pure Go драйвер — легка компіляція без CGO
- WAL дозволяє concurrent reads

### Негативні
- Обмежена конкурентність запису (1 connection)
- Не масштабується горизонтально
- Для high-load потрібна міграція на PostgreSQL
