# Universal Business Engine - Execution Plan

## Overview

Цей документ описує план трансформації skeleton з базового API starter в універсальний Business Engine, придатний для будь-якого B2B/B2C бізнесу.

**Мета:** Створити reusable foundation, який містить універсальні бізнес-модулі, спільні для всіх бізнес-систем:

- **Parties** - Customers, Suppliers, Partners, Employees
- **Contracts** - Agreements between parties
- **Accounting** - Financial operations, transactions
- **Ordering** - Universal order system
- **Catalog** - Products/Services/Properties

## Architecture Principles

### Core Principles

1. **Domain-Driven Design (DDD)** - Bounded contexts, aggregates, domain events
2. **Hexagonal Architecture** - Domain in center, infrastructure as adapters
3. **Event-Driven** - Async communication between contexts
4. **PostgreSQL 16** - Native UUID v7, JSONB, full-text search
5. **Repository Pattern** - scany v2 + squirrel for data access
6. **Clean Code** - Separation of concerns, clear boundaries

### Context Map

```
┌─────────────────────────────────────────────────────────────────┐
│                      Business Engine                             │
│                                                                  │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐        │
│  │   Identity   │  │    Audit     │  │ Notifications│        │
│  │  (Existing) │  │  (Existing)  │  │  (Existing)  │        │
│  └──────────────┘  └──────────────┘  └──────────────┘        │
│                                                                  │
│  ┌──────────────────────────────────────────────────────┐      │
│  │                    Parties                            │      │
│  │   Customers, Suppliers, Partners, Employees         │      │
│  └──────────────────────────────────────────────────────┘      │
│                           │                                      │
│                           │ PartyID                              │
│                           ▼                                      │
│  ┌──────────────────────────────────────────────────────┐      │
│  │                   Contracts                           │      │
│  │   Agreements, Terms, Documents                        │      │
│  └──────────────────────────────────────────────────────┘      │
│                           │                                      │
│                           │ ContractID                            │
│                           ▼                                      │
│  ┌──────────────────────────────────────────────────────┐      │
│  │                  Accounting                          │      │
│  │   Accounts, Transactions, Payables, Receivables      │      │
│  └──────────────────────────────────────────────────────┘      │
│                           │                                      │
│                           │ Atomic transactions                   │
│                           ▼                                      │
│  ┌──────────────────────────────────────────────────────┐      │
│  │                   Ordering                           │      │
│  │   Orders, Quotes, Carts, Order Lines                │      │
│  └──────────────────────────────────────────────────────┘      │
│                           │                                      │
│                           │ ItemID                                │
│                           ▼                                      │
│  ┌──────────────────────────────────────────────────────┐      │
│  │                    Catalog                           │      │
│  │   Items, Categories, Prices, Attributes              │      │
│  └──────────────────────────────────────────────────────┘      │
│                                                                  │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐        │
│  │    Files     │  │    Tasks     │  │  EventBus    │        │
│  │  (Existing) │  │  (Existing)  │  │  (Existing)  │        │
│  └──────────────┘  └──────────────┘  └──────────────┘        │
└─────────────────────────────────────────────────────────────────┘
```

## Phase-by-Phase Implementation

### Phase 1: Foundation & Planning (Current) ✅

**Status:** Completed
- ✅ Skeleton repository optimization (scany v2 + squirrel)
- ✅ ADR-016: Database Stack Standard
- ✅ Environment configuration cleanup
- ✅ Documentation suite (45K+ words)
- ✅ PostgreSQL 16 migrations (001-017)
- ✅ CHANGELOG with breaking changes

**Next:** Start Phase 2

---

### Phase 2: Parties Context 🔜

**Estimated Time:** 2-3 days
**Priority:** HIGH (Required by all other contexts)

#### 2.1 Domain Layer

```
internal/parties/domain/
├── party.go                 # Base party interface
├── party_type.go           # CUSTOMER, SUPPLIER, PARTNER, EMPLOYEE
├── customer.go              # Customer aggregate root
├── supplier.go              # Supplier aggregate root
├── partner.go               # Partner aggregate root
├── employee.go              # Employee aggregate root
├── contact_info.go          # Value object: email, phone, address
├── bank_account.go          # Value object: bank details
├── errors.go                # Domain errors
├── events.go                # Domain events
└── repository.go           # Repository interfaces
```

**Key Domain Objects:**

