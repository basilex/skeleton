# Action Plan: CRM Critical Issues Resolution

**Created:** 2026-04-08
**Priority:** CRITICAL - Must complete before frontend development
**Estimated Time:** 12-14 hours (Phase 1: Foundation)
**Team:** 1 developer

---

## 📋 Phase 1: FOUNDATION (CRITICAL) - 12-14 hours

### ✅ TASK 1: Create Catalog Domain Events (2 hours)

**Objective:** Enable Catalog context to publish domain events

**Files to Create:**
```
internal/catalog/domain/events.go
```

**Implementation Steps:**

1. **Create events file:**
```bash
touch internal/catalog/domain/events.go
```

2. **Add events:**
```go
// internal/catalog/domain/events.go
package catalog

import "time"

type DomainEvent interface {
    EventName() string
    OccurredAt() time.Time
}

// ItemCreated is published when a new item is created
type ItemCreated struct {
    ItemID      ItemID
    SKU         string
    Name        string
    CategoryID  *CategoryID
    BasePrice   float64
    Currency    string
    OccurredAt  time.Time
}

func (e ItemCreated) EventName() string { return "catalog.item_created" }
func (e ItemCreated) OccurredAt() time.Time { return e.OccurredAt }

// ItemPriceChanged is published when item price is updated
type ItemPriceChanged struct {
    ItemID      ItemID
    OldPrice    float64
    NewPrice    float64
    Currency    string
    OccurredAt  time.Time
}

func (e ItemPriceChanged) EventName() string { return "catalog.item_price_changed" }
func (e ItemPriceChanged) OccurredAt() time.Time { return e.OccurredAt }

// ItemStatusChanged is published when item status changes
type ItemStatusChanged struct {
    ItemID      ItemID
    OldStatus   ItemStatus
    NewStatus   ItemStatus
    OccurredAt  time.Time
}

func (e ItemStatusChanged) EventName() string { return "catalog.item_status_changed" }
func (e ItemStatusChanged) OccurredAt() time.Time { return e.OccurredAt }

// ItemDeactivated is published when item is deactivated
type ItemDeactivated struct {
    ItemID      ItemID
    Reason      string
    OccurredAt  time.Time
}

func (e ItemDeactivated) EventName() string { return "catalog.item_deactivated" }
func (e ItemDeactivated) OccurredAt() time.Time { return e.OccurredAt }
```

3. **Update Item aggregate to publish events:**
```go
// Add to internal/catalog/domain/item.go

import "github.com/basilex/skeleton/pkg/eventbus"

type Item struct {
    // ... existing fields
    events []eventbus.Event
}

// Update CreateItem to publish event
func NewItem(...) (*Item, error) {
    // ... existing code
    item.events = append(item.events, ItemCreated{
        ItemID: item.id,
        SKU: sku,
        Name: name,
        // ... etc
        OccurredAt: time.Now(),
    })
    return item, nil
}

// Add PullEvents method
func (i *Item) PullEvents() []eventbus.Event {
    events := i.events
    i.events = make([]eventbus.Event, 0)
    return events
}
```

**Verification:**
```bash
cd internal/catalog/domain && go test -v
# Should pass all tests

go build ./internal/catalog/...
# Should compile successfully
```

**Commit:** `git commit -m "feat(catalog): add domain events for item lifecycle"`

---

### ✅ TASK 2: Create Order Event Handlers (1 hour)

**Objective:** Subscribe Inventory to Order events for stock reservation

**Files to Create:**
```
internal/inventory/infrastructure/eventhandler/order_handler.go
```

**Implementation:**

