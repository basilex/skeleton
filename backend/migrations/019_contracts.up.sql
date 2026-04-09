CREATE TYPE contract_type AS ENUM (
    'supply',
    'service',
    'employment',
    'partnership',
    'lease',
    'license'
);

CREATE TYPE contract_status AS ENUM (
    'draft',
    'pending_approval',
    'active',
    'expired',
    'terminated'
);

CREATE TYPE payment_type AS ENUM ('prepaid', 'postpaid', 'credit');

CREATE TABLE contracts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    contract_type contract_type NOT NULL,
    status contract_status NOT NULL DEFAULT 'draft',
    
    -- Party reference (links to parties.id)
    party_id UUID NOT NULL,
    
    -- Terms
    payment_terms JSONB NOT NULL,
    delivery_terms JSONB,
    
    -- Validity
    validity_period DATERANGE NOT NULL,
    
    -- Documents (array of file IDs)
    documents UUID[] DEFAULT '{}',
    
    -- Credit limit
    credit_limit BIGINT,
    currency VARCHAR(3) DEFAULT 'UAH',
    
    -- Metadata
    metadata JSONB DEFAULT '{}',
    
    -- Audit
    created_by UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    signed_at TIMESTAMPTZ,
    terminated_at TIMESTAMPTZ
);

-- Indexes for performance
CREATE INDEX idx_contracts_party ON contracts(party_id);
CREATE INDEX idx_contracts_status ON contracts(status);
CREATE INDEX idx_contracts_type ON contracts(contract_type);
CREATE INDEX idx_contracts_validity ON contracts USING GIST(validity_period);
CREATE INDEX idx_contracts_active ON contracts(party_id, status) WHERE status = 'active';
CREATE INDEX idx_contracts_created ON contracts(created_at);

-- Trigger for updated_at
CREATE TRIGGER contracts_updated_at
    BEFORE UPDATE ON contracts
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at();

-- Comments
COMMENT ON TABLE contracts IS 'Contracts and agreements between parties';
COMMENT ON COLUMN contracts.party_id IS 'Reference to party (customer, supplier, partner, employee)';
COMMENT ON COLUMN contracts.payment_terms IS 'Payment conditions: type, credit days, penalties, discounts';
COMMENT ON COLUMN contracts.delivery_terms IS 'Delivery conditions: type, estimated days, shipping cost';
COMMENT ON COLUMN contracts.validity_period IS 'Contract validity period (start and end dates)';
COMMENT ON COLUMN contracts.documents IS 'Array of file IDs (links to files context)';
COMMENT ON COLUMN contracts.credit_limit IS 'Credit limit for this contract (if applicable)';