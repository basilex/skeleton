# Improvement Plan - Bring to Perfection

**Created:** 2026-04-09
**Status:** Active
**Goal:** Fix all critical and important issues identified in comprehensive code review

---

## 📊 Executive Summary

**Overall Score:** 7.5/10 → Target: 9.5/10

**Critical Issues:** 3
**Important Issues:** 7
**Medium Issues:** 4
**Estimated Effort:** 4-6 weeks

---

## 🎯 Improvement Phases

### Phase 1: Critical Fixes (Week 1-2)
**Priority:** P0 - Must fix before production deployment
**Impact:** Data integrity, system consistency

| Issue | Status | Effort | Owner |
|-------|--------|--------|-------|
| Transaction Management | ⬜ Not Started | 3 days | TBD |
| Stock Domain Events | ⬜ Not Started | 2 days | TBD |
| Money Value Object | ⬜ Not Started | 3 days | TBD |

### Phase 2: Important Improvements (Week 3-4)
**Priority:** P1 - Fix within 2 weeks
**Impact:** Code quality, maintainability

| Issue | Status | Effort | Owner |
|-------|--------|--------|-------|
| Event Publishing Standardization | ⬜ Not Started | 2 days | TBD |
| Application Error Types | ⬜ Not Started | 1 day | TBD |
| Silent Failures Fix | ⬜ Not Started | 1 day | TBD |
| Input Validation | ⬜ Not Started | 2 days | TBD |
| Rate Limiting | ⬜ Not Started | 1 day | TBD |

### Phase 3: Test Coverage (Week 5-6)
**Priority:** P2 - Fix within 1 month
**Impact:** Code reliability, regression prevention

| Issue | Status | Effort | Owner |
|-------|--------|--------|-------|
| Application Layer Tests | ⬜ Not Started | 5 days | TBD |
| Infrastructure Layer Tests | ⬜ Not Started | 4 days | TBD |
| HTTP Handler Tests | ⬜ Not Started | 3 days | TBD |

### Phase 4: Documentation & Polish (Ongoing)
**Priority:** P3 - Improvements over time
**Impact:** Developer experience, maintainability

| Issue | Status | Effort | Owner |
|-------|--------|--------|-------|
| OpenAPI/Swagger | ⬜ Not Started | 2 days | TBD |
| Deployment Runbooks | ⬜ Not Started | 1 day | TBD |
| Troubleshooting Guides | ⬜ Not Started | 1 day | TBD |

---

## 📋 Detailed Fix Plans

### P0-1: Transaction Management

**Problem:** Multi-aggregate operations without transactions lead to inconsistent state.

**Affected Files:**
- `internal/accounting/application/command/record_transaction.go`
- `internal/inventory/application/command/transfer_stock.go`
- `internal/inventory/application/command/reserve_stock.go`
- `internal/ordering/application/command/*.go` (potentially)

**Solution:** Implement Unit of Work pattern with transaction manager.

**Implementation Steps:**

1. **Create Transaction Manager Interface**

```go
// pkg/transaction/manager.go
package transaction

import "context"

type Manager interface {
    Execute(ctx context.Context, fn func(ctx context.Context) error) error
    ExecuteWithResult(ctx context.Context, fn func(ctx context.Context) (interface{}, error)) (interface{}, error)
}
```

2. **Implement PostgreSQL Transaction Manager**

```go
// pkg/transaction/pgx_manager.go
package transaction

import (
    "context"
    "fmt"
    
    "github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgxpool"
)

type PgxTransactionManager struct {
    pool *pgxpool.Pool
}

func NewPgxTransactionManager(pool *pgxpool.Pool) *PgxTransactionManager {
    return &PgxTransactionManager{pool: pool}
}

func (m *PgxTransactionManager) Execute(ctx context.Context, fn func(ctx context.Context) error) error {
    tx, err := m.pool.Begin(ctx)
    if err != nil {
        return fmt.Errorf("begin transaction: %w", err)
    }
    
    defer func() {
        if p := recover(); p != nil {
            tx.Rollback(ctx)
            panic(p)
        }
    }()
    
    // Inject transaction into context
    txCtx := context.WithValue(ctx, txKey{}, tx)
    
    if err := fn(txCtx); err != nil {
        if rbErr := tx.Rollback(ctx); rbErr != nil {
            return fmt.Errorf("rollback: %v, original: %w", rbErr, err)
        }
        return err
    }
    
    if err := tx.Commit(ctx); err != nil {
        return fmt.Errorf("commit transaction: %w", err)
    }
    
    return nil
}

type txKey struct{}

// FromContext extracts transaction from context for repository use
func FromContext(ctx context.Context) (pgx.Tx, bool) {
    tx, ok := ctx.Value(txKey{}).(pgx.Tx)
    return tx, ok
}
```