```go
// internal/inventory/infrastructure/eventhandler/order_handler.go
package eventhandler

import (
    "context"
    
    orderingdomain "github.com/basilex/skeleton/internal/ordering/domain"
    "github.com/basilex/skeleton/internal/inventory/domain"
)

type OrderEventHandler struct {
    stockRepo        domain.StockRepository
    reservationRepo  domain.StockReservationRepository
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

// HandleOrderConfirmed reserves stock for order lines
func (h *OrderEventHandler) HandleOrderConfirmed(ctx context.Context, event orderingdomain.OrderConfirmed) error {
    // For each order line, find or create stock and reserve quantity
    for _, line := range event.Lines {
        // Find stock by item_id
        stock, err := h.stockRepo.FindByItemAndWarehouse(ctx, line.ItemID, event.WarehouseID)
        if err != nil {
            // Create stock record if doesn't exist
            stock, err = domain.NewStock(line.ItemID, event.WarehouseID)
            if err != nil {
                return err
            }
            if err := h.stockRepo.Save(ctx, stock); err != nil {
                return err
            }
        }
        
        // Reserve stock
        if err := stock.Reserve(line.Quantity); err != nil {
            return err
        }
        
        // Create reservation
        reservation, err := domain.NewStockReservation(
            stock.GetID(),
            event.OrderID,
            line.Quantity,
            nil, // No expiration by default
        )
        if err != nil {
            return err
        }
        
        if err := h.reservationRepo.Save(ctx, reservation); err != nil {
            return err
        }
        
        if err := h.stockRepo.Save(ctx, stock); err != nil {
            return err
        }
    }
    
    return nil
}

// HandleOrderCancelled releases stock reservations
func (h *OrderEventHandler) HandleOrderCancelled(ctx context.Context, event orderingdomain.OrderCancelled) error {
    // Find all reservations for this order
    reservations, err := h.reservationRepo.FindByOrder(ctx, event.OrderID)
    if err != nil {
        return err
    }
    
    // Cancel each reservation and release stock
    for _, reservation := range reservations {
        stock, err := h.stockRepo.FindByID(ctx, reservation.GetStockID())
        if err != nil {
            return err
        }
        
        if err := reservation.Cancel(); err != nil {
            return err
        }
        
        stock.ReleaseReservation(reservation.GetQuantity())
        
        if err := h.reservationRepo.Save(ctx, reservation); err != nil {
            return err
        }
        
        if err := h.stockRepo.Save(ctx, stock); err != nil {
            return err
        }
    }
    
    return nil
}

// HandleOrderCompleted fulfills reservations and deducts stock
func (h *OrderEventHandler) HandleOrderCompleted(ctx context.Context, event orderingdomain.OrderCompleted) error {
    // Find all reservations for this order
    reservations, err := h.reservationRepo.FindByOrder(ctx, event.OrderID)
    if err != nil {
        return err
    }
    
    // Fulfill each reservation
    for _, reservation := range reservations {
        stock, err := h.stockRepo.FindByID(ctx, reservation.GetStockID())
        if err != nil {
            return err
        }
        
        if err := reservation.Fulfill(); err != nil {
            return err
        }
        
        stock.FulfillReservation(reservation.GetQuantity())
        
        if err := h.reservationRepo.Save(ctx, reservation); err != nil {
            return err
        }
        
        if err := h.stockRepo.Save(ctx, stock); err != nil {
            return err
        }
    }
    
    return nil
}

func (h *OrderEventHandler) Register(bus eventbus.Bus) {
    bus.Subscribe("ordering.order_confirmed", h.HandleOrderConfirmed)
    bus.Subscribe("ordering.order_cancelled", h.HandleOrderCancelled)
    bus.Subscribe("ordering.order_completed", h.HandleOrderCompleted)
}
```

**Verification:**
```bash
go build ./internal/inventory/infrastructure/eventhandler/...
```

**Commit:** `git commit -m "feat(inventory): add order event handlers for stock reservation"`

---

### ✅ TASK 3: Create Invoice Event Handlers (1 hour)

**Objective:** Auto-create invoice when order is confirmed

**Files to Create:**
```
internal/invoicing/infrastructure/eventhandler/order_handler.go
```

**Implementation:**

```go
// internal/invoicing/infrastructure/eventhandler/order_handler.go
package eventhandler

import (
    "context"
    "time"
    
    "github.com/basilex/skeleton/internal/invoicing/domain"
    orderingdomain "github.com/basilex/skeleton/internal/ordering/domain"
)

type OrderEventHandler struct {
    invoiceRepo domain.InvoiceRepository
}

func NewOrderEventHandler(invoiceRepo domain.InvoiceRepository) *OrderEventHandler {
    return &OrderEventHandler{
        invoiceRepo: invoiceRepo,
    }
}

func (h *OrderEventHandler) HandleOrderConfirmed(ctx context.Context, event orderingdomain.OrderConfirmed) error {
    // Generate invoice number (simple sequential for now)
    invoiceNumber := generateInvoiceNumber(event.OrderID)
    
    // Create invoice from order
    invoice, err := domain.NewInvoice(
        invoiceNumber,
        event.CustomerID,
        event.Currency,
        time.Now().Add(30 * 24 * time.Hour), // 30 days due date
    )
    if err != nil {
        return err
    }
    
    // Link to order
    invoice.LinkOrder(event.OrderID)
    
    // Add invoice lines from order lines
    for _, line := range event.Lines {
        invoiceLine := domain.NewInvoiceLine(
            invoice.GetID(),
            line.Description,
            line.Quantity,
            line.UnitPrice,
            line.Unit,
            line.Discount,
        )
        invoice.AddLine(invoiceLine)
    }
    
    // Calculate totals
    invoice.CalculateTotals()
    
    // Save invoice
    if err := h.invoiceRepo.Save(ctx, invoice); err != nil {
        return err
    }
    
    return nil
}

func generateInvoiceNumber(orderID string) string {
    // Simple: INV-{timestamp}
    // TODO: Use proper sequence generator
    return fmt.Sprintf("INV-%d", time.Now().Unix())
}

func (h *OrderEventHandler) Register(bus eventbus.Bus) {
    bus.Subscribe("ordering.order_confirmed", h.HandleOrderConfirmed)
}
```

