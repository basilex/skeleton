# Cross-Context Integration

This document describes the event-driven architecture that enables communication between bounded contexts in the Skeleton Business Engine.

## Overview

The system uses domain events to implement cross-context integration following the principles of Domain-Driven Design (DDD). Each bounded context publishes domain events when significant state changes occur, and other contexts can subscribe to these events to trigger their own business logic.

## Architecture

### Cross-Context Communication Flow

**Order Confirmed triggers:**
1. → **Inventory Context**: Reserve stock
2. → **Invoicing Context**: Create invoice
   - → **Accounting Context**: Create journal entry

### Communication Pattern

| Source Context | Event | Target Context | Action |
|---------------|-------|----------------|--------|
| Ordering | `OrderConfirmed` | Inventory | Reserve stock quantities |
| Ordering | `OrderConfirmed` | Invoicing | Create invoice automatically |
| Ordering | `OrderCancelled` | Inventory | Release stock reservations |
| Ordering | `OrderCompleted` | Inventory | Fulfill reservations, adjust stock |
| Invoicing | `InvoiceCreated` | Accounting | Create journal entries |

## Event Flow

### 1. Order → Inventory Integration

**Trigger:** Order status changes to `Confirmed`

**Event:** `OrderConfirmed`

**Handler:** `inventory/infrastructure/eventhandler/order_handler.go`

**Actions:**
1. Find stock for each order line item
2. Reserve quantity from available stock
3. Create `StockReservation` record linked to order
4. Publish `StockReserved` event (future)

**Handler Methods:**
```go
HandleOrderConfirmed(ctx, event)   // Reserve stock
HandleOrderCancelled(ctx, event)    // Release reservations
HandleOrderCompleted(ctx, event)    // Fulfill reservations
```

**Business Rules:**
- Stock is reserved when order is confirmed
- Reservations are released when order is cancelled
- Reservations are fulfilled when order is completed
- Reserved quantity is subtracted from available quantity

### 2. Order → Invoice Integration

**Trigger:** Order status changes to `Confirmed`

**Event:** `OrderConfirmed`

**Handler:** `invoicing/infrastructure/eventhandler/order_handler.go`

**Actions:**
1. Generate invoice number (INV-{timestamp})
2. Create invoice from order data
3. Add invoice lines from order lines
4. Link invoice to order
5. Set 30-day payment due date
6. Publish `InvoiceCreated` event

**Handler Methods:**
```go
HandleOrderConfirmed(ctx, event)   // Create invoice
```

**Business Rules:**
- Invoice is automatically created when order is confirmed
- Invoice inherits customer, currency, and lines from order
- Payment due date is set to 30 days from creation
- Invoice starts in `Draft` status

### 3. Invoice → Accounting Integration

**Trigger:** Invoice is created

**Event:** `InvoiceCreated`

**Handler:** `accounting/infrastructure/eventhandler/invoice_handler.go`

**Actions:**
1. Find Accounts Receivable account (code: 1300)
2. Find Revenue account (code: 4000)
3. Create transaction for invoice amount
4. Publish `TransactionRecorded` event

**Handler Methods:**
```go
HandleInvoiceCreated(ctx, event)    // Create journal entry
```

**Business Rules:**
- Debit: Accounts Receivable (1300)
- Credit: Revenue (4000)
- Transaction reference: INV-{invoice_number}
- Posted by: SYSTEM

## Domain Events Structure

### Ordering Domain Events

```go
type OrderConfirmed struct {
    OrderID     OrderID
    CustomerID  string
    SupplierID  string
    WarehouseID string              // Future: warehouse assignment
    Lines       []OrderConfirmedLine
    Total       float64
    Currency    string
    occurredAt  time.Time
}

type OrderConfirmedLine struct {
    ItemID    string
    ItemName  string
    Quantity  float64
    Unit      string
    UnitPrice float64
    Discount  float64
    Total     float64
}
```

### Invoicing Domain Events

```go
type InvoiceCreated struct {
    InvoiceID     InvoiceID
    InvoiceNumber string
    CustomerID    string
    Total         float64
    Currency      string
    occurredAt    time.Time
}
```

## Event Handler Implementation Pattern

### Handler Structure

```go
type OrderEventHandler struct {
    stockRepo       domain.StockRepository
    reservationRepo domain.StockReservationRepository
}

func NewOrderEventHandler(
    stockRepo domain.StockRepository,
    reservationRepo domain.StockReservationRepository,
) *OrderEventHandler {
    return &OrderEventHandler{
        stockRepo:       stockRepo,
        reservationRepo: reservationRepo,
    }
}
```

### Registration Pattern

```go
func (h *OrderEventHandler) Register(bus eventbus.Bus) {
    bus.Subscribe("ordering.order_confirmed", h.handleOrderConfirmed)
    bus.Subscribe("ordering.order_cancelled", h.handleOrderCancelled)
    bus.Subscribe("ordering.order_completed", h.handleOrderCompleted)
}

// Wrapper for type safety
func (h *OrderEventHandler) handleOrderConfirmed(ctx context.Context, event eventbus.Event) error {
    e, ok := event.(orderingDomain.OrderConfirmed)
    if !ok {
        return fmt.Errorf("invalid event type: expected OrderConfirmed")
    }
    return h.HandleOrderConfirmed(ctx, e)
}
```

### Wire Integration

