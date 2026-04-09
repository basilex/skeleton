-- Migration: 025_inventory.down.sql
-- Description: Drops inventory tables
-- Created: 2025-01-08

-- Drop foreign key constraint from stock table
ALTER TABLE stock DROP CONSTRAINT IF EXISTS stock_item_warehouse_key;
ALTER TABLE stock DROP CONSTRAINT IF EXISTS stock_item_id_fkey;
ALTER TABLE stock DROP CONSTRAINT IF EXISTS stock_warehouse_id_fkey;
ALTER TABLE stock DROP CONSTRAINT IF EXISTS stock_last_movement_id_fkey;

-- Drop indexes
DROP INDEX IF EXISTS idx_stock_reservations_created_at;
DROP INDEX IF EXISTS idx_stock_reservations_reserved_at;
DROP INDEX IF EXISTS idx_stock_reservations_status;
DROP INDEX IF EXISTS idx_stock_reservations_order;
DROP INDEX IF EXISTS idx_stock_reservations_stock;

DROP INDEX IF EXISTS idx_stock_movements_created_at;
DROP INDEX IF EXISTS idx_stock_movements_occurred_at;
DROP INDEX IF EXISTS idx_stock_movements_type;
DROP INDEX IF EXISTS idx_stock_movements_to_warehouse;
DROP INDEX IF EXISTS idx_stock_movements_from_warehouse;
DROP INDEX IF EXISTS idx_stock_movements_item;

DROP INDEX IF EXISTS idx_stock_created_at;
DROP INDEX IF EXISTS idx_stock_available;
DROP INDEX IF EXISTS idx_stock_warehouse;
DROP INDEX IF EXISTS idx_stock_item;

DROP INDEX IF EXISTS idx_warehouses_created_at;
DROP INDEX IF EXISTS idx_warehouses_code;
DROP INDEX IF EXISTS idx_warehouses_status;

-- Drop tables
DROP TABLE IF EXISTS stock_reservations;
DROP TABLE IF EXISTS stock_movements;
DROP TABLE IF EXISTS stock;
DROP TABLE IF EXISTS warehouses;

-- Drop enums
DROP TYPE IF EXISTS reservation_status;
DROP TYPE IF EXISTS movement_type;
DROP TYPE IF EXISTS warehouse_status;