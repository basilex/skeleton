# Testing Strategy

**Created:** 2026-04-09
**Status:** Active
**Goal:** Achieve comprehensive test coverage across all layers

---

## 📊 Current Testing State

### Coverageby Layer

| Layer | Current Coverage | Target Coverage | Gap |
|-------|-------------------|-----------------|-----|
| Domain | 40-70% | 80% | -10-30% |
| Application | 0% | 70% | -70% |
| Infrastructure | 0% | 60% | -60% |
| HTTP Handlers | 0% | 60% | -60% |
| Integration | 0% | 50% | -50% |

### Test Distribution

```
Total Test Files: 51
Domain Tests: ~40 files
Application Tests: 0 files  ⚠️ Critical
Infrastructure Tests: 11 files (session/preferences)
HTTP Tests: 0 files  ⚠️ Critical
```

---

## 🎯 Testing Strategy Pyramid

```
        /\
       /  \      End-to-End Tests (10%)
      /----\     - Key user journeys
     /      \
    /--------\   Integration Tests (30%)
   /          \  - Repository tests
  /            \ - Handler tests with real DB
 /--------------\ 
/   Unit Tests    \ Unit Tests (60%)
------------------- - Domain logic tests
                    - Application handler tests
                    - Value object tests
```

---

## 📋 Phase 1: Domain Layer Testing (80% Target)

### Current Issues
- Missing tests for critical business rules
- Some domains have minimal coverage
- Missing edge case tests

### Required Tests by Context

#### 1. Accounting Domain

**Files to Test:**
- `internal/accounting/domain/journal_entry_test.go` (exists, needs expansion)
- `internal/accounting/domain/account_test.go` (exists, needs expansion)
- `internal/accounting/domain/accounting_period_test.go` (MISSING)
- `internal/accounting/domain/reconciliation_test.go` (MISSING)

**Test Cases Needed:**

```go
// internal/accounting/domain/journal_entry_test.go

func TestJournalEntry_Post_Requires_BalancedLines(t *testing.T) {
    // Arrange: Create unbalanced journal entry
    entry := createUnbalancedJournalEntry()
    
    // Act: Try to post
    err := entry.Post()
    
    // Assert: Should fail
    require.Error(t, err)
    assert.Contains(t, err.Error(), "not balanced")
}

func TestJournalEntry_CannotPostDraftTwice(t *testing.T) {
    // Arrange: Posted journal entry
    entry := createPostedJournalEntry()
    
    // Act: Try to post again
    err := entry.Post()
    
    // Assert: Should fail
    require.Error(t, err)
    assert.Contains(t, err.Error(), "already posted")
}

func TestJournalEntry_Void_RequiresPosted(t *testing.T) {
    // Arrange: Draft journal entry
    entry := createDraftJournalEntry()
    
    // Act: Try to void
    err := entry.Void()
    
    // Assert: Should fail
    require.Error(t, err)
    assert.Contains(t, err.Error(), "only posted")
}

func TestJournalEntry_Balanced_CreditsEqualDebits(t *testing.T) {
    // Arrange: Create entry with balanced lines
    entry := createBalancedJournalEntry()
    
    // Act: Check balance
    isBalanced := entry.IsBalanced()
    
    // Assert: Should be balanced
    assert.True(t, isBalanced)
}
```

**New Test Files to Create:**

```go
// internal/accounting/domain/accounting_period_test.go

func TestAccountingPeriod_CannotClose_WithPostedJournalEntries(t *testing.T)
func TestAccountingPeriod_CannotReopen_OnceLocked(t *testing.T)
func TestAccountingPeriod_Duration_Validation(t *testing.T)

// internal/accounting/domain/reconciliation_test.go

func TestReconciliation_CannotComplete_WithUnmatched_items(t *testing.T)
func TestReconciliation_Match_Validation(t *testing.T)
```

#### 2. Inventory Domain

**Files to Test:**
- `internal/inventory/domain/stock_test.go` (MISSING - CRITICAL)
- `internal/inventory/domain/lot_test.go` (MISSING)
- `internal/inventory/domain/stock_take_test.go` (MISSING)

**Critical Tests Needed:**