3. **Update Repositories to Use Transaction from Context**

```go
// internal/accounting/infrastructure/persistence/account_repository.go

func (r *AccountRepository) Save(ctx context.Context, account *domain.Account) error {
    query := `UPDATE accounts SET balance = $1 WHERE id = $2`
    
    // Check if transaction exists in context
    if tx, ok := transaction.FromContext(ctx); ok {
        // Use transaction
        _, err := tx.Exec(ctx, query, account.Balance(), account.ID())
        return err
    }
    
    // Use regular connection
    _, err := r.pool.Exec(ctx, query, account.Balance(), account.ID())
    return err
}
```

4. **Update Command Handlers**

```go
// internal/accounting/application/command/record_transaction.go

type RecordTransactionHandler struct {
    accounts     domain.AccountRepository
    transactions domain.TransactionRepository
    bus          eventbus.Bus
    txManager    transaction.Manager  // Add this
}

func NewRecordTransactionHandler(
    accounts domain.AccountRepository,
    transactions domain.TransactionRepository,
    bus eventbus.Bus,
    txManager transaction.Manager,
) *RecordTransactionHandler {
    return &RecordTransactionHandler{
        accounts:     accounts,
        transactions: transactions,
        bus:          bus,
        txManager:    txManager,
    }
}

func (h *RecordTransactionHandler) Handle(ctx context.Context, cmd RecordTransactionCommand) (RecordTransactionResult, error) {
    var result RecordTransactionResult
    
    err := h.txManager.Execute(ctx, func(ctx context.Context) error {
        // Parse and validate
        fromAccountID, err := domain.ParseAccountID(cmd.FromAccountID)
        if err != nil {
            return fmt.Errorf("parse from account ID: %w", err)
        }
        
        toAccountID, err := domain.ParseAccountID(cmd.ToAccountID)
        if err != nil {
            return fmt.Errorf("parse to account ID: %w", err)
        }
        
        // Load aggregates
        fromAccount, err := h.accounts.FindByID(ctx, fromAccountID)
        if err != nil {
            return fmt.Errorf("find from account: %w", err)
        }
        
        toAccount, err := h.accounts.FindByID(ctx, toAccountID)
        if err != nil {
            return fmt.Errorf("find to account: %w", err)
        }
        
        // Perform business operations
        money := domain.Money{Amount: cmd.Amount, Currency: cmd.Currency}
        
        if err := fromAccount.Credit(money); err != nil {
            return fmt.Errorf("credit from account: %w", err)
        }
        
        if err := toAccount.Debit(money); err != nil {
            return fmt.Errorf("debit to account: %w", err)
        }
        
        // Create transaction record
        transaction, err := domain.NewTransaction(fromAccountID, toAccountID, money, cmd.Description)
        if err != nil {
            return fmt.Errorf("create transaction: %w", err)
        }
        
        // Save all within transaction
        if err := h.accounts.Save(ctx, fromAccount); err != nil {
            return fmt.Errorf("save from account: %w", err)
        }
        
        if err := h.accounts.Save(ctx, toAccount); err != nil {
            return fmt.Errorf("save to account: %w", err)
        }
        
        if err := h.transactions.Save(ctx, transaction); err != nil {
            return fmt.Errorf("save transaction: %w", err)
        }
        
        // Publish events
        for _, event := range fromAccount.PullEvents() {
            if err := h.bus.Publish(ctx, event); err != nil {
                return fmt.Errorf("publish from account event: %w", err)
            }
        }
        
        for _, event := range toAccount.PullEvents() {
            if err := h.bus.Publish(ctx, event); err != nil {
                return fmt.Errorf("publish to account event: %w", err)
            }
        }
        
        for _, event := range transaction.PullEvents() {
            if err := h.bus.Publish(ctx, event); err != nil {
                return fmt.Errorf("publish transaction event: %w", err)
            }
        }
        
        result = RecordTransactionResult{
            TransactionID: transaction.ID().String(),
        }
        
        return nil
    })
    
    return result, err
}
```