```go
// cmd/api/wire.go

// Initialize event handlers
inventoryOrderEventHandler := inventoryEventHandler.NewOrderEventHandler(stockRepo, stockReservationRepo)
invoicingOrderEventHandler := invoicingEventHandler.NewOrderEventHandler(invoiceRepo)
accountingInvoiceEventHandler := accountingEventHandler.NewInvoiceEventHandler(accountRepo, transactionRepo)

// Register handlers with event bus
inventoryOrderEventHandler.Register(bus)
invoicingOrderEventHandler.Register(bus)
accountingInvoiceEventHandler.Register(bus)

// Add to dependencies
return &Dependencies{
    // ... other dependencies
    InventoryOrderEventHandler: inventoryOrderEventHandler,
    InvoicingOrderEventHandler: invoicingOrderEventHandler,
    AccountingInvoiceEventHandler: accountingInvoiceEventHandler,
}
```

## Event Bus Configuration

The event bus is configurable based on environment:

### Development (In-Memory)

```go
// pkg/eventbus/memory/bus.go
bus := membus.New()
```

**Pros:** Simple, fast, no infrastructure
**Cons:** Events lost on restart, no cross-instance communication

### Production (Redis)

```go
// pkg/eventbus/redis/bus.go
bus := redisbus.New(redisClient)
```

**Pros:** Persistent, cross-instance, reliable
**Cons:** Requires Redis infrastructure

## Testing Event Handlers

### Unit Test Example

```go
func TestOrderEventHandler_HandleOrderConfirmed(t *testing.T) {
    // Setup
    mockStockRepo := &MockStockRepository{}
    mockReservationRepo := &MockStockReservationRepository{}
    handler := NewOrderEventHandler(mockStockRepo, mockReservationRepo)
    
    // Create event
    event := orderingDomain.OrderConfirmed{
        OrderID:    orderingDomain.NewOrderID(),
        CustomerID: "customer-123",
        Lines: []orderingDomain.OrderConfirmedLine{
            {ItemID: "item-1", Quantity: 10},
        },
    }
    
    // Execute
    err := handler.HandleOrderConfirmed(context.Background(), event)
    
    // Assert
    assert.NoError(t, err)
    // Verify stock was reserved
}
```

### Integration Test Example

```go
func TestOrderToInvoiceIntegration(t *testing.T) {
    // Setup test database
    pgContainer := testcontainers.Postgres()
    db := pgxpool.Connect(pgContainer.ConnectionString())
    
    // Create order
    order := createTestOrder()
    repo.Save(ctx, order)
    
    // Publish event
    bus.Publish(ctx, orderingDomain.OrderConfirmed{OrderID: order.ID})
    
    // Wait for handler
    time.Sleep(100 * time.Millisecond)
    
    // Verify invoice was created
    invoices := invoiceRepo.FindByOrder(ctx, order.ID)
    assert.Len(t, invoices, 1)
}
```

## Future Enhancements

### 1. Warehouse Assignment

Currently, order confirmation requires manual warehouse assignment. Future enhancement:

```go
type OrderConfirmed struct {
    // ...
    WarehouseID string    // Auto-assigned based on shipping address
    // ...
}
```

### 2. Event Sourcing

Persist all events for replay and audit:

```go
type EventStore interface {
    Save(ctx context.Context, event eventbus.Event) error
    Load(ctx context.Context, aggregateID string) ([]eventbus.Event, error)
}
```

### 3. Saga Pattern

Implement long-running business processes:

```go
type OrderSaga struct {
    OrderID     string
    CurrentStep string
    Status      string
    Compensations []Compensation
}
```

### 4. Event Versioning

Support event schema evolution:

```go
type OrderConfirmed struct {
    Version     int
    OrderID     OrderID
    // v1 fields
    // v2 fields
}
```

## Benefits

1. **Loose Coupling**: Contexts don't depend on each other's internal implementation
2. **Scalability**: Handlers can be deployed independently
3. **Extensibility**: Easy to add new handlers without modifying existing code
4. **Audit Trail**: Full history of business events
5. **Testability**: Each handler can be tested in isolation
6. **Flexibility**: Handlers can have different transaction boundaries

## Best Practices

1. **Idempotency**: Handlers must handle duplicate events safely
2. **Error Handling**: Log errors but don't block event processing
3. **Async Processing**: Don't block in handlers for long operations
4. **Transaction Boundaries**: Use separate transactions per context
5. **Event Naming**: Use past tense (OrderConfirmed, not OrderConfirm)
6. **Event Payload**: Include all necessary data, minimize lookups
7. **Version Events**: Plan for schema evolution from day one

## Monitoring

### Metrics to Track

```
# Event processing latency
event_handler_duration_seconds{handler="order", event="confirmed"}

# Event processing errors
event_handler_errors_total{handler="order", event="confirmed"}

# Event bus lag
event_bus_lag_seconds{bus="redis"}

# Active event handlers
event_handlers_active{context="inventory"}
```

### Logging

```go
log.InfoContext(ctx, "event handled",
    "event", "OrderConfirmed",
    "order_id", event.OrderID,
    "duration", time.Since(start),
)
```

## References

- [Domain-Driven Design](https://domainlanguage.com/ddd/)
- [Event Sourcing Pattern](https://martinfowler.com/eaaDev/EventSourcing.html)
- [CQRS](https://martinfowler.com/bliki/CQRS.html)
- [Saga Pattern](https://microservices.io/patterns/data/saga.html)