-- Migration: 023_invoicing.up.sql
-- Description: Creates invoicing tables for invoice management
-- Created: 2025-01-08

-- Invoice status enum
CREATE TYPE invoice_status AS ENUM (
    'draft',      -- Initial status
    'sent',       -- Invoice sent to customer
    'viewed',     -- Customer viewed invoice
    'paid',       -- Fully paid
    'overdue',    -- Past due date
    'cancelled'   -- Invoice cancelled
);

-- Payment method enum
CREATE TYPE payment_method AS ENUM (
    'bank_transfer',
    'card',
    'cash',
    'check',
    'crypto'
);

-- Invoices table
CREATE TABLE invoices (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    invoice_number VARCHAR(100) NOT NULL UNIQUE,
    order_id UUID REFERENCES orders(id) ON DELETE SET NULL,
    contract_id UUID REFERENCES contracts(id) ON DELETE SET NULL,
    customer_id UUID NOT NULL REFERENCES parties(id) ON DELETE RESTRICT,
    supplier_id UUID REFERENCES parties(id) ON DELETE SET NULL,
    issue_date DATE NOT NULL DEFAULT CURRENT_DATE,
    due_date DATE NOT NULL,
    status invoice_status NOT NULL DEFAULT 'draft',
    subtotal DECIMAL(15, 2) NOT NULL DEFAULT 0,
    tax_amount DECIMAL(15, 2) NOT NULL DEFAULT 0,
    discount DECIMAL(15, 2) NOT NULL DEFAULT 0,
    total DECIMAL(15, 2) NOT NULL DEFAULT 0,
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    notes TEXT,
    paid_amount DECIMAL(15, 2) NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Invoice lines table
CREATE TABLE invoice_lines (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    invoice_id UUID NOT NULL REFERENCES invoices(id) ON DELETE CASCADE,
    description TEXT NOT NULL,
    quantity DECIMAL(15, 3) NOT NULL CHECK (quantity > 0),
    unit_price DECIMAL(15, 2) NOT NULL CHECK (unit_price >= 0),
    unit VARCHAR(50) NOT NULL,
    discount DECIMAL(15, 2) NOT NULL DEFAULT 0 CHECK (discount >= 0),
    total DECIMAL(15, 2) NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Payments table
CREATE TABLE payments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    invoice_id UUID NOT NULL REFERENCES invoices(id) ON DELETE CASCADE,
    amount DECIMAL(15, 2) NOT NULL CHECK (amount > 0),
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    method payment_method NOT NULL,
    reference VARCHAR(255),
    paid_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Indexes for efficient querying
CREATE INDEX idx_invoices_customer ON invoices(customer_id);
CREATE INDEX idx_invoices_supplier ON invoices(supplier_id);
CREATE INDEX idx_invoices_order ON invoices(order_id);
CREATE INDEX idx_invoices_contract ON invoices(contract_id);
CREATE INDEX idx_invoices_status ON invoices(status);
CREATE INDEX idx_invoices_due_date ON invoices(due_date);
CREATE INDEX idx_invoices_issue_date ON invoices(issue_date);
CREATE INDEX idx_invoices_created_at ON invoices(created_at);

CREATE INDEX idx_invoice_lines_invoice ON invoice_lines(invoice_id);

CREATE INDEX idx_payments_invoice ON payments(invoice_id);
CREATE INDEX idx_payments_paid_at ON payments(paid_at);

-- Comments for documentation
COMMENT ON TABLE invoices IS 'Invoice management for billing customers and suppliers';
COMMENT ON TABLE invoice_lines IS 'Line items for invoices (products/services)';
COMMENT ON TABLE payments IS 'Payment records linked to invoices';

COMMENT ON COLUMN invoices.subtotal IS 'Sum of all line totals before tax and discount';
COMMENT ON COLUMN invoices.paid_amount IS 'Total amount paid towards this invoice';
COMMENT ON COLUMN invoices.total IS 'Final invoice amount: subtotal + tax - discount';
COMMENT ON COLUMN invoice_lines.total IS 'Line total: quantity * unit_price - discount';
COMMENT ON COLUMN payments.reference IS 'Payment reference number (bank ref, transaction ID, etc.)';