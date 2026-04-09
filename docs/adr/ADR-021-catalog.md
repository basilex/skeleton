# ADR-021: Catalog Bounded Context

## Status
Accepted

## Context
Product catalog with hierarchical categories and flexible item attributes.

## Decision
Implement Catalog as a bounded context with LTREE for category hierarchy and JSONB for item attributes.

### Domain Model
- **Item**: Product/Service in catalog
- **Category**: Hierarchical category (LTREE)
- **Attributes**: Flexible key-value attributes (JSONB)
- **ItemStatus**: Active, Inactive, Discontinued

### Architecture
```
internal/catalog/
├── domain/
│   ├── item.go              # Item aggregate
│   ├── category.go           # Category aggregate
│   ├── item_status.go        # Status enum
│   ├── ids.go                # Identifiers
│   ├── errors.go             # Domain errors
│   └── repository.go         # Repository interfaces
├── infrastructure/
│   └── persistence/
│       ├── models.go
│       ├── item_repository.go
│       └── category_repository.go
├── application/
│   ├── command/
│   │   ├── create_item.go
│   │   └── update_item.go
│   └── query/
│       ├── item.go
│       └── list_items.go
└── ports/http/
    ├── handler.go
    └── dto.go
```

### Database Schema

#### Categories (LTREE Hierarchy)
```sql
CREATE TABLE catalog_categories (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    path LTREE NOT NULL,              -- Hierarchy path
    is_active BOOLEAN NOT NULL DEFAULT true,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    CONSTRAINT unique_path UNIQUE (path)
);

CREATE INDEX idx_categories_path ON catalog_categories USING GIST (path);
CREATE INDEX idx_categories_name ON catalog_categories(name);

-- Examples:
-- 'electronics'                           -- Top level
-- 'electronics.computers'                 -- Level 2
-- 'electronics.computers.laptops'          -- Level 3
-- 'electronics.computers.laptops.gaming'  -- Level 4
```

#### Items
```sql
CREATE TYPE item_status AS ENUM (
    'active',
    'inactive',
    'discontinued'
);

CREATE TYPE price_type AS ENUM (
    'fixed',
    'negotiable',
    'on_request'
);

CREATE TABLE catalog_items (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    category_id UUID REFERENCES catalog_categories(id),
    
    -- Basic info
    name VARCHAR(255) NOT NULL,
    description TEXT,
    sku VARCHAR(100) UNIQUE,
    
    -- Pricing
    base_price DECIMAL(15,2) NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'UAH',
    price_type price_type NOT NULL DEFAULT 'fixed',
    
    -- Status
    status item_status NOT NULL DEFAULT 'active',
    
    -- Flexible attributes
    attributes JSONB DEFAULT '{}',
    
    -- Metadata
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_items_category ON catalog_items(category_id);
CREATE INDEX idx_items_status ON catalog_items(status);
CREATE INDEX idx_items_sku ON catalog_items(sku);
CREATE INDEX idx_items_name ON catalog_items USING GIN (to_tsvector('english', name));

-- Example attributes:
{
    "brand": "Apple",
    "model": "MacBook Pro 16\"",
    "screen_size": "16 inch",
    "ram": "32GB",
    "storage": "1TB SSD",
    "processor": "M2 Max",
    "color": "Space Gray",
    "weight": "2.1 kg"
}
```

### Key Design Decisions

#### 1. LTREE for Category Hierarchy
```sql
-- Insert categories
INSERT INTO catalog_categories (name, path) VALUES
    ('Electronics', 'electronics'),
    ('Computers', 'electronics.computers'),
    ('Laptops', 'electronics.computers.laptops'),
    ('Gaming Laptops', 'electronics.computers.laptops.gaming');

-- Query: Get all subcategories under 'Computers'
SELECT * FROM catalog_categories 
WHERE path <@ 'electronics.computers';

-- Query: Get all ancestors of 'Gaming Laptops'
SELECT * FROM catalog_categories 
WHERE 'electronics.computers.laptops.gaming' <@ path;

-- Query: Get immediate children of 'Computers'
SELECT * FROM catalog_categories 
WHERE path ~ 'electronics.computers.*{1}'
  AND nlevel(path) = nlevel('electronics.computers') + 1;
```

**Benefits:**
- ✅ Unlimited hierarchy depth
- ✅ Fast parent/child queries
- ✅ Path-based navigation
- ✅ No recursive CTEs needed
- ✅ Built-in operators

#### 2. JSONB for Flexible Attributes
```go
type Attributes map[string]interface{}

// Flexible schema - different per category
electronicsAttrs := Attributes{
    "brand": "Apple",
    "model": "MacBook Pro",
    "specs": map[string]interface{}{
        "ram": "32GB",
        "storage": "1TB",
    },
}

furnitureAttrs := Attributes{
    "material": "Oak",
    "dimensions": map[string]float64{
        "width": 120.5,
        "height": 75.0,
        "depth": 50.0,
    },
    "color": "Natural",
}

// Query JSONB
SELECT * FROM catalog_items 
WHERE attributes->>'brand' = 'Apple'
  AND attributes->'specs'->>'ram' = '32GB';

-- GIN index for fast JSON queries
CREATE INDEX idx_items_attributes ON catalog_items USING GIN (attributes);
```

