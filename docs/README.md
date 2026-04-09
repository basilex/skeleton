# Skeleton Business Engine - Documentation

A production-ready business engine with Domain-Driven Design, Hexagonal Architecture, and CQRS.

## 📚 Architecture Documentation

### Architecture Decision Records (ADR)
- [ADR-017: Parties Bounded Context](adr/ADR-017-parties.md) - Customer, Supplier, Partner, Employee management
- [ADR-018: Contracts Bounded Context](adr/ADR-018-contracts.md) - Contract lifecycle with DATERANGE
- [ADR-019: Accounting Bounded Context](adr/ADR-019-accounting.md) - Chart of Accounts & Double-Entry
- [ADR-020: Ordering Bounded Context](adr/ADR-020-ordering.md) - Order management with state machine
- [ADR-021: Catalog Bounded Context](adr/ADR-021-catalog.md) - Product catalog with LTREE & JSONB

### Quick Start

```bash
# Build
go build ./cmd/api

# Run tests
go test ./internal/parties/domain/... ./internal/contracts/domain/... ./internal/accounting/domain/... ./internal/ordering/domain/... ./internal/catalog/domain/... -v

# Run migrations
psql -d skeleton -f migrations/018_parties.up.sql
psql -d skeleton -f migrations/019_contracts.up.sql
psql -d skeleton -f migrations/020_accounting.up.sql
psql -d skeleton -f migrations/021_ordering.up.sql
psql -d skeleton -f migrations/022_catalog.up.sql
```

## 🏗️ Architecture Overview

### Bounded Contexts

#### 1. Parties (Universal)
Multi-tenant party management for any business type.
- **Entities**: Customer, Supplier, Partner, Employee
- **Key Features**: Loyalty system, JSONB contact info, status management
- **Database**: Single table with discriminator (party_type)
- [Details](adr/ADR-017-parties.md)

#### 2. Contracts
Contract lifecycle management with validity periods.
- **Entities**: Contract, PaymentTerms, DeliveryTerms
- **Key Features**: DATERANGE validity, state transitions, party references
- **Database**: Native date range type with exclusion constraints
- [Details](adr/ADR-018-contracts.md)

#### 3. Accounting
Chart of Accounts with double-entry bookkeeping.
- **Entities**: Account, Transaction, Money
- **Key Features**: Account hierarchy, debit/credit logic, transaction tracking
- **Database**: Account types with parent-child relationships
- [Details](adr/ADR-019-accounting.md)

#### 4. Ordering
Order management with lines and state machine.
- **Entities**: Order, OrderLine, Quote
- **Key Features**: State machine, line management, automatic totals
- **Database**: Order lines with FK cascade
- [Details](adr/ADR-020-ordering.md)

#### 5. Catalog
Product catalog with hierarchical categories.
- **Entities**: Item, Category, Attributes
- **Key Features**: LTREE hierarchy, JSONB attributes, flexible schema
- **Database**: LTREE for categories, JSONB for item attributes
- [Details](adr/ADR-021-catalog.md)

### Design Patterns

#### Domain-Driven Design (DDD)

**Domain Layer Components:**
- Aggregates: Customer, Contract, Account, Order, Item
- Entities: OrderLine, Transaction
- Value Objects: Money, ContactInfo, Address, PaymentTerms
- Domain Events: CustomerCreated, OrderPlaced, TransactionRecorded
- Repository Interfaces: CustomerRepository, OrderRepository

#### Hexagonal Architecture (Ports & Adapters)
```
Ports (Interfaces)
├── HTTP (REST API)
├── Application (Commands/Queries)
└── Domain (Business Logic)

Adapters (Implementations)
├── Infrastructure (PostgreSQL, Redis)
├── Persistence (Repositories)
└── External (EventBus)
```

#### CQRS (Command Query Separation)
```
├── Commands (Write)
│   ├── CreateCustomer
│   ├── ConfirmOrder
│   └── RecordTransaction
│
└── Queries (Read)
    ├── GetCustomer
    ├── ListOrders
    └── ListAccounts
```

### Technology Stack

#### Core Technologies
- **Language**: Go 1.21+
- **Database**: PostgreSQL 16
- **Router**: Gin-Gonic
- **SQL**: squirrel (query builder) + scany v2 (scanner)
- **Validation**: Gin binding
- **Events**: EventBus (memory/Redis)

#### PostgreSQL Features Used
- **UUID v7**: Time-sortable identifiers
- **DATERANGE**: Contract validity periods
- **LTREE**: Hierarchical category paths
- **JSONB**: Flexible contact info & attributes
- **ENUM Types**: Type-safe status fields
- **GIN Indexes**: Fast JSONB & LTREE queries

## 🚀 Getting Started

### Prerequisites
- Go 1.21+
- PostgreSQL 16+
- Redis (optional, for distributed events)