```go
type PartyType string

const (
    PartyTypeCustomer  PartyType = "customer"
    PartyTypeSupplier   PartyType = "supplier"
    PartyTypePartner    PartyType = "partner"
    PartyTypeEmployee   PartyType = "employee"
)

type Party interface {
    GetID() PartyID
    GetType() PartyType
    GetName() string
    GetContactInfo() ContactInfo
    GetTaxID() string
    GetStatus() PartyStatus
}

type Customer struct {
    id              PartyID
    partyType       PartyType     // Always CUSTOMER
    name            string
    taxID           string
    contactInfo     ContactInfo
    bankAccount    BankAccount
    status         PartyStatus
    loyaltyLevel   LoyaltyLevel    // Optional
    totalPurchases Money           // Calculated
    createdAt      time.Time
    updatedAt      time.Time
    events         []domain.Event
}

type Supplier struct {
    id              PartyID
    partyType       PartyType     // Always SUPPLIER
    name            string
    taxID           string
    contactInfo     ContactInfo
    bankAccount    BankAccount
    status         PartyStatus
    rating         SupplierRating
    contracts      []ContractID    // Active contracts
    createdAt      time.Time
    updatedAt      time.Time
    events         []domain.Event
}

type ContactInfo struct {
    Email       string
    Phone       string
    Address     Address
    Website     string
    SocialMedia map[string]string
}
```

#### 2.2 Application Layer

```
internal/parties/application/
├── commands/
│   ├── create_customer.go
│   ├── update_customer.go
│   ├── create_supplier.go
│   ├── update_supplier.go
│   ├── assign_contract.go
│   └── change_status.go
└── queries/
    ├── get_customer.go
    ├── list_customers.go
    ├── get_supplier.go
    ├── list_suppliers.go
    ├── search_parties.go
    └── get_party_stats.go
```

#### 2.3 Infrastructure Layer

```
internal/parties/infrastructure/
└── persistence/
    ├── party_repository.go       # Implements domain repository
    ├── customer_repository.go    # Customer-specific queries
    ├── supplier_repository.go    # Supplier-specific queries
    └── models.go                 # DTOs for scany v2
```

**Repository Pattern (using scany v2 + squirrel):**

```go
type partyDTO struct {
    ID           string         `db:"id"`
    PartyType    string         `db:"party_type"`
    Name         string         `db:"name"`
    TaxID        string         `db:"tax_id"`
    ContactInfo  json.RawMessage `db:"contact_info"`
    BankAccount  json.RawMessage `db:"bank_account"`
    Status      string         `db:"status"`
    Metadata    json.RawMessage `db:"metadata"`
    CreatedAt   time.Time      `db:"created_at"`
    UpdatedAt   time.Time      `db:"updated_at"`
}

type PartyRepository struct {
    pool *pgxpool.Pool
    psql sq.StatementBuilderType
}

func (r *PartyRepository) FindByID(ctx context.Context, id PartyID) (*Party, error) {
    var dto partyDTO
    err := pgxscan.Get(ctx, r.pool, &dto,
        `SELECT * FROM parties WHERE id = $1`, id)
    if err != nil {
        if pgxscan.NotFound(err) {
            return nil, ErrPartyNotFound
        }
        return nil, fmt.Errorf("find party: %w", err)
    }
    return r.dtoToDomain(dto)
}
```

#### 2.4 HTTP Layer

```
internal/parties/ports/http/
├── handler.go                  # HTTP handlers
├── dto.go                      # Request/Response DTOs
├── routes.go                   # Route registration
└── validation.go               # Request validation
```

**API Endpoints:**

```
POST   /api/v1/customers                 # Create customer
GET    /api/v1/customers/:id             # Get customer
GET    /api/v1/customers                 # List customers (paginated)
PUT    /api/v1/customers/:id             # Update customer
DELETE /api/v1/customers/:id             # Deactivate customer

POST   /api/v1/suppliers                 # Create supplier
GET    /api/v1/suppliers/:id             # Get supplier
GET    /api/v1/suppliers                 # List suppliers
PUT    /api/v1/suppliers/:id             # Update supplier
DELETE /api/v1/suppliers/:id             # Deactivate supplier

POST   /api/v1/partners                  # Create partner
GET    /api/v1/partners/:id              # Get partner
GET    /api/v1/partners                  # List partners
PUT    /api/v1/partners/:id              # Update partner

POST   /api/v1/employees                 # Create employee
GET    /api/v1/employees/:id              # Get employee
GET    /api/v1/employees                  # List employees
PUT    /api/v1/employees/:id             # Update employee
```

#### 2.5 Database Migration