#### 3. Item State Machine
```
Active ──Deactivate──► Inactive
   ▲                        │
   │                        │
   └──────Activate──────────┘
   │
   │ Discontinue
   ▼
Discontinued
```

```go
func (i *Item) Activate() {
    i.status = ItemStatusActive
}

func (i *Item) Deactivate() {
    i.status = ItemStatusInactive
}

func (i *Item) Discontinue() {
    i.status = ItemStatusDiscontinued
}

func (i *Item) SetAttribute(key string, value interface{}) {
    if i.attributes == nil {
        i.attributes = make(Attributes)
    }
    i.attributes[key] = value
}

func (i *Item) GetAttribute(key string) (interface{}, bool) {
    if i.attributes == nil {
        return nil, false
    }
    val, ok := i.attributes[key]
    return val, ok
}
```

### API Endpoints

```
POST   /api/v1/catalog/items              # Create item
GET    /api/v1/catalog/items/:id           # Get item
GET    /api/v1/catalog/items                # List items (paginated)
PUT    /api/v1/catalog/items/:id            # Update item

POST   /api/v1/catalog/categories          # Create category
GET    /api/v1/catalog/categories/:id      # Get category
GET    /api/v1/catalog/categories          # List categories (tree)
GET    /api/v1/catalog/categories/:id/items # Items in category
```

### Usage Example

```go
// Create category hierarchy
POST /api/v1/catalog/categories
{
    "name": "Electronics",
    "path": "electronics"
}

POST /api/v1/catalog/categories
{
    "name": "Computers",
    "path": "electronics.computers"
}

// Create item with flexible attributes
POST /api/v1/catalog/items
{
    "category_id": "computers-category-id",
    "name": "MacBook Pro 16\"",
    "sku": "MBP-16-M2MAX",
    "base_price": 249999.00,
    "currency": "UAH",
    "attributes": {
        "brand": "Apple",
        "model": "MacBook Pro 16\"",
        "screen_size": "16 inch",
        "ram": "32GB",
        "storage": "1TB SSD",
        "processor": "M2 Max",
        "color": "Space Gray"
    }
}

// Query items by attributes
POST /api/v1/catalog/items/search
{
    "category_path": "electronics.computers.laptops",
    "attributes": {
        "brand": "Apple",
        "ram": "32GB"
    },
    "price_range": {
        "min": 200000,
        "max": 300000
    }
}

// Get category tree
GET /api/v1/catalog/categories?tree=true

// Response:
{
    "id": "electronics-id",
    "name": "Electronics",
    "path": "electronics",
    "children": [
        {
            "id": "computers-id",
            "name": "Computers",
            "path": "electronics.computers",
            "children": [...]
        }
    ]
}
```

### LTREE Operations

#### Query Examples
```sql
-- Direct path lookup (very fast)
SELECT * FROM catalog_items 
WHERE category_id IN (
    SELECT id FROM catalog_categories 
    WHERE path = 'electronics.computers.laptops'
);

-- All items in category tree (includes descendants)
SELECT * FROM catalog_items 
WHERE category_id IN (
    SELECT id FROM catalog_categories 
    WHERE path <@ 'electronics.computers'
);

-- Category ancestors
SELECT * FROM catalog_categories 
WHERE 'electronics.computers.laptops.gaming' <@ path;

-- Category descendants
SELECT * FROM catalog_categories 
WHERE path <@ 'electronics.computers';

-- Move category (update path)
UPDATE catalog_categories 
SET path = 'electronics.gaming.laptops' 
WHERE path = 'electronics.computers.laptops.gaming';

-- Update all descendants
UPDATE catalog_categories 
SET path = text2ltree(
    replace(ltree2text(path), 'electronics.computers', 'electronics.tech')
)
WHERE path <@ 'electronics.computers';
```

### Consequences

#### Positive
- ✅ Unlimited category hierarchy
- ✅ Flexible item attributes
- ✅ Type-safe status management
- ✅ Fast tree queries with LTREE
- ✅ JSONB for flexible schema
- ✅ GIN indexes for attribute search

#### Negative
- ⚠️ PostgreSQL-specific (LTREE)
- ⚠️ Complex JSONB queries
- ⚠️ No built-in validation for attributes
- ⚠️ Attribute search performance depends on GIN index

### Integration Points

1. **Ordering Context**: Item reference in order lines
2. **Accounting Context**: Item pricing for transactions
3. **Files Context**: Item images and documents
4. **Search Context**: Full-text search on items

### Performance Considerations

- GiST index on `path` for LTREE operations
- GIN index on `attributes` for JSONB queries
- Partial index on `(status, category_id)` for active items
- Consider separate table for frequently searched attributes
- Text search index on `name` for fuzzy search

### Future Enhancements

1. **Search Integration**: Elasticsearch/OpenSearch
2. **Pricing Rules**: Dynamic pricing based on rules
3. **Inventory**: Stock levels, reservations
4. **Variants**: Product variants (size, color)
5. **Bundles**: Product bundles/kits
6. **Localization**: Multi-language support
7. **Media**: Images, videos, documents per item

### References
- PostgreSQL LTREE documentation
- JSONB design patterns
- Catalog management patterns
