# ADR-004: RBAC Model

## Статус: Accepted

## Контекст

Потрібна система контролю доступу:
- Різні рівні доступу для різних користувачів
- Гнучкість у призначенні прав
- Проста інтеграція з HTTP middleware

## Рішення

Role-Based Access Control (RBAC):
- User → UserRole → Role → RolePermission → Permission
- Permission format: `resource:action`
- Wildcard підтримка: `*:*`, `resource:*`
- Middleware chain: Authenticate → RequirePermission

## Наслідки

### Позитивні
- Стандартний, зрозумілий підхід
- Легко додавати нові permissions
- Wildcard зменшує дублювання правил

### Негативні
- Не підтримує attribute-based rules (ABAC)
- Для fine-grained access потрібне розширення