**Verification:**
```bash
go build ./internal/invoicing/infrastructure/eventhandler/...
```

**Commit:** `git commit -m "feat(invoicing): add order event handler for auto-invoice creation"`

---

### ✅ TASK 4: Create Accounting Event Handlers (1 hour)

**Objective:** Create journal entries when invoice is created

**Files to Create:**
```
internal/accounting/infrastructure/eventhandler/invoice_handler.go
```

**Implementation:**

```go
// internal/accounting/infrastructure/eventhandler/invoice_handler.go
package eventhandler

import (
    "context"
    
    "github.com/basilex/skeleton/internal/accounting/domain"
    invoicingdomain "github.com/basilex/skeleton/internal/invoicing/domain"
)

type InvoiceEventHandler struct {
    accountRepo domain.AccountRepository
    transactionRepo domain.TransactionRepository
}

func NewInvoiceEventHandler(
    accountRepo domain.AccountRepository,
    transactionRepo domain.TransactionRepository,
) *InvoiceEventHandler {
    return &InvoiceEventHandler{
        accountRepo:      accountRepo,
        transactionRepo:   transactionRepo,
    }
}

func (h *InvoiceEventHandler) HandleInvoiceCreated(ctx context.Context, event invoicingdomain.InvoiceCreated) error {
    // Create journal entry for invoice
    // Debit: Accounts Receivable (1300)
    // Credit: Sales Revenue (4000)
    // Credit: Sales Tax Payable (2500)
    
    transaction, err := domain.NewTransaction(
        fmt.Sprintf("Invoice %s", event.InvoiceNumber),
        "Invoice generated from order",
        "USD", // Should match invoice currency
    )
    if err != nil {
        return err
    }
    
    // Add debit line (Accounts Receivable)
    transaction.AddLine(
        "1300", // Account Code for Accounts Receivable
        event.Total, // Debit amount
        0, // Credit
    )
    
    // Add credit lines
    transaction.AddLine(
        "4000", // Sales Revenue
        0,
        event.Subtotal, // Credit amount
    )
    
    if event.TaxAmount > 0 {
        transaction.AddLine(
            "2500", // Sales Tax Payable
            0,
            event.TaxAmount,
        )
    }
    
    // Save transaction
    if err := h.transactionRepo.Save(ctx, transaction); err != nil {
        return err
    }
    
    return nil
}

func (h *InvoiceEventHandler) Register(bus eventbus.Bus) {
    bus.Subscribe("invoicing.invoice_created", h.HandleInvoiceCreated)
}
```

**Verification:**
```bash
go build ./internal/accounting/infrastructure/eventhandler/...
```

**Commit:** `git commit -m "feat(accounting): add invoice event handler for journal entries"`

---

### ✅ TASK 5: Wire Event Handlers (30 minutes)

**Objective:** Register all event handlers in wire.go

**File to Modify:**
```
cmd/api/wire.go
```

**Changes:**

```go
// Add to imports section
inventoryEventHandler "github.com/basilex/skeleton/internal/inventory/infrastructure/eventhandler"
invoicingEventHandler "github.com/basilex/skeleton/internal/invoicing/infrastructure/eventhandler"
accountingEventHandler "github.com/basilex/skeleton/internal/accounting/infrastructure/eventhandler"

// In Dependencies struct, add
type Dependencies struct {
    // ... existing fields
    CatalogEventHandler *catalogEventHandler.CatalogEventHandler
    OrderEventHandler *inventoryEventHandler.OrderEventHandler
    InvoiceEventHandler *invoicingEventHandler.OrderEventHandler
    AccountingInvoiceHandler *accountingEventHandler.InvoiceEventHandler
}

// In wireDependencies function, add after handlers initialization:

// Initialize event handlers
orderEventHandler := inventoryEventHandler.NewOrderEventHandler(
    stockRepo,
    stockReservationRepo,
)
orderEventHandler.Register(bus) // Subscribe to ordering events

invoiceOrderHandler := invoicingEventHandler.NewOrderEventHandler(
    invoiceRepo,
)
invoiceOrderHandler.Register(bus) // Subscribe to ordering events

accountingInvoiceHandler := accountingEventHandler.NewInvoiceEventHandler(
    accountRepo,
    transactionRepo,
)
accountingInvoiceHandler.Register(bus) // Subscribe to invoicing events

// Add to return statement
return &Dependencies{
    // ... existing fields
    OrderEventHandler: orderEventHandler,
    InvoiceEventHandler: invoiceOrderHandler,
    AccountingInvoiceHandler: accountingInvoiceHandler,
}
```