```sql
-- migrations/018_parties.up.sql

CREATE TYPE party_type AS ENUM ('customer', 'supplier', 'partner', 'employee');
CREATE TYPE party_status AS ENUM ('active', 'inactive', 'blacklisted');
CREATE TYPE loyalty_level AS ENUM ('bronze', 'silver', 'gold', 'platinum');

CREATE TABLE parties (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    party_type party_type NOT NULL,
    name VARCHAR(255) NOT NULL,
    tax_id VARCHAR(50),
    
    -- Contact info (JSONB for flexibility)
    contact_info JSONB NOT NULL DEFAULT '{}',
    
    -- Bank account (optional)
    bank_account JSONB,
    
    -- Status
    status party_status NOT NULL DEFAULT 'active',
    
    -- Loyalty (for customers)
    loyalty_level loyalty_level,
    total_purchases DECIMAL(15,2) DEFAULT 0,
    
    -- Rating (for suppliers)
    rating JSONB,
    
    -- Metadata
    metadata JSONB DEFAULT '{}',
    
    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    -- Constraints
    CONSTRAINT valid_tax_id CHECK (tax_id IS NULL OR LENGTH(tax_id) >= 8)
);

-- Indexes
CREATE INDEX idx_parties_type ON parties(party_type);
CREATE INDEX idx_parties_status ON parties(status);
CREATE INDEX idx_parties_tax_id ON parties(tax_id) WHERE tax_id IS NOT NULL;
CREATE INDEX idx_parties_name ON parties(name);
CREATE INDEX idx_parties_created ON parties(created_at);

-- GIN index for JSONB
CREATE INDEX idx_parties_contact ON parties USING GIN(contact_info);

-- Updated_at trigger
CREATE OR REPLACE FUNCTION update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER parties_updated_at
    BEFORE UPDATE ON parties
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at();

-- Comments
COMMENT ON TABLE parties IS 'Universal parties: customers, suppliers, partners, employees';
COMMENT ON COLUMN parties.party_type IS 'Type: customer, supplier, partner, employee';
COMMENT ON COLUMN parties.tax_id IS 'Tax identification number (IPN, EDRPOU, etc.)';
COMMENT ON COLUMN parties.total_purchases IS 'Total purchases amount (for customers)';
COMMENT ON COLUMN parties.rating IS 'Supplier rating: quality, delivery speed, price';
```

#### 2.6 Tests

```
internal/parties/
├── domain/
│   ├── customer_test.go
│   ├── supplier_test.go
│   └── contact_info_test.go
├── application/
│   ├── commands/
│   │   ├── create_customer_test.go
│   │   └── create_supplier_test.go
│   └── queries/
│       ├── get_customer_test.go
│       └── list_customers_test.go
└── infrastructure/
    └── persistence/
        ├── party_repository_test.go
        ├── customer_repository_test.go
        └── supplier_repository_test.go
```

#### 2.7 ADR Document

```
docs/adr/ADR-017-parties.md
```

---

### Phase 3: Contracts Context 🔜

**Estimated Time:** 2-3 days
**Priority:** HIGH (Required by Accounting, Ordering)
**Dependencies:** Phase 2 (Parties)

#### 3.1 Domain Layer

```
internal/contracts/domain/
├── contract.go              # Contract aggregate root
├── contract_type.go         # SUPPLY, SERVICE, EMPLOYMENT, PARTNERSHIP
├── contract_status.go       # DRAFT, ACTIVE, EXPIRED, TERMINATED
├── payment_terms.go         # Value object: payment conditions
├── delivery_terms.go        # Value object: delivery conditions
├── document.go              # Attached documents
├── errors.go
├── events.go
└── repository.go
```

**Key Domain Objects:**

```go
type Contract struct {
    id              ContractID
    contractType    ContractType
    status         ContractStatus
    
    // Parties
    partyID        PartyID           // From parties context
    
    // Terms
    paymentTerms   PaymentTerms      // Days, penalties, discounts
    deliveryTerms  DeliveryTerms     // Delivery conditions
    
    // Validity
    validityPeriod DateRange
    
    // Documents
    documents      []DocumentID      // From files context
    
    // Audit
    createdBy      UserID            // From identity context
    createdAt      time.Time
    updatedAt      time.Time
    
    events         []domain.Event
}

type PaymentTerms struct {
    PaymentType    PaymentType    // PREPAID, POSTPAID, CREDIT
    CreditDays     int            // Days before payment due
    PenaltyRate   Percentage     // Late payment penalty
    DiscountRate   Percentage     // Early payment discount
    Currency      Currency
}

type DeliveryTerms struct {
    DeliveryType   DeliveryType   // PICKUP, DELIVERY, DIGITAL
    EstimatedDays   int
    ShippingCost   Money
    Insurance      bool
}
```

#### 3.2 Database Migration

