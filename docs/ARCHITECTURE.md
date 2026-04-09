# Architecture Overview

This document describes the architecture of the Skeleton Business Engine, following Domain-Driven Design (DDD) and Hexagonal Architecture principles.

---

## Table of Contents

1. [Architecture Style](#architecture-style)
2. [Bounded Contexts](#bounded-contexts)
3. [Domain Model](#domain-model)
4. [Infrastructure](#infrastructure)
5. [Cross-Cutting Concerns](#cross-cutting-concerns)
6. [Technology Choices](#technology-choices)

---

## Architecture Style

### Hexagonal Architecture (Ports & Adapters)

```
┌─────────────────────────────────────────────────────────────┐
│                     Presentation Layer                      │
│                    (HTTP Handlers/Ports)                    │
└─────────────────────────┬───────────────────────────────────┘
                          │
┌─────────────────────────┴───────────────────────────────────┐
│                    Application Layer                        │
│              (Commands, Queries, Services)                  │
└─────────────────────────┬───────────────────────────────────┘
                          │
┌─────────────────────────┴───────────────────────────────────┐
│                      Domain Layer                           │
│          (Aggregates, Entities, Value Objects)              │
└─────────────────────────┬───────────────────────────────────┘
                          │
┌─────────────────────────┴───────────────────────────────────┐
│                  Infrastructure Layer                       │
│          (Repositories, Event Bus, Database)                │
└─────────────────────────────────────────────────────────────┘
```

### Key Principles

1. **Domain-centric**: Business logic lives in the domain layer
2. **Dependency inversion**: Domain doesn't depend on infrastructure
3. **Ports & Adapters**: Interfaces (ports) define contracts, implementations (adapters) provide concrete solutions
4. **Testability**: Domain logic can be tested without infrastructure

---

## Bounded Contexts

### 1. Identity Context

**Purpose**: Authentication and Authorization

**Responsibilities**:
- User authentication (email/password)
- Session management (devices, expiration)
- Role-Based Access Control (RBAC)
- User preferences

**Key Entities**:
- User
- Session
- Role
- Permission

**Database Tables**:
- users
- sessions
- roles
- permissions
- user_preferences

**Events**:
- UserCreated
- UserActivated
- SessionCreated
- SessionRevoked

---

### 2. Accounting Context

**Purpose**: Financial Management and Double-Entry Bookkeeping

**Responsibilities**:
- Chart of Accounts management
- Journal entries validation
- Transaction recording
- Account balance tracking

**Key Entities**:
- Account (hierarchical)
- Transaction
- JournalEntry
- JournalLine

**Money Integration**:
```go
type Account struct {
    id          AccountID
    code        string
    name        string
    accountType AccountType
    currency    Currency
    balance     money.Money  // int64 cents
    parentID    *AccountID
}
```

**Database Tables**:
- accounts (hierarchical)
- transactions
- journal_entries
- journal_lines

**Events**:
- AccountCreated
- TransactionRecorded
- JournalEntryPosted

---

### 3. Parties Context

**Purpose**: Universal party management (Customers, Suppliers, Partners, Employees)

**Responsibilities**:
- Customer management with credit limits
- Supplier management with performance ratings
- Partner management
- Contact information
- Loyalty programs

**Key Entities**:
- Customer
- Supplier
- Partner
- Employee

**Money Integration**:
```go
type Customer struct {
    id             PartyID
    name           string
    totalPurchases  money.Money  // Total purchase amount
    creditLimit    money.Money  // Credit limit
    currentCredit  money.Money  // Current credit used
}
```

**Database Tables**:
- parties (single table with party_type discriminator)

**Events**:
- CustomerCreated
- CustomerCreditLimitChanged

---

### 4. Contracts Context

**Purpose**: Contract Lifecycle Management

**Responsibilities**:
- Contract creation and renewal
- Amendment tracking
- Status management
- Validity period tracking

**Key Entities**:
- Contract
- PaymentTerms
- DeliveryTerms

**Database Tables**:
- contracts (with DATERANGE validity_period)

---

### 5. Invoicing Context

**Purpose**: Invoice Processing and Payment Tracking

**Responsibilities**:
- Invoice creation with line items
- Payment recording
- Credit note processing
- Installment plans

**Key Entities**:
- Invoice
- InvoiceLine
- Payment
- CreditNote
- InstallmentPlan

**Money Integration**:
```go
type Invoice struct {
    id          InvoiceID
    subtotal    money.Money
    taxAmount   money.Money
    discount    money.Money
    total       money.Money
    paidAmount  money.Money
}
```

**Database Tables**:
- invoices
- invoice_lines
- payments
- credit_notes

**Events**:
- InvoiceCreated
- InvoicePaid
- PaymentRecorded

---

### 6. Inventory Context

**Purpose**: Stock and Warehouse Management

**Responsibilities**:
- Stock tracking (quantity, lot, serial)
- Warehouse zones and locations
- Stock take with variance detection
- Stock adjustments (gain/loss)

**Key Entities**:
- Stock
- Warehouse
- Zone
- StockTake
- StockAdjustment

**Database Tables**:
- stocks
- warehouses
- zones
- stock_takes
- stock_adjustments

**Events**:
- StockAdjusted
- StockReserved
- StockReservationReleased

---

### 7. Ordering Context

**Purpose**: Order Processing

**Responsibilities**:
- Order creation and management
- Order line items
- Status workflow
- Order confirmation

**Key Entities**:
- Order
- OrderLine
- Quote

**Money Integration**:
```go
type Order struct {
    id          OrderID
    subtotal    money.Money
    taxAmount   money.Money
    discount    money.Money
    total       money.Money
    status      OrderStatus
}
```

**Events**:
- OrderCreated
- OrderConfirmed
- OrderCompleted

---

### 8. Catalog Context

**Purpose**: Product Catalog Management

**Responsibilities**:
- Product and service definitions
- Product variants (size, color, material)
- Pricing rules
- Category hierarchy

**Key Entities**:
- Product
- Service
- Variant
- Category (LTREE hierarchy)

**Database Tables**:
- catalog_items
- catalog_variants
- catalog_categories (with LTREE path)
- catalog_prices

---

### 9. Documents Context

**Purpose**: Document Management

**Responsibilities**:
- File uploads with metadata
- Version control
- Approval workflows
- Document relationships

**Key Entities**:
- File
- Version
- Approval

---

### 10. Notifications Context

**Purpose**: Multi-channel Notifications

**Responsibilities**:
- Email notifications
- SMS notifications
- Push notifications
- In-app notifications
- Templates

---

## Domain Model

### Aggregates

Eachbounded context defines aggregates that ensure consistency boundaries:

```go
type Invoice struct {
    id          InvoiceID
    lines       []*InvoiceLine
    payments    []*Payment
    // ... value objects
    events      []eventbus.Event  // Domain events
}

func (i *Invoice) AddLine(line *InvoiceLine) error {
    // Business logic
    // Validate invariants
    // Update state
    // Publish events
    i.events = append(i.events, InvoiceLineAdded{...})
}
```

### Value Objects

Value objects are immutable and compared by value:

```go
type Money struct {
    Amount   int64
    Currency string
}

func (m Money) Add(other Money) (Money, error) {
    if m.Currency != other.Currency {
        return Money{}, ErrDifferentCurrencies
    }
    return Money{Amount: m.Amount + other.Amount, Currency: m.Currency}, nil
}
```

### Domain Events

Rich domain events for cross-context communication:

```go
type OrderConfirmed struct {
    OrderID     OrderID
    CustomerID  string
    Lines       []OrderConfirmedLine
    Total       money.Money
    Currency    string
    OccurredAt  time.Time
}
```

---

## Infrastructure

### Database (PostgreSQL)

**Key Features**:
- JSONB for flexible JSON fields
- LTREE for hierarchical categories
- DATERANGE for contract validity periods
- UUID v7 for primary keys

**Schema Organization**:
- Migrations: sequential numbered files
- Each bounded context has its own migration file
- Foreign keys across contexts (loose coupling)

**Example Migration**:
```sql
-- 018_parties.up.sql
CREATE TABLE parties (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    party_type party_type NOT NULL,
    name VARCHAR(255) NOT NULL,
    total_purchases BIGINT DEFAULT 0,  -- Money as int64 cents
    credit_limit BIGINT DEFAULT 0,
    -- ...
);
```

### Event Bus

**Implementation**: In-memory event bus for monolithic deployment

**Interface**:
```go
type EventBus interface {
    Subscribe(eventName string, handler Handler) error
    Publish(ctx context.Context, event Event) error
}
```

**Event Flow**:
1. Aggregate publishes domain event
2. Application layer pulls events and publishes to event bus
3. Event handlers process events asynchronously
4. Cross-context communication via events

**Future**: Ready for extraction to message queue (Kafka, RabbitMQ)

---

## Cross-Cutting Concerns

### 1. Money Value Object

**Implementation**: All monetary values use `pkg/money.Money`

```go
// Storage: int64 cents (BIGINT in database)
// API: float64 for JSON serialization
// Domain: Money value object with operations

// Create
amount, _ := money.NewFromFloat(100.50, "USD")

// Operations
total, _ := amount.Add(another)
total, err := amount.Multiply(1.1)  // 10% increase

// Storage
db.Exec("INSERT ... VALUES (?, ?)", id, amount.GetAmount())

// API Response
json.Marshal(struct{ Amount float64 }{Amount: amount.ToFloat64()})
```

### 2. Transaction Management

**Implementation**: Transaction context propagation

```go
type TransactionManager interface {
    Execute(ctx context.Context, fn func(ctx context.Context) error) error
}

// Usage in command handler
func (h *Handler) Handle(ctx context.Context, cmd Command) error {
    return h.txManager.Execute(ctx, func(ctx context.Context) error {
        // All operations in same transaction
        account, _ := h.accountRepo.FindByID(ctx, accountID)
        account.Debit(amount)
        return h.accountRepo.Save(ctx, account)
    })
}
```

### 3. Error Handling

**Domain Errors**:
```go
var (
    ErrAccountNotFound         = errors.New("account not found")
    ErrAccountInactive         = errors.New("account is inactive")
    ErrDifferentCurrencies     = errors.New("different currencies")
    ErrInsufficientCredit      = errors.New("insufficient credit")
)
```

**Application Errors**:
```go
type ApplicationError struct {
    Code    string
    Message string
    Err     error
}
```

### 4. Logging

**Structured logging** with context:
```go
log.Info("Account created",
    "account_id", account.GetID(),
    "account_code", account.GetCode(),
    "account_type", account.GetType(),
)
```

---

## Technology Choices

### Backend Stack

| Component | Technology | Purpose |
|-----------|------------|---------|
| Language | Go 1.25+ | High performance, simplicity |
| Web Framework | Gin | HTTP routing and middleware |
| Database | PostgreSQL 16+ | Relational database with JSONB, LTREE |
| Migrations | golang-migrate | Database schema versioning |
| UUID | UUIDv7 | Time-ordered unique IDs |
| Money | Custom `pkg/money` | Monetary calculations in cents |

### Frontend Stack (Integration Point)

Frontend integration options:
- **REST API**: JSON over HTTP
- **OpenAPI/Swagger**: API documentation
- **CORS**: Cross-origin requests

### Deployment Options

- **Monolith**: All contexts in one service
- **Modular Monolith**: Ready for extraction
- **Microservices**: Each context can become a service

### Future Considerations

- Message Queue: Kafka or RabbitMQ for event bus
- CQRS: Separate read/write models
- Event Sourcing: Event store for audit
- GraphQL: Alternative API layer

---

## Integration Points

### Frontend Integration

**API Structure**:
```
/api/v1/
├── /auth        # Authentication
├── /users       # User management
├── /accounts    # Accounting
├── /parties     # Parties
├── /contracts   # Contracts
├── /invoices    # Invoicing
├── /orders      # Orders
└── /stock       # Inventory
```

**Money Representation**:
```json
{
  "amount": 100.50,
  "currency": "USD"
}
```

### Cross-Context Integration

**Event-Driven Communication**:

```
Ordering Context
     │
     ├─ publishes OrderCreated
     │
     ▼
Event Bus
     │
     ├─ InvoiceCreated handler (Invoicing)
     ├─ StockReservation handler (Inventory)
     └─ Notification handler (Notifications)
```

**Integration Events**:
1. `OrderCreated` → Creates invoice
2. `OrderConfirmed` → Reserves stock
3. `PaymentRecorded` → Updates account balance

---

## Architectural Decisions

See [ADR Index](adr/) for detailed Architecture Decision Records:

- [ADR-017: Parties Bounded Context](adr/ADR-017-parties.md)
- [ADR-018: Contracts Bounded Context](adr/ADR-018-contracts.md)
- [ADR-019: Accounting Bounded Context](adr/ADR-019-accounting.md)
- [ADR-020: Ordering Bounded Context](adr/ADR-020-ordering.md)
- [ADR-021: Catalog Bounded Context](adr/ADR-021-catalog.md)

---

## See Also

- **[Database Schema](DATABASE.md)** - Detailed database design
- **[Development Guide](DEVELOPMENT.md)** - Development workflow
- **[Testing Strategy](TESTING.md)** - Testing approach
- **[Main README](../README.md)** - Project overview