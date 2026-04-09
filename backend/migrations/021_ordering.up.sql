-- Orders (Universal order system)
CREATE TYPE order_status AS ENUM (
    'draft',
    'pending',
    'confirmed',
    'processing',
    'completed',
    'cancelled',
    'refunded'
);

CREATE TABLE orders (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    order_number VARCHAR(50) UNIQUE NOT NULL,
    
    -- Parties (references parties.id from parties context)
    customer_id UUID NOT NULL,
    supplier_id UUID NOT NULL,
    contract_id UUID,  -- References contracts.id from contracts context
    
    -- Amounts
    subtotal BIGINT NOT NULL,
    tax_amount BIGINT DEFAULT 0,
    discount BIGINT DEFAULT 0,
    total BIGINT NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'UAH',
    
    -- Status
    status order_status NOT NULL DEFAULT 'draft',
    
    -- Dates
    order_date TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    due_date TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    cancelled_at TIMESTAMPTZ,
    
    -- Notes
    notes TEXT,
    
    -- Metadata
    metadata JSONB DEFAULT '{}',
    
    -- Audit
    created_by UUID,  -- References users.id
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    CONSTRAINT valid_total CHECK (total >= 0),
    CONSTRAINT valid_discount CHECK (discount >= 0)
);

CREATE INDEX idx_orders_customer ON orders(customer_id);
CREATE INDEX idx_orders_supplier ON orders(supplier_id);
CREATE INDEX idx_orders_status ON orders(status);
CREATE INDEX idx_orders_date ON orders(order_date);
CREATE INDEX idx_orders_number ON orders(order_number);

CREATE TRIGGER orders_updated_at
    BEFORE UPDATE ON orders
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at();

COMMENT ON TABLE orders IS 'Universal orders: purchase orders, sales orders, bookings, appointments';
COMMENT ON COLUMN orders.customer_id IS 'Party placing the order (customer)';
COMMENT ON COLUMN orders.supplier_id IS 'Party fulfilling the order (supplier)';
COMMENT ON COLUMN orders.total IS 'Total order amount = subtotal + tax - discount';

-- Order Lines
CREATE TABLE order_lines (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    order_id UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    
    -- Item
    item_id UUID NOT NULL,  -- References catalog_items.id from catalog context
    item_name VARCHAR(255) NOT NULL,
    
    -- Quantities
    quantity DECIMAL(10,2) NOT NULL,
    unit VARCHAR(20),  -- 'piece', 'kg', 'hour', etc.
    
    -- Pricing
    unit_price BIGINT NOT NULL,
    discount BIGINT DEFAULT 0,
    total BIGINT NOT NULL,
    
    -- Metadata
    metadata JSONB DEFAULT '{}',
    
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    CONSTRAINT valid_quantity CHECK (quantity > 0),
    CONSTRAINT valid_unit_price CHECK (unit_price >= 0),
    CONSTRAINT valid_line_total CHECK (total >= 0)
);

CREATE INDEX idx_order_lines_order ON order_lines(order_id);
CREATE INDEX idx_order_lines_item ON order_lines(item_id);

COMMENT ON TABLE order_lines IS 'Order line items';
COMMENT ON COLUMN order_lines.item_id IS 'Reference to catalog item';

-- Quotes (Price quotes)
CREATE TYPE quote_status AS ENUM (
    'draft',
    'sent',
    'accepted',
    'rejected',
    'expired'
);

CREATE TABLE quotes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    quote_number VARCHAR(50) UNIQUE NOT NULL,
    
    -- Parties
    customer_id UUID NOT NULL,  -- References parties.id
    supplier_id UUID NOT NULL,  -- References parties.id
    
    -- Amounts
    subtotal BIGINT NOT NULL,
    tax_amount BIGINT DEFAULT 0,
    total BIGINT NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'UAH',
    
    -- Validity
    valid_from DATE NOT NULL,
    valid_until DATE NOT NULL,
    
    -- Status
    status quote_status NOT NULL DEFAULT 'draft',
    
    -- Converted to order
    order_id UUID REFERENCES orders(id),
    
    -- Notes
    notes TEXT,
    
    -- Metadata
    metadata JSONB DEFAULT '{}',
    
    -- Audit
    created_by UUID,  -- References users.id
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_quotes_customer ON quotes(customer_id);
CREATE INDEX idx_quotes_supplier ON quotes(supplier_id);
CREATE INDEX idx_quotes_status ON quotes(status);

COMMENT ON TABLE quotes IS 'Price quotes (commercial proposals)';

-- TRIGGER
CREATE TRIGGER quotes_updated_at
    BEFORE UPDATE ON quotes
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at();