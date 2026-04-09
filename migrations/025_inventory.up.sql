-- Migration: 025_inventory.up.sql
-- Description: Creates inventory tables for warehouse and stock management
-- Created: 2025-01-08

-- Warehouse status enum
CREATE TYPE warehouse_status AS ENUM (
    'active',       -- Warehouse is operational
    'inactive',     -- Warehouse is deactivated
    'maintenance'   -- Warehouse is under maintenance
);

-- Movement type enum
CREATE TYPE movement_type AS ENUM (
    'receipt',      -- Stock received into warehouse
    'issue',        -- Stock issued from warehouse
    'transfer',     -- Stock transferred between warehouses
    'adjustment',   -- Stock adjustment (correction)
    'return'        -- Stock returned to warehouse
);

-- Reservation status enum
CREATE TYPE reservation_status AS ENUM (
    'active',       -- Reservation is active
    'fulfilled',    -- Reservation has been fulfilled
    'cancelled',    -- Reservation was cancelled
    'expired'       -- Reservation has expired
);

-- Warehouses table
CREATE TABLE warehouses (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    name VARCHAR(255) NOT NULL,
    code VARCHAR(50) UNIQUE,
    location VARCHAR(500),
    capacity DECIMAL(15, 2) DEFAULT 0 CHECK (capacity >= 0),
    status warehouse_status NOT NULL DEFAULT 'active',
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Stock table
CREATE TABLE stock (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    item_id UUID NOT NULL REFERENCES items(id) ON DELETE RESTRICT,
    warehouse_id UUID NOT NULL REFERENCES warehouses(id) ON DELETE RESTRICT,
    quantity DECIMAL(15, 3) NOT NULL DEFAULT 0 CHECK (quantity >= 0),
    reserved_qty DECIMAL(15, 3) NOT NULL DEFAULT 0 CHECK (reserved_qty >= 0),
    available_qty DECIMAL(15, 3) NOT NULL DEFAULT 0 CHECK (available_qty >= 0),
    reorder_point DECIMAL(15, 3) NOT NULL DEFAULT 0 CHECK (reorder_point >= 0),
    reorder_quantity DECIMAL(15, 3) NOT NULL DEFAULT 0 CHECK (reorder_quantity >= 0),
    last_movement_id UUID REFERENCES stock_movements(id) ON DELETE SET NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_available_qty CHECK (available_qty = quantity - reserved_qty),
    CONSTRAINT uk_item_warehouse UNIQUE (item_id, warehouse_id)
);

-- Stock movements table
CREATE TABLE stock_movements (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    movement_type movement_type NOT NULL,
    item_id UUID NOT NULL REFERENCES items(id) ON DELETE RESTRICT,
    from_warehouse UUID REFERENCES warehouses(id) ON DELETE RESTRICT,
    to_warehouse UUID REFERENCES warehouses(id) ON DELETE RESTRICT,
    quantity DECIMAL(15, 3) NOT NULL CHECK (quantity > 0),
    reference_id VARCHAR(255),
    reference_type VARCHAR(50),
    notes TEXT,
    occurred_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_movement_warehouse CHECK (
        (movement_type = 'transfer' AND from_warehouse IS NOT NULL AND to_warehouse IS NOT NULL) OR
        (movement_type = 'receipt' AND to_warehouse IS NOT NULL) OR
        (movement_type = 'issue' AND from_warehouse IS NOT NULL) OR
        (movement_type = 'adjustment' AND from_warehouse IS NOT NULL) OR
        (movement_type = 'return' AND to_warehouse IS NOT NULL)
    )
);

-- Stock reservations table
CREATE TABLE stock_reservations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    stock_id UUID NOT NULL REFERENCES stock(id) ON DELETE RESTRICT,
    order_id UUID NOT NULL REFERENCES orders(id) ON DELETE RESTRICT,
    quantity DECIMAL(15, 3) NOT NULL CHECK (quantity > 0),
    status reservation_status NOT NULL DEFAULT 'active',
    reserved_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMP WITH TIME ZONE,
    fulfilled_at TIMESTAMP WITH TIME ZONE,
    cancelled_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Indexes for efficient querying
CREATE INDEX idx_warehouses_status ON warehouses(status);
CREATE INDEX idx_warehouses_code ON warehouses(code);
CREATE INDEX idx_warehouses_created_at ON warehouses(created_at);

CREATE INDEX idx_stock_item ON stock(item_id);
CREATE INDEX idx_stock_warehouse ON stock(warehouse_id);
CREATE INDEX idx_stock_available ON stock(available_qty);
CREATE INDEX idx_stock_created_at ON stock(created_at);

CREATE INDEX idx_stock_movements_item ON stock_movements(item_id);
CREATE INDEX idx_stock_movements_from_warehouse ON stock_movements(from_warehouse);
CREATE INDEX idx_stock_movements_to_warehouse ON stock_movements(to_warehouse);
CREATE INDEX idx_stock_movements_type ON stock_movements(movement_type);
CREATE INDEX idx_stock_movements_occurred_at ON stock_movements(occurred_at);
CREATE INDEX idx_stock_movements_created_at ON stock_movements(created_at);

CREATE INDEX idx_stock_reservations_stock ON stock_reservations(stock_id);
CREATE INDEX idx_stock_reservations_order ON stock_reservations(order_id);
CREATE INDEX idx_stock_reservations_status ON stock_reservations(status);
CREATE INDEX idx_stock_reservations_reserved_at ON stock_reservations(reserved_at);
CREATE INDEX idx_stock_reservations_created_at ON stock_reservations(created_at);

-- Comments for documentation
COMMENT ON TABLE warehouses IS 'Warehouse management for storage locations';
COMMENT ON TABLE stock IS 'Stock levels for items in warehouses';
COMMENT ON TABLE stock_movements IS 'Stock movement history (receipts, issues, transfers, adjustments)';
COMMENT ON TABLE stock_reservations IS 'Stock reservations for orders';

COMMENT ON COLUMN warehouses.code IS 'Unique warehouse code for quick reference';
COMMENT ON COLUMN warehouses.capacity IS 'Maximum storage capacity';
COMMENT ON COLUMN warehouses.metadata IS 'Additional metadata (contact info, address, etc.)';

COMMENT ON COLUMN stock.quantity IS 'Total quantity in stock';
COMMENT ON COLUMN stock.reserved_qty IS 'Quantity reserved for orders';
COMMENT ON COLUMN stock.available_qty IS 'Available quantity for new orders (quantity - reserved_qty)';
COMMENT ON COLUMN stock.reorder_point IS 'Reorder threshold level';
COMMENT ON COLUMN stock.reorder_quantity IS 'Quantity to reorder when stock reaches reorder_point';

COMMENT ON COLUMN stock_movements.from_warehouse IS 'Source warehouse (for issue/transfer/adjustment)';
COMMENT ON COLUMN stock_movements.to_warehouse IS 'Destination warehouse (for receipt/transfer/return)';
COMMENT ON COLUMN stock_movements.reference_id IS 'Reference ID (order ID, receipt ID, etc.)';
COMMENT ON COLUMN stock_movements.reference_type IS 'Reference type (order, receipt, transfer, adjustment)';

COMMENT ON COLUMN stock_reservations.order_id IS 'Order this reservation is for';
COMMENT ON COLUMN stock_reservations.expires_at IS 'Expiration time for temporary reservations';
COMMENT ON COLUMN stock_reservations.fulfilled_at IS 'When reservation was fulfilled';
COMMENT ON COLUMN stock_reservations.cancelled_at IS 'When reservation was cancelled';