# ADR-022: Inventory Bounded Context

## Status

Accepted

## Context

The skeleton Business Engine requires inventory management capabilities to support warehouse operations, stock tracking, and order fulfillment. This context needs to handle:

1. **Warehouse Management** - Multiple warehouses with status tracking (active/inactive/maintenance)
2. **Stock Levels** - Real-time inventory quantities with availability calculations
3. **Stock Movements** - Receipts, issues, transfers, and adjustments with full history
4. **Stock Reservations** - Order-based reservations with expiration tracking

The inventory context must integrate with:
- **Ordering Context** - For order fulfillment and stock reservations
- **Catalog Context** - For item/product references
- **Accounting Context** - For inventory valuation (future)

## Decision

We implement the Inventory bounded context following Domain-Driven Design with Hexagonal Architecture:

### Domain Layer

**Aggregates:**
- `Warehouse` - Warehouse management with status transitions
- `Stock` - Inventory levels with quantity tracking
- `StockMovement` - Movement history (receipt/issue/transfer/adjustment/return)
- `StockReservation` - Order-based reservations

**Value Objects:**
- `WarehouseID`, `StockID`, `StockMovementID`, `StockReservationID` - Type-safe identifiers
- `WarehouseStatus` - enum (active/inactive/maintenance)
- `MovementType` - enum (receipt/issue/transfer/adjustment/return)
- `ReservationStatus` - enum (active/fulfilled/cancelled/expired)

**Domain Events:**
- `WarehouseCreated` - Published when new warehouse is created
- `StockAdjusted` - Published when stock quantity changes
- `StockReserved` - Published when stock is reserved for order
- `StockMoved` - Published when stock is transferred between warehouses

### Application Layer

**Commands (10):**
- CreateWarehouse, UpdateWarehouse
- CreateStock, AdjustStock
- ReceiptStock, IssueStock, TransferStock
- ReserveStock, FulfillReservation, CancelReservation

**Queries (8):**
- GetWarehouse, ListWarehouses
- GetStock, ListStock
- GetStockMovement, ListStockMovements
- GetReservation, ListReservations

### Infrastructure Layer

**Repositories:**
- `WarehouseRepository` - Warehouse persistence
- `StockRepository` - Stock levels persistence
- `StockMovementRepository` - Movement history persistence
- `StockReservationRepository` - Reservations persistence

**Implementation:**
- PostgreSQL with `scany v2` for scanning
- `squirrel` for dynamic queries
- UUID v7 for time-sortable IDs

### Database Design

**Tables:**
```sql
-- Warehouses with status management
CREATE TABLE warehouses (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    name VARCHAR(255) NOT NULL,
    code VARCHAR(50) UNIQUE,
    location VARCHAR(500),
    capacity DECIMAL(15, 2) DEFAULT 0,
    status warehouse_status NOT NULL DEFAULT 'active',
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Stock levels with availability tracking
CREATE TABLE stock (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    item_id UUID NOT NULL REFERENCES items(id),
    warehouse_id UUID NOT NULL REFERENCES warehouses(id),
    quantity DECIMAL(15, 3) NOT NULL DEFAULT 0,
    reserved_qty DECIMAL(15, 3) NOT NULL DEFAULT 0,
    available_qty DECIMAL(15, 3) NOT NULL DEFAULT 0,
    reorder_point DECIMAL(15, 3) NOT NULL DEFAULT 0,
    reorder_quantity DECIMAL(15, 3) NOT NULL DEFAULT 0,
    last_movement_id UUID REFERENCES stock_movements(id),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_available_qty CHECK (available_qty = quantity - reserved_qty),
    CONSTRAINT uk_item_warehouse UNIQUE (item_id, warehouse_id)
);

-- Movement history with type discrimination
CREATE TABLE stock_movements (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    movement_type movement_type NOT NULL,
    item_id UUID NOT NULL REFERENCES items(id),
    from_warehouse UUID REFERENCES warehouses(id),
    to_warehouse UUID REFERENCES warehouses(id),
    quantity DECIMAL(15, 3) NOT NULL CHECK (quantity > 0),
    reference_id VARCHAR(255),
    reference_type VARCHAR(50),
    notes TEXT,
    occurred_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Reservations with expiration tracking
CREATE TABLE stock_reservations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    stock_id UUID NOT NULL REFERENCES stock(id),
    order_id UUID NOT NULL REFERENCES orders(id),
    quantity DECIMAL(15, 3) NOT NULL CHECK (quantity > 0),
    status reservation_status NOT NULL DEFAULT 'active',
    reserved_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMP WITH TIME ZONE,
    fulfilled_at TIMESTAMP WITH TIME ZONE,
    cancelled_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);
```

