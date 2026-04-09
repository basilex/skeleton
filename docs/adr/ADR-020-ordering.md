# ADR-020: Ordering Bounded Context

## Status
Accepted

## Context
Universal order management system for B2B/B2C with order lines, quotes, and state machine.

## Decision
Implement Ordering as a bounded context with order state machine and order lines.

### Domain Model
- **Order**: Order aggregate with state machine
- **OrderLine**: Order line value object
- **Quote**: Price quote before order
- **OrderStatus**: Draft, Pending, Confirmed, Processing, Completed, Cancelled, Refunded

### Architecture
```
internal/ordering/
├── domain/
│   ├── order.go            # Order aggregate with lines
│   ├── order_status.go     # Status enum
│   ├── ids.go               # Identifiers
│   ├── errors.go            # Domain errors
│   ├── events.go            # Domain events
│   └── repository.go         # Repository interfaces
├── infrastructure/
│   └── persistence/
│       ├── models.go
│       └── order_repository.go
├── application/
│   ├── command/
│   │   ├── create_order.go
│   │   ├── add_order_line.go
│   │   └── update_order_status.go
│   └── query/
│       ├── order.go
│       └── list_orders.go
└── ports/http/
    ├── handler.go
    └── dto.go
```

### Database Schema

#### Orders
```sql
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
    
    -- Parties (references parties.id from parties context)
    customer_id UUID NOT NULL,
    supplier_id UUID NOT NULL,
    contract_id UUID,             -- References contracts.id
    
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
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    CONSTRAINT valid_total CHECK (total >= 0)
);
```

#### Order Lines
```sql
CREATE TABLE order_lines (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    order_id UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    
    -- Item
    item_id UUID NOT NULL,         -- References catalog_items.id
    item_name VARCHAR(255) NOT NULL,
    
    -- Quantities
    quantity DECIMAL(10,2) NOT NULL,
    unit VARCHAR(20),               -- 'piece', 'kg', 'hour', etc.
    
    -- Pricing
    unit_price DECIMAL(15,2) NOT NULL,
    discount DECIMAL(15,2) DEFAULT 0,
    total DECIMAL(15,2) NOT NULL,
    
    -- Metadata
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    CONSTRAINT valid_quantity CHECK (quantity > 0),
    CONSTRAINT valid_unit_price CHECK (unit_price >= 0)
);
```

### Key Design Decisions

#### 1. Order State Machine
```
            ┌──────┐
            │Draft │
            └──┬───┘
               │ Confirm
            ┌──▼───┐
            │Pending│
            └──┬───┘
               │ Process
            ┌──▼──────┐
            │Processing│
            └──┬──────┘
               │ Complete
            ┌──▼──────┐
            │Completed│
            └────────┘
            
            Any ──Cancel──► Cancelled
```

```go
func (o *Order) Confirm() error {
    if o.status != OrderStatusDraft && o.status != OrderStatusPending {
        return fmt.Errorf("cannot confirm order in %s status", o.status)
    }
    if len(o.lines) == 0 {
        return fmt.Errorf("cannot confirm order without lines")
    }
    o.status = OrderStatusConfirmed
    return nil
}

func (o *Order) Complete() error {
    if o.status != OrderStatusConfirmed && o.status != OrderStatusProcessing {
        return ErrOrderCannotComplete
    }
    o.status = OrderStatusCompleted
    now := time.Now().UTC()
    o.completedAt = &now
    return nil
}

func (o *Order) Cancel(reason string) error {
    if o.status == OrderStatusCompleted || o.status == OrderStatusCancelled {
        return ErrOrderCannotCancel
    }
    o.status = OrderStatusCancelled
    now := time.Now().UTC()
    o.cancelledAt = &now
    o.notes = reason
    return nil
}
```

