# ADR-009: Mandatory Swagger Annotations for HTTP Handlers

## Status

Accepted

## Context

API documentation is critical for:
1. **Developer experience** — нові розробники швидко розуміють API
2. **Client integration** — сторонні сервіси можуть інтегруватися без додаткової документації
3. **Testing** — Swagger UI дозволяє тестувати endpoints вручну
4. **Consistency** — єдиний формат для всіх endpoints

Проблема: відсутність swagger анотацій призводить до:
- Неповної документації API
- Складнощів у підтримці (документація відстає від коду)
- Відсутності specs для external integrators

## Decision

**Всі HTTP handlers обов'язково повинні мати swagger анотації.**

### Вимоги

#### Обов'язкові поля для кожного handler:

```go
// ListRecords godoc
// @Summary List audit records
// @Description Returns paginated list of audit records with optional filters
// @Tags audit
// @Produce json
// @Security BearerAuth
// @Param actor_id query string false "Filter by actor ID"
// @Param resource query string false "Filter by resource"
// @Param action query string false "Filter by action"
// @Success 200 {object} query.ListRecordsResult
// @Failure 401 {object} apierror.APIError
// @Failure 403 {object} apierror.APIError
// @Failure 500 {object} apierror.APIError
// @Router /api/v1/audit/records [get]
func (h *Handler) ListRecords(c *gin.Context) {
    // ...
}
```

#### Чек-лист для кожної анотації:

- [ ] `@Summary` — короткий опис (1-2 sentences)
- [ ]`@Description` — детальний опис якщо потрібно
- [ ] `@Tags` — група endpoint'ів (auth, users, roles, audit, status)
- [ ] `@Produce` — content-type (json)
- [ ] `@Security` — тип авторизації (BearerAuth, SessionAuth)
- [ ] `@Param` — параметри (path, query, body)
  - Name, type, required, description
- [ ] `@Success` — успішна відповідь
  - Status code, response type, description
- [ ] `@Failure` — помилки
  - Status codes для різних помилок
  - Тип відповіді (apierror.APIError)
- [ ] `@Router` — маршрут і HTTP method

### Приклади для різних HTTP methods

#### GET (list)

```go
// ListUsers godoc
// @Summary List users
// @Description Returns paginated list of users
// @Tags users
// @Produce json
// @Security BearerAuth
// @Param cursor query string false "Pagination cursor"
// @Param limit query int false "Items per page"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} apierror.APIError
// @Failure 403 {object} apierror.APIError
// @Router /api/v1/users [get]
```

#### POST (create)

```go
// Register godoc
// @Summary Register a new user
// @Description Creates a new user account with email and password
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "Registration request"
// @Success 201 {object} map[string]string
// @Failure 422 {object} apierror.APIError
// @Failure 500 {object} apierror.APIError
// @Router /api/v1/auth/register [post]
```

#### GET (single resource)

```go
// GetUser godoc
// @Summary Get user by ID
// @Description Returns user details by ID
// @Tags users
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID"
// @Success 200 {object} query.UserDTO
// @Failure 401 {object} apierror.APIError
// @Failure 403 {object} apierror.APIError
// @Failure 404 {object} apierror.APIError
// @Router /api/v1/users/{id} [get]
```

### Tags

Використовувати наступні tags:
- `auth` — authentication operations (register, login, logout)
- `users` — user management operations
- `roles` — role management operations
- `audit` — audit log operations
- `status` — system status and health checks

### Security

Використовувати:
- `@Security BearerAuth` — для JWT token authentication
- `@Security SessionAuth` — для cookie-based session authentication

### Response Types

Для складних типів відповідей використовувати:
1. DTO structs з `json` tags: `{object} query.UserDTO`
2. Для map/simple types: `{object} map[string]interface{}`
3. Для помилок: `{object} apierror.APIError`

## Implementation

### Автоматична генерація

```bash
make swagger         # Згенерувати docs
make swagger-serve   # Згенерувати і запустити Swagger UI
```

Swagger UI доступний за адресою: http://localhost:8080/swagger/index.html

### CI/CD Integration

```yaml
# GitHub Actions
- name: Generate Swagger
  run: make swagger
  
- name: Check Swagger Changes
  run: |
    if [[ -n $(git status --porcelain docs/api/) ]]; then
      echo "Swagger docs are out of date. Run: make swagger"
      exit 1
    fi
```

## Consequences

### Positive

- ✓ Автоматично генерована документація
- ✓ Swagger UI для тестування API
- ✓ Documentation as code (не відстає від коду)
- ✓ Можна генерувати client SDKs
- ✓ Легко інтегрувати з Postman, Insomnia

### Negative

- Потрібно підтримувати анотації при зміні endpoints
- Swagger generation додає час до build process

### Neutral

- Response types мають бути serializable to JSON
- Всі DTOs мають мати `json` tags

## Checklist for Code Review

При review PR з новим handler'ом перевірити:

- [ ] Всі обов'язкові поля анотації присутні
- [ ] Tags відповідають bounded context
- [ ] Security type правильний (BearerAuth / SessionAuth)
- [ ] Всі параметри задокументовані (@Param)
- [ ] Усі response codes покриті (@Success, @Failure)
- [ ] Router path与方法 HTTP співпадають
- [ ] `make swagger` виконано без помилок
- [ ] Swagger changes закомічені

## Examples from Codebase

### Audit handler

```go
// internal/audit/ports/http/handler.go
// ListRecords godoc
// @Summary List audit records
// @Description Returns paginated list of audit records with optional filters
// @Tags audit
// @Produce json
// @Security BearerAuth
// @Param actor_id query string false "Filter by actor ID"
// @Param resource query string false "Filter by resource (user, role, auth)"
// @Param action query string false "Filter by action"
// @Param date_from query string false "Filter from date (RFC3339)"
// @Param date_to query string false "Filter to date (RFC3339)"
// @Param cursor query string false "Pagination cursor"
// @Param limit query int false "Items per page"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} apierror.APIError
// @Failure 403 {object} apierror.APIError
// @Failure 500 {object} apierror.APIError
// @Router /api/v1/audit/records [get]
```

## References

- [Swag Documentation](https://github.com/swaggo/swag)
- [OpenAPI Specification](https://swagger.io/specification/)
- [Swagger UI](https://swagger.io/tools/swagger-ui/)