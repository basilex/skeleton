# Contexts Hardening - Working Plan

**Created:** 2026-04-09  
**Status:** In Progress  
**Goal:** Make all bounded contexts "domain-rich" with proper business logic

---

## 📊 Overall Progress

**Total Tasks:** 42  
**Completed:** 6  
**In Progress:** 0  
**Pending:** 36

---

## 🔴 PRIORITY 1 - CRITICAL (Must Fix Before Production)

### 1. Accounting Context (Score: 58/100 → Target: 85/100)

**Status:** ✅ Completed  
**Effort:** 5 days  
**Priority:** CRITICAL

**Why Critical:** Double-entry bookkeeping requires proper journal entry validation

#### Tasks:

- [x] Create `JournalEntry` aggregate
  - File: `internal/accounting/domain/journal_entry.go`
  - Status: ✅ Completed
  - Effort: 1 day
  - Requires:
    - `JournalLine` value object
    - Double-entry validation (Debits == Credits)
    - Status: draft/posted/voided
    - Business rules

- [x] Create `AccountingPeriod` aggregate
  - File: `internal/accounting/domain/accounting_period.go`
  - Status: ✅ Completed
  - Effort: 1 day
  - Requires:
    - Period status: open/closed/locked
    - Date range validation
    - Close period business rules

- [x] Create `Reconciliation` aggregate
  - File: `internal/accounting/domain/reconciliation.go`
  - Status: ✅ Completed
  - Effort: 1 day
  - Requires:
    - Bank statement matching
    - Discrepancy tracking
    - Status management

- [ ] Add hierarchy management to `Account`
  - File: `internal/accounting/domain/account.go` (modify)
  - Status: ⬜ Not Started
  - Effort: 0.5 day
  - Requires:
    - Parent-child relationships
    - Account tree validation
    - Descendant/balance validation

- [x] Add domain events
  - Files: `internal/accounting/domain/events.go` (modify)
  - Status: ✅ Completed
  - Effort: 0.5 day
  - Events to add:
    - `JournalEntryPosted`
    - `AccountPeriodClosed`
    - `ReconciliationCompleted`

- [x] Add tests
  - File: `internal/accounting/domain/journal_entry_test.go` (create)
  - Status: ✅ Completed
  - Effort: 1 day

**Definition of Done:**
- [x] All aggregates implemented
- [x] Business rules validated in domain
- [x] Domain events published
- [x] Tests passing
- [ ] Wire.go updated

---

### 2. Documents Context (Score: 63/100 → Target: 80/100)

**Status:** ✅ Completed  
**Effort:** 3 days  
**Priority:** CRITICAL

**Why Critical:** Contract workflow requires approval process

#### Tasks:

- [x] Create `ApprovalWorkflow` aggregate
  - File: `internal/documents/domain/approval_workflow.go`
  - Status: ✅ Completed
  - Effort: 1.5 days
  - Requires:
    - Workflow steps
    - Approval/rejection logic
    - Status transitions
    - Approver assignment

- [x] Create `DocumentVersion` value object
  - File: `internal/documents/domain/document_version.go`
  - Status: ✅ Completed
  - Effort: 0.5 day
  - Requires:
    - Version numbering
    - Change tracking
    - Diff logic

- [x] Add versioning to `Document`
  - File: `internal/documents/domain/document.go` (modify)
  - Status: ✅ Completed
  - Effort: 0.5 day
  - Requires:
    - Version history
    - GetVersion() method
    - CreateVersion() method

- [x] Add domain events
  - Files: `internal/documents/domain/events.go` (modify)
  - Status: ✅ Completed
  - Effort: 0.5 day
  - Events to add:
    - `ApprovalRequested`
    - `ApprovalCompleted`
    - `DocumentVersionCreated`

**Definition of Done:**
- [x] ApprovalWorkflow implemented
- [x] Tests passing
- [ ] Wire.go updated

---

## 🟡 PRIORITY 2 - HIGH (Should Fix Soon)

### 3. Parties Context (Score: 88/100 → Target: 92/100)

**Status:** ✅ Completed  
**Effort:** 1 day  
**Priority:** HIGH

#### Tasks:

- [x] Add `CreditLimit` to Customer
  - File: `internal/parties/domain/customer.go` (modify)
  - Status: ✅ Completed
  - Effort: 2 hours
  - Requires:
    - `Money` value object
    - Credit limit validation
    - Current credit tracking
    - Business rule: credit <= limit

- [x] Add `PerformanceScore` to Supplier
  - File: `internal/parties/domain/supplier.go` (modify)
  - Status: ✅ Completed
  - Effort: 2 hours
  - Requires:
    - Score calculation
    - Rating thresholds
    - Business rules for rating

