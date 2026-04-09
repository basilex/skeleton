# ADR-019: Accounting Bounded Context

## Status
Accepted

## Context
Chart of Accounts (Plan рахунків) and double-entry bookkeeping system for financial tracking.

## Decision
Implement Accounting as a bounded context with double-entry transactions.

### Domain Model
- **Account**: Chart of accounts entity (Asset, Liability, Equity, Revenue, Expense)
- **Transaction**: Double-entry bookkeeping transaction
- **Money**: Value object for amounts with currency
- **AccountType**: Asset, Liability, Equity, Revenue, Expense

### Architecture

- `internal/accounting/`
  - `domain/`
    - `account.go` - Account aggregate
    - `transaction.go` - Transaction aggregate
    - `money.go` - Money value object
    - `ids.go` - Identifiers
    - `events.go` - Domain events
    - `repository.go` - Repository interfaces
  - `infrastructure/`
    - `persistence/`
      - `models.go`
      - `account_repository.go`
      - `transaction_repository.go`
  - `application/`
    - `command/`
      - `create_account.go`
      - `record_transaction.go`
    - `query/`
      - `account.go`
      - `list_accounts.go`
  - `ports/http/`
    - `handler.go`
    - `dto.go`

### Database Schema

#### Chart of Accounts
```sql
CREATE TYPE account_type AS ENUM (
    'asset',       -- Активи (1000-1999)
    'liability',   -- Пасиви (2000-2999)
    'equity',      -- Капітал (3000-3999)
    'revenue',     -- Доходи (4000-4999)
    'expense'      -- Витрати (5000-5999)
);

CREATE TABLE accounts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    code VARCHAR(20) UNIQUE NOT NULL,      -- Account code (e.g., '1010')
    name VARCHAR(255) NOT NULL,              -- Account name
    account_type account_type NOT NULL,
    currency VARCHAR(3) DEFAULT 'UAH',
    balance DECIMAL(15,2) DEFAULT 0,
    parent_id UUID REFERENCES accounts(id),  -- Hierarchical structure
    is_active BOOLEAN NOT NULL DEFAULT true,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_accounts_type ON accounts(account_type);
CREATE INDEX idx_accounts_parent ON accounts(parent_id);
```

#### Transactions (Double-Entry)
```sql
CREATE TABLE transactions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    from_account UUID NOT NULL REFERENCES accounts(id),  -- Credit
    to_account UUID NOT NULL REFERENCES accounts(id),    -- Debit
    amount DECIMAL(15,2) NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'UAH',
    
    -- Reference (what is this transaction for?)
    reference_type VARCHAR(50),  -- 'invoice', 'payment', 'order'
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
```

### Key Design Decisions

#### 1. Hierarchical Chart of Accounts
```go
// Parent-child relationships
Account {
    ID:       AccountID
    Code:     "1010"            // Cash on hand
    Name:     "Готівка в касі"
    Type:     Asset
    ParentID: nil               // Root account
}

Account {
    ID:       AccountID
    Code:     "1011"            // Sub-account
    Name:     "Готівка в національній валюті"
    Type:     Asset
    ParentID: AccountID("1010") // Child of Cash
}
```

#### 2. Double-Entry Bookkeeping
```go
type Transaction struct {
    fromAccount AccountID   // Credit account
    toAccount   AccountID   // Debit account
    amount      Money
    currency    Currency
    reference   string      // What is this for?
    description string
    occurredAt  time.Time
}

// Every transaction affects two accounts
func (t *Transaction) Apply() error {
    // Debit destination account
    if err := toAccount.Debit(amount); err != nil {
        return err
    }
    
    // Credit source account
    if err := fromAccount.Credit(amount); err != nil {
        return err
    }
    
    return nil
}
```

#### 3. Account Types & Balance Direction

**Debit Increases:**
- Asset accounts (1010: Cash +Debit increases balance)
- Expense accounts (5000: Expenses +Debit increases)