#### 2. Order Lines as Value Objects
```go
type OrderLine struct {
    id        OrderLineID
    orderID   OrderID
    itemID    string
    itemName  string
    quantity  float64
    unit      string
    unitPrice float64
    discount  float64
    total     float64
}

// Automatically calculated on line creation
total = (quantity * unitPrice) - discount

// Lines managed within Order aggregate
func (o *Order) AddLine(line *OrderLine) error {
    if o.status != OrderStatusDraft {
        return fmt.Errorf("cannot add lines to order in %s status", o.status)
    }
    o.lines = append(o.lines, line)
    o.recalculateTotals()
    return nil
}
```

#### 3. Order Number Generation
```go
// Sequential order numbers
ORD-2024-000001
ORD-2024-000002
ORD-2024-000003

// In application layer
orderNumber := fmt.Sprintf("ORD-%d-%06d", time.Now().Year(), sequence)
```

### API Endpoints

```
POST   /api/v1/orders                # Create order
GET    /api/v1/orders/:id             # Get order with lines
GET    /api/v1/orders                # List orders (paginated)
POST   /api/v1/orders/:id/lines      # Add order line
PUT    /api/v1/orders/:id/status     # Update order status
```

### Usage Example

```go
// Create draft order
POST /api/v1/orders
{
    "order_number": "ORD-2024-001",
    "customer_id": "019f1234-...",
    "supplier_id": "019f5678-...",
    "currency": "UAH"
}

// Add line items
POST /api/v1/orders/{id}/lines
{
    "item_id": "catalog-item-id",
    "item_name": "Product Name",
    "quantity": 10,
    "unit": "piece",
    "unit_price": 150.00,
    "discount": 0
}

// Confirm order
PUT /api/v1/orders/{id}/status
{
    "status": "confirmed"
}

// Complete order
PUT /api/v1/orders/{id}/status
{
    "status": "completed"
}

// Cancel order
PUT /api/v1/orders/{id}/status
{
    "status": "cancelled",
    "reason": "Customer request"
}

// Query orders
GET /api/v1/orders?customer_id=xxx&status=completed&start_date=2024-01-01
```

### Order Lifecycle

#### Draft State
- Order created
- Lines can be added/removed
- Totals recalculated automatically
- No party notifications

#### Confirmed State
- Order confirmed by customer
- Lines locked (no changes)
- Supplier notification sent
- Inventory reserved

#### Processing State
- Supplier processing order
- Fulfillment in progress
- Status updates tracked

#### Completed State
- Order fulfilled
- Invoice generated
- Transaction recorded (Accounting context)
- Stock updated (Catalog context)

#### Cancelled State
- Order cancelled
- Lines removed
- Inventory released
- Cancellation reason recorded

### Consequences

#### Positive
- ✅ Clear state machine with validation
- ✅ Automatic total calculation
- ✅ Integration-ready with Parties, Contracts, Accounting, Catalog
- ✅ Line-level tracking with items
- ✅ Audit trail for all status changes

#### Negative
- ⚠️ No modification after confirmation
- ⚠️ Requires manual order number generation
- ⚠️ No partial fulfillment
- ⚠️ No return/refund flow (yet)

### Integration Points

1. **Parties Context**: Customer/Supplier validation
2. **Contracts Context**: Contract reference for terms
3. **Catalog Context**: Item validation, stock reservation
4. **Accounting Context**: Transaction generation on completion
5. **Notifications Context**: Status change notifications

### Performance Considerations

- Composite index on `(customer_id, status)`
- Composite index on `(supplier_id, status)`
- Index on `order_date` for date range queries
- Consider partitioning by `order_date` for large volumes

### Future Enhancements

1. **Quotes**: Convert quote to order
2. **Recurring Orders**: Scheduled recurring orders
3. **Partial Fulfillment**: Split shipments
4. **Returns**: Return order flow
5. **Discounts**: Line-level and order-level discounts
6. **Approvals**: Multi-level approval workflow

### References
- Order management patterns
- State machine design
- DDD aggregate design