- [x] Add domain events
  - File: `internal/parties/domain/events.go` (modify)
  - Status: ✅ Completed
  - Effort: 1 hour
  - Events to add:
    - `CustomerCreditLimitChanged`
    - `SupplierRatingUpdated`

- [x] Add tests
  - File: `internal/parties/domain/customer_test.go` (modify)
  - Status: ✅ Completed
  - Effort: 3 hours

---

### 4. Catalog Context (Score: 76/100 → Target: 85/100)

**Status:** ✅ Completed  
**Effort:** 2 days  
**Priority:** HIGH

#### Tasks:

- [x] Create `ProductVariant` aggregate
  - File: `internal/catalog/domain/product_variant.go`
  - Status: ✅ Completed
  - Effort: 1 day
  - Requires:
    - Attributes map (Size, Color, etc.)
    - SKU generation
    - Price override
    - Inventory tracking

- [x] Create `PricingRule` aggregate
  - File: `internal/catalog/domain/pricing_rule.go`
  - Status: ✅ Completed
  - Effort: 0.5 day
  - Requires:
    - Volume discount rules
    - Date range validity
    - Customer group targeting

- [x] Add variant support to `Item`
  - File: `internal/catalog/domain/item.go` (modify)
  - Status: ✅ Completed
  - Effort: 0.5 day
  - Requires:
    - HasVariants flag
    - GetVariants() method
    - CreateVariant() method

---

### 5. Inventory Context (Score: 90/100 → Target: 95/100)

**Status:** ✅ Completed  
**Effort:** 2 days  
**Priority:** HIGH

#### Tasks:

- [x] Create `Lot` aggregate
  - File: `internal/inventory/domain/lot.go`
  - Status: ✅ Completed
  - Effort: 1 day
  - Requires:
    - Lot number
    - Serial numbers
    - Expiry date
    - Stock tracking

- [x] Create `StockTake` aggregate
  - File: `internal/inventory/domain/stock_take.go`
  - Status: ✅ Completed
  - Effort: 0.5 day
  - Requires:
    - Count session
    - Variance tracking
    - Adjustment creation

- [x] Create `StockLocation` value object
  - File: `internal/inventory/domain/stock_location.go`
  - Status: ✅ Completed
  - Effort: 0.5 day
  - Requires:
    - Warehouse ID
    - Zone
    - Aisle
    - Bin/Shelf

---

### 6. Invoicing Context (Score: 90/100 → Target: 95/100)

**Status:** ✅ Completed  
**Effort:** 2 days  
**Priority:** HIGH

#### Tasks:

- [x] Create `CreditNote` aggregate
  - File: `internal/invoicing/domain/credit_note.go`
  - Status: ✅ Completed
  - Effort: 1 day
  - Requires:
    - Link to invoice
    - Credit amount
    - Reason
    - Status management

- [x] Create `Installment` aggregate
  - File: `internal/invoicing/domain/installment.go`
  - Status: ✅ Completed
  - Effort: 0.5 day
  - Requires:
    - Payment schedule
    - Due dates
    - Amount tracking

- [x] Add tax calculation to `Invoice`
  - File: `internal/invoicing/domain/invoice.go` (modify)
  - Status: ✅ Completed
  - Effort: 0.5 day
  - Requires:
    - Tax rate
    - Tax amount calculation
    - Net/Gross amounts

---

## 🟢 PRIORITY 3 - MEDIUM (Nice to Have)

### 7. Contracts Context (Score: 80/100 → Target: 85/100)

**Status:** ⬜ Not Started  
**Effort:** 1 day  
**Priority:** MEDIUM

#### Tasks:

- [ ] Add renewal logic to `Contract`
  - File: `internal/contracts/domain/contract.go` (modify)
  - Status: ⬜ Not Started
  - Effort: 4 hours
  - Requires:
    - Renew method
    - Auto-renewal flag
    - Renewal count

- [ ] Add amendment tracking
  - File: `internal/contracts/domain/contract.go` (modify)
  - Status: ⬜ Not Started
  - Effort: 4 hours
  - Requires:
    - Amendment history
    - Version tracking

---

### 8. Notifications Context (Score: 76/100 → Target: 80/100)

**Status:** ⬜ Not Started  
**Effort:** 4 hours  
**Priority:** MEDIUM

#### Tasks:

- [ ] Add quiet hours to preferences
  - File: `internal/notifications/domain/notification.go` (modify)
  - Status: ⬜ Not Started
  - Effort: 2 hours
  - Requires:
    - QuietHours value object
    - Timezone support
    - Delivery hold logic