**Verification:**
```bash
go build ./cmd/api/...
# Should compile successfully
```

**Commit:** `git commit -m "feat: wire event handlers for cross-context integration"`

---

### ✅ TASK 6: Create Integration Tests (2 hours)

**Objective:** Verify end-to-end workflows work

**Files to Create:**
```
tests/integration/order_invoice_test.go
tests/integration/invoice_accounting_test.go
tests/integration/catalog_stock_test.go
```

**Example Integration Test:**

```go
// tests/integration/order_invoice_test.go
package integration

import (
    "context"
    "testing"
    
    "github.com/basilex/skeleton/internal/ordering/domain"
    "github.com/basilex/skeleton/internal/invoicing/domain"
)

func TestOrderConfirmedCreatesInvoice(t *testing.T) {
    // Setup test database
    // Create order
    // Confirm order
    // Verify invoice created
    // Verify invoice lines match order lines
    // Verify invoice totals calculated
}

func TestOrderCompletedFulfillsReservation(t *testing.T) {
    // Setup:
    // Create item
    // Create stock
    // Create order with that item
    // Confirm order (should reserve stock)
    
    // Test:
    // Complete order
    // Verify stock reserved_qty decreased
    // Verify reservation status = fulfilled
    // Verify stock quantity decreased
}

func TestInvoiceCreatedCreatesJournalEntry(t *testing.T) {
    // Create invoice
    // Verify journal entry created
    // Verify debit line (Accounts Receivable)
    // Verify credit lines (Revenue, Tax)
    // Verify transaction balances
}
```

**Verification:**
```bash
make test-integration
# All integration tests should pass
```

**Commit:** `git commit -m "test: add integration tests for cross-context workflows"`

---

## 📊 Phase 1 Checklist

- [ ] **Task 1:** Catalog events file created and working
- [ ] **Task 2:** Order event handlers created and wired
- [ ] **Task 3:** Invoice event handlers created and wired
- [ ] **Task 4:** Accounting event handlers created and wired
- [ ] **Task 5:** All handlers registered in wire.go
- [ ] **Task 6:** Integration tests passing

**Success Criteria:**
```bash
# All should pass:
go test ./internal/catalog/domain/... -v
go test ./internal/inventory/infrastructure/eventhandler/... -v
go test ./internal/invoicing/infrastructure/eventhandler/... -v
go test ./internal/accounting/infrastructure/eventhandler/... -v
make test-integration
go build ./cmd/api/...
```

---

## 🎯 Definition of Done

Phase 1 is complete when:

1. **Catalog publishes events** ✅
   - ItemCreated event fires when item created
   - Item price/status changes publish events

2. **Order → Inventory integration works** ✅
   - Order confirmed → Stock reserved
   - Order cancelled → Reservation released
   - Order completed → Stock deducted

3. **Order → Invoice integration works** ✅
   - Order confirmed → Invoice auto-created
   - Invoice lines match order lines

4. **Invoice → Accounting integration works** ✅
   - Invoice created → Journal entry created
   - Debit/credit balances

5. **All tests passing** ✅
   - Unit tests pass
   - Integration tests pass
   - Build succeeds

6. **Documentation updated** ✅
   - README updated
   - CHANGELOG updated
   - Integration guide created

---

## 🚀 After Phase 1 Completion

**Start Phase 2 (Type Safety)** - 6-8 hours

Then Phase 3 (Automation) - 5-7 hours

Then ready for Next.js frontend development!

---

## 📝 Notes

- Each task should be committed separately
- Run tests after each task
- Keep commits small and focused
- Update documentation as you go
- Create branches per task if needed

**Estimated Timeline:**
- Day 1: Tasks 1-3 (4-6 hours)
- Day 2: Tasks 4-6 (6-8 hours)
- Total: 2 days for complete Phase 1