```sql
-- migrations/019_contracts.up.sql

CREATE TYPE contract_type AS ENUM (
    'supply',          -- Supply of goods
    'service',         -- Service provision
    'employment',      -- Employment contract
    'partnership',     -- Partnership agreement
    'lease',          -- Lease/rent
    'license'         -- License agreement
);

CREATE TYPE contract_status AS ENUM (
    'draft',
    'pending_approval',
    'active',
    'expired',
    'terminated'
);

CREATE TYPE payment_type AS ENUM ('prepaid', 'postpaid', 'credit');

CREATE TABLE contracts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    contract_type contract_type NOT NULL,
    status contract_status NOT NULL DEFAULT 'draft',
    
    -- Parties
    party_id UUID NOT NULL REFERENCES parties(id),
    
    -- Terms
    payment_terms JSONB NOT NULL,
    delivery_terms JSONB,
    
    -- Validity
    validity_period DATERANGE NOT NULL,
    
    -- Documents
    documents UUID[] DEFAULT '{}',  -- References files.id
    
    -- Credit limit
    credit_limit DECIMAL(15,2),
    currency VARCHAR(3) DEFAULT 'UAH',
    
    -- Metadata
    metadata JSONB DEFAULT '{}',
    
    -- Audit
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    signed_at TIMESTAMPTZ,
    terminated_at TIMESTAMPTZ
);

CREATE INDEX idx_contracts_party ON contracts(party_id);
CREATE INDEX idx_contracts_status ON contracts(status);
CREATE INDEX idx_contracts_type ON contracts(contract_type);
CREATE INDEX idx_contracts_validity ON contracts USING GIST(validity_period);
CREATE INDEX idx_contracts_active ON contracts(party_id, status) WHERE status = 'active';

-- Trigger for updated_at
CREATE TRIGGER contracts_updated_at
    BEFORE UPDATE ON contracts
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at();

COMMENT ON TABLE contracts IS 'Contracts and agreements between parties';
```

---

### Phase 4: Accounting Context 🔜

**Estimated Time:** 4-5 days
**Priority:** HIGH (Core financial engine)
**Dependencies:** Phase 2 (Parties), Phase 3 (Contracts)

#### 4.1 Domain Layer

```
internal/accounting/domain/
├── account.go               # Chart of accounts
├── account_type.go          # ASSET, LIABILITY, EQUITY, REVENUE, EXPENSE
├── transaction.go           # Double-entry transaction
├── journal.go               # Journal entry
├── invoice.go               # Invoice (in/out)
├── payment.go               # Payment
├── payable.go               # Money owed to suppliers
├── receivable.go            # Money owed by customers
├── money.go                 # Value object: amount + currency
├── currency.go              # Currency codes
├── errors.go
├── events.go
└── repository.go
```

**Key Domain Objects:**

```go
type Account struct {
    id          AccountID
    code        string         // Account code (e.g., "1010")
    name        string
    accountType AccountType    // ASSET, LIABILITY, EQUITY, REVENUE, EXPENSE
    currency    Currency
    balance     Money
    parentID    AccountID      // Hierarchy
    isActive    bool
}

type Transaction struct {
    id            TransactionID
    fromAccount   AccountID      // Credit
    toAccount     AccountID      // Debit
    amount        Money
    currency      Currency
    reference     Reference      // Invoice, Payment, Order, etc.
    description   string
    occurredAt    time.Time
    postedAt      time.Time
    postedBy      UserID         // From identity context
}

type Payable struct {
    id              PayableID
    supplierID      PartyID        // From parties context
    contractID      ContractID     // From contracts context
    invoiceID       InvoiceID
    invoiceNumber   string
    invoiceDate     time.Time
    
    amount          Money          // Total amount
    taxAmount      Money          // VAT/Tax
    dueAmount      Money          // Amount to pay
    
    dueDate        time.Time
    paidAmount    Money          // Amount already paid
    paymentStatus PaymentStatus   // UNPAID, PARTIAL, PAID, OVERDUE
    
    documents      []DocumentID   // From files context
    
    createdAt      time.Time
    updatedAt      time.Time
}
```

#### 4.2 Database Migration

