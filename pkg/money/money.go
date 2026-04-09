// Package money provides a Money value object for handling monetary values.
// It stores amounts in the smallest currency unit (cents) to avoid floating-point precision issues.
package money

import (
	"errors"
	"fmt"
	"strings"
)

var (
	// ErrNegativeAmount is returned when attempting to create money with a negative amount.
	ErrNegativeAmount = errors.New("amount cannot be negative")
	// ErrDifferentCurrencies is returned when attempting operations on different currencies.
	ErrDifferentCurrencies = errors.New("cannot operate on different currencies")
	// ErrInvalidCurrency is returned when currency code is invalid.
	ErrInvalidCurrency = errors.New("invalid currency code: must be 3-letter ISO 4217 code")
)

// Money represents a monetary value in the smallest currency unit (cents/minor units).
// This avoids floating-point precision issues common with currency calculations.
type Money struct {
	Amount   int64  // Amount in smallest currency unit (e.g., cents for USD)
	Currency string // ISO 4217 currency code (e.g., USD, EUR, UAH)
}

// New creates a new Money instance from an amount in the smallest currency unit and a currency code.
func New(amount int64, currency string) (Money, error) {
	if amount < 0 {
		return Money{}, ErrNegativeAmount
	}
	if currency == "" {
		return Money{}, errors.New("currency is required")
	}
	currency = strings.ToUpper(currency)
	if len(currency) != 3 {
		return Money{}, ErrInvalidCurrency
	}

	return Money{
		Amount:   amount,
		Currency: currency,
	}, nil
}

// NewFromFloat creates Money from a float64 amount (e.g., 12.34 -> 1234 cents).
// This is convenient for converting user input or API responses.
func NewFromFloat(amount float64, currency string) (Money, error) {
	if amount < 0 {
		return Money{}, ErrNegativeAmount
	}
	// Multiply by 100 and round to avoid floating-point errors
	return New(int64(amount*100+0.5), currency)
}

// Add adds two Money values. Returns error if currencies don't match.
func (m Money) Add(other Money) (Money, error) {
	if m.Currency != other.Currency {
		return Money{}, ErrDifferentCurrencies
	}
	return Money{
		Amount:   m.Amount + other.Amount,
		Currency: m.Currency,
	}, nil
}

// Subtract subtracts two Money values. Returns error if currencies don't match or result would be negative.
func (m Money) Subtract(other Money) (Money, error) {
	if m.Currency != other.Currency {
		return Money{}, ErrDifferentCurrencies
	}
	if m.Amount < other.Amount {
		return Money{}, ErrNegativeAmount
	}
	return Money{
		Amount:   m.Amount - other.Amount,
		Currency: m.Currency,
	}, nil
}

// Multiply multiplies Money by a factor. Returns error if factor is negative.
func (m Money) Multiply(factor float64) (Money, error) {
	if factor < 0 {
		return Money{}, ErrNegativeAmount
	}
	result := int64(float64(m.Amount)*factor + 0.5)
	return Money{
		Amount:   result,
		Currency: m.Currency,
	}, nil
}

// Divide divides Money by a factor. Returns error if factor is zero or negative.
func (m Money) Divide(factor float64) (Money, error) {
	if factor <= 0 {
		return Money{}, errors.New("division factor must be positive")
	}
	result := int64(float64(m.Amount)/factor + 0.5)
	return Money{
		Amount:   result,
		Currency: m.Currency,
	}, nil
}

// ToFloat64 converts Money to float64 for display or API responses (e.g., 1234 cents -> 12.34).
func (m Money) ToFloat64() float64 {
	return float64(m.Amount) / 100.0
}

// String returns formatted string representation (e.g., "12.34 USD").
func (m Money) String() string {
	return fmt.Sprintf("%.2f %s", m.ToFloat64(), m.Currency)
}

// Equals checks if two Money values are equal.
func (m Money) Equals(other Money) bool {
	return m.Amount == other.Amount && m.Currency == other.Currency
}

// IsZero checks if amount is zero.
func (m Money) IsZero() bool {
	return m.Amount == 0
}

// IsPositive checks if amount is positive.
func (m Money) IsPositive() bool {
	return m.Amount > 0
}

// CompareTo compares two Money values.
// Returns: -1 if m < other, 0 if m == other, 1 if m > other
// Panics if currencies don't match.
func (m Money) CompareTo(other Money) int {
	if m.Currency != other.Currency {
		panic(fmt.Sprintf("cannot compare different currencies: %s vs %s", m.Currency, other.Currency))
	}
	if m.Amount < other.Amount {
		return -1
	} else if m.Amount > other.Amount {
		return 1
	}
	return 0
}

// GreaterThan checks if this Money is greater than other.
func (m Money) GreaterThan(other Money) bool {
	return m.CompareTo(other) > 0
}

// LessThan checks if this Money is less than other.
func (m Money) LessThan(other Money) bool {
	return m.CompareTo(other) < 0
}

// GreaterThanOrEqual checks if this Money is greater than or equal to other.
func (m Money) GreaterThanOrEqual(other Money) bool {
	return m.CompareTo(other) >= 0
}

// LessThanOrEqual checks if this Money is less than or equal to other.
func (m Money) LessThanOrEqual(other Money) bool {
	return m.CompareTo(other) <= 0
}

// GetAmount returns the amount in smallest currency unit.
func (m Money) GetAmount() int64 {
	return m.Amount
}

// GetCurrency returns the ISO 4217 currency code.
func (m Money) GetCurrency() string {
	return m.Currency
}
