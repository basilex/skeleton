# Makefile Quick Reference

## Найбільш вживані команди

### Docker - Основні операції

```bash
# Запустити контейнери в background
make docker-up

# Зупинити і видалити контейнери (зберігає дані)
make docker-down

# Повний reset з видаленням даних ⚠️ DESTRUCTIVE
make docker-drop

# Статус контейнерів
make docker-status

# Перегляд логів
make docker-logs          # Всі логи
make docker-logs-app      # Тільки API
make docker-logs-db       # Тільки PostgreSQL

# Shell в контейнері
make docker-shell         # В API контейнері
make docker-shell-root    # З root правами

# Повний reset з нуля (видалення volumes + rebuild + migrate)
make fresh-start
```

### База даних - Основні операції

```bash
# Підключення до бази
make psql                # Interactive psql

# Перегляд таблиць та статистики
make db-tables           # Список таблиць
make db-stats            # Розміри та статистика
make db-migrations       # Історія міграцій
make db-connections      # Активні з'єднання

# Міграції
make migrate-up          # Застосувати всі міграції
make migrate-down        # Відкотити останню міграцію
make migrate-status      # Статус міграцій
make migrate-reset       # Повний reset (⚠️ DESTRUCTIVE)

# Backup & Restore
make db-backup                           # Створити backup
make db-restore BACKUP_FILE=backups/... # Відновити з backup

# SQL команди
make db-sql SQL='SELECT * FROM users LIMIT 5;'

# Небезпечні операції ⚠️
make db-truncate         # Очистити всі таблиці (зберегти схему)
make db-drop-tables      # Видалити всі таблиці
```

### Розробка

```bash
# Швидкий старт
make dev                 # migrate-up + seed + run

# Початковий setup
make setup               # keys + migrate-up + seed

# Тестування
make test                # Всі тести
make test-unit           # Тільки unit тести
make test-integration    # Integration тести (потрібен Docker)

# Якість коду
make lint                # golangci-lint
make fmt                 # Форматування коду

# Документація
make swagger             # Генерувати Swagger
make swagger-serve       # Запустити Swagger UI
```

### Моніторинг та статус

```bash
# Загальний статус системи
make status

# Health endpoints
make health

# Database performance
make db-slow-queries     # Повільні запити
make db-index-usage      # Використання індексів
make db-cache-ratio      # Cache hit ratio
```

## Приклади використання

### Початок роботи з проектом

```bash
# 1. Перший запуск (повний setup)
make fresh-start

# або покроково:
make docker-up
make migrate-up
make seed
```

### Розробка функціоналу

```bash
# Створити нову міграцію
# редактувати файл migrations/XXX_description.up.sql
make migrate-up

# Перевірити що таблиці створені
make db-tables

# Заповнити тестовими даними
make seed

# Запустити додаток
make run
```

### Debugging

```bash
# Перевірити статус всього
make status

# Переглянути логи API
make docker-logs-app

# Підключитися до бази
make psql

# Перевірити повільні запити
make db-slow-queries

# Виконати SQL команду
make db-sql SQL='SELECT COUNT(*) FROM users;'
```

### Reset environment

```bash
# М'який reset (зберегти дані)
make migrate-reset      # Скинути міграції

# Повний reset (видалити дані) ⚠️
make docker-drop        # Зупинити і видалити volumes
make docker-up          # Запустити з нуля
make migrate-up         # Застосувати міграції
make seed               # Заповнити даними

# або однією командою:
make fresh-start
```

### CI/CD

```bash
# Перевірити код перед комітом
make ci                  # lint + test

# Запустити тести з coverage
make test-cover

# Бенчмарки
make bench
```

## Структура Makefile

```
Makefile
├── BUILD              # Компіляція, запуск
├── TESTING            # Тести (unit, integration)
├── CODE QUALITY       # Лінтери, форматування
├── DOCKER
│   ├── Container Management  # up, down, stop, start, restart
│   ├── Utility Commands      # shell, exec, logs
│   ├── Environment           # dev, staging, prod
│   └── Cleanup              # clean, drop, reset
├── DATABASE
│   ├── Connection & Queries # psql, tables, stats
│   ├── Performance          # slow-queries, index-usage
│   ├── Backup & Restore     # backup, restore, export
│   └── Advanced Operations  # truncate, drop-tables
├── MIGRATIONS         # up, down, status, reset
├── BENCHMARKS         # perf testing
├── DOCUMENTATION      # swagger
└── UTILITIES          # fmt, deps, health, status
```

## Troubleshooting

### Контейнери не запускаються

```bash
# Перевірити статус
make docker-status

# Зупинити все і запустити з нуля
make docker-drop
make docker-up
```

### Міграції не виконуються

```bash
# Перевірити статус міграцій
make migrate-status

# Перевірити чи контейні запущений
make docker-check

# Скинути базу і почати з нуля
make migrate-reset
```

### Помилки в логах

```bash
# Дивитися логи в реальному часі
make watch-logs

# Або конкретний сервіс
make docker-logs-app
make docker-logs-db
```

### База даних повільна

```bash
# Перевірити статистику
make db-stats

# Знайти повільні запити
make db-slow-queries

# Перевірити індекси
make db-index-usage

# Cache hit ratio (має бути >99%)
make db-cache-ratio
```

## Корисні поради

1. **Використовуйте `make help`** - показує всі доступні команди
2. **Перевіряйте статус перед змінами** - `make status`
3. **Створюйте backup перед небезпечними операціями** - `make db-backup`
4. **Використовуйте `make fresh-start` для повного reset** - безпечніше ніж вручну
5. **Моніторте production через `make db-stats` та `make db-slow-queries`**

## Небезпечні команди ⚠️

- `make docker-drop` - видаляє ВСІ дані
- `make migrate-reset` - скидає ВСІ міграції
- `make db-truncate` - очищає ВСІ таблиці
- `make db-drop-tables` - видаляє ВСІ таблиці
- `make fresh-start` - повний reset з нуля

Всі ці команди просять підтвердження перед виконанням.