**Credit Increases:**
- Liability accounts (2000: Accounts Payable +Credit increases)
- Equity accounts (3000: Owner's Equity +Credit increases)
- Revenue accounts (4000: Sales Revenue +Credit increases)

```go
func (a *Account) Debit(amount Money) error {
    switch a.accountType {
    case AccountTypeAsset, AccountTypeExpense:
        // Debit increases asset/expense
        a.balance.Amount += amount.Amount
    case AccountTypeLiability, AccountTypeEquity, AccountTypeRevenue:
        // Debit decreases liability/equity/revenue
        a.balance.Amount -= amount.Amount
    }
}

func (a *Account) Credit(amount Money) error {
    switch a.accountType {
    case AccountTypeAsset, AccountTypeExpense:
        // Credit decreases asset/expense
        a.balance.Amount -= amount.Amount
    case AccountTypeLiability, AccountTypeEquity, AccountTypeRevenue:
        // Credit increases liability/equity/revenue
        a.balance.Amount += amount.Amount
    }
}
```

### API Endpoints

```
POST   /api/v1/accounts              # Create account
GET    /api/v1/accounts/:id           # Get account
GET    /api/v1/accounts              # List accounts (paginated)
POST   /api/v1/transactions          # Record transaction
```

### Usage Example

```go
// Create Chart of Accounts
POST /api/v1/accounts
{
    "code": "1010",
    "name": "Готівка в касі",
    "account_type": "asset",
    "currency": "UAH"
}

POST /api/v1/accounts
{
    "code": "3010",
    "name": "Статутний капітал",
    "account_type": "equity",
    "currency": "UAH"
}

// Record transaction: Owner invests 100,000 UAH
POST /api/v1/transactions
{
    "from_account_id": "3010-account-id",  // Credit: Equity
    "to_account_id": "1010-account-id",     // Debit: Cash
    "amount": 100000,
    "currency": "UAH",
    "description": "Внесення статутного капіталу"
}

// Query: Account balance
GET /api/v1/accounts/1010-account-id

// Query: Account transactions
GET /api/v1/accounts/1010-account-id/transactions
```

### Standard Chart of Accounts (Ukraine)

```
1000-1999  Активи (Assets)
  1010      Готівка в касі
  1030      Готівка на розрахунковому рахунку
  2010      Запаси
  2600      Готова продукція

2000-2999  Пасиви (Liabilities)
  2010      Короткострокові кредити
  3010      Розрахунки з постачальниками

3000-3999  Капітал (Equity)
  3010      Статутний капітал
  3500      Нерозподілений прибуток

4000-4999  Доходи (Revenue)
  4010      Дохід від реалізації
  4300      Інші доходи

5000-5999  Витрати (Expenses)
  5010      Собівартість реалізації
  5200      Адміністративні витрати
```

### Consequences

#### Positive
- ✅ Double-entry integrity enforced
- ✅ Hierarchical account structure
- ✅ Type-safe account types
- ✅ Automatic balance tracking
- ✅ Audit trail for all transactions

#### Negative
- ⚠️ Complex for non-accountants
- ⚠️ Requires chart of accounts setup
- ⚠️ No rollback mechanism (need reversal transactions)
- ⚠️ Currency conversion not handled

### Integration Points

1. **Ordering Context**: Order payments generate transactions
2. **Parties Context**: Payables/receivables per party
3. **Audit Context**: All transactions audited
4. **Reports**: Financial statements generation

### Performance Considerations

- Index on `(from_account, occurred_at)` for account history
- Index on `(to_account, occurred_at)` for account history
- Composite index on `(reference_type, reference_id)` for lookups
- Consider partitioning by `occurred_at` for large volumes

### Future Enhancements

1. **Recurring Transactions**: Scheduled transactions
2. **Budgets**: Budget tracking per account
3. **Multi-currency**: Currency conversion
4. **Closing Entries**: Period closing automation
5. **Financial Reports**: Balance sheet, P&L generation

### References
- Double-entry bookkeeping principles
- Ukrainian Chart of Accounts (Plan рахунків)
- Accounting equation: Assets = Liabilities + Equity