**ENUMs:**
- `warehouse_status` - Warehouse lifecycle states
- `movement_type` - Type of stock movement
- `reservation_status` - Reservation lifecycle states

### HTTP API

**Endpoints (18):**
```
POST   /api/v1/warehouses                    Create warehouse
GET    /api/v1/warehouses/:id                 Get warehouse
GET    /api/v1/warehouses                     List warehouses
PUT    /api/v1/warehouses/:id                 Update warehouse

POST   /api/v1/stock                          Create stock
GET    /api/v1/stock/:id                      Get stock
GET    /api/v1/stock                          List stock
POST   /api/v1/stock/:id/adjust                Adjust stock quantity
POST   /api/v1/stock/receipt                  Receive stock into warehouse
POST   /api/v1/stock/issue                    Issue stock from warehouse
POST   /api/v1/stock/transfer                 Transfer stock between warehouses
POST   /api/v1/stock/reserve                  Reserve stock for order

POST   /api/v1/reservations/fulfill            Fulfill reservation
POST   /api/v1/reservations/cancel             Cancel reservation
GET    /api/v1/reservations/:id                Get reservation
GET    /api/v1/reservations                    List reservations

GET    /api/v1/movements/:id                   Get movement
GET    /api/v1/movements                       List movements
```

**Authorization:**
- All endpoints require authentication
- Read operations require `inventory:read` permission
- Write operations require `inventory:write` permission

### Business Rules

**Stock Availability:**
```go
// available_qty = quantity - reserved_qty
// Calculated automatically, never stored directly
func (s *Stock) IsAvailable(quantity float64) bool {
    return s.availableQty >= quantity
}
```

**Movement Constraints:**
- Receipt: must have `to_warehouse`, no `from_warehouse`
- Issue: must have `from_warehouse`, no `to_warehouse`
- Transfer: must have both `from_warehouse` and `to_warehouse`
- Adjustment: must have `from_warehouse == to_warehouse`

**Reservation Lifecycle:**
- Created in `active` status
- Can be `fulfilled` (stock consumed) or `cancelled` (stock released)
- Automatically `expired` when `expires_at` is reached
- Cannot double-fulfill or double-cancel

### Testing

**Domain Tests (26):**
- Warehouse: creation, status transitions, capacity management
- Stock: quantity adjustments, reservations, availability checks
- StockMovement: all movement types, validation
- StockReservation: lifecycle, expiration

**Test Results:**
```bash
$ go test ./internal/inventory/domain/... -v
PASS
ok  github.com/basilex/skeleton/internal/inventory/domain
```

## Consequences

### Positive

1. **Complete Inventory Management** - Warehouses, stock levels, movements, reservations
2. **Type Safety** - Compile-time checks for all IDs and enums
3. **Event-Driven** - Domain events for cross-context communication
4. **Testability** - 26 domain tests covering all aggregates
5. **Performance** - Optimized queries with proper indexes
6. **Flexibility** - JSONB metadata for warehouse attributes
7. **Traceability** - Complete movement history with references

### Negative

1. **Complexity** - Multiple aggregates with relationships
2. **Integration** - Requires coordination with Ordering and Catalog contexts
3. **Data Consistency** - Need to ensure stock availability calculations are correct
4. **Performance** - Complex queries for stock movements across multiple warehouses

### Risks

1. **Distributed Transactions** - Stock reservation and order creation may need distributed transaction
2. **Race Conditions** - Stock availability checks need proper locking
3. **Migration Complexity** - Need to carefully handle existing data

## Alternatives Considered

1. **Single Stock Table** - Rejected because separate tables for movements and reservations provide better traceability
2. **Event Sourcing** - Rejected because simpler CRUD operations are sufficient for stock management
3. **Separate Service** - Rejected because inventory is tightly coupled with ordering

## Implementation

**Files Created (44):**
- Domain Layer: 13 files (aggregates, value objects, events, interfaces, tests)
- Infrastructure Layer: 5 files (repositories)
- Application Layer: 18 files (10 commands, 8 queries)
- HTTP Layer: 2 files (handler, DTO)
- Migration: 2 files (up/down)
- Documentation: 4 files (ADR, tests, Wire, Routes)

**Migration:** `025_inventory.up/down.sql`

**Integration:**
- `cmd/api/wire.go` - Dependency injection
- `cmd/api/routes.go` - HTTP routes

## References

- [ADR-020: Ordering Bounded Context](ADR-020-ordering.md) - Order management
- [ADR-021: Catalog Bounded Context](ADR-021-catalog.md) - Product catalog
- [ADR-006: UUID v7](ADR-006-uuid-v7.md) - Time-sortable identifiers
- [ADR-007: Cursor Pagination](ADR-007-cursor-pagination.md) - List endpoints