5. **Wire Dependencies**

```go
// cmd/api/wire.go

func InitializeDependencies(
    ctx context.Context,
    cfg *config.Config,
    pool *pgxpool.Pool,
    redisClient *redis.Client,
) (*Dependencies, error) {
    // ... existing dependencies ...
    
    // Add transaction manager
    txManager := transaction.NewPgxTransactionManager(pool)
    
    // Update handlers
    recordTransactionHandler := accountingCommand.NewRecordTransactionHandler(
        accountRepo,
        transactionRepo,
        eventBus,
        txManager,  // Inject
    )
    
    // ... rest of dependencies ...
}
```

**Test Plan:**
```go
// internal/accounting/application/command/record_transaction_test.go

func TestRecordTransaction_RollbackOnError(t *testing.T) {
    // Setup: Create two accounts
    // Action: Transfer money, but fail on second save
    // Assert: Both accounts unchanged
}

func TestRecordTransaction_Commit(t *testing.T) {
    // Setup: Create two accounts
    // Action: Transfer money successfully
    // Assert: Both accounts updated correctly
}
```

**Completion Criteria:**
- [ ] Transaction Manager interface created
- [ ] PgxTransactionManager implemented
- [ ] All repositories support transaction context
- [ ] All multi-aggregate handlers use transaction manager
- [ ] Integration tests pass
- [ ] Manual testing confirms rollback works

**Files to Create/Modify:**
- Create: `pkg/transaction/manager.go`
- Create: `pkg/transaction/pgx_manager.go`
- Modify: All repositories (add transaction support)
- Modify: All command handlers (wrap in transaction)
- Create: Integration tests

---

### P0-2: Stock Domain Events

**Problem:** Stock aggregate doesn't publish domain events, breaking event-driven architecture.

**Affected Files:**
- `internal/inventory/domain/stock.go`
- `internal/inventory/application/command/*.go`

**Solution:** Add domain events to Stock aggregate and update all command handlers.

**Implementation Steps:**

1. **Update Stock Aggregate**

```go
// internal/inventory/domain/stock.go

type Stock struct {
    id              StockID
    itemID          string
    warehouseID     string
    quantity        float64
    reservedQty     float64
    availableQty    float64
    reorderPoint    float64
    lastMovementID  StockMovementID
    createdAt       time.Time
    updatedAt       time.Time
    events          []eventbus.Event  // Add this
}

// Add PullEvents method
func (s *Stock) PullEvents() []eventbus.Event {
    events := s.events
    s.events = make([]eventbus.Event, 0)
    return events
}

// Helper to publish events
func (s *Stock) publishEvent(event eventbus.Event) {
    s.events = append(s.events, event)
    s.updatedAt = time.Now().UTC()
}

// Update methods to publish events
func (s *Stock) AdjustQuantity(quantity float64, movementID StockMovementID) error {
    oldQty := s.quantity
    s.quantity += quantity
    
    if s.quantity < 0 {
        return ErrInsufficientStock
    }
    
    s.availableQty = s.quantity - s.reservedQty
    s.lastMovementID = movementID
    s.updatedAt = time.Now().UTC()
    
    s.publishEvent(StockAdjusted{
        StockID:     s.id,
        ItemID:      s.itemID,
        WarehouseID: s.warehouseID,
        OldQty:      oldQty,
        NewQty:      s.quantity,
        MovementID:  movementID,
        Reason:      "adjustment",
        OccurredAt:  time.Now().UTC(),
    })
    
    return nil
}

func (s *Stock) Reserve(quantity float64, reservationID StockReservationID) error {
    if s.availableQty < quantity {
        return ErrInsufficientStock
    }
    
    oldQty := s.quantity
    oldReserved := s.reservedQty
    
    s.reservedQty += quantity
    s.availableQty = s.quantity - s.reservedQty
    
    s.publishEvent(StockReserved{
        StockID:        s.id,
        ItemID:         s.itemID,
        WarehouseID:    s.warehouseID,
        Quantity:       quantity,
        ReservationID:  reservationID,
        OldAvailable:   s.availableQty + quantity,
        NewAvailable:   s.availableQty,
        OccurredAt:     time.Now().UTC(),
    })
    
    return nil
}

func (s *Stock) FulfillReservation(quantity float64) {
    oldQty := s.quantity
    oldReserved := s.reservedQty
    
    s.quantity -= quantity
    s.reservedQty -= quantity
    s.availableQty = s.quantity - s.reservedQty
    
    s.publishEvent(StockReservationFulfilled{
        StockID:     s.id,
        ItemID:      s.itemID,
        Quantity:    quantity,
        OldQty:      oldQty,
        NewQty:      s.quantity,
        OccurredAt:  time.Now().UTC(),
    })
}

func (s *Stock) ReleaseReservation(quantity float64) {
    s.reservedQty -= quantity
    s.availableQty = s.quantity - s.reservedQty
    
    s.publishEvent(StockReservationReleased{
        StockID:     s.id,
        ItemID:      s.itemID,
        Quantity:    quantity,
        Available:   s.availableQty,
        OccurredAt:  time.Now().UTC(),
    })
}
```

