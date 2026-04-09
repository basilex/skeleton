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

func (m Money) Add(other Money) (Money, error) {
	if m.Currency != other.Currency {
		return Money{}, fmt.Errorf("cannot add different currencies")
	}
	return Money{Amount: m.Amount + other.Amount, Currency: m.Currency}, nil
}

func (m Money) Subtract(other Money) (Money, error) {
	if m.Currency != other.Currency {
		return Money{}, fmt.Errorf("cannot subtract different currencies")
	}
	if m.Amount < other.Amount {
		return Money{}, fmt.Errorf("insufficient funds")
	}
	return Money{Amount: m.Amount - other.Amount, Currency: m.Currency}, nil
}
