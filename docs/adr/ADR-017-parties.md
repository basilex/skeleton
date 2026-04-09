# ADR-017: Parties Bounded Context

## Status
Accepted

## Context
Universally applicable Party Management for B2B/B2C business with multiple party types: Customers, Suppliers, Partners, Employees.

## Decision
Implement Parties as a bounded context with:

### Domain Model
- **Party (Abstract)**: Base entity with common attributes
- **Customer**: B2B/B2C customers with loyalty tracking
- **Supplier**: Vendors/suppliers with contract relationships
- **Partner**: Business partners with rating system
- **Employee**: Internal staff with position tracking

### Architecture

- `internal/parties/`
  - `domain/`
    - `party.go` - Abstract base
    - `customer.go` - Customer aggregate
    - `supplier.go` - Supplier aggregate
    - `partner.go` - Partner aggregate
    - `employee.go` - Employee aggregate
    - `contact_info.go` - Value object (JSONB)
    - `bank_account.go` - Value object (JSONB)
    - `ids.go` - Identifiers
    - `events.go` - Domain events
    - `repository.go` - Repository interfaces
  - `infrastructure/`
    - `persistence/`
      - `models.go` - DTO mapping
      - `customer_repository.go`
      - `supplier_repository.go`
      - `partner_repository.go`
      - `employee_repository.go`
  - `application/`
    - `command/`
      - `create_customer.go`
      - `update_customer.go`
      - `create_supplier.go`
    - `query/`
      - `customer.go`
      - `supplier.go`
  - `ports/http/`
    - `handler.go` - REST handlers
    - `dto.go` - Request/response DTOs

### Database Schema
```sql
CREATE TYPE party_type AS ENUM ('customer', 'supplier', 'partner', 'employee');
CREATE TYPE party_status AS ENUM ('active', 'inactive', 'blacklisted');
CREATE TYPE loyalty_level AS ENUM ('bronze', 'silver', 'gold', 'platinum');

CREATE TABLE parties (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    party_type party_type NOT NULL,
    name VARCHAR(255) NOT NULL,
    tax_id VARCHAR(50),
    contact_info JSONB NOT NULL,
    bank_account JSONB,
    status party_status NOT NULL DEFAULT 'active',
    loyalty_level loyalty_level DEFAULT 'bronze',
    total_purchases DECIMAL(15,2) DEFAULT 0,
    rating JSONB,
    contracts TEXT[],
    position VARCHAR(100),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_parties_type ON parties(party_type);
CREATE INDEX idx_parties_status ON parties(status);
CREATE INDEX idx_parties_email ON parties USING GIN ((contact_info->>'email'));
```

### Key Design Decisions

#### 1. Single Table with Discriminator
- Uses `party_type` enum as discriminator
- Shared table for all party types
- Type-specific fields stored in JSONB columns
- Pros: Simple queries, referential integrity
- Cons: Sparse table for party-specific fields

#### 2. JSONB Value Objects
- `contact_info`: Email, phone, address, social media
- `bank_account`: Bank details, IBAN, SWIFT
- GIN indexes for fast JSON queries
- Flexible schema evolution

#### 3. Loyalty System
- Bronze → Silver → Gold → Platinum progression
- Thresholds: 20K, 50K, 100K total purchases
- Automatic upgrade via domain logic

#### 4. Status Management
- Active/Inactive/Blacklisted states
- State machine prevents illegal transitions
- Blacklisted parties cannot activate

### API Endpoints

#### Customers
```
POST   /api/v1/customers          # Create customer
GET    /api/v1/customers/:id      # Get customer
GET    /api/v1/customers          # List customers (paginated)
PUT    /api/v1/customers/:id      # Update customer
```

#### Suppliers
```
POST   /api/v1/suppliers          # Create supplier
GET    /api/v1/suppliers/:id      # Get supplier
GET    /api/v1/suppliers          # List suppliers (paginated)
```

### Usage Example

```go
// Create customer
handler := partiesHTTP.NewHandler(
    createCustomerCmd,
    updateCustomerCmd,
    getCustomerQuery,
    listCustomersQuery,
    createSupplierCmd,
    getSupplierQuery,
    listSuppliersQuery,
)

// REST call
POST /api/v1/customers
{
    "name": "ACME Corp",
    "tax_id": "12345678",
    "email": "info@acme.com",
    "phone": "+1234567890",
    "address": {
        "city": "Kyiv",
        "country": "Ukraine"
    }
}

// Response
{
    "id": "019f1234-5678-7abc-def0-123456789abc",
    "name": "ACME Corp",
    "status": "active",
    "loyalty_level": "bronze"
}
```

### Consequences

#### Positive
- ✅ Universal party management for any business type
- ✅ JSONB provides schema flexibility
- ✅ Domain events enable integration with other contexts
- ✅ Type-safe party identification
- ✅ Pagination with cursor-based approach

#### Negative
- ⚠️ Sparse table due to party-type-specific fields
- ⚠️ JSONB queries slower than normalized columns
- ⚠️ Complex migrations for nested JSON structures

### Integration Points

1. **Contracts Context**: Party ID referenced in contracts
2. **Accounting Context**: Party ID for payables/receivables
3. **Ordering Context**: Customer/Supplier for orders
4. **Audit Context**: Automatic audit logging via events

### Performance Considerations

- GIN index on `contact_info->>'email'` for fast lookup
- Cursor-based pagination for large datasets
- Composite index on `(party_type, status)` for filtering
- JSONB compression for large contact info

### Migration Path

To add a new party type:
1. Add type to `party_type` enum
2. Create aggregate in domain/
3. Add repository in infrastructure/
4. Add commands/queries in application/
5. Add HTTP handlers in ports/http/
6. Update wire.go and routes.go

### References
- DDD: Bounded Context, Aggregate, Value Object
- PostgreSQL: JSONB, GIN indexes, UUID v7
- Hexagonal Architecture: Ports & Adapters
