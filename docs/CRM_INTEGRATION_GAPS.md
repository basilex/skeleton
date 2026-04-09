# CRM Integration Gaps - Action Plan

**Audit Date:** 2026-04-08
**Status:** CRITICAL - CRM workflows are blocked

---

## 📊 Integration Status Matrix

| From Context | To Context | Event | Status | Priority |
|--------------|------------|-------|--------|----------|
| Catalog | Inventory | ItemCreated → Stock | ❌ Missing | HIGH |
| Ordering | Inventory | OrderConfirmed → ReserveStock | ❌ Missing | **CRITICAL** |
| Ordering | Invoicing | OrderConfirmed → CreateInvoice | ❌ Missing | **CRITICAL** |
| Ordering | Documents | OrderConfirmed → CreateDocument | ❌ Missing | HIGH |
| Ordering | Parties | UpdateCustomerPurchases | ❌ Missing | HIGH |
| Invoicing | Accounting | InvoiceCreated → CreateEntry | ❌ Missing | **CRITICAL** |
| Invoicing | Documents | InvoiceCreated → CreatePDF | ❌ Missing | HIGH |
| Parties | Identity | CustomerCreated → CreateUser | ❌ Missing | MEDIUM |
| Inventory | Parties | LowStock → NotifySuppliers | ❌ Missing | MEDIUM |
| Documents | (Multiple) | Auto-create documents | ❌ Missing | HIGH |

---

## 🔴 CRITICAL: Must Fix Before CRM Launch

### 1. Catalog Events (2-3 hours)

**Problem:** Catalog has NO domain events at all.

**Impact:**
- New items don't create stock records
- Price changes don't propagate
- Category changes not tracked

**Fix:**
```bash
# Create events file
touch internal/catalog/domain/events.go
```

**Required Events:**
- `ItemCreated` - When new product added
- `ItemPriceChanged` - Price updates
- `ItemStatusChanged` - Status updates (active/inactive/discontinued)
- `CategoryCreated` - New category

**Implementation:**
1. Create `catalog/domain/events.go` with all events
2. Update `Item` aggregate to call `PullEvents()`
3. Add events in Item methods: `NewItem()`, `UpdatePrice()`, `Activate()`, `Deactivate()`

---

### 2. Order → Inventory Integration (4-5 hours)

**Problem:** No stock reservation when order is confirmed.

**Impact:**
- Overselling (more orders than stock)
- Manual inventory management
- Customer disappointment

**Required Event Handlers:**

```go
// File: internal/inventory/infrastructure/eventhandler/order_handler.go

func (h *OrderHandler) HandleOrderConfirmed(ctx context.Context, event ordering.OrderConfirmed) error {
    // Reserve stock for each order line
    for _, line := range event.Lines {
        _, err := h.reserveStock.Handle(ctx, inventory.ReserveStockCommand{
            StockID: line.ItemID, // Need to find stock by item_id
            OrderID: event.OrderID,
            Quantity: line.Quantity,
        })
        if err != nil {
            // Compensating transaction: release all reservations
            return err
        }
    }
    return nil
}
```

**Affected Events:**
- `OrderConfirmed` → Reserve stock
- `OrderCancelled` → Release reservation
- `OrderCompleted` → Fulfill reservation (deduct from stock)

---

### 3. Order → Invoice Integration (2-3 hours)

**Problem:** Invoices created manually, no automation.

**Impact:**
- Delay in invoice generation
- Human errors
- Missed invoices

**Required Event Handlers:**

```go
// File: internal/invoicing/infrastructure/eventhandler/order_handler.go

func (h *OrderHandler) HandleOrderConfirmed(ctx context.Context, event ordering.OrderConfirmed) error {
    // Auto-create invoice from order
    invoice, err := h.createInvoice.Handle(ctx, invoicing.CreateInvoiceCommand{
        OrderID: event.OrderID,
        CustomerID: event.CustomerID,
        Lines: convertLines(event.Lines),
        DueDate: time.Now().Add(30 * 24 * time.Hour),
    })
    return err
}
```

---

### 4. Invoice → Accounting Integration (3-4 hours)

**Problem:** No automatic journal entries.

**Impact:**
- Manual bookkeeping
- Delayed financial statements
- Compliance issues

**Required Event Handlers:**

```go
// File: internal/accounting/infrastructure/eventhandler/invoice_handler.go

func (h *InvoiceHandler) HandleInvoiceCreated(ctx context.Context, event invoicing.InvoiceCreated) error {
    // Create journal entry for invoice
    entry := accounting.JournalEntry{
        Description: fmt.Sprintf("Invoice %s", event.InvoiceNumber),
        Lines: []accounting.EntryLine{
            {AccountCode: "1300", Debit: event.Total}, // Accounts Receivable
            {AccountCode: "4000", Credit: event.Subtotal}, // Sales Revenue
            {AccountCode: "2500", Credit: event.TaxAmount}, // Sales Tax Payable
        },
    }
    // Save entry...
    return nil
}
```

---

## ⚠️ HIGH PRIORITY: Should Fix Before Launch

### 5. Typed IDs Throughout (6-8 hours)

**Problem:** All IDs are strings, no type safety.

**Impact:**
- Runtime errors (passing wrong ID type)
- Data integrity issues
- No compile-time checking

**Example:**
```go
// Current (❌ Bad)
type Order struct {
    CustomerID string // Could be any string!
    Lines      []OrderLine
}

// Should be (✅ Good)
type Order struct {
    CustomerID parties.PartyID // Type-safe!
    Lines      []OrderLine
}
```