- [ ] Add template validation
  - File: `internal/notifications/domain/template.go` (modify)
  - Status: ⬜ Not Started
  - Effort: 2 hours
  - Requires:
    - Variable validation
    - Required fields check

---

### 9. Files Context (Score: 74/100 → Target: 78/100)

**Status:** ⬜ Not Started  
**Effort:** 4 hours  
**Priority:** MEDIUM

#### Tasks:

- [ ] Add virus scanning status
  - File: `internal/files/domain/file.go` (modify)
  - Status: ⬜ Not Started
  - Effort: 2 hours
  - Requires:
    - ScanStatus enum
    - ScannedAt timestamp
    - ThreatInfo field

- [ ] Add file type validation
  - File: `internal/files/domain/file.go` (modify)
  - Status: ⬜ Not Started
  - Effort: 2 hours
  - Requires:
    - AllowedTypes list
    - Validation in domain

---

## 🔵 PRIORITY 4 - LOW (Can Wait)

### 10. Identity Context (Score: 84/100 → Target: 88/100)

**Status:** ⬜ Not Started  
**Effort:** 1 day  
**Priority:** LOW

#### Tasks:

- [ ] Create `Session` aggregate
  - File: `internal/identity/domain/session.go`
  - Status: ⬜ Not Started
  - Effort: 6 hours
  - Requires:
    - Session management
    - Expiration logic
    - Device tracking

- [ ] Add `UserPreferences`
  - File: `internal/identity/domain/preferences.go`
  - Status: ⬜ Not Started
  - Effort: 2 hours
  - Requires:
    - Theme preferences
    - Language
    - Notification settings

---

## 📋 Implementation Checklist

### Before Starting Each Task:

- [ ] Read existing domain code
- [ ] Identify business rules from requirements
- [ ] Design value objects first
- [ ] Plan domain events

### During Implementation:

- [ ] All business rules in domain layer
- [ ] Value objects for complex types
- [ ] Domain events published
- [ ] Error types defined
- [ ] Tests written

### After Completion:

- [ ] Unit tests passing
- [ ] Integration tests added
- [ ] Wire.go updated
- [ ]Routes.go updated
- [ ] Documentation updated

---

## 📊 Progress Tracking

### Week 1 Plan (Critical)

| Day | Task | Context | Status |
|-----|------|---------|--------|
| Mon | Accounting -JournalEntry | Accounting | ⬜ |
| Tue | Accounting - AccountingPeriod | Accounting | ⬜ |
| Wed | Accounting - Reconciliation | Accounting | ⬜ |
| Thu | Documents - ApprovalWorkflow | Documents | ⬜ |
| Fri | Documents - Versioning | Documents | ⬜ |

### Week 2 Plan (High Priority)

| Day | Task | Context | Status |
|-----|------|---------|--------|
| Mon | Parties - CreditLimit | Parties | ⬜ |
| Tue | Catalog - ProductVariant | Catalog | ⬜ |
| Wed | Catalog - PricingRule | Catalog | ⬜ |
| Thu | Inventory - Lot tracking | Inventory | ⬜ |
| Fri | Invoicing - CreditNote | Invoicing | ⬜ |

### Week 3 Plan (Medium/Low)

| Day | Task | Context | Status |
|-----|------|---------|--------|
| Mon | Contracts - Renewal | Contracts | ⬜ |
| Tue | Notifications - QuietHours | Notifications | ⬜ |
| Wed | Files - VirusScan | Files | ⬜ |
| Thu | Identity - Session | Identity | ⬜ |
| Fri | Buffer / Catch-up | - | ⬜ |

---

## 🎯 Success Criteria

Each context must achieve:

1. **Domain Model Completeness:** ≥ 85/100
2. **State Machines:** All states with valid transitions
3. **Business Rules:** Enforced in domain layer
4. **Test Coverage:** ≥ 80% for domain
5. **Domain Events:** Published for all state changes
6. **Documentation:** ADR updated

---

## 📝 Notes

- Each completed task should have a commit
- Update this file after each task completion
- Mark status with: ⬜ Not Started, 🔄 In Progress, ✅ Completed
- Add blockers in respective task comments

---

## 🔗 Related Documents

- [CRM_INTEGRATION_GAPS.md](./docs/CRM_INTEGRATION_GAPS.md) - Integration gaps
- [CROSS_CONTEXT_INTEGRATION.md](./docs/CROSS_CONTEXT_INTEGRATION.md) - Event flows
- [ACTION_PLAN_PHASE1.md](./docs/ACTION_PLAN_PHASE1.md) - Phase 1 plan
- [ADR Files](./docs/adr/) - Architecture Decisions

---

**Last Updated:** 2026-04-09  
**Next Review:** Weekly