-- Chart of Accounts (План рахунків)
CREATE TYPE account_type AS ENUM (
    'asset',       -- Активи
    'liability',   -- Пасиви
    'equity',      -- Капітал
    'revenue',     -- Доходи
    'expense'      -- Витрати
);

CREATE TABLE accounts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    code VARCHAR(20) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    account_type account_type NOT NULL,
    currency VARCHAR(3) DEFAULT 'UAH',
    balance DECIMAL(15,2) DEFAULT 0,
    parent_id UUID REFERENCES accounts(id),
    is_active BOOLEAN NOT NULL DEFAULT true,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_accounts_type ON accounts(account_type);
CREATE INDEX idx_accounts_parent ON accounts(parent_id);
CREATE INDEX idx_accounts_code ON accounts(code);

COMMENT ON TABLE accounts IS 'Chart of accounts (Plan рахунків)';
COMMENT ON COLUMN accounts.code IS 'Account code (e.g., 1010 for Cash)';
COMMENT ON COLUMN accounts.account_type IS 'Type: asset, liability, equity, revenue, expense';
COMMENT ON COLUMN accounts.balance IS 'Current balance in the account';

-- Transactions (Double-entry bookkeeping)
CREATE TABLE transactions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    from_account UUID NOT NULL REFERENCES accounts(id),  -- Credit
    to_account UUID NOT NULL REFERENCES accounts(id),    -- Debit
    amount DECIMAL(15,2) NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'UAH',
    
    -- Reference (what is this transaction for?)
    reference_type VARCHAR(50),  -- 'invoice', 'payment', 'order', etc.
    reference_id UUID,
    
    description TEXT,
    occurred_at TIMESTAMPTZ NOT NULL,
    posted_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    posted_by UUID REFERENCES users(id),
    
    metadata JSONB DEFAULT '{}'
);

CREATE INDEX idx_transactions_from ON transactions(from_account);
CREATE INDEX idx_transactions_to ON transactions(to_account);
CREATE INDEX idx_transactions_date ON transactions(occurred_at);
CREATE INDEX idx_transactions_reference ON transactions(reference_type, reference_id);

COMMENT ON TABLE transactions IS 'Double-entry bookkeeping transactions';
COMMENT ON COLUMN transactions.from_account IS 'Credit account (source)';
COMMENT ON COLUMN transactions.to_account IS 'Debit account (destination)';

-- Invoices
CREATE TYPE invoice_status AS ENUM ('draft', 'sent', 'paid', 'overdue', 'cancelled');
CREATE TYPE invoice_direction AS ENUM ('incoming', 'outgoing');

CREATE TABLE invoices (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    invoice_number VARCHAR(50) NOT NULL,
    direction invoice_direction NOT NULL,  -- incoming from supplier / outgoing to customer
    
    -- Party
    party_id UUID NOT NULL,  -- References parties.id (from parties context)
    contract_id UUID,         -- References contracts.id (from contracts context)
    
    -- Amounts
    subtotal DECIMAL(15,2) NOT NULL,
    tax_amount DECIMAL(15,2) DEFAULT 0,
    total DECIMAL(15,2) NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'UAH',
    
    -- Dates
    invoice_date DATE NOT NULL,
    due_date DATE NOT NULL,
    
    -- Status
    status invoice_status NOT NULL DEFAULT 'draft',
    
    -- Documents
    documents UUID[] DEFAULT '{}',  -- References files.id
    
    -- Audit
    created_by UUID,  -- References users.id
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_invoices_party ON invoices(party_id);
CREATE INDEX idx_invoices_status ON invoices(status);
CREATE INDEX idx_invoices_due ON invoices(due_date);

COMMENT ON TABLE invoices IS 'Invoices (incoming from suppliers, outgoing to customers)';

-- Payables (Money owed to suppliers)
CREATE TYPE payment_status AS ENUM ('unpaid', 'partially_paid', 'paid', 'overdue');

CREATE TABLE payables (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    supplier_id UUID NOT NULL,  -- References parties.id
    contract_id UUID,            -- References contracts.id
    invoice_id UUID REFERENCES invoices(id),
    
    -- Amounts
    amount DECIMAL(15,2) NOT NULL,
    tax_amount DECIMAL(15,2) DEFAULT 0,
    due_amount DECIMAL(15,2) NOT NULL,
    paid_amount DECIMAL(15,2) DEFAULT 0,
    currency VARCHAR(3) NOT NULL DEFAULT 'UAH',
    
    -- Dates
    invoice_date DATE NOT NULL,
    due_date DATE NOT NULL,
    
    -- Status
    payment_status payment_status NOT NULL DEFAULT 'unpaid',
    
    -- Documents
    documents UUID[] DEFAULT '{}',
    
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_payables_supplier ON payables(supplier_id);
CREATE INDEX idx_payables_status ON payables(payment_status);
CREATE INDEX idx_payables_due ON payables(due_date);

COMMENT ON TABLE payables IS 'Accounts payable (Кредиторська заборгованість)';

-- Receivables (Money owed by customers)
CREATE TABLE receivables (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    customer_id UUID NOT NULL,   -- References parties.id
    contract_id UUID,             -- References contracts.id
    invoice_id UUID REFERENCES invoices(id),
    
    -- Amounts
    amount DECIMAL(15,2) NOT NULL,
    tax_amount DECIMAL(15,2) DEFAULT 0,
    due_amount DECIMAL(15,2) NOT NULL,
    paid_amount DECIMAL(15,2) DEFAULT 0,
    currency VARCHAR(3) NOT NULL DEFAULT 'UAH',
    
    -- Dates
    invoice_date DATE NOT NULL,
    due_date DATE NOT NULL,
    
    -- Status
    payment_status payment_status NOT NULL DEFAULT 'unpaid',
    
    -- Documents
    documents UUID[] DEFAULT '{}',
    
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_receivables_customer ON receivables(customer_id);
CREATE INDEX idx_receivables_status ON receivables(payment_status);
CREATE INDEX idx_receivables_due ON receivables(due_date);

COMMENT ON TABLE receivables IS 'Accounts receivable (Дебіторська заборгованість)';

-- Payments
CREATE TYPE payment_method AS ENUM ('cash', 'bank_transfer', 'card', 'check');

CREATE TABLE payments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    payable_id UUID REFERENCES payables(id),
    receivable_id UUID REFERENCES receivables(id),
    invoice_id UUID REFERENCES invoices(id),
    
    -- Amount
    amount DECIMAL(15,2) NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'UAH',
    
    -- Method
    payment_method payment_method NOT NULL,
    payment_date DATE NOT NULL,
    
    -- Reference
    transaction_id UUID REFERENCES transactions(id),
    reference_number VARCHAR(100),
    
    -- Audit
    created_by UUID,  -- References users.id
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_payments_payable ON payments(payable_id);
CREATE INDEX idx_payments_receivable ON payments(receivable_id);
CREATE INDEX idx_payments_date ON payments(payment_date);

COMMENT ON TABLE payments IS 'Payment records';

-- Triggers for updated_at
CREATE TRIGGER invoices_updated_at
    BEFORE UPDATE ON invoices
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER payables_updated_at
    BEFORE UPDATE ON payables
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER receivables_updated_at
    BEFORE UPDATE ON receivables
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at();