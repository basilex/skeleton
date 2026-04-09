# API Reference

> **Skeleton CRM REST API Documentation**

## 🔐 Authentication

All API requests require authentication (except registration/login):

```http
Authorization: Bearer <access_token>
```

### Get Access Token

```http
POST /api/v1/auth/login
Content-Type: application/json

{
  "email": "admin@skeleton.local",
  "password": "Admin1234!"
}
```

**Response:**
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIs...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIs...",
  "user": {
    "id": "01234567-...",
    "email": "admin@skeleton.local",
    "name": "Admin User",
    "roles": ["super_admin"]
  }
}
```

### Refresh Token

```http
POST /api/v1/auth/refresh
Content-Type: application/json

{
  "refresh_token": "eyJhbGciOiJIUzI1NiIs..."
}
```

---

## 📊 Common Patterns

### Pagination

```http
GET /api/v1/customers?page=2&limit=20&sort_by=name&sort_order=asc
```

**Response:**
```json
{
  "items": [...],
  "total": 150,
  "page": 2,
  "limit": 20,
  "has_more": true
}
```

### Money Format

All monetary values use **int64 cents** internally:

```json
{
  "total": 10000,      // $100.00 USD (in cents)
  "currency": "USD"
}
```

Frontend should use `Money` class from `@shared/types/money`:

```typescript
const total = Money.fromCents(10000, 'USD')
total.toFloat()        // 100.00
total.format()         // "$100.00"
```

### Error Responses

```json
{
  "error": "validation_error",
  "message": "Invalid input",
  "details": {
    "email": "Invalid email format"
  }
}
```

---

## 🎯 API Endpoints

### Authentication

#### Register User

```http
POST /api/v1/auth/register
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "SecurePass123!",
  "name": "John Doe"
}
```

**Response:** `201 Created`

```json
{
  "access_token": "...",
  "refresh_token": "...",
  "user": { ... }
}
```

#### Login

```http
POST /api/v1/auth/login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "SecurePass123!"
}
```

#### Refresh Token

```http
POST /api/v1/auth/refresh
Content-Type: application/json

{
  "refresh_token": "..."
}
```

#### Logout

```http
POST /api/v1/auth/logout
Authorization: Bearer <token>
```

#### Get Current User

```http
GET /api/v1/auth/me
Authorization: Bearer <token>
```

---

### Users

#### List Users

```http
GET /api/v1/users?page=1&limit=20
Authorization: Bearer <token>
```

**Required Permission:** `users:read`

**Response:**
```json
{
  "items": [
    {
      "id": "01234567-...",
      "email": "user@example.com",
      "name": "John Doe",
      "active": true,
      "created_at": "2026-01-01T00:00:00Z"
    }
  ],
  "total": 50,
  "page": 1,
  "limit": 20
}
```

#### Get User

```http
GET /api/v1/users/:id
Authorization: Bearer <token>
```

#### Assign Role

```http
POST /api/v1/users/:id/roles
Authorization: Bearer <token>
Content-Type: application/json

{
  "role_id": "..."
}
```

**Required Permission:** `roles:manage`

#### Deactivate User

```http
PATCH /api/v1/users/:id/deactivate
Authorization: Bearer <token>
```

**Required Permission:** `users:write`

---

### Customers

#### List Customers

```http
GET /api/v1/customers?page=1&limit=20&sort_by=name&sort_order=asc
Authorization: Bearer <token>
```

**Response:**
```json
{
  "items": [
    {
      "id": "01234567-...",
      "name": "Acme Corp",
      "email": "contact@acme.com",
      "phone": "+1-555-0100",
      "credit_limit": 50000,
      "total_purchases": 125000,
      "currency": "USD",
      "active": true,
      "created_at": "2026-01-01T00:00:00Z"
    }
  ],
  "total": 250,
  "page": 1,
  "limit": 20
}
```

#### Create Customer

```http
POST /api/v1/customers
Authorization: Bearer <token>
Content-Type: application/json

{
  "name": "Acme Corporation",
  "email": "contact@acme.com",
  "phone": "+1-555-0100",
  "credit_limit": 50000,
  "currency": "USD",
  "addresses": [
    {
      "type": "billing",
      "street": "123 Main St",
      "city": "New York",
      "state": "NY",
      "postal_code": "10001",
      "country": "US"
    }
  ]
}
```

#### Get Customer

```http
GET /api/v1/customers/:id
Authorization: Bearer <token>
```

#### Update Customer

```http
PUT /api/v1/customers/:id
Authorization: Bearer <token>
Content-Type: application/json