```sql
-- migrations/020_accounting.up.sql

-- Chart of Accounts
CREATE TYPE account_type AS ENUM (
    'asset',       -- Активи
    'liability',   -- Пасиви
    'equity',      -- Капітал
    'revenue',     -- Доходи
    'expense'      -- Витрати
);

CREATE TABLE accounts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    code VARCHAR(20) UNIQUE NOT NULL,  -- Account code
    name VARCHAR(255) NOT NULL,
    account_type account_type NOT NULL,
    currency VARCHAR(3) DEFAULT 'UAH',
    balance DECIMAL(15,2) DEFAULT 0,
    parent_id UUID REFERENCES accounts(id),
    is_active BOOLEAN NOT NULL DEFAULT true,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_accounts_type ON accounts(account_type);
CREATE INDEX idx_accounts_parent ON accounts(parent_id);

-- Transactions (Double-entry)
CREATE TABLE transactions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    from_account UUID NOT NULL REFERENCES accounts(id),  -- Credit
    to_account UUID NOT NULL REFERENCES accounts(id),    -- Debit
    amount DECIMAL(15,2) NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'UAH',
    
    -- Reference (what is this transaction for?)
    reference_type VARCHAR(50),  -- 'invoice', 'payment', 'order', etc.
    reference_id UUID,
    
    description TEXT,
    occurred_at TIMESTAMPTZ NOT NULL,
    posted_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    posted_by UUID REFERENCES users(id),
    
    metadata JSONB DEFAULT '{}'
);

CREATE INDEX idx_transactions_from ON transactions(from_account);
CREATE INDEX idx_transactions_to ON transactions(to_account);
CREATE INDEX idx_transactions_date ON transactions(occurred_at);
CREATE INDEX idx_transactions_reference ON transactions(reference_type, reference_id);

-- Invoices
CREATE TYPE invoice_status AS ENUM ('draft', 'sent', 'paid', 'overdue', 'cancelled');
CREATE TYPE invoice_direction AS ENUM ('incoming', 'outgoing');

CREATE TABLE invoices (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    invoice_number VARCHAR(50) NOT NULL,
    direction invoice_direction NOT NULL,  -- incoming from supplier / outgoing to customer
    
    -- Party
    party_id UUID NOT NULL REFERENCES parties(id),
    contract_id UUID REFERENCES contracts(id),
    
    -- Amounts
    subtotal DECIMAL(15,2) NOT NULL,
    tax_amount DECIMAL(15,2) DEFAULT 0,
    total DECIMAL(15,2) NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'UAH',
    
    -- Dates
    invoice_date DATE NOT NULL,
    due_date DATE NOT NULL,
    
    -- Status
    status invoice_status NOT NULL DEFAULT 'draft',
    
    -- Documents
    documents UUID[] DEFAULT '{}',
    
    -- Audit
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_invoices_party ON invoices(party_id);
CREATE INDEX idx_invoices_status ON invoices(status);
CREATE INDEX idx_invoices_due ON invoices(due_date);

-- Payables (Money owed to suppliers)
CREATE TYPE payment_status AS ENUM ('unpaid', 'partially_paid', 'paid', 'overdue');

CREATE TABLE payables (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    supplier_id UUID NOT NULL REFERENCES parties(id),
    contract_id UUID REFERENCES contracts(id),
    invoice_id UUID REFERENCES invoices(id),
    
    -- Amounts
    amount DECIMAL(15,2) NOT NULL,
    tax_amount DECIMAL(15,2) DEFAULT 0,
    due_amount DECIMAL(15,2) NOT NULL,
    paid_amount DECIMAL(15,2) DEFAULT 0,
    currency VARCHAR(3) NOT NULL DEFAULT 'UAH',
    
    -- Dates
    invoice_date DATE NOT NULL,
    due_date DATE NOT NULL,
    
    -- Status
    payment_status payment_status NOT NULL DEFAULT 'unpaid',
    
    -- Documents
    documents UUID[] DEFAULT '{}',
    
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_payables_supplier ON payables(supplier_id);
CREATE INDEX idx_payables_status ON payables(payment_status);
CREATE INDEX idx_payables_due ON payables(due_date);

-- Receivables (Money owed by customers)
CREATE TABLE receivables (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    customer_id UUID NOT NULL REFERENCES parties(id),
    contract_id UUID REFERENCES contracts(id),
    invoice_id UUID REFERENCES invoices(id),
    
    -- Amounts
    amount DECIMAL(15,2) NOT NULL,
    tax_amount DECIMAL(15,2) DEFAULT 0,
    due_amount DECIMAL(15,2) NOT NULL,
    paid_amount DECIMAL(15,2) DEFAULT 0,
    currency VARCHAR(3) NOT NULL DEFAULT 'UAH',
    
    -- Dates
    invoice_date DATE NOT NULL,
    due_date DATE NOT NULL,
    
    -- Status
    payment_status payment_status NOT NULL DEFAULT 'unpaid',
    
    -- Documents
    documents UUID[] DEFAULT '{}',
    
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_receivables_customer ON receivables(customer_id);
CREATE INDEX idx_receivables_status ON receivables(payment_status);
CREATE INDEX idx_receivables_due ON receivables(due_date);

-- Payments
CREATE TYPE payment_method AS ENUM ('cash', 'bank_transfer', 'card', 'check');

CREATE TABLE payments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    payable_id UUID REFERENCES payables(id),
    receivable_id UUID REFERENCES receivables(id),
    invoice_id UUID REFERENCES invoices(id),
    
    -- Amount
    amount DECIMAL(15,2) NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'UAH',
    
    -- Method
    payment_method payment_method NOT NULL,
    payment_date DATE NOT NULL,
    
    -- Reference
    transaction_id UUID REFERENCES transactions(id),
    reference_number VARCHAR(100),
    
    -- Audit
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_payments_payable ON payments(payable_id);
CREATE INDEX idx_payments_receivable ON payments(receivable_id);
CREATE INDEX idx_payments_date ON payments(payment_date);

-- Comments
COMMENT ON TABLE accounts IS 'Chart of accounts (Plan рахунків)';
COMMENT ON TABLE transactions IS 'Double-entry bookkeeping transactions';
COMMENT ON TABLE invoices IS 'Invoices (incoming from suppliers, outgoing to customers)';
COMMENT ON TABLE payables IS 'Accounts payable (Кредиторська заборгованість)';
COMMENT ON TABLE receivables IS 'Accounts receivable (Дебіторська заборгованість)';
COMMENT ON TABLE payments IS 'Payment records';
```