**Affected Contexts:**
- Parties: `PartyID` instead of `string`
- Catalog: `ItemID` instead of `string`
- Ordering: `OrderID` instead of `string`
- Invoicing: All `string` ID fields
- Inventory: Already uses typed IDs ✅

**Implementation Strategy:**
1. Create typed ID types in each context's `domain/ids.go`
2. Update all aggregates to use typed IDs
3. Update repositories to convert to/from string for DB
4. Update HTTP DTOs to use string (for JSON)

---

### 6. Customer Purchase Tracking (2-3 hours)

**Problem:** `Customer.totalPurchases` updated manually.

**Impact:**
- Customer statistics delayed
- Loyalty levels not automatic
- Segmentation broken

**Required Event Handlers:**

```go
// File: internal/parties/infrastructure/eventhandler/order_handler.go

func (h *OrderHandler) HandleOrderCompleted(ctx context.Context, event ordering.OrderCompleted) error {
    customer, err := h.customers.FindByID(ctx, event.CustomerID)
    if err != nil {
        return err
    }
    
    customer.AddPurchase(event.Total)
    return h.customers.Save(ctx, customer)
}
```

---

## 📝 MEDIUM PRIORITY: Quality of Life

### 7. Document Automation (3-4 hours)

Auto-generate PDF documents when invoices/orders created.

### 8. Customer Portal Integration (2-3 hours)

Link Identity users with Customer entities.

### 9. Low Stock Notifications (2-3 hours)

Notify suppliers when stock drops below reorder point.

---

## 🗓️ Implementation Timeline

### Phase 1: FOUNDATION (Week 1) - 12-14 hours
**Goal:** Enable cross-context communication

1. ✅ Catalog events (2-3h)
2. ✅ Order → Inventory handlers (4-5h)
3. ✅ Order → Invoice handlers (2-3h)
4. ✅ Invoice → Accounting handlers (3-4h)

**Deliverable:** Basic CRM workflows work end-to-end

---

### Phase 2: TYPE SAFETY (Week 1-2) - 6-8 hours
**Goal:** Compile-time ID checking

5. ✅ Typed IDs in all contexts (6-8h)

**Deliverable:** No more ID mix-ups, compile-time safety

---

### Phase 3: AUTOMATION (Week 2) - 5-7 hours
**Goal:** Remove manual work

6. ✅ Customer purchase tracking (2-3h)
7. ✅ Document automation (3-4h)

**Deliverable:** Hands-free operation

---

### Phase 4: ENHANCEMENT (Week 3) - 4-6 hours
**Goal:** Quality of life improvements

8. ✅ Customer portal integration (2-3h)
9. ✅ Low stock notifications (2-3h)

**Deliverable:** Production-ready CRM

---

## 📊 Estimated Total Effort

| Phase | Hours | Priority |
|-------|-------|----------|
| Phase 1: Foundation | 12-14h | **CRITICAL** |
| Phase 2: Type Safety | 6-8h | HIGH |
| Phase 3: Automation | 5-7h | HIGH |
| Phase 4: Enhancement | 4-6h | MEDIUM |
| **Total** | **27-35h** | |

**Team Size:** 1 developer
**Duration:** 1-3 weeks depending on availability

---

## ✅ What's Already Working

Good news! These are already implemented correctly:

1. **Event Bus Infrastructure** - EventBus is set up and working
2. **Event Publishing Pattern** - Identity, Parties, Ordering, Invoicing, Documents, Inventory all publish events
3. **Event Subscription** - Wire.go has proper event handler registration
4. **Repository Pattern** - All contexts have clean repository interfaces
5. **UUID v7** - All IDs use time-sortable UUIDs
6. **Domain Model** - Strong aggregates, value objects, domain rules
7. **CQRS-lite** - Commands and queries properly separated
8. **Inventory Context** - Freshly implemented with full event support ✅

---

## 🎯 Success Criteria

After completing these integrations, the system will support:

✅ **Customer Registration → Order Flow:**
```
Customer registers (Identity)
  → CustomerCreated (Parties)
  → Customer can browse catalog (Catalog)
  → Customer places order (Ordering)
  → OrderConfirmed reserves stock (Inventory)
  → Invoice auto-created (Invoicing)
  → Payment processed (Invoicing)
  → Journal entries created (Accounting)
  → Order shipped
  → Stock deducted (Inventory)
```

✅ **Complete Order Cycle:**
1. Order confirmed → Stock reserved
2. Invoice created → PDF generated
3. Payment received → Journal entries created
4. Order shipped → Stock deducted
5. Customer stats updated → Loyalty calculated

✅ **Financial Integrity:**
- Every transaction creates audit trail
- Double-entry bookkeeping maintained
- Real-time financial statements possible

---

## 🚦 Risk Assessment

**If we DON'T implement these integrations:**

| Risk | Impact | Probability |
|------|--------|-------------|
| Overselling | Customer disappointment | **HIGH** |
| Manual Data Entry | Slower operations | **HIGH** |
| Data Inconsistency | Financial errors | **MEDIUM** |
| Compliance Issues | Audit failures | **MEDIUM** |
| Customer Churn | Loss of business | **LOW-MEDIUM** |

**Recommendation:** Implement Phase 1 (Foundation) immediately before starting frontend work.

---

## 📝 Next Steps

1. **Review this audit report with team**
2. **Prioritize Phase 1 tasks** (all CRITICAL)
3. **Allocate developer(s)** for integration work
4. **Create integration tests** for each workflow
5. **Document event schemas** for frontend team

**Ready to start frontend AFTER Phase 1 is complete.**