2. **Define Stock Events**

```go
// internal/inventory/domain/events.go

type StockAdjusted struct {
    StockID     StockID
    ItemID      string
    WarehouseID string
    OldQty      float64
    NewQty      float64
    MovementID  StockMovementID
    Reason      string
    OccurredAt  time.Time
}

func (e StockAdjusted) OccurredAt() time.Time { return e.OccurredAt }
func (e StockAdjusted) AggregateID() string   { return e.StockID.String() }
func (e StockAdjusted) AggregateType() string  { return "Stock" }
func (e StockAdjusted) EventType() string      { return "inventory.stock.adjusted" }

type StockReserved struct {
    StockID        StockID
    ItemID         string
    WarehouseID    string
    Quantity       float64
    ReservationID  StockReservationID
    OldAvailable   float64
    NewAvailable   float64
    OccurredAt     time.Time
}

func (e StockReserved) OccurredAt() time.Time { return e.OccurredAt }
func (e StockReserved) AggregateID() string   { return e.StockID.String() }
func (e StockReserved) AggregateType() string  { return "Stock" }
func (e StockReserved) EventType() string      { return "inventory.stock.reserved" }

type StockReservationFulfilled struct {
    StockID     StockID
    ItemID      string
    Quantity    float64
    OldQty      float64
    NewQty      float64
    OccurredAt  time.Time
}

func (e StockReservationFulfilled) OccurredAt() time.Time { return e.OccurredAt }
func (e StockReservationFulfilled) AggregateID() string   { return e.StockID.String() }
func (e StockReservationFulfilled) AggregateType() string  { return "Stock" }
func (e StockReservationFulfilled) EventType() string      { return "inventory.stock.reservation.fulfilled" }

type StockReservationReleased struct {
    StockID     StockID
    ItemID      string
    Quantity    float64
    Available   float64
    OccurredAt  time.Time
}

func (e StockReservationReleased) OccurredAt() time.Time { return e.OccurredAt }
func (e StockReservationReleased) AggregateID() string   { return e.StockID.String() }
func (e StockReservationReleased) AggregateType() string  { return "Stock" }
func (e StockReservationReleased) EventType() string      { return "inventory.stock.reservation.released" }
```

3. **Update Command Handlers**

```go
// internal/inventory/application/command/reserve_stock.go

func (h *ReserveStockHandler) Handle(ctx context.Context, cmd ReserveStockCommand) (*ReserveStockResult, error) {
    // Parse IDs
    stockID, err := domain.ParseStockID(cmd.StockID)
    if err != nil {
        return nil, fmt.Errorf("parse stock ID: %w", err)
    }
    
    reservationID := domain.NewStockReservationID()
    
    // Load aggregate
    stock, err := h.stock.FindByID(ctx, stockID)
    if err != nil {
        return nil, fmt.Errorf("find stock: %w", err)
    }
    
    // Business logic
    if err := stock.Reserve(cmd.Quantity, reservationID); err != nil {
        return nil, fmt.Errorf("reserve stock: %w", err)
    }
    
    // Save
    if err := h.stock.Save(ctx, stock); err != nil {
        return nil, fmt.Errorf("save stock: %w", err)
    }
    
    if err := h.reservations.Save(ctx, reservation); err != nil {
        return nil, fmt.Errorf("save reservation: %w", err)
    }
    
    // Publish events
    for _, event := range stock.PullEvents() {
        if err := h.bus.Publish(ctx, event); err != nil {
            return nil, fmt.Errorf("publish stock event: %w", err)
        }
    }
    
    return &ReserveStockResult{
        ReservationID: reservationID.String(),
    }, nil
}
```

