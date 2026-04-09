-- Catalog (Products/Services/Properties)
CREATE TYPE item_status AS ENUM ('active', 'inactive', 'discontinued');

CREATE TABLE catalog_categories (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    parent_id UUID REFERENCES catalog_categories(id),
    path LTREE,  -- Hierarchical path
    is_active BOOLEAN NOT NULL DEFAULT true,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_categories_parent ON catalog_categories(parent_id);
CREATE INDEX idx_categories_path ON catalog_categories USING GIST(path);
CREATE INDEX idx_categories_active ON catalog_categories(is_active);

COMMENT ON TABLE catalog_categories IS 'Taxonomy: product/service categories';
COMMENT ON COLUMN catalog_categories.path IS 'Hierarchical path using LTREE';

CREATE TABLE catalog_items (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    category_id UUID REFERENCES catalog_categories(id),
    
    -- Basic info
    sku VARCHAR(100) UNIQUE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    
    -- Pricing
    base_price DECIMAL(15,2) NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'UAH',
    
    -- Status
    status item_status NOT NULL DEFAULT 'active',
    
    -- Attributes (flexible JSONB)
    attributes JSONB DEFAULT '{}',
    
    -- Metadata
    metadata JSONB DEFAULT '{}',
    
    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_items_category ON catalog_items(category_id);
CREATE INDEX idx_items_sku ON catalog_items(sku);
CREATE INDEX idx_items_status ON catalog_items(status);
CREATE INDEX idx_items_name ON catalog_items USING GIN(to_tsvector('english', name));
CREATE INDEX idx_items_attrs ON catalog_items USING GIN(attributes);

COMMENT ON TABLE catalog_items IS 'Catalog items: products, services, properties';
COMMENT ON COLUMN catalog_items.attributes IS 'Flexible attributes per item type';

-- Prices (support for multiple prices per item)
CREATE TYPE price_type AS ENUM ('base', 'sale', 'wholesale', 'partner');

CREATE TABLE catalog_prices (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    item_id UUID NOT NULL REFERENCES catalog_items(id) ON DELETE CASCADE,
    price_type price_type NOT NULL DEFAULT 'base',
    amount DECIMAL(15,2) NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'UAH',
    valid_from DATE NOT NULL DEFAULT CURRENT_DATE,
    valid_until DATE,
    is_active BOOLEAN NOT NULL DEFAULT true,
    
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_prices_item ON catalog_prices(item_id);
CREATE INDEX idx_prices_type ON catalog_prices(price_type);
CREATE INDEX idx_prices_valid ON catalog_prices(valid_from, valid_until);

COMMENT ON TABLE catalog_prices IS 'Multiple prices per item: base, sale, wholesale';

-- TRIGGERS
CREATE TRIGGER catalog_items_updated_at
    BEFORE UPDATE ON catalog_items
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER catalog_categories_updated_at
    BEFORE UPDATE ON catalog_categories
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at();