### Configuration

```yaml
# config.yaml
app:
  env: development
  port: 8080

database:
  host: localhost
  port: 5432
  name: skeleton
  user: postgres
  password: password

redis:
  url: redis://localhost:6379
```

### Run Application

```bash
# Install dependencies
go mod download

# Run migrations
make migrate-up

# Start server
make run
```

## 📖 API Documentation

### Swagger UI
```
http://localhost:8080/swagger/index.html
```

### Endpoints Overview

#### Parties
```
POST   /api/v1/customers          # Create customer
GET    /api/v1/customers/:id      # Get customer
GET    /api/v1/customers          # List customers
PUT    /api/v1/customers/:id      # Update customer
```

#### Contracts
```
POST   /api/v1/contracts              # Create contract
GET    /api/v1/contracts/:id            # Get contract
GET    /api/v1/contracts               # List contracts
PUT    /api/v1/contracts/:id/activate  # Activate contract
PUT    /api/v1/contracts/:id/terminate # Terminate contract
```

#### Accounting
```
POST   /api/v1/accounts          # Create account
GET    /api/v1/accounts/:id       # Get account
GET    /api/v1/accounts          # List accounts
POST   /api/v1/transactions      # Record transaction
```

#### Ordering
```
POST   /api/v1/orders            # Create order
GET    /api/v1/orders/:id        # Get order
GET    /api/v1/orders            # List orders
POST   /api/v1/orders/:id/lines # Add order line
PUT    /api/v1/orders/:id/status # Update status
```

#### Catalog
```
POST   /api/v1/catalog/items     # Create item
GET    /api/v1/catalog/items/:id  # Get item
GET    /api/v1/catalog/items      # List items
PUT    /api/v1/catalog/items/:id  # Update item
```

## 🧪 Testing

### Unit Tests
```bash
# Run domain tests
go test ./internal/parties/domain/... -v
go test ./internal/contracts/domain/... -v
go test ./internal/accounting/domain/... -v
go test ./internal/ordering/domain/... -v
go test ./internal/catalog/domain/... -v

# Run all tests
go test ./... -v
```

### Integration Tests
```go
// tests/integration/accounting_test.go
func TestAccountingIntegration(t *testing.T) {
    // Use testcontainers for PostgreSQL
    // Test repository operations
}
```

## 📊 Database Schema

### Entity Relationship Diagram

```
┌─────────────┐        ┌──────────────┐
│   parties   │        │  contracts   │
├─────────────┤        ├──────────────┤
│ id (PK)     │◄───────│ party_id (FK)│
│ party_type  │        │ validity     │
│ contact_info│        │ payment_terms│
│ status      │        │ status       │
└─────────────┘        └──────────────┘
        │                     │
        │                     │
        ▼                     ▼
┌─────────────┐        ┌──────────────┐
│   orders    │        │  accounting  │
├─────────────┤        ├──────────────┤
│ customer_id │        │ accounts     │
│ supplier_id │        │ transactions │
│ contract_id │        └──────────────┘
│ total       │
└─────────────┘
        │
        ▼
┌─────────────┐
│ order_lines │
├─────────────┤
│ item_id     │
│ quantity    │
│ unit_price  │
└─────────────┘
        │
        ▼
┌─────────────┐
│catalog_items│
├─────────────┤
│ category_id │
│ attributes  │
└─────────────┘
```

## 🔧 Development

### Project Structure
```
skeleton/
├── cmd/api/              # Application entrypoint
├── internal/
│   ├── parties/          # Parties bounded context
│   ├── contracts/        # Contracts bounded context
│   ├── accounting/       # Accounting bounded context
│   ├── ordering/         # Ordering bounded context
│   ├── catalog/          # Catalog bounded context
│   ├── identity/         # Auth & Users
│   ├── audit/             # Audit logging
│   ├── files/             # File management
│   ├── notifications/     # Notifications
│   └── tasks/             # Background tasks
├── migrations/           # Database migrations
├── pkg/                  # Shared packages
├── docs/                 # Documentation
└── tests/                # Integration tests
```

### Make Commands
```bash
make run          # Start server
make test         # Run tests
make migrate-up   # Run migrations
make migrate-down # Rollback migrations
make swagger     # Generate Swagger docs
make lint         # Run linter
make build        # Build binary
```

## 📝 License

MIT License - See LICENSE file for details.

## 🤝 Contributing

1. Fork the repository
2. Create feature branch (`git checkout -b feature/amazing-feature`)
3. Commit changes (`git commit -m 'Add amazing feature'`)
4. Push to branch (`git push origin feature/amazing-feature`)
5. Open Pull Request

## 📞 Support

- **Issues**: GitHub Issues
- **Documentation**: `/docs` directory
- **Architecture**: ADR documents in `/docs/adr`
