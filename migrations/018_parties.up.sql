CREATE TYPE party_type AS ENUM ('customer', 'supplier', 'partner', 'employee');
CREATE TYPE party_status AS ENUM ('active', 'inactive', 'blacklisted');
CREATE TYPE loyalty_level AS ENUM ('bronze', 'silver', 'gold', 'platinum');

CREATE TABLE parties (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    party_type party_type NOT NULL,
    name VARCHAR(255) NOT NULL,
    tax_id VARCHAR(50),
    
    contact_info JSONB NOT NULL DEFAULT '{}',
    bank_account JSONB,
    
    status party_status NOT NULL DEFAULT 'active',
    
    loyalty_level loyalty_level,
    total_purchases DECIMAL(15,2) DEFAULT 0,
    
    rating JSONB,
    contracts UUID[] DEFAULT '{}',
    
    position VARCHAR(255),
    
    metadata JSONB DEFAULT '{}',
    
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    CONSTRAINT valid_tax_id CHECK (tax_id IS NULL OR LENGTH(tax_id) >= 8)
);

CREATE INDEX idx_parties_type ON parties(party_type);
CREATE INDEX idx_parties_status ON parties(status);
CREATE INDEX idx_parties_tax_id ON parties(tax_id) WHERE tax_id IS NOT NULL;
CREATE INDEX idx_parties_name ON parties USING GIN(to_tsvector('english', name));
CREATE INDEX idx_parties_created ON parties(created_at);
CREATE INDEX idx_parties_contact ON parties USING GIN(contact_info);

CREATE TRIGGER parties_updated_at
    BEFORE UPDATE ON parties
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at();

COMMENT ON TABLE parties IS 'Universal parties: customers, suppliers, partners, employees';
COMMENT ON COLUMN parties.party_type IS 'Type: customer, supplier, partner, employee';
COMMENT ON COLUMN parties.tax_id IS 'Tax identification number (IPN, EDRPOU, etc.)';
COMMENT ON COLUMN parties.contact_info IS 'Contact information (email, phone, address)';
COMMENT ON COLUMN parties.bank_account IS 'Bank account details';
COMMENT ON COLUMN parties.total_purchases IS 'Total purchases amount (for customers)';
COMMENT ON COLUMN parties.rating IS 'Supplier rating: quality, delivery speed, price';
COMMENT ON COLUMN parties.contracts IS 'List of active contract IDs';
COMMENT ON COLUMN parties.position IS 'Employee position';