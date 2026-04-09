package domain

import (
	"fmt"

	"github.com/basilex/skeleton/pkg/uuid"
)

type AccountID uuid.UUID

func NewAccountID() AccountID {
	return AccountID(uuid.NewV7())
}

func ParseAccountID(s string) (AccountID, error) {
	u, err := uuid.Parse(s)
	if err != nil {
		return AccountID{}, fmt.Errorf("invalid account id: %w", err)
	}
	return AccountID(u), nil
}

func (id AccountID) String() string {
	return uuid.UUID(id).String()
}

type AccountType string

const (
	AccountTypeAsset     AccountType = "asset"
	AccountTypeLiability AccountType = "liability"
	AccountTypeEquity    AccountType = "equity"
	AccountTypeRevenue   AccountType = "revenue"
	AccountTypeExpense   AccountType = "expense"
)

func (at AccountType) String() string {
	return string(at)
}

func ParseAccountType(s string) (AccountType, error) {
	switch AccountType(s) {
	case AccountTypeAsset, AccountTypeLiability, AccountTypeEquity,
		AccountTypeRevenue, AccountTypeExpense:
		return AccountType(s), nil
	default:
		return "", fmt.Errorf("invalid account type: %s", s)
	}
}

type Currency string

const (
	CurrencyUAH Currency = "UAH"
	CurrencyUSD Currency = "USD"
	CurrencyEUR Currency = "EUR"
	CurrencyGBP Currency = "GBP"
)

func (c Currency) String() string {
	return string(c)
}

func ParseCurrency(s string) (Currency, error) {
	switch Currency(s) {
	case CurrencyUAH, CurrencyUSD, CurrencyEUR, CurrencyGBP:
		return Currency(s), nil
	default:
		return "", fmt.Errorf("invalid currency: %s", s)
	}
}

type Money struct {
	Amount   float64
	Currency Currency
}

func NewMoney(amount float64, currency Currency) (Money, error) {
	if amount < 0 {
		return Money{}, fmt.Errorf("amount cannot be negative")
	}
	return Money{Amount: amount, Currency: currency}, nil
}

func (m Money) Add(other Money) Money {
	return Money{Amount: m.Amount + other.Amount, Currency: m.Currency}
}

func (m Money) Subtract(other Money) Money {
	return Money{Amount: m.Amount - other.Amount, Currency: m.Currency}
}

func (m Money) IsZero() bool {
	return m.Amount == 0
}

func (m Money) Equals(other Money) bool {
	return m.Amount == other.Amount && m.Currency == other.Currency
}

func (m Money) IsNegative() bool {
	return m.Amount < 0
}

type JournalEntryID string

func NewJournalEntryID() JournalEntryID {
	return JournalEntryID(uuid.NewV7().String())
}

func ParseJournalEntryID(s string) (JournalEntryID, error) {
	return JournalEntryID(s), nil
}

func (id JournalEntryID) String() string {
	return string(id)
}

type AccountingPeriodID string

func NewAccountingPeriodID() AccountingPeriodID {
	return AccountingPeriodID(uuid.NewV7().String())
}

func ParseAccountingPeriodID(s string) (AccountingPeriodID, error) {
	return AccountingPeriodID(s), nil
}

func (id AccountingPeriodID) String() string {
	return string(id)
}

type ReconciliationID string

func NewReconciliationID() ReconciliationID {
	return ReconciliationID(uuid.NewV7().String())
}

func ParseReconciliationID(s string) (ReconciliationID, error) {
	return ReconciliationID(s), nil
}

func (id ReconciliationID) String() string {
	return string(id)
}
