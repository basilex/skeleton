# Improvement Checklist

**Created:** 2026-04-09
**Last Updated:** 2026-04-09

Use this checklist to track progress of improvements. Mark items as completed `[x]` when done.

---

## 🔴 PRIORITY 0 - CRITICAL (Week 1-2)

### P0-1: Transaction Management

**Status:** ⬜ Not Started

**Files to Create:**
- [ ] `pkg/transaction/manager.go` - Interface definition
- [ ] `pkg/transaction/pgx_manager.go` - PostgreSQL implementation
- [ ] `pkg/transaction/manager_test.go` - Unit tests

**Files to Modify:**
- [ ] `internal/accounting/infrastructure/persistence/account_repository.go` - Add transaction support
- [ ] `internal/accounting/infrastructure/persistence/transaction_repository.go` - Add transaction support
- [ ] `internal/accounting/application/command/record_transaction.go` - Wrap in transaction
- [ ] `internal/inventory/infrastructure/persistence/stock_repository.go` - Add transaction support
- [ ] `internal/inventory/infrastructure/persistence/reservation_repository.go` - Add transaction support
- [ ] `internal/inventory/application/command/transfer_stock.go` - Wrap in transaction
- [ ] `internal/inventory/application/command/reserve_stock.go` - Wrap in transaction
- [ ] `cmd/api/wire.go` - Inject transaction manager

**Tests to Create:**
- [ ] `pkg/transaction/manager_test.go` - Unit tests
- [ ] `internal/accounting/application/command/record_transaction_test.go` - Integration test with rollback
- [ ] `internal/inventory/application/command/transfer_stock_test.go` - Integration test with rollback

**Verification:**
- [ ] Run accounting integration tests - all pass
- [ ] Manual test: Transfer money between accounts, verify atomicity
- [ ] Manual test: Transfer stock, verify rollback on error

**Effort:** 3 days
**Owner:** TBD
**Due Date:** Week 1

---

### P0-2: Stock Domain Events

**Status:** ⬜ Not Started

**Files to Modify:**
- [ ] `internal/inventory/domain/stock.go` - Add events field, PullEvents(), publishEvent()
- [ ] `internal/inventory/domain/events.go` - Add Stock events
- [ ] `internal/inventory/application/command/create_stock.go` - Publish events
- [ ] `internal/inventory/application/command/reserve_stock.go` - Publish events
- [ ] `internal/inventory/application/command/release_stock.go` - Publish events
- [ ] `internal/inventory/application/command/fulfill_reservation.go` - Publish events
- [ ] `internal/inventory/application/command/adjust_stock.go` - Publish events
- [ ] `internal/inventory/application/command/transfer_stock.go` - Publish events
- [ ] `internal/inventory/application/command/issue_stock.go` - Publish events
- [ ] `internal/inventory/application/command/receipt_stock.go` - Publish events

**Tests to Create:**
- [ ] `internal/inventory/domain/stock_test.go` - Test event publishing
- [ ] `internal/inventory/application/command/stock_events_test.go` - Integration test

**Verification:**
- [ ] Stock.AdjustQuantity publishes StockAdjusted event
- [ ] Stock.Reserve publishes StockReserved event
- [ ] Stock.FulfillReservation publishes StockReservationFulfilled event
- [ ] Stock.ReleaseReservation publishes StockReservationReleased event
- [ ] All command handlers publish events after save

**Effort:** 2 days
**Owner:** TBD
**Due Date:** Week 1

---

### P0-3: Money Value Object

**Status:** ⬜ Not Started

**Files to Create:**
- [ ] `pkg/money/money.go` - Money value object
- [ ] `pkg/money/money_test.go` - Unit tests

**Files to Modify:**
- [ ] `internal/invoicing/domain/invoice.go` - Use Money for amounts
- [ ] `internal/invoicing/domain/invoice_line.go` - Use Money for price
- [ ] `internal/invoicing/domain/payment.go` - Use Money for amount
- [ ] `internal/invoicing/domain/credit_note.go` - Use Money for amounts
- [ ] `internal/invoicing/domain/installment.go` - Use Money for amounts
- [ ] `internal/invoicing/infrastructure/persistence/invoice_repository.go` - Store amounts as int64
- [ ] `internal/ordering/domain/order.go` - Use Money for amounts
- [ ] `internal/ordering/infrastructure/persistence/order_repository.go` - Store amounts as int64
- [ ] `internal/parties/domain/customer.go` - Use Money for credit
- [ ] `internal/parties/infrastructure/persistence/customer_repository.go` - Store credit as int64
- [ ] `internal/accounting/domain/account.go` - Use Money for balance
- [ ] `internal/accounting/infrastructure/persistence/account_repository.go` - Store balance as int64