**Completion Criteria:**
- [ ] Stock aggregate has events field
- [ ] PullEvents() method implemented
- [ ] All state changes publish events
- [ ] All command handlers updated
- [ ] Events defined in domain/events.go
- [ ] Integration tests verify events published

---

### P0-3: Money Value Object

**Problem:** Using float64 for monetary values causes precision errors.

**Files to Create/Modify:**

1. **Create Money Value Object**

```go
// pkg/money/money.go

package money

import (
    "errors"
    "fmt"
    "strings"
)

var (
    ErrInvalidAmount       = errors.New("invalid amount: must be non-negative")
    ErrDifferentCurrencies = errors.New("cannot operate on different currencies")
    ErrNegativeAmount      = errors.New("amount cannot be negative")
)

// Money represents a monetary value in smallest currency unit (cents/minor units)
type Money struct {
    Amount   int64  // Amount in smallest currency unit (cents for USD)
    Currency string // ISO 4217 currency code (USD, EUR, etc.)
}

// New creates a new Money instance
func New(amount int64, currency string) (Money, error) {
    if amount < 0 {
        return Money{}, ErrNegativeAmount
    }
    if currency == "" {
        return Money{}, errors.New("currency is required")
    }
    currency = strings.ToUpper(currency)
    if len(currency) != 3 {
        return Money{}, errors.New("currency must be 3-letter ISO code")
    }
    
    return Money{
        Amount:   amount,
        Currency: currency,
    }, nil
}

// NewFromFloat creates Money from float64 (e.g., 12.34 -> 1234 cents)
func NewFromFloat(amount float64, currency string) (Money, error) {
    if amount < 0 {
        return Money{}, ErrNegativeAmount
    }
    return New(int64(amount*100), currency)
}

// Add adds two Money values
func (m Money) Add(other Money) (Money, error) {
    if m.Currency != other.Currency {
        return Money{}, ErrDifferentCurrencies
    }
    return Money{
        Amount:   m.Amount + other.Amount,
        Currency: m.Currency,
    }, nil
}

// Subtract subtracts two Money values
func (m Money) Subtract(other Money) (Money, error) {
    if m.Currency != other.Currency {
        return Money{}, ErrDifferentCurrencies
    }
    if m.Amount < other.Amount {
        return Money{}, ErrNegativeAmount
    }
    return Money{
        Amount:   m.Amount - other.Amount,
        Currency: m.Currency,
    }, nil
}

// Multiply multiplies Money by a factor
func (m Money) Multiply(factor float64) Money {
    return Money{
        Amount:   int64(float64(m.Amount) * factor),
        Currency: m.Currency,
    }
}

// ToFloat64 converts Money to float64 for display (e.g., 1234 cents -> 12.34)
func (m Money) ToFloat64() float64 {
    return float64(m.Amount) / 100.0
}

// String returns formatted string representation
func (m Money) String() string {
    return fmt.Sprintf("%.2f %s", m.ToFloat64(), m.Currency)
}

// Equals checks if two Money values are equal
func (m Money) Equals(other Money) bool {
    return m.Amount == other.Amount && m.Currency == other.Currency
}

// IsZero checks if amount is zero
func (m Money) IsZero() bool {
    return m.Amount == 0
}

// IsPositive checks if amount is positive
func (m Money) IsPositive() bool {
    return m.Amount > 0
}
```

2. **Update Invoice to Use Money**