```go
// internal/inventory/domain/stock_test.go

func TestStock_Reserve_InsufficientQuantity(t *testing.T) {
    stock := createStockWithQuantity(100.0)
    
    err := stock.Reserve(150.0, reservationID)
    
    require.Error(t, err)
    assert.Equal(t, ErrInsufficientStock, err)
}

func TestStock_Reserve_AvailableDecreases(t *testing.T) {
    stock := createStockWithQuantity(100.0)
    
    err := stock.Reserve(30.0, reservationID)
    
    require.NoError(t, err)
    assert.Equal(t, 100.0, stock.Quantity())
    assert.Equal(t, 30.0, stock.ReservedQuantity())
    assert.Equal(t, 70.0, stock.AvailableQuantity())
}

func TestStock_FulfillReservation_QuantityDecreases(t *testing.T) {
    stock := createStockWithQuantity(100.0)
    stock.Reserve(30.0, reservationID)
    
    stock.FulfillReservation(30.0)
    
    assert.Equal(t, 70.0, stock.Quantity())
    assert.Equal(t, 0.0, stock.ReservedQuantity())
    assert.Equal(t, 70.0, stock.AvailableQuantity())
}

func TestStock_AdjustQuantity_NegativeNotAllowed(t *testing.T) {
    stock := createStockWithQuantity(50.0)
    
    err := stock.AdjustQuantity(-60.0, movementID)
    
    require.Error(t, err)
    assert.Equal(t, ErrInsufficientStock, err)
}

func TestStock_ReorderPoint_Logic(t *testing.T) {
    stock := createStockWithQuantity(100.0)
    stock.SetReorderPoint(20.0)
    
    stock.AdjustQuantity(-85.0, movementID)
    
    assert.True(t, stock.NeedsReorder())
}
```

#### 3. Invoicing Domain

**Files to Test:**
- `internal/invoicing/domain/invoice_test.go` (exists, needs expansion)
- `internal/invoicing/domain/credit_note_test.go` (exists, needs expansion)
- `internal/invoicing/domain/installment_test.go` (MISSING)

**Additional Test Cases:**

```go
// internal/invoicing/domain/invoice_test.go

func TestInvoice_RecordPayment_ExceedsTotal(t *testing.T) {
    invoice := createInvoiceWithTotal(100.00)
    
    payment, err := invoice.RecordPayment(150.00, PaymentMethodBankTransfer, "ref")
    
    require.Error(t, err)
    assert.Nil(t, payment)
    assert.Contains(t, err.Error(), "exceeds")
}
```

---

## 📋 Phase 2: Application Layer Testing (70% Target)

### Current State
- **0 test files** for application layer
- All command/query handlers untested
- Critical for verifying business workflows

### Test Structure

```
internal/
  invoicing/
    application/
      command/
        create_invoice_test.go
        send_invoice_test.go
        record_payment_test.go
        cancel_invoice_test.go
      query/
        get_invoice_test.go
        list_invoices_test.go
```

### Test Pattern: Command Handlers

```go
// internal/invoicing/application/command/create_invoice_test.go

func TestCreateInvoiceHandler_Handle_ValidCommand(t *testing.T) {
    // Arrange
    repo := &mocks.MockInvoiceRepository{}
    bus := &mocks.MockEventBus{}
    handler := NewCreateInvoiceHandler(repo, bus)
    
    cmd := CreateInvoiceCommand{
        CustomerID: "customer-123",
        Lines: []InvoiceLineDTO{
            {ItemID: "item-1", Quantity: 2, Price: 100.00},
        },
    }
    
    // Act
    result, err := handler.Handle(context.Background(), cmd)
    
    // Assert
    require.NoError(t, err)
    assert.NotEmpty(t, result.InvoiceID)
    repo.AssertCalled(t, "Save", mock.Anything, mock.Anything)
    bus.AssertCalled(t, "Publish", mock.Anything, mock.Anything)
}

func TestCreateInvoiceHandler_Handle_InvalidCustomerID(t *testing.T) {
    // Arrange
    repo := &mocks.MockInvoiceRepository{}
    bus := &mocks.MockEventBus{}
    handler := NewCreateInvoiceHandler(repo, bus)
    
    cmd := CreateInvoiceCommand{
        CustomerID: "",  // Invalid
        Lines: []InvoiceLineDTO{
            {ItemID: "item-1", Quantity: 1, Price: 100.00},
        },
    }
    
    // Act
    result, err := handler.Handle(context.Background(), cmd)
    
    // Assert
    require.Error(t, err)
    assert.Nil(t, result)
    repo.AssertNotCalled(t, "Save")
}

func TestCreateInvoiceHandler_Handle_RepositoryError(t *testing.T) {
    // Arrange
    repo := &mocks.MockInvoiceRepository{}
    repo.On("Save", mock.Anything, mock.Anything).Return(errors.New("db error"))
    bus := &mocks.MockEventBus{}
    handler := NewCreateInvoiceHandler(repo, bus)
    
    cmd := CreateInvoiceCommand{
        CustomerID: "customer-123",
        Lines: []InvoiceLineDTO{
            {ItemID: "item-1", Quantity: 1, Price: 100.00},
        },
    }
    
    // Act
    result, err := handler.Handle(context.Background(), cmd)
    
    // Assert
    require.Error(t, err)
    assert.Contains(t, err.Error(), "save invoice")
}
```