---

### Phase 5: Ordering Context 🔜

**Estimated Time:** 3-4 days
**Priority:** HIGH
**Dependencies:** Phase 2 (Parties), Phase 3 (Contracts), Phase 4 (Accounting)

#### 5.1 Domain Layer

```
internal/ordering/domain/
├── order.go                 # Base order aggregate
├── order_line.go            # Order line
├── order_status.go          # DRAFT, PENDING, CONFIRMED, COMPLETED, CANCELLED
├── quote.go                 # Price quote
├── cart.go                  # Shopping cart
├── errors.go
├── events.go
└── repository.go
```

#### 5.2 Database Migration

```sql
-- migrations/021_ordering.up.sql

CREATE TYPE order_status AS ENUM (
    'draft',
    'pending',
    'confirmed',
    'processing',
    'completed',
    'cancelled',
    'refunded'
);

CREATE TABLE orders (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    order_number VARCHAR(50) UNIQUE NOT NULL,
    
    -- Parties
    customer_id UUID NOT NULL REFERENCES parties(id),
    supplier_id UUID NOT NULL REFERENCES parties(id),
    contract_id UUID REFERENCES contracts(id),
    
    -- Amounts
    subtotal DECIMAL(15,2) NOT NULL,
    tax_amount DECIMAL(15,2) DEFAULT 0,
    discount DECIMAL(15,2) DEFAULT 0,
    total DECIMAL(15,2) NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'UAH',
    
    -- Status
    status order_status NOT NULL DEFAULT 'draft',
    
    -- Dates
    order_date TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    due_date TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    cancelled_at TIMESTAMPTZ,
    
    -- Notes
    notes TEXT,
    
    -- Metadata
    metadata JSONB DEFAULT '{}',
    
    -- Audit
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_orders_customer ON orders(customer_id);
CREATE INDEX idx_orders_supplier ON orders(supplier_id);
CREATE INDEX idx_orders_status ON orders(status);
CREATE INDEX idx_orders_date ON orders(order_date);

CREATE TABLE order_lines (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    order_id UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    
    -- Item
    item_id UUID NOT NULL,  -- Reference to catalog item
    item_name VARCHAR(255) NOT NULL,
    
    -- Quantities
    quantity DECIMAL(10,2) NOT NULL,
    unit VARCHAR(20),  -- 'piece', 'kg', 'hour', etc.
    
    -- Pricing
    unit_price DECIMAL(15,2) NOT NULL,
    discount DECIMAL(15,2) DEFAULT 0,
    total DECIMAL(15,2) NOT NULL,
    
    -- Metadata
    metadata JSONB DEFAULT '{}',
    
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_order_lines_order ON order_lines(order_id);

-- Quotes (price quotes)
CREATE TYPE quote_status AS ENUM ('draft', 'sent', 'accepted', 'rejected', 'expired');

CREATE TABLE quotes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    quote_number VARCHAR(50) UNIQUE NOT NULL,
    
    -- Parties
    customer_id UUID NOT NULL REFERENCES parties(id),
    supplier_id UUID NOT NULL REFERENCES parties(id),
    
    -- Amounts
    subtotal DECIMAL(15,2) NOT NULL,
    tax_amount DECIMAL(15,2) DEFAULT 0,
    total DECIMAL(15,2) NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'UAH',
    
    -- Validity
    valid_from DATE NOT NULL,
    valid_until DATE NOT NULL,
    
    -- Status
    status quote_status NOT NULL DEFAULT 'draft',
    
    -- Converted to order
    order_id UUID REFERENCES orders(id),
    
    -- Notes
    notes TEXT,
    
    -- Metadata
    metadata JSONB DEFAULT '{}',
    
    -- Audit
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_quotes_customer ON quotes(customer_id);
CREATE INDEX idx_quotes_status ON quotes(status);

-- Comments
COMMENT ON TABLE orders IS 'Universal orders: purchase orders, sales orders, bookings, appointments';
COMMENT ON TABLE order_lines IS 'Order line items';
COMMENT ON TABLE quotes IS 'Price quotes (commercial proposals)';
```