**Database Migrations:**
- [ ] `migrations/027_money_type_migration.up.sql` - Convert float columns to bigint
- [ ] `migrations/027_money_type_migration.down.sql` - Rollback migration

**Tests to Modify:**
- [ ] `internal/invoicing/domain/invoice_test.go` - Use Money
- [ ] `internal/ordering/domain/order_test.go` - Use Money
- [ ] `internal/parties/domain/customer_test.go` - Use Money
- [ ] `internal/accounting/domain/account_test.go` - Use Money

**Verification:**
- [ ] All amounts stored as integers (cents/minor units)
- [ ] API responses still show decimal format
- [ ] Precision errors eliminated in calculations
- [ ] All integration tests pass

**Effort:** 3 days
**Owner:** TBD
**Due Date:** Week 2

---

## 🟡 PRIORITY 1 - HIGH (Week 3-4)

### P1-1: Event Publishing Standardization

**Status:** ⬜ Not Started

**Files to Audit:**
- [ ] Audit all command handlers for event publishing
- [ ] List handlers missing event publishing
- [ ] List handlers with inconsistent error handling

**Files to Modify:**
- [ ] Create standard pattern for event publishing
- [ ] Update all handlers to use consistent pattern
- [ ] Add proper error handling for event publishing failures

**Pattern to Implement:**
```go
// Standard pattern
events := aggregate.PullEvents()
for _, event := range events {
    if err := h.bus.Publish(ctx, event); err != nil {
        // Log and fail operation for consistency
        return Result{}, fmt.Errorf("publish event: %w", err)
    }
}
```

**Verification:**
- [ ] All command handlers publish events
- [ ] All handlers fail on event publishing error
- [ ] No silent failures

**Effort:** 2 days
**Owner:** TBD
**Due Date:** Week 3

---

### P1-2: Application Error Types

**Status:** ⬜ Not Started

**Files to Create:**
- [ ] `pkg/errors/application.go` - Application error types
- [ ] `pkg/errors/codes.go` - Error codes
- [ ] `pkg/errors/application_test.go` - Unit tests

**Error Types:**
```go
type ErrorCode string

const (
    ErrCodeNotFound      ErrorCode = "NOT_FOUND"
    ErrCodeInvalidInput  ErrorCode = "INVALID_INPUT"
    ErrCodeConflict      ErrorCode = "CONFLICT"
    ErrCodeUnauthorized  ErrorCode = "UNAUTHORIZED"
    ErrCodeInternal      ErrorCode = "INTERNAL"
)

type ApplicationError struct {
    Code    ErrorCode
    Message string
    Details map[string]interface{}
    Cause   error
}
```

**Files to Modify:**
- [ ] All command handlers - Return ApplicationError instead of domain errors
- [ ] All query handlers - Return ApplicationError for not found cases
- [ ] HTTP handlers - Convert ApplicationError to HTTP responses

**Verification:**
- [ ] All handlers return ApplicationError types
- [ ] HTTP responses have consistent error format
- [ ] Error codes documented

**Effort:** 1 day
**Owner:** TBD
**Due Date:** Week 3

---

### P1-3: Silent Failures Fix

**Status:** ⬜ Not Started

**Files to Fix:**
- [ ] `internal/identity/application/command/revoke_user_sessions.go` - Return aggregated errors
- [ ] `internal/identity/application/command/preferences.go` - Handle SetTheme/SetLanguage errors
- [ ] `internal/notifications/application/command/update_preferences.go` - Handle errors properly

**Pattern to Implement:**
```go
// For batch operations, aggregate errors
type MultiError struct {
    Errors []error
}

func (e *MultiError) Error() string {
    var msgs []string
    for _, err := range e.Errors {
        msgs = append(msgs, err.Error())
    }
    return strings.Join(msgs, "; ")
}

// Usage
var errors []error
for _, item := range items {
    if err := process(item); err != nil {
        errors = append(errors, err)
    }
}
if len(errors) > 0 {
    return &MultiError{Errors: errors}
}
```

**Verification:**
- [ ] All batch operations return aggregated errors
- [ ] All preference updates propagate errors
- [ ] No `_ = fn()` patterns ignoring errors

**Effort:** 1 day
**Owner:** TBD
**Due Date:** Week 3

---

### P1-4: Input Validation

**Status:** ⬜ Not Started

**Files to Create:**
- [ ] `pkg/validation/validator.go` - Validation middleware
- [ ] `pkg/validation/rules.go` - Common validation rules
- [ ] `pkg/validation/validator_test.go` - Unit tests