```go
// internal/invoicing/domain/invoice.go

type Invoice struct {
    id            InvoiceID
    number        string
    customerID    string
    lines         []InvoiceLine
    subtotal      Money  // Changed from float64
    taxAmount     Money  // Changed from float64
    total         Money  // Changed from float64
    paidAmount    Money  // Changed from float64
    status        InvoiceStatus
    // ...
}

func NewInvoice(number string, customerID string, lines []InvoiceLine) (*Invoice, error) {
    // Validate inputs
    if number == "" {
        return nil, ErrInvalidInvoiceNumber
    }
    if customerID == "" {
        return nil, ErrInvalidCustomerID
    }
    if len(lines) == 0 {
        return nil, ErrInvoiceRequiresLines
    }
    
    // Calculate totals using Money
    subtotal := Money{Amount: 0, Currency: "USD"}
    for _, line := range lines {
        var err error
        subtotal, err = subtotal.Add(line.Total())
        if err != nil {
            return nil, err
        }
    }
    
    // Calculate tax (10%)
    taxAmount := subtotal.Multiply(0.10)
    
    // Calculate total
    total, err := subtotal.Add(taxAmount)
    if err != nil {
        return nil, err
    }
    
    return &Invoice{
        id:         NewInvoiceID(),
        number:     number,
        customerID: customerID,
        lines:      lines,
        subtotal:   subtotal,
        taxAmount:  taxAmount,
        total:      total,
        paidAmount: Money{Amount: 0, Currency: "USD"},
        status:     InvoiceStatusDraft,
        createdAt:  time.Now().UTC(),
    }, nil
}

func (i *Invoice) RecordPayment(amount Money, method PaymentMethod, reference string) (*Payment, error) {
    if i.status == InvoiceStatusDraft {
        return nil, errors.New("cannot record payment for draft invoice")
    }
    if i.status == InvoiceStatusCancelled {
        return nil, ErrInvoiceAlreadyCancelled
    }
    if i.status == InvoiceStatusPaid {
        return nil, ErrInvoiceAlreadyPaid
    }
    
    // Validate currency
    if amount.Currency != i.total.Currency {
        return nil, ErrDifferentCurrencies
    }
    
    // Validate amount
    newPaidAmount, err := i.paidAmount.Add(amount)
    if err != nil {
        return nil, err
    }
    
    if newPaidAmount.Amount > i.total.Amount {
        return nil, ErrPaymentExceedsAmount
    }
    
    i.paidAmount = newPaidAmount
    
    // Check if fully paid
    if i.paidAmount.Equals(i.total) {
        i.status = InvoiceStatusPaid
        i.paidAt.Time = time.Now().UTC()
        i.paidAt.Valid = true
    }
    
    payment := NewPayment(i.id, amount, method, reference)
    i.payments = append(i.payments, payment)
    
    i.events = append(i.events, InvoicePaymentRecorded{
        InvoiceID:    i.id,
        PaymentID:   payment.ID(),
        Amount:      amount.Amount,
        Method:      string(method),
        OccurredAt:  time.Now().UTC(),
    })
    
    return payment, nil
}
```

3. **Update Storage Layer**

```go
// internal/invoicing/infrastructure/persistence/invoice_repository.go

func (r *InvoiceRepository) Save(ctx context.Context, invoice *domain.Invoice) error {
    query, args, err := r.psql.Insert("invoices").
        Columns("id", "number", "customer_id", "subtotal", "tax_amount", "total", "status").
        Values(
            invoice.ID().String(),
            invoice.Number(),
            invoice.CustomerID(),
            invoice.Subtotal().Amount,  // Store as integer (cents)
            invoice.TaxAmount().Amount,
            invoice.Total().Amount,
            invoice.Status(),
        ).
        Suffix("ON CONFLICT(id) DO UPDATE SET ...").
        ToSql()
    
    _, err = r.pool.Exec(ctx, query, args...)
    return err
}

func (r *InvoiceRepository) FindByID(ctx context.Context, id domain.InvoiceID) (*domain.Invoice, error) {
    var dto invoiceDTO
    err := pgxscan.Get(ctx, r.pool, &dto,
        `SELECT id, number, customer_id, subtotal, tax_amount, total FROM invoices WHERE id = $1`,
        id.String())
    if err != nil {
        return nil, fmt.Errorf("find invoice: %w", err)
    }
    
    return r.dtoToDomain(dto)
}

func (r *InvoiceRepository) dtoToDomain(dto invoiceDTO) (*domain.Invoice, error) {
    // Convert stored amounts back to Money
    subtotal := Money{Amount: dto.Subtotal, Currency: "USD"}
    taxAmount := Money{Amount: dto.TaxAmount, Currency: "USD"}
    total := Money{Amount: dto.Total, Currency: "USD"}
    
    return domain.ReconstituteInvoice(
        dto.ID,
        dto.Number,
        dto.CustomerID,
        lines,
        subtotal,
        taxAmount,
        total,
        // ...
    ), nil
}
```

**Migration Plan:**