---

### Phase 6: Catalog Context 🔜

**Estimated Time:** 2-3 days
**Priority:** MEDIUM
**Dependencies:** None (can be developed in parallel)

#### 6.1 Domain Layer

```
internal/catalog/domain/
├── item.go                 # Base catalog item (abstract)
├── item_type.go           # PRODUCT, SERVICE, PROPERTY
├── category.go            # Categories
├── price.go               # Pricing
├── attribute.go            # Dynamic attributes
├── errors.go
├── events.go
└── repository.go
```

#### 6.2 Database Migration

```sql
-- migrations/022_catalog.up.sql

CREATE TYPE item_status AS ENUM ('active', 'inactive', 'discontinued');

CREATE TABLE catalog_categories (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    parent_id UUID REFERENCES catalog_categories(id),
    path LTREE,  -- Hierarchical path
    is_active BOOLEAN NOT NULL DEFAULT true,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_categories_parent ON catalog_categories(parent_id);
CREATE INDEX idx_categories_path ON catalog_categories USING GIST(path);

CREATE TABLE catalog_items (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    category_id UUID REFERENCES catalog_categories(id),
    
    -- Basic info
    sku VARCHAR(100) UNIQUE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    
    -- Pricing
    base_price DECIMAL(15,2) NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'UAH',
    
    -- Status
    status item_status NOT NULL DEFAULT 'active',
    
    -- Attributes (flexible JSONB)
    attributes JSONB DEFAULT '{}',
    
    -- Metadata
    metadata JSONB DEFAULT '{}',
    
    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_items_category ON catalog_items(category_id);
CREATE INDEX idx_items_sku ON catalog_items(sku);
CREATE INDEX idx_items_status ON catalog_items(status);
CREATE INDEX idx_items_attrs ON catalog_items USING GIN(attributes);

-- Prices (support for multiple prices per item)
CREATE TYPE price_type AS ENUM ('base', 'sale', 'wholesale', 'partner');

CREATE TABLE catalog_prices (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    item_id UUID NOT NULL REFERENCES catalog_items(id) ON DELETE CASCADE,
    price_type price_type NOT NULL DEFAULT 'base',
    amount DECIMAL(15,2) NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'UAH',
    valid_from DATE NOT NULL DEFAULT CURRENT_DATE,
    valid_until DATE,
    is_active BOOLEAN NOT NULL DEFAULT true,
    
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_prices_item ON catalog_prices(item_id);
CREATE INDEX idx_prices_type ON catalog_prices(price_type);
CREATE INDEX idx_prices_valid ON catalog_prices(valid_from, valid_until);

-- Comments
COMMENT ON TABLE catalog_categories IS 'Taxonomy: product/service categories';
COMMENT ON TABLE catalog_items IS 'Catalog items: products, services, properties';
COMMENT ON COLUMN catalog_items.attributes IS 'Flexible attributes per item type';
COMMENT ON TABLE catalog_prices IS 'Multiple prices per item: base, sale, wholesale';
```

---

## Implementation Sequence

### Recommended Order

```
Phase 2: Parties (2-3 days)
    ↓
Phase 3: Contracts (2-3 days)
    ↓
Phase 4: Accounting (4-5 days)
    ↓
Phase 5: Ordering (3-4 days)
    ↓
Phase 6: Catalog (2-3 days)

Total: 13-18 days
```

### Parallel Development

Some phases can be developed in parallel:

```
Phase 2: Parties ────────────────┐
                                   ↓
Phase 3: Contracts ─────────────┼── Phase 4: Accounting
                                   ↓
Phase 6: Catalog (independent) ──┘
                                   ↓
                              Phase 5: Ordering
```

---

## Testing Strategy

### Unit Tests

- **Domain layer**: Entity creation, business rules, invariants
- **Value objects**: Money operations, DateRange, ContactInfo

### Integration Tests

- **Repository layer**: Database operations, queries
- **Use Testcontainers** for PostgreSQL 16

### End-to-End Tests

- **HTTP layer**: API endpoints, authentication
- **Event flow**: Domain events propagation

### Test Coverage Goals

- Domain: 100%
- Application: 80%+
- Infrastructure: 80%+
- HTTP: 70%+

---

## Documentation Requirements

For each context, create:

1. **ADR (Architecture Decision Record)**
   - Why this context exists
   - Key design decisions
   - Trade-offs

2. **API Documentation**
   - OpenAPI/Swagger specs
   - Request/Response examples

3. **Developer Guide**
   - How to use the context
   - Code examples
   - Common patterns

4. **Database Diagram**
   - ERD for context tables
   - Relationships

---

## Success Criteria

### Phase 2 Success
- ✅ Can create/read/update/delete parties
- ✅ Can search parties by type, name, tax_id
- ✅ Parties can have multiple contact methods
- ✅ Parties can be linked to contracts

### Phase 3 Success
- ✅ Can create contracts with payment/delivery terms
- ✅ Can track contract lifecycle (draft→active→expired)
- ✅ Contracts are linked to parties
- ✅ Can attach documents to contracts

### Phase 4 Success
- ✅ Double-entry bookkeeping works
- ✅ Can record transactions between accounts
- ✅ Can create invoices (incoming/outgoing)
- ✅ Can track payables and receivables
- ✅ Can record payments
- ✅ Balance sheet calculates correctly

### Phase 5 Success
- ✅ Can create orders with multiple lines
- ✅ Orders are linked to parties and contract
- ✅ Can create quotes and convert to orders
- ✅ Order status transitions work correctly
- ✅ Can calculate totals, taxes, discounts

### Phase 6 Success
- ✅ Can create catalog categories (hierarchical)
- ✅ Can create catalog items with attributes
- ✅ Can set multiple prices per item
- ✅ Can search items by category, attributes, price
- ✅ Items can be linked to orders

---

## Migration Path for Existing Businesses

### Food Store
```
business-engine (skeleton)
    ├── internal/
    │   ├── parties/       ✅ Universal
    │   ├── contracts/      ✅ Universal
    │   ├── accounting/     ✅ Universal
    │   ├── ordering/       ✅ Universal
    │   ├── catalog/        ✅ Universal
    │   └── food/          🆕 Business-specific
    │       ├── domain/
    │       │   ├── product.go      (extends catalog.Item)
    │       │   ├── inventory.go
    │       │   ├── expiration.go
    │       │   └── warehouse.go
    │       └── ...
```

### Real Estate Agency
```
business-engine (skeleton)
    ├── internal/
    │   ├── parties/       ✅ Universal
    │   ├── contracts/      ✅ Universal
    │   ├── accounting/     ✅ Universal
    │   ├── ordering/       ✅ Universal (deals)
    │   ├── catalog/        ✅ Universal (properties)
    │   └── realestate/    🆕 Business-specific
    │       ├── domain/
    │       │   ├── property.go    (extends catalog.Item)
    │       │   ├── viewing.go
    │       │   ├── deal.go
    │       │   └── agent.go
    │       └── ...
```

### Travel Agency
```
business-engine (skeleton)
    ├── internal/
    │   ├── parties/       ✅ Universal
    │   ├── contracts/      ✅ Universal
    │   ├── accounting/     ✅ Universal
    │   ├── ordering/       ✅ Universal (bookings)
    │   ├── catalog/        ✅ Universal (tours)
    │   └── travel/         🆕 Business-specific
    │       ├── domain/
    │       │   ├── tour.go       (extends catalog.Item)
    │       │   ├── booking.go
    │       │   ├── itinerary.go
    │       │   └── traveler.go
    │       └── ...
```

### Healthcare
```
business-engine (skeleton)
    ├── internal/
    │   ├── parties/       ✅ Universal
    │   ├── contracts/      ✅ Universal
    │   ├── accounting/     ✅ Universal
    │   ├── ordering/       ✅ Universal (appointments)
    │   ├── catalog/        ✅ Universal (services)
    │   └── healthcare/     🆕 Business-specific
    │       ├── domain/
    │       │   ├── service.go    (extends catalog.Item)
    │       │   ├── appointment.go
    │       │   ├── patient.go     (extends parties.Customer)
    │       │   ├── doctor.go      (extends parties.Employee)
    │       │   ├── diagnosis.go
    │       │   └── prescription.go
    │       └── ...
```

---

## Next Steps

1. **Review this plan** and approve
2. **Start Phase 2: Parties** implementation
3. **Follow the sequence**: Parties → Contracts → Accounting → Ordering → Catalog
4. **Test each phase** before moving to next
5. **Update documentation** as we go

---

## Notes

- All migrations use UUID v7
- All JSONB columns have GIN indexes
- All currency amounts are DECIMAL(15,2)
- All tables have created_at/updated_at
- All entities use domain events for cross-context communication
- All repositories use scany v2 + squirrel pattern
- All HTTP handlers use the same structure as skeleton

---

**Ready to start Phase 2: Parties?** 🚀