### Test Pattern: Query Handlers

```go
// internal/invoicing/application/query/get_invoice_test.go

func TestGetInvoiceHandler_Handle_ExistingInvoice(t *testing.T) {
    // Arrange
    repo := &mocks.MockInvoiceRepository{}
    expectedInvoice := createTestInvoice()
    repo.On("FindByID", mock.Anything, mock.Anything).Return(expectedInvoice, nil)
    
    handler := NewGetInvoiceHandler(repo)
    query := GetInvoiceQuery{InvoiceID: "invoice-123"}
    
    // Act
    result, err := handler.Handle(context.Background(), query)
    
    // Assert
    require.NoError(t, err)
    assert.NotNil(t, result)
    assert.Equal(t, "invoice-123", result.ID)
}

func TestGetInvoiceHandler_Handle_NonExistingInvoice(t *testing.T) {
    // Arrange
    repo := &mocks.MockInvoiceRepository{}
    repo.On("FindByID", mock.Anything, mock.Anything).Return(nil, domain.ErrInvoiceNotFound)
    
    handler := NewGetInvoiceHandler(repo)
    query := GetInvoiceQuery{InvoiceID: "non-existing"}
    
    // Act
    result, err := handler.Handle(context.Background(), query)
    
    // Assert
    require.Error(t, err)
    assert.Nil(t, result)
}
```

### Mocking Framework

Using `testify/mock` for all mocks:

```go
// internal/mocks/invoice_repository.go

package mocks

import (
    "context"
    "github.com/stretchr/testify/mock"
    "github.com/basilex/skeleton/internal/invoicing/domain"
)

type MockInvoiceRepository struct {
    mock.Mock
}

func (m *MockInvoiceRepository) Save(ctx context.Context, invoice *domain.Invoice) error {
    args := m.Called(ctx, invoice)
    return args.Error(0)
}

func (m *MockInvoiceRepository) FindByID(ctx context.Context, id domain.InvoiceID) (*domain.Invoice, error) {
    args := m.Called(ctx, id)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*domain.Invoice), args.Error(1)
}

// ... other methods
```

---

## 📋 Phase 3: Infrastructure Layer Testing (60% Target)

### Integration Tests with Testcontainers

Already have good examples in:
- `internal/identity/infrastructure/persistence/session_repository_test.go`
- `internal/identity/infrastructure/persistence/preferences_repository_test.go`

### Pattern to Follow

```go
// internal/invoicing/infrastructure/persistence/invoice_repository_test.go

func TestInvoiceRepository_Save(t *testing.T) {
    pool := testutil.SetupPostgres(t)
    testutil.RunMigrations(t, pool, testutil.DefaultSchema)
    
    repo := persistence.NewInvoiceRepository(pool)
    ctx := context.Background()
    
    invoice := createTestInvoice()
    
    t.Run("save new invoice", func(t *testing.T) {
        err := repo.Save(ctx, invoice)
        require.NoError(t, err)
        
        found, err := repo.FindByID(ctx, invoice.ID())
        require.NoError(t, err)
        assert.Equal(t, invoice.ID(), found.ID())
        assert.Equal(t, invoice.Number(), found.Number())
    })
    
    t.Run("update existing invoice", func(t *testing.T) {
        invoice.Send()
        err := repo.Save(ctx, invoice)
        require.NoError(t, err)
        
        found, err := repo.FindByID(ctx, invoice.ID())
        require.NoError(t, err)
        assert.Equal(t, domain.InvoiceStatusSent, found.Status())
    })
}

func TestInvoiceRepository_FindByCustomer(t *testing.T) {
    pool := testutil.SetupPostgres(t)
    testutil.RunMigrations(t, pool, testutil.DefaultSchema)
    
    repo := persistence.NewInvoiceRepository(pool)
    ctx := context.Background()
    
    // Create test data
    customerID := "customer-1"
    for i := 0; i < 3; i++ {
        invoice := createTestInvoiceForCustomer(customerID)
        repo.Save(ctx, invoice)
    }
    
    t.Run("find invoices by customer", func(t *testing.T) {
        invoices, err := repo.FindByCustomer(ctx, customerID)
        require.NoError(t, err)
        assert.Len(t, invoices, 3)
    })
}
```