{
  "name": "Acme Corporation Updated",
  "email": "new-contact@acme.com"
}
```

---

### Invoices

#### List Invoices

```http
GET /api/v1/invoices?status=draft&page=1&limit=20
Authorization: Bearer <token>
```

**Query Parameters:**
- `status` - Filter by status: `draft`, `sent`, `paid`, `cancelled`
- `customer_id` - Filter by customer
- `from_date`, `to_date` - Filter by date range

**Response:**
```json
{
  "items": [
    {
      "id": "01234567-...",
      "number": "INV-2026-001",
      "customer_id": "...",
      "customer_name": "Acme Corp",
      "total": 10000,
      "currency": "USD",
      "status": "sent",
      "due_date": "2026-02-15",
      "created_at": "2026-01-15T00:00:00Z"
    }
  ],
  "total": 45,
  "page": 1,
  "limit": 20
}
```

#### Create Invoice

```http
POST /api/v1/invoices
Authorization: Bearer <token>
Content-Type: application/json

{
  "customer_id": "01234567-...",
  "due_date": "2026-02-15",
  "notes": "Thank you for your business",
  "lines": [
    {
      "description": "Consulting services",
      "quantity": 10,
      "unit_price": 15000,
      "currency": "USD"
    },
    {
      "description": "Software license",
      "quantity": 1,
      "unit_price": 5000,
      "currency": "USD"
    }
  ]
}
```

**Response:** `201 Created`

```json
{
  "id": "...",
  "number": "INV-2026-002",
  "customer_id": "...",
  "total": 155000,
  "subtotal": 155000,
  "currency": "USD",
  "status": "draft",
  "lines": [ ... ]
}
```

#### Get Invoice

```http
GET /api/v1/invoices/:id
Authorization: Bearer <token>
```

#### Add Invoice Line

```http
POST /api/v1/invoices/:id/lines
Authorization: Bearer <token>
Content-Type: application/json

{
  "description": "Additional service",
  "quantity": 5,
  "unit_price": 2000,
  "currency": "USD"
}
```

#### Send Invoice

```http
POST /api/v1/invoices/:id/send
Authorization: Bearer <token>
```

Changes status from `draft` to `sent`.

#### Record Payment

```http
POST /api/v1/invoices/:id/payments
Authorization: Bearer <token>
Content-Type: application/json

{
  "amount": 50000,
  "currency": "USD",
  "method": "bank_transfer",
  "reference": "TXN-12345"
}
```

#### Cancel Invoice

```http
POST /api/v1/invoices/:id/cancel
Authorization: Bearer <token>
```

---

### Orders

#### List Orders

```http
GET /api/v1/orders?status=pending&page=1&limit=20
Authorization: Bearer <token>
```

**Query Parameters:**
- `status` - Filter by status: `draft`, `pending`, `confirmed`, `shipped`, `delivered`, `cancelled`
- `customer_id` - Filter by customer

#### Create Order

```http
POST /api/v1/orders
Authorization: Bearer <token>
Content-Type: application/json

{
  "customer_id": "...",
  "shipping_address": {
    "street": "456 Oak St",
    "city": "Los Angeles",
    "state": "CA",
    "postal_code": "90001",
    "country": "US"
  },
  "lines": [
    {
      "item_id": "...",
      "quantity": 5,
      "unit_price": 10000,
      "currency": "USD"
    }
  ]
}
```

#### Update Order Status

```http
PATCH /api/v1/orders/:id/status
Authorization: Bearer <token>
Content-Type: application/json

{
  "status": "confirmed"
}
```

---

### Accounting

#### List Accounts

```http
GET /api/v1/accounts?type=asset&page=1&limit=20
Authorization: Bearer <token>
```

**Query Parameters:**
- `type` - Filter by type: `asset`, `liability`, `equity`, `revenue`, `expense`

#### Create Account

```http
POST /api/v1/accounts
Authorization: Bearer <token>
Content-Type: application/json

{
  "code": "1010",
  "name": "Cash",
  "type": "asset",
  "parent_id": "..."  // optional for hierarchy
}
```

#### Record Transaction

```http
POST /api/v1/transactions
Authorization: Bearer <token>
Content-Type: application/json

{
  "description": "Invoice payment",
  "lines": [
    {
      "account_id": "1010",  // Cash
      "debit": 10000
    },
    {
      "account_id": "4000",  // Revenue
      "credit": 10000
    }
  ]
}
```

**Note:** Debits must equal credits (double-entry bookkeeping).

---

### Inventory

#### List Warehouses

```http
GET /api/v1/warehouses
Authorization: Bearer <token>
```

#### Create Warehouse

```http
POST /api/v1/warehouses
Authorization: Bearer <token>
Content-Type: application/json

{
  "name": "Main Warehouse",
  "location": "New York, NY",
  "capacity": 10000
}
```

#### List Stock

```http
GET /api/v1/stock?warehouse_id=...&item_id=...
Authorization: Bearer <token>
```

#### Adjust Stock

```http
POST /api/v1/stock/:id/adjust
Authorization: Bearer <token>
Content-Type: application/json