**Files to Modify:**
- [ ] All HTTP handlers - Add validation middleware
- [ ] All command/query DTOs - Add validation tags

**Example:**
```go
type CreateInvoiceRequest struct {
    Number     string            `json:"number" validate:"required,min=1,max=50"`
    CustomerID string            `json:"customer_id" validate:"required,uuid"`
    Lines      []InvoiceLineDTO  `json:"lines" validate:"required,min=1,dive"`
}

type InvoiceLineDTO struct {
    ItemID   string  `json:"item_id" validate:"required,uuid"`
    Quantity int     `json:"quantity" validate:"required,min=1"`
    Price    float64 `json:"price" validate:"required,min=0.01"`
}
```

**Verification:**
- [ ] All inputs validated before processing
- [ ] Validation errors return 400 with details
- [ ] No invalid data reaches domain layer

**Effort:** 2 days
**Owner:** TBD
**Due Date:** Week 4

---

### P1-5: Rate Limiting

**Status:** ⬜ Not Started

**Files to Modify:**
- [ ] `cmd/api/routes.go` - Add rate limiting to auth endpoints
- [ ] `pkg/middleware/rate_limit.go` - Rate limiting middleware (already exists)

**Endpoints to Protect:**
- [ ] POST /auth/login - 5 attempts per minute per IP
- [ ] POST /auth/register - 5 attempts per hour per IP
- [ ] POST /auth/forgot-password - 3 attempts per hour per email
- [ ] POST /auth/reset-password - 3 attempts per hour per token

**Configuration:**
```go
// Different limits for different endpoints
var rateLimits = map[string]RateLimitConfig{
    "/auth/login": {
        Requests: 5,
        Window:   time.Minute,
        KeyFunc:  func(c *gin.Context) string { return c.ClientIP() },
    },
    "/auth/register": {
        Requests: 5,
        Window:   time.Hour,
        KeyFunc:  func(c *gin.Context) string { return c.ClientIP() },
    },
}
```

**Verification:**
- [ ] Rate limiting active on auth endpoints
- [ ] 429 Too Many Requests returned when limit exceeded
- [ ] Rate limit headers included in response

**Effort:** 1 day
**Owner:** TBD
**Due Date:** Week 4

---

## 🟢 PRIORITY 2 - MEDIUM (Week 5-6)

### P2-1: Application Layer Tests

**Status:** ⬜ Not Started

**Test Coverage Target:** 70%

**Tests to Create:**

**Accounting:**
- [ ] `internal/accounting/application/command/create_account_test.go`
- [ ] `internal/accounting/application/command/record_transaction_test.go`
- [ ] `internal/accounting/application/query/get_account_test.go`

**Invoicing:**
- [ ] `internal/invoicing/application/command/create_invoice_test.go`
- [ ] `internal/invoicing/application/command/send_invoice_test.go`
- [ ] `internal/invoicing/application/command/record_payment_test.go`
- [ ] `internal/invoicing/application/query/get_invoice_test.go`

**Inventory:**
- [ ] `internal/inventory/application/command/reserve_stock_test.go`
- [ ] `internal/inventory/application/command/fulfill_reservation_test.go`
- [ ] `internal/inventory/application/command/transfer_stock_test.go`

**Ordering:**
- [ ] `internal/ordering/application/command/create_order_test.go`
- [ ] `internal/ordering/application/command/update_order_status_test.go`

**Mock Files to Create:**
- [ ] Use mockery to generate all repository mocks

**Verification:**
- [ ] Application layer coverage >70%
- [ ] All command handlers tested
- [ ] All query handlers tested
- [ ] Mocks generated for all dependencies

**Effort:** 5 days
**Owner:** TBD
**Due Date:** Week 5

---

### P2-2: Infrastructure Layer Tests

**Status:** ⬜ Not Started

**Test Coverage Target:** 60%

**Tests to Create:**

**Accounting:**
- [ ] `internal/accounting/infrastructure/persistence/account_repository_test.go`
- [ ] `internal/accounting/infrastructure/persistence/journal_entry_repository_test.go`

**Invoicing:**
- [ ] `internal/invoicing/infrastructure/persistence/invoice_repository_test.go`
- [ ] `internal/invoicing/infrastructure/persistence/credit_note_repository_test.go`

**Inventory:**
- [ ] `internal/inventory/infrastructure/persistence/stock_repository_test.go`
- [ ] `internal/inventory/infrastructure/persistence/reservation_repository_test.go`

**Ordering:**
- [ ] `internal/ordering/infrastructure/persistence/order_repository_test.go`

**Verification:**
- [ ] Infrastructure layer coverage >60%
- [ ] All repositories tested
- [ ] Integration tests use testcontainers