### Required Integration Tests by Context

**Accounting:**
- [ ] `account_repository_test.go` - Account CRUD, hierarchy operations
- [ ] `journal_entry_repository_test.go` - Journal entry operations
- [ ] `transaction_repository_test.go` - Transaction recording

**Invoicing:**
- [ ] `invoice_repository_test.go` - Invoice operations
- [ ] `credit_note_repository_test.go` - Credit note operations

**Inventory:**
- [ ] `stock_repository_test.go` - Stock operations
- [ ] `stock_reservation_repository_test.go` - Reservation operations
- [ ] `stock_movement_repository_test.go` - Movement tracking

**Ordering:**
- [ ] `order_repository_test.go` - Order operations

---

## 📋 Phase 4: HTTP Handler Testing (60% Target)

### Pattern: HTTP Integration Tests

```go
// internal/invoicing/ports/http/invoice_handler_test.go

func TestInvoiceHandler_CreateInvoice(t *testing.T) {
    // Setup
    router := setupTestRouter()
    handler := NewInvoiceHandler(/* mocked dependencies */)
    router.POST("/invoices", handler.Create)
    
    t.Run("valid invoice creation", func(t *testing.T) {
        body := `{
            "customer_id": "customer-123",
            "lines": [
                {
                    "item_id": "item-1",
                    "quantity": 2,
                    "price": 100.00
                }
            ]
        }`
        
        req := httptest.NewRequest("POST", "/invoices", strings.NewReader(body))
        req.Header.Set("Content-Type", "application/json")
        w := httptest.NewRecorder()
        
        router.ServeHTTP(w, req)
        
        assert.Equal(t, http.StatusCreated, w.Code)
        
        var response map[string]interface{}
        json.Unmarshal(w.Body.Bytes(), &response)
        assert.NotEmpty(t, response["id"])
    })
    
    t.Run("invalid customer_id", func(t *testing.T) {
        body := `{
            "customer_id": "",
            "lines": []
        }`
        
        req := httptest.NewRequest("POST", "/invoices", strings.NewReader(body))
        w := httptest.NewRecorder()
        
        router.ServeHTTP(w, req)
        
        assert.Equal(t, http.StatusBadRequest, w.Code)
    })
}
```

---

## 🛠️ Test Infrastructure

### 1. Test Utilities Package

Create comprehensive test utilities:

```go
// pkg/testutil/builder.go

package testutil

import (
    "time"
    
    "github.com/basilex/skeleton/internal/invoicing/domain"
)

// InvoiceBuilder for test data
type InvoiceBuilder struct {
    invoice *domain.Invoice
}

func NewInvoiceBuilder() *InvoiceBuilder {
    invoice, _ := domain.NewInvoice("INV-001", "customer-1", []domain.InvoiceLine{
        {
            ItemID:   "item-1",
            Quantity: 1,
            Price:    100.00,
        },
    })
    return &InvoiceBuilder{invoice: invoice}
}

func (b *InvoiceBuilder) WithNumber(number string) *InvoiceBuilder {
    // Use reflection or create new invoice
    return b
}

func (b *InvoiceBuilder) WithCustomerID(customerID string) *InvoiceBuilder {
    return b
}

func (b *InvoiceBuilder) WithLines(lines []domain.InvoiceLine) *InvoiceBuilder {
    return b
}

func (b *InvoiceBuilder) WithStatus(status domain.InvoiceStatus) *InvoiceBuilder {
    // Use domain methods to change status
    return b
}

func (b *InvoiceBuilder) Build() *domain.Invoice {
    return b.invoice
}
```