{
  "quantity": 100,
  "reason": "Inventory count adjustment"
}
```

#### Transfer Stock

```http
POST /api/v1/stock/transfer
Authorization: Bearer <token>
Content-Type: application/json

{
  "item_id": "...",
  "from_warehouse_id": "...",
  "to_warehouse_id": "...",
  "quantity": 50
}
```

---

### Catalog

#### List Items

```http
GET /api/v1/catalog/items?category_id=...
Authorization: Bearer <token>
```

#### Create Item

```http
POST /api/v1/catalog/items
Authorization: Bearer <token>
Content-Type: application/json

{
  "sku": "PROD-001",
  "name": "Product Name",
  "description": "Product description",
  "category_id": "...",
  "unit_price": 5000,
  "currency": "USD",
  "status": "active"
}
```

---

### Files

#### Upload File

```http
POST /api/v1/files
Authorization: Bearer <token>
Content-Type: multipart/form-data

file: <binary>
```

**Response:**
```json
{
  "id": "...",
  "name": "document.pdf",
  "mime_type": "application/pdf",
  "size": 102400,
  "url": "/api/v1/files/.../download",
  "created_at": "..."
}
```

#### List Files

```http
GET /api/v1/files?page=1&limit=20
Authorization: Bearer <token>
```

#### Delete File

```http
DELETE /api/v1/files/:id
Authorization: Bearer <token>
```

---

### Notifications

#### List Notifications

```http
GET /api/v1/notifications?unread_only=true
Authorization: Bearer <token>
```

#### Get Notification

```http
GET /api/v1/notifications/:id
Authorization: Bearer <token>
```

#### Get Notification Preferences

```http
GET /api/v1/notifications/preferences
Authorization: Bearer <token>
```

#### Update Preferences

```http
PATCH /api/v1/notifications/preferences
Authorization: Bearer <token>
Content-Type: application/json

{
  "email_enabled": true,
  "sms_enabled": false,
  "push_enabled": true,
  "quiet_hours_start": "22:00",
  "quiet_hours_end": "08:00"
}
```

---

## 🔒 Permissions

### Permission Format

```
<resource>:<action>

Examples:
- users:read
- users:write
- users:delete
- invoices:*          // all actions on invoices
- *:*                  // super admin (all permissions)
```

### Default Roles

#### super_admin
```
*:*  // Full access to everything
```

#### admin
```
users:read, users:write
roles:read, roles:manage
invoices:*, orders:*
accounting:*, inventory:*
catalog:*, files:*
notifications:*, audit:read
```

#### viewer
```
users:read
invoices:read, orders:read
inventory:read, catalog:read
files:read
notifications:read
```

---

## 📝 Request/Response Examples

### Example: Create Customer with Invoice

```bash
# 1. Create customer
curl -X POST http://localhost:8080/api/v1/customers \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Acme Corp",
    "email": "contact@acme.com",
    "credit_limit": 100000,
    "currency": "USD"
  }'

# Response: {"id": "cust-123", ...}

# 2. Create invoice
curl -X POST http://localhost:8080/api/v1/invoices \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "customer_id": "cust-123",
    "due_date": "2026-02-15",
    "lines": [
      {
        "description": "Services",
        "quantity": 1,
        "unit_price": 50000,
        "currency": "USD"
      }
    ]
  }'

# Response: {"id": "inv-456", "number": "INV-2026-001", ...}

# 3. Send invoice
curl -X POST http://localhost:8080/api/v1/invoices/inv-456/send \
  -H "Authorization: Bearer <token>"

# 4. Record payment
curl -X POST http://localhost:8080/api/v1/invoices/inv-456/payments \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "amount": 50000,
    "currency": "USD",
    "method": "bank_transfer"
  }'
```

---

## 🚫 Error Codes

| Code | Description |
|------|-------------|
| 400  | Bad Request - Invalid input |
| 401  | Unauthorized - Authentication required |
| 403  | Forbidden - Insufficient permissions |
| 404  | Not Found - Resource doesn't exist |
| 409  | Conflict - Resource already exists |
| 422  | Unprocessable Entity - Validation failed |
| 429  | Too Many Requests - Rate limit exceeded |
| 500  | Internal Server Error |
| 503  | Service Unavailable |

---

## 📚 OpenAPI/Swagger

Interactive API documentation available at:

- **Swagger UI**: http://localhost:8080/swagger/index.html
- **OpenAPI Spec**: http://localhost:8080/swagger/doc.json

---

**API Version:** v2.0  
**Base URL:** `/api/v1`  
**Content-Type:** `application/json`
