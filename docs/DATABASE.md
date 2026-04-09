# Database Schema & Migrations

Complete guide to the database schema and migration strategy.

---

## Table of Contents

1. [Overview](#overview)
2. [Schema Design](#schema-design)
3. [Migrations](#migrations)
4. [Money Storage](#money-storage)
5. [UUID Strategy](#uuid-strategy)
6. [PostgreSQL Features](#postgresql-features)
7. [Best Practices](#best-practices)

---

## Overview

### Database Technology

- **PostgreSQL 16+** with extensions
- **Connection**: Go pgx driver (v5)
- **ORM**: None (native SQL with squirrel query builder)
- **Migrations**: golang-migrate

### Key Design Principles

1. **Money values**: BIGINT (int64 cents)
2. **Primary keys**: UUIDv7 (time-ordered)
3. **Timestamps**: TIMESTAMPTZ (with timezone)
4. **Soft deletes**: Boolean flags instead of actual deletion
5. **Audit fields**: created_at, updated_at on all tables

---

## Schema Design

### Bounded Contexts → Schemas

Each bounded context maps to a logical schema domain:

```
Identity       → users, sessions, roles, permissions
Accounting    → accounts, transactions, journal_entries
Parties       → parties (customers, suppliers, employees, partners)
Contracts     → contracts, contract_terms
Invoicing     → invoices, payments, credit_notes
Inventory     → stocks, warehouses, stock_adjustments
Ordering      → orders, order_lines, quotes
Catalog       → catalog_items, catalog_categories
Notifications → notifications, notification_templates
```

### Entity Relationships

```
┌──────────────┐
│   User       │
│  (Identity)  │
└──────┬───────┘
       │
       │ creates
       ▼
┌──────────────┐      ┌──────────────┐
│   Customer   │──────│   Contract   │
│  (Parties)   │      │ (Contracts)  │
└──────┬───────┘      └──────┬───────┘
       │                     │
       │ places              │ linked to
       ▼                     ▼
┌──────────────┐      ┌──────────────┐
│    Order     │──────│   Invoice    │
│  (Ordering)  │      │ (Invoicing)  │
└──────────────┘      └──────────────┘
       │                     │
       │ reserves            │ generates
       ▼                     ▼
┌──────────────┐      ┌──────────────┐
│    Stock     │      │ Transaction  │
│ (Inventory)  │      │ (Accounting) │
└──────────────┘      └──────────────┘
```

---

## Migrations

### Migration Files

Migrations are sequential SQL files in `migrations/`:

```
migrations/
├── 001_uuid.up.sql           # UUID extension
├── 002_audit.up.sql          # Audit functions
├── 003_identity.up.sql      # Users, roles, sessions
├── 018_parties.up.sql       # Parties context
├── 019_contracts.up.sql     # Contracts context
├── 020_accounting.up.sql    # Accounting context
├── 021_ordering.up.sql      # Ordering context
├── 022_catalog.up.sql       # Catalog context
└── 023_invoicing.up.sql     # Invoicing context
```

### Naming Convention

```
NNN_description.up.sql   # Apply migration
NNN_description.down.sql # Rollback migration
```

### Running Migrations

```bash
# Apply all migrations
make migrate

# Or manually
migrate -path migrations -database "postgres://..." up

# Rollback last migration
migrate -path migrations -database "postgres://..." down 1

# Check version
migrate -path migrations -database "postgres://..." version
```

### Creating NewMigration

```bash
# Create migration files
migrate create -ext sql -dir migrations -seq add_new_table

# This creates:
# migrations/024_add_new_table.up.sql
# migrations/024_add_new_table.down.sql
```

### Migration Template

```sql
-- migrations/024_example.up.sql
CREATE TABLE example (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    name VARCHAR(255) NOT NULL,
    amount BIGINT NOT NULL DEFAULT 0,  -- Money as BIGINT
    
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TRIGGER example_updated_at
    BEFORE UPDATE ON example
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at();
```

```sql
-- migrations/024_example.down.sql
DROP TABLE IF EXISTS example;
```

---

## Money Storage

### Why BIGINT?

Money is stored as **BIGINT** (int64) representing cents:

```
$100.50 → 10050 cents (int64)
```

**Rationale**:
- ✅ No floating-point precision errors
- ✅ Exact arithmetic
- ✅ Database-agnostic
- ✅ Fast calculations
- ✅ Perfect for accounting

### Database Schema

```sql
CREATE TABLE invoices (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    
    -- Money fields as BIGINT
    subtotal BIGINT NOT NULL,
    tax_amount BIGINT DEFAULT 0,
    discount BIGINT DEFAULT 0,
    total BIGINT NOT NULL,
    paid_amount BIGINT DEFAULT 0,
    
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

### Application Layer

```go
// Domain: Money value object
type Invoice struct {
    subtotal   money.Money  // pkg/money.Money
    taxAmount  money.Money
    total      money.Money
}

// Persistence: int64 conversion
type invoiceDTO struct {
    Subtotal int64 `db:"subtotal"`  // Database reads int64
}

func (dto *invoiceDTO) toDomain() (*Invoice, error) {
    subtotal, _ := money.New(dto.Subtotal, "USD")  // BIGINT → Money
    return &Invoice{subtotal: subtotal}, nil
}

// Repository: int64 conversion
func (r *Repo) Save(ctx context.Context, inv *Invoice) error {
    _, err := r.pool.Exec(ctx,
        "INSERT INTO invoices (...) VALUES ($1, ...)",
        inv.GetSubtotal().GetAmount(),  // Money → int64
    )
    return err
}
```

### API Layer

```go
// API Response (JSON)
type InvoiceResponse struct {
    Subtotal float64 `json:"subtotal"`  // 100.50 USD
}

// Handler
func (h *Handler) GetInvoice(ctx *gin.Context) {
    invoice, _ := h.repo.FindByID(ctx, id)
    response := InvoiceResponse{
        Subtotal: invoice.GetSubtotal().ToFloat64(),  // Money → float64
    }
    ctx.JSON(200, response)
}
```

---

## UUID Strategy

### Why UUIDv7?

- **Time-ordered**: Sortable by creation time
- **No collision**: Distributed safe
- **No coordination**: No central authority needed

### Implementation

```sql
-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create function for UUIDv7
CREATE OR REPLACE FUNCTION uuid_generate_v7() RETURNS uuid
AS $$
BEGIN
    RETURN (SELECT gen_random_uuid());
END;
$$ LANGUAGE plpgsql;

-- Use in tables
CREATE TABLE invoices (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    -- ...
);
```

### Foreign Keys

```sql
CREATE TABLE invoice_lines (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    invoice_id UUID NOT NULL REFERENCES invoices(id) ON DELETE CASCADE,
    -- ...
);
```

---

## PostgreSQL Features

### JSONB for Flexible Data

```sql
-- Structured JSON fields
CREATE TABLE parties (
    id UUID PRIMARY KEY,
    contact_info JSONB NOT NULL DEFAULT '{}',
    metadata JSONB DEFAULT '{}',
    -- ...
);

-- Query JSON
SELECT * FROM parties
WHERE contact_info->>'email' = 'user@example.com';

-- Update JSON
UPDATE parties
SET contact_info = jsonb_set(contact_info, '{phone}', '"1234567890"')
WHERE id = '...';
```

### LTREE for Hierarchies

```sql
-- Hierarchical categories
CREATE EXTENSION IF NOT EXISTS ltree;

CREATE TABLE catalog_categories (
    id UUID PRIMARY KEY,
    path LTREE NOT NULL,  -- 'electronics.computers.laptops'
    -- ...
);

-- Query hierarchy
SELECT * FROM catalog_categories
WHERE path <@ 'electronics.computers'::ltree;  -- All descendants
```

### DATERANGE for Validity

```sql
-- Date ranges for contracts
CREATE TABLE contracts (
    id UUID PRIMARY KEY,
    validity_period DATERANGE NOT NULL,
    -- ...
);

-- Query overlapping contracts
SELECT * FROM contracts
WHERE validity_period && DATERANGE('[2024-01-01, 2024-12-31)');
```

### Array Types

```sql
-- Arrays for collections
CREATE TABLE parties (
    contracts UUID[] DEFAULT '{}',  -- Array of contract IDs
    -- ...
);

-- Query array
SELECT * FROM parties WHERE 'contract-123' = ANY(contracts);

-- Update array
UPDATE parties
SET contracts = array_append(contracts, 'contract-456')
WHERE id = '...';
```

---

## Best Practices

### 1. Transactions

Always use transactions for operations affecting multiple tables:

```go
func (s *Service) CreateInvoice(ctx context.Context, cmd CreateInvoice) error {
    return s.txManager.Execute(ctx, func(ctx context.Context) error {
        // Create invoice
        invoice, _ := domain.NewInvoice(cmd...)
        if err := s.invoiceRepo.Save(ctx, invoice); err != nil {
            return err
        }
        
        // Create transaction
        transaction, _ := domain.NewTransaction(...)
        return s.transactionRepo.Save(ctx, transaction)
    })
}
```

### 2. Indexing

```sql
-- Composite indexes for common queries
CREATE INDEX idx_invoices_customer_status ON invoices(customer_id, status);
CREATE INDEX idx_invoices_due_date ON invoices(due_date);

-- JSONB path indexes
CREATE INDEX idx_parties_contact_email ON parties
    USING GIN (contact_info jsonb_path_ops);

-- LTREE indexes
CREATE INDEX idx_categories_path ON catalog_categories USING GIST(path);
```

### 3. Soft Deletes

```sql
CREATE TABLE parties (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    -- ...
);

-- Instead of DELETE
UPDATE parties SET is_active = false WHERE id = '...';

-- Query active records
SELECT * FROM parties WHERE is_active = true;
```

### 4. Auditing

```sql
-- All tables have audit triggers
CREATE TRIGGER invoices_updated_at
    BEFORE UPDATE ON invoices
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at();

-- Audit function
CREATE OR REPLACE FUNCTION update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
```

### 5. Connection Pooling

```go
// Database configuration
config := pgxpool.Config{
    MaxConns:        25,
    MinConns:        5,
    MaxConnLifetime: 1 * time.Hour,
    MaxConnIdleTime: 30 * time.Minute,
    HealthCheckPeriod: 1 * time.Minute,
}
```

---

## Common Queries

### Paginated List

```go
func (r *Repo) FindAll(ctx context.Context, filter Filter) (PageResult, error) {
    query := r.psql.Select("*").From("invoices").
        Where(squirrel.Eq{"customer_id": filter.CustomerID}).
        OrderBy("created_at DESC").
        Limit(uint64(filter.Limit + 1))
    
    // Return with next cursor for pagination
    return PageResult{Items: invoices, HasMore: hasMore}, nil
}
```

### Insert on Conflict

```sql
INSERT INTO accounts (id, code, name, balance, updated_at)
VALUES ($1, $2, $3, $4, NOW())
ON CONFLICT(id) DO UPDATE SET
    balance = EXCLUDED.balance,
    updated_at = NOW();
```

### Join with JSONB

```sql
SELECT
    i.*,
    p.contact_info->>'email' as customer_email,
    p.contact_info->>'phone' as customer_phone
FROM invoices i
JOIN parties p ON i.customer_id = p.id
WHERE i.status = 'pending';
```

---

## Database Diagrams

### Core Tables

```
users                   parties              invoices
┌──────────────┐       ┌──────────────┐    ┌──────────────┐
│ id (UUID)    │       │ id (UUID)    │    │ id (UUID)    │
│ email        │       │ party_type   │    │ customer_id  │───┐
│ password_hash│       │ name         │    │ total (INT)  │   │
│ role_id      │──┐    │ total_purchas│    │ status       │   │
└──────────────┘  │    └──────────────┘    └──────────────┘   │
                   │                                              │
roles              │    accounts             orders               │
┌──────────────┐   │    ┌──────────────┐    ┌──────────────┐     │
│ id (UUID)    │◀──┘    │ id (UUID)    │    │ id (UUID)    │     │
│ name         │        │ code         │    │ customer_id  │─────┘
│ permissions  │        │ balance (INT)│    │ total (INT)  │
└──────────────┘        └──────────────┘    └──────────────┘
```

---

## See Also

- **[Architecture Overview](ARCHITECTURE.md)** - System architecture
- **[Development Guide](DEVELOPMENT.md)** - Database development workflow
- **[Main README](../README.md)** - Project overview