### 2. Mock Generation

Use `mockery` to generate mocks:

```yaml
# .mockery.yaml
quiet: False
dry-run: False
with-expecter: True
all: True
filename: "mock_{{.InterfaceName}}.go"
dir: "{{.InterfaceDir}}/mocks"
mockname: "Mock{{.InterfaceName}}"
outpkg: "mocks"
```

Run mock generation:

```bash
mockery --config=.mockery.yaml
```

### 3. Test Fixtures

```go
// pkg/testutil/fixtures/fixtures.go

package fixtures

import (
    "time"
    
    "github.com/basilex/skeleton/internal/invoicing/domain"
    "github.com/basilex/skeleton/pkg/money"
)

func Invoice() *domain.Invoice {
    invoice, _ := domain.NewInvoice("INV-001", "customer-1", []domain.InvoiceLine{
        {
            ItemID:   "item-1",
            Quantity: 2,
            Price:    money.New(10000, "USD"), // $100.00
        },
    })
    return invoice
}

func InvoiceSent() *domain.Invoice {
    invoice := Invoice()
    invoice.Send()
    return invoice
}

func InvoicePaid() *domain.Invoice {
    invoice := InvoiceSent()
    payment := money.New(22000, "USD") // $220.00 (including tax)
    invoice.RecordPayment(payment, domain.PaymentMethodBankTransfer, "ref-123")
    return invoice
}
```

---

## 📊 Coverage Goals by Week

### Week 1 Targets
- [ ] Domain layer tests: 60% coverage
- [ ] Create missing test files for inventory domain
- [ ] Add edge case tests for accounting domain

### Week 2 Targets
- [ ] Application layer tests: 50% coverage
- [ ] Create command handler tests
- [ ] Create query handler tests

### Week 3 Targets
- [ ] Infrastructure layer tests: 40% coverage
- [ ] Add integration tests for repositories
- [ ] Remove integration test flakiness

### Week 4 Targets
- [ ] HTTP handler tests: 40% coverage
- [ ] Add API integration tests
- [ ] Coverage reports automated

---

## 🔍 Test Quality Standards

### 1. Test Naming Convention

```go
// Pattern: Test<Unit>_<Scenario>_<ExpectedResult>

func TestInvoice_RecordPayment_ValidPayment_Success(t *testing.T) {}
func TestInvoice_RecordPayment_InvalidAmount_Error(t *testing.T) {}
func TestInvoice_RecordPayment_ExceedsTotal_Error(t *testing.T) {}
```

### 2. Test Structure: Arrange-Act-Assert

```go
func TestJournalEntry_Post_BalancedLines_Success(t *testing.T) {
    // Arrange
    entry := createJournalEntryWithLines([]JournalLine{
        {AccountID: "acc-1", Debit: 100.0, Credit: 0.0},
        {AccountID: "acc-2", Debit: 0.0, Credit: 100.0},
    })
    
    // Act
    err := entry.Post()
    
    // Assert
    require.NoError(t, err)
    assert.Equal(t, JournalEntryStatusPosted, entry.Status())
}
```

### 3. Test Coverage Requirements

Each test file must cover:
- [ ] Happy path (success scenario)
- [ ] Edge cases (boundary conditions)
- [ ] Error cases (validation failures)
- [ ] State transitions (lifecycle changes)
- [ ] Invariant enforcement (business rules)

---

## 🚀 Running Tests

### Unit Tests
```bash
make test-unit
# or
go test ./internal/.../domain/... -v -cover
```

### Integration Tests
```bash
make test-integration
# or
go test ./internal/.../infrastructure/... -v -cover -tags=integration
```

### Coverage Report
```bash
make test-coverage
# or
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html
```

### Specific Context
```bash
go test ./internal/invoicing/... -v -cover
```

---

## 📈 Continuous Improvement

### Monthly Review
1. Analyze test coverage report
2. Identify uncovered critical paths
3. Add flaky test detection
4. Refactor test utilities
5. Update test documentation

### CI/CD Integration
```yaml
# .github/workflows/test.yml
name: Tests
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v3
      - run: make test-unit
      - run: make test-integration
      - run: make test-coverage
      - uses: codecov/codecov-action@v3
```

---

**Last Updated:** 2026-04-09
**Next Review:** Weekly during testing phase