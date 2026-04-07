# ADR-001: Hexagonal Architecture

## Статус: Accepted

## Контекст

Потрібна архітектура, що забезпечує:
- Чітке розділення бізнес-логіки та інфраструктури
- Незалежність від фреймворків, БД, зовнішніх сервісів
- Легкість тестування domain шару без інфраструктури
- Можливість заміни технологій без зміни бізнес-логіки

## Рішення

Використовувати Hexagonal Architecture (Ports & Adapters) з DDD:
- **Domain** — aggregates, value objects, domain events, repository interfaces
- **Application** — command/query handlers (use cases)
- **Infrastructure** — реалізації репозиторіїв, сервісів
- **Ports** — HTTP handlers як entry points

Dependency rule: залежності йдуть всередину, domain не знає нічого зовнішнього.

## Наслідки

### Позитивні
- Domain шар повністю ізольований і легко тестується
- Можна замінити БД/фреймворк без зміни бізнес-логіки
- Чіткі межі між шарами

### Негативні
- Більше boilerplate коду
- Потрібна дисципліна для дотримання правил
- Може бути overkill для простих CRUD додатків
