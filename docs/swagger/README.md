# Swagger API Documentation

This directory contains the OpenAPI/Swagger specification for the Skeleton Business Engine API.

## Files

- `swagger.json` - OpenAPI 2.0 specification (JSON format)
- `swagger.yaml` - OpenAPI 2.0 specification (YAML format)
- `index.html` - Swagger UI interface

## Viewing Documentation

### Via Docker Compose

```bash
make docker-up
open http://localhost:8081
```

### Via API Server

The API server serves Swagger UI at:

```bash
make run
open http://localhost:8080/swagger/index.html
```

## Updating Documentation

After modifying API handlers, regenerate the documentation:

```bash
# Install swag CLI (if not already installed)
go install github.com/swaggo/swag/cmd/swag@latest

# Generate documentation
swag init -g cmd/api/main.go -o docs/swagger

# Or use Make
make swagger-gen
```

## API Structure

The API is organized into the following bounded contexts:

### Core Services
- **auth** - Authentication and authorization
- **users** - User management
- **roles** - Role and permission management
- **audit** - Audit logging
- **status** - System health and build information

### Business Contexts
- **parties** - Customer, supplier, partner, employee management
- **contracts** - Contract lifecycle management
- **accounting** - Chart of accounts, transactions
- **ordering** - Order and quote management
- **catalog** - Product catalog with categories
- **invoicing** - Invoice and payment processing
- **inventory** - Warehouse, stock, movements, reservations
- **documents** - Document generation and signatures

### Supporting Services
- **files** - File upload and processing
- **notifications** - Notification management

## Cross-Context Integration

The API implements event-driven architecture for cross-context integration:

- **Order Confirmed** → Stock Reservation (Inventory)
- **Order Confirmed** → Invoice Creation (Invoicing)
- **Invoice Created** → Journal Entry (Accounting)

See `docs/CROSS_CONTEXT_INTEGRATION.md` for detailed information.

## Authentication

The API supports two authentication methods:

### Session Authentication (Cookie)
Used for web clients with session cookies:
```
Cookie: session=<session_token>
```

### Bearer Authentication (JWT)
Used for API clients and mobile apps:
```
Authorization: Bearer <jwt_token>
```

## Authorization

The API uses Role-Based Access Control (RBAC):
- Admin role has full access
- Manager role has limited administrative access
- User role has basic access

## Rate Limiting

API endpoints are rate-limited:
- 100 requests per minute for authenticated users
- 20 requests per minute for anonymous users

## Pagination

List endpoints use cursor-based pagination:
```
GET /api/v1/orders?limit=20&cursor=eyJpZCI6IjAxOS...
```

## Versioning

The API uses URL versioning:
```
/api/v1/...  # Current stable version
/api/v2/...  # Future version (when available)
```

## Content Types

- Request: `application/json`
- Response: `application/json`
- File upload: `multipart/form-data`

## Error Responses

All errors follow a consistent format:

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid request parameters",
    "details": [
      {
        "field": "email",
        "message": "must be a valid email address"
      }
    ]
  }
}
```

## See Also

- [Architecture Documentation](../architecture/ARCHITECTURE.md)
- [Docker Development Guide](../DOCKER_DEVELOPMENT.md)
- [Cross-Context Integration](../CROSS_CONTEXT_INTEGRATION.md)
- [ADR Documents](../adr/README.md)