```sql
-- migrations/027_money_type_migration.up.sql

-- Step 1: Add new columns with _cents suffix
ALTER TABLE invoices ADD COLUMN subtotal_cents BIGINT;
ALTER TABLE invoices ADD COLUMN tax_amount_cents BIGINT;
ALTER TABLE invoices ADD COLUMN total_cents BIGINT;
ALTER TABLE invoices ADD COLUMN paid_amount_cents BIGINT;

ALTER TABLE orders ADD COLUMN subtotal_cents BIGINT;
ALTER TABLE orders ADD COLUMN total_cents BIGINT;

-- Step 2: Migrate data (multiply by 100 to convert to cents)
UPDATE invoices SET 
    subtotal_cents = CAST(subtotal * 100 AS BIGINT),
    tax_amount_cents = CAST(tax_amount * 100 AS BIGINT),
    total_cents = CAST(total * 100 AS BIGINT),
    paid_amount_cents = CAST(paid_amount * 100 AS BIGINT);

UPDATE orders SET
    subtotal_cents = CAST(subtotal * 100 AS BIGINT),
    total_cents = CAST(total * 100 AS BIGINT);

-- Step 3: Drop old columns
ALTER TABLE invoices DROP COLUMN subtotal;
ALTER TABLE invoices DROP COLUMN tax_amount;
ALTER TABLE invoices DROP COLUMN total;
ALTER TABLE invoices DROP COLUMN paid_amount;

ALTER TABLE orders DROP COLUMN subtotal;
ALTER TABLE orders DROP COLUMN total;

-- Step 4: Rename new columns
ALTER TABLE invoices RENAME COLUMN subtotal_cents TO subtotal;
ALTER TABLE invoices RENAME COLUMN tax_amount_cents TO tax_amount;
ALTER TABLE invoices RENAME COLUMN total_cents TO total;
ALTER TABLE invoices RENAME COLUMN paid_amount_cents TO paid_amount;

ALTER TABLE orders RENAME COLUMN subtotal_cents TO subtotal;
ALTER TABLE orders RENAME COLUMN total_cents TO total;
```

**Completion Criteria:**
- [ ] Money value object created in pkg/money
- [ ] All financial contexts updated (invoicing, ordering, parties, accounting)
- [ ] Database migrations created
- [ ] Money arithmetic operations tested
- [ ] Precision issues resolved
- [ ] API responses still show decimal format

---

## 📝 Progress Tracking

### Week 1 Checklist
- [ ] Transaction Manager interface created
- [ ] PgxTransactionManager implemented
- [ ] Stock events added
- [ ] Money value object created
- [ ] Invoice updated to use Money

### Week 2 Checklist
- [ ] All handlers use transactions
- [ ] All inventory handlers publish events
- [ ] Money integrated across contexts
- [ ] Integration tests passing

---

## 🔍 Verification Checklist

After completing each priority level:

**P0 Verification:**
- [ ] Run accounting integration tests - verify double-entry consistency
- [ ] Run inventory integration tests - verify stock availability
- [ ] Run invoicing integration tests - verify payment calculations
- [ ] Manual test: Transaction rollback scenarios
- [ ] Manual test: Event publishing verification

**P1 Verification:**
- [ ] All command handlers publish events consistently
- [ ] Application errors have proper error codes
- [ ] No silent failures in logs
- [ ] Input validation tests passing

**P2 Verification:**
- [ ] Application layer test coverage >70%
- [ ] Infrastructure layer test coverage >60%
- [ ] All integration tests passing
- [ ] Code coverage report generated

---

## 📚 Related Documents

- [P0_CRITICAL_FIXES.md](./P0_CRITICAL_FIXES.md) - Detailed critical fixes
- [P1_IMPORTANT_FIXES.md](./P1_IMPORTANT_FIXES.md) - Important improvements
- [TESTING_STRATEGY.md](./TESTING_STRATEGY.md) - Comprehensive testing plan
- [TRANSACTION_MANAGEMENT.md](./TRANSACTION_MANAGEMENT.md) - Transaction patterns
- [MONEY_VALUE_OBJECT.md](./MONEY_VALUE_OBJECT.md) - Money implementation guide

---

## 📊 Metrics Dashboard

Track these metrics before and after improvements:

| Metric | Before | Target | Current |
|--------|--------|--------|---------|
| Domain Test Coverage | 40-70% | 80% | - |
| Application Test Coverage | 0% | 70% | - |
| Infrastructure Test Coverage | 0% | 60% | - |
| Transaction Coverage | 40% | 100% | - |
| Event Publishing Coverage | 70% | 100% | - |
| Money Value Object Usage | 20% | 100% | - |
| Event Publishing Consistency | 60% | 100% | - |

---

**Last Updated:** 2026-04-09
**Next Review:** Weekly during implementation