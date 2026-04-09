-- Migration: 023_invoicing.down.sql
-- Description: Drops invoicing tables
-- Created: 2025-01-08

-- Drop in reverse order of creation
DROP TABLE IF EXISTS payments CASCADE;
DROP TABLE IF EXISTS invoice_lines CASCADE;
DROP TABLE IF EXISTS invoices CASCADE;

-- Drop enums
DROP TYPE IF EXISTS payment_method CASCADE;
DROP TYPE IF EXISTS invoice_status CASCADE;