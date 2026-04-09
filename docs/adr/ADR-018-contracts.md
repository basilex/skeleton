# ADR-018: Contracts Bounded Context

## Status
Accepted

## Context
Contract management between parties with validity periods, payment terms, and delivery terms.

## Decision
Implement Contracts as a bounded context with DATERANGE for validity periods.

### Domain Model
- **Contract**: Agreement between parties
- **PaymentTerms**: Prepaid/credit payment conditions
- **DeliveryTerms**: Delivery conditions (INCOTERMS)
- **ContractType**: Framework, purchase, service contracts
- **ContractStatus**: Draft, active, suspended, terminated

### Architecture
```
internal/contracts/
├── domain/
│   ├── contract.go          # Contract aggregate
│   ├── payment_terms.go     # Value object
│   ├── delivery_terms.go    # Value object
│   ├── ids.go                # Identifiers
│   ├── events.go             # Domain events
│   └── repository.go         # Repository interfaces
├── infrastructure/
│   └── persistence/
│       ├── models.go
│       └── contract_repository.go
├── application/
│   ├── command/
│   │   ├── create_contract.go
│   │   ├── activate_contract.go
│   │   └── terminate_contract.go
│   └── query/
│       └── contract.go
└── ports/http/
    ├── handler.go
    └── dto.go
```

### Database Schema
```sql
CREATE TYPE contract_type AS ENUM (
    'framework',     -- Framework agreement
    'purchase',      -- Purchase contract
    'service',       -- Service agreement
    'rental',        -- Rental contract
    'licensing'      -- License agreement
);

CREATE TYPE contract_status AS ENUM (
    'draft',
    'active',
    'suspended',
    'terminated'
);

CREATE TYPE payment_type AS ENUM (
    'prepaid',       -- Prepayment required
    'credit_7',      -- Net 7 days
    'credit_15',     -- Net 15 days
    'credit_30',     -- Net 30 days
    'credit_45',     -- Net 45 days
    'credit_60',     -- Net 60 days
    'credit_90'      -- Net 90 days
);

CREATE TABLE contracts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    contract_number VARCHAR(100) UNIQUE NOT NULL,
    contract_type contract_type NOT NULL,
    
    -- Parties
    party_id UUID NOT NULL,           -- References parties.id
    counterparty_id UUID,             -- For bilateral contracts
    
    -- Terms
    validity_period DATERANGE NOT NULL,
    payment_terms JSONB NOT NULL,
    delivery_terms JSONB,
    
    -- Status
    status contract_status NOT NULL DEFAULT 'draft',
    
    -- Amounts
    total_value DECIMAL(15,2),
    currency VARCHAR(3) DEFAULT 'UAH',
    
    -- Metadata
    metadata JSONB DEFAULT '{}',
    
    -- Audit
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    EXCLUDE USING GIST (
        validity_period WITH &&,
        CAST(party_id AS TEXT) WITH =
    )
);
```

### Key Design Decisions

#### 1. DATERANGE for Validity Periods
```go
// Native PostgreSQL date range type
validity_period DATERANGE NOT NULL

// Date range operations
SELECT * FROM contracts 
WHERE validity_period && '[2024-01-01, 2024-12-31]'::daterange;

// Overlap detection built-in
SELECT * FROM contracts 
WHERE party_id = $1 
  AND validity_period && $2::daterange;
```

**Benefits:**
- ✅ Native overlap detection
- ✅ Temporal queries optimized
- ✅ Prevents contract date conflicts
- ✅ Clean expiration queries

#### 2. Payment Terms as Value Object
```go
type PaymentTerms struct {
    Type     PaymentType
    Days     int
    Currency Currency
}

// Prepaid: full payment before delivery
// Credit N: payment within N days after delivery
```

#### 3. Contract State Machine
```
Draft → Active → Suspended → Active
                ↓
            Terminated
```

```go
// State transitions
func (c *Contract) Activate() error {
    if c.status != ContractStatusDraft {
        return ErrContractCannotActivate
    }
    c.status = ContractStatusActive
    return nil
}

func (c *Contract) Terminate() error {
    if c.status == ContractStatusTerminated {
        return ErrContractAlreadyTerminated
    }
    c.status = ContractStatusTerminated
    return nil
}
```

### API Endpoints

```
POST   /api/v1/contracts                # Create contract
GET    /api/v1/contracts/:id             # Get contract
GET    /api/v1/contracts                # List contracts (paginated)
PUT    /api/v1/contracts/:id/activate   # Activate contract
PUT    /api/v1/contracts/:id/terminate  # Terminate contract
```

### Usage Example

```go
// Create contract
POST /api/v1/contracts
{
    "contract_number": "CTR-2024-001",
    "contract_type": "purchase",
    "party_id": "019f1234-...",
    "validity_period": {
        "start": "2024-01-01",
        "end": "2024-12-31"
    },
    "payment_terms": {
        "type": "credit_30",
        "currency": "UAH"
    },
    "delivery_terms": {
        "incoterms": "DAP",
        "delivery_days": 14
    }
}

// Find active contracts for date
GET /api/v1/contracts?status=active&date=2024-06-15

// Find overlapping contracts
GET /api/v1/contracts?party_id=xxx&overlap=true
```

### DATERANGE Benefits

#### Temporal Queries
```sql
-- Active contracts on specific date
SELECT * FROM contracts 
WHERE validity_period @> '2024-06-15'::date;

-- Contracts expiring in next 30 days
SELECT * FROM contracts 
WHERE upper(validity_period) <= CURRENT_DATE + 30;

-- Conflicting contracts for same party
SELECT * FROM contracts c1, contracts c2
WHERE c1.party_id = c2.party_id
  AND c1.id != c2.id
  AND c1.validity_period && c2.validity_period;
```

### Consequences

#### Positive
- ✅ Native date range operations
- ✅ Constraint checking at database level
- ✅ Optimized temporal queries
- ✅ Prevents contract overlaps
- ✅ Clean expiration handling

#### Negative
- ⚠️ PostgreSQL-specific (not portable)
- ⚠️ Requires EXCLUDE constraint for overlaps
- ⚠️ Complex date range parsing in Go

### Integration Points

1. **Parties Context**: Party validation
2. **Ordering Context**: Contract reference in orders
3. **Accounting Context**: Payables/receivables linked
4. **Audit Context**: Contract lifecycle events

### Performance Considerations

- GiST index on `validity_period` for fast range queries
- Composite index on `(party_id, validity_period)`
- Index on `(status, validity_period)` for active contracts
- Consider partial index for active contracts only

### References
- PostgreSQL DATERANGE documentation
- Temporal data patterns
- Contract lifecycle management