**Effort:** 4 days
**Owner:** TBD
**Due Date:** Week 5

---

### P2-3: HTTP Handler Tests

**Status:** ⬜ Not Started

**Test Coverage Target:** 60%

**Tests to Create:**
- [ ] `internal/invoicing/ports/http/invoice_handler_test.go`
- [ ] `internal/inventory/ports/http/stock_handler_test.go`
- [ ] `internal/ordering/ports/http/order_handler_test.go`
- [ ] `internal/identity/ports/http/auth_handler_test.go`

**Verification:**
- [ ] HTTP handler coverage >60%
- [ ] All endpoints tested
- [ ] Error responses tested

**Effort:** 3 days
**Owner:** TBD
**Due Date:** Week 6

---

## 🔵 PRIORITY 3 - LOW (Ongoing)

### P3-1: OpenAPI/Swagger Documentation

**Status:** ⬜ Not Started

**Files to Create:**
- [ ] `docs/api/openapi.yaml` - OpenAPI specification
- [ ] `docs/api/README.md` - API documentation guide

**Files to Modify:**
- [ ] Add swagger comments to all HTTP handlers
- [ ] Generate swagger docs automatically

**Effort:** 2 days
**Owner:** TBD
**Due Date:** Ongoing

---

### P3-2: Deployment Runbooks

**Status:** ⬜ Not Started

**Files to Create:**
- [ ] `docs/DEPLOYMENT_RUNBOOK.md` - Step-by-step deployment guide
- [ ] `docs/ROLLBACK_PROCEDURE.md` - Rollback procedures
- [ ] `docs/MONITORING_SETUP.md` - Monitoring configuration

**Effort:** 1 day
**Owner:** TBD
**Due Date:** Ongoing

---

### P3-3: Troubleshooting Guides

**Status:** ⬜ Not Started

**Files to Create:**
- [ ] `docs/TROUBLESHOOTING.md` - Common issues and solutions
- [ ] `docs/FAQ.md` - Frequently asked questions

**Effort:** 1 day
**Owner:** TBD
**Due Date:** Ongoing

---

## 📈 Progress Tracking

### Weekly Metrics

| Week | Domain Tests | App Tests | Infra Tests | HTTP Tests | Overall |
|------|--------------|-----------|-------------|------------|---------|
| 1 | 40-70% | 0% | 0% | 0% | 0-25% |
| 2 | 80% | 0% | 10% | 0% | 25% |
| 3 | 80% | 30% | 30% | 0% | 40% |
| 4 | 80% | 50% | 40% | 20% | 50% |
| 5 | 80% | 60% | 50% | 40% | 60% |
| 6 | 80% | 70% | 60% | 60% | 70% |

### Monthly Metrics

| Month | Transaction Coverage | Event Coverage | Money Usage | Score |
|-------|---------------------|----------------|-------------|-------|
| 1 | 40% | 70% | 20% | 43% |
| 2 | 100% | 100% | 100% | 100% |

---

## ✅ Completion Criteria

### P0 Completion
- [ ] All transaction operations wrapped in transaction manager
- [ ] All inventory commands publish domain events
- [ ] All monetary values use Money value object
- [ ] Integration tests passing for accounting, inventory
- [ ] Code review approved

### P1 Completion
- [ ] All command handlers publish events consistently
- [ ] Application errors have proper error codes
- [ ] No silent failures in handlers
- [ ] Input validation active on all endpoints
- [ ] Rate limiting active on auth endpoints

### P2 Completion
- [ ] Application layer test coverage >70%
- [ ] Infrastructure layer test coverage >60%
- [ ] HTTP handler test coverage >60%
- [ ] Integration tests stable

### P3 Completion
- [ ] OpenAPI specification generated
- [ ] Deployment runbooks created
- [ ] Troubleshooting guides available

---

## 🎯 Definition of Done

Each task is considered done when:

1. **Code Complete**
   - [ ] All files created/modified
   - [ ] Code follows project conventions
   - [ ] No linting errors
   - [ ] No LSP errors

2. **Tests Complete**
   - [ ] Unit tests written
   - [ ] Integration tests written (where applicable)
   - [ ] All tests passing
   - [ ] Coverage meets target

3. **Documentation**
   - [ ] Code comments added
   - [ ] API documentation updated
   - [ ] Migration files documented

4. **Review**
   - [ ] Self-review completed
   - [ ] Code review approved
   - [ ] QA verification passed

5. **Integration**
   - [ ] Merged to main branch
   - [ ] CI/CD pipeline passing
   - [ ] Deployed to staging
   - [ ] Manual testing passed

---

**Last Updated:** 2026-04-09
**Update Frequency:** Weekly