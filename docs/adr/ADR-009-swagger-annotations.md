# ADR-009: Mandatory Swagger Annotations for HTTP Handlers

## Status

Accepted

## Context

API documentation is critical for:
1. **Developer experience** — new developers quickly understand API
2. **Client integration** — third-party services can integrate without additional documentation
3. **Testing** — Swagger UI allows testing endpoints manually
4. **Consistency** — unified format for all endpoints

Problem: lack of swagger annotations leads to:
- Incomplete API documentation
- Difficulties in maintenance (documentation lags behind code)
- Lack of specs for external integrators

## Decision

**All HTTP handlers must have swagger annotations.**

##Requirements

### Mandatory fields for each handler:

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

#### Checklist for each annotation:

- [ ] `@Summary` — short description (1-2 sentences)
- [ ] `@Description` — detailed description if needed
- [ ] `@Tags` — endpoint group (auth, users, roles, audit, status)
- [ ] `@Produce` — content-type (json)
- [ ] `@Security` — authorization type (BearerAuth, SessionAuth)
- [ ] `@Param` — parameters (path, query, body)
  - Name, type, required, description
- [ ] `@Success` — successful response
  - Status code, response type, description
- [ ] `@Failure` — errors
  - Status codes for different errors
  - Response type (apierror.APIError)
- [ ] `@Router` — route and HTTP method

### Examples for different HTTP methods

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

Use the following tags:
- `auth` — authentication operations (register, login, logout)
- `users` — user management operations
- `roles` — role management operations
- `audit` — audit log operations
- `status` — system status and health checks

### Security

Use:
- `@Security BearerAuth` — for JWT token authentication
- `@Security SessionAuth` — for cookie-based session authentication

### Response Types

For complex response types use:
1. DTO structs with `json` tags: `{object} query.UserDTO`
2. For map/simple types: `{object} map[string]interface{}`
3. For errors: `{object} apierror.APIError`

## Implementation

### Automatic generation

```bash
make swagger         # Generate docs
make swagger-serve   # Generate and launch Swagger UI
```

Swagger UI is available at: http://localhost:8080/swagger/index.html

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

- ✓ Automatically generated documentation
- ✓ Swagger UI for API testing
- ✓ Documentation as code (doesn't lag behind code)
- ✓ Can generate client SDKs
- ✓ Easy integration with Postman, Insomnia

### Negative

- Need to maintain annotations when changing endpoints
- Swagger generation adds time to build process

### Neutral

- Response types must be serializable to JSON
- All DTOs must have `json` tags

## Checklist for Code Review

When reviewing PR with new handler check:

- [ ] All mandatory annotation fields present
- [ ] Tags match bounded context
- [ ] Security type correct (BearerAuth / SessionAuth)
- [ ] All parameters documented (@Param)
- [ ] All response codes covered (@Success, @Failure)
- [ ] Router path and HTTP method match
- [ ] `make swagger` executed without errors
- [ ] Swagger changes committed

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