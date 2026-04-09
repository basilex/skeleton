package domain

import (
	"fmt"
	"time"
)

type Account struct {
	id          AccountID
	code        string
	name        string
	accountType AccountType
	currency    Currency
	balance     Money
	parentID    *AccountID
	isActive    bool
	createdAt   time.Time
	updatedAt   time.Time
	events      []DomainEvent
}

func NewAccount(code, name string, accountType AccountType, currency Currency, parentID *AccountID) (*Account, error) {
	if code == "" {
		return nil, fmt.Errorf("account code is required")
	}
	if name == "" {
		return nil, fmt.Errorf("account name is required")
	}

	now := time.Now().UTC()
	a := &Account{
		id:          NewAccountID(),
		code:        code,
		name:        name,
		accountType: accountType,
		currency:    currency,
		balance:     Money{Amount: 0, Currency: currency},
		parentID:    parentID,
		isActive:    true,
		createdAt:   now,
		updatedAt:   now,
		events:      make([]DomainEvent, 0),
	}

	a.events = append(a.events, AccountCreated{
		AccountID:   a.id,
		AccountCode: a.code,
		AccountName: a.name,
		AccountType: a.accountType,
		OcurredAt:   now,
	})

	return a, nil
}

func (a *Account) GetID() AccountID        { return a.id }
func (a *Account) GetCode() string         { return a.code }
func (a *Account) GetName() string         { return a.name }
func (a *Account) GetType() AccountType    { return a.accountType }
func (a *Account) GetCurrency() Currency   { return a.currency }
func (a *Account) GetBalance() Money       { return a.balance }
func (a *Account) GetParentID() *AccountID { return a.parentID }
func (a *Account) IsActive() bool          { return a.isActive }
func (a *Account) GetCreatedAt() time.Time { return a.createdAt }
func (a *Account) GetUpdatedAt() time.Time { return a.updatedAt }

func (a *Account) Debit(amount Money) error {
	if !a.isActive {
		return ErrAccountInactive
	}
	if a.currency != amount.Currency {
		return ErrDifferentCurrencies
	}

	switch a.accountType {
	case AccountTypeAsset, AccountTypeExpense:
		a.balance.Amount += amount.Amount
	case AccountTypeLiability, AccountTypeEquity, AccountTypeRevenue:
		a.balance.Amount -= amount.Amount
	}

	a.updatedAt = time.Now().UTC()
	return nil
}

func (a *Account) Credit(amount Money) error {
	if !a.isActive {
		return ErrAccountInactive
	}
	if a.currency != amount.Currency {
		return ErrDifferentCurrencies
	}

	switch a.accountType {
	case AccountTypeAsset, AccountTypeExpense:
		a.balance.Amount -= amount.Amount
	case AccountTypeLiability, AccountTypeEquity, AccountTypeRevenue:
		a.balance.Amount += amount.Amount
	}

	a.updatedAt = time.Now().UTC()
	return nil
}

func (a *Account) Deactivate() {
	a.isActive = false
	a.updatedAt = time.Now().UTC()
}

func (a *Account) Activate() {
	a.isActive = true
	a.updatedAt = time.Now().UTC()
}

func (a *Account) PullEvents() []DomainEvent {
	events := a.events
	a.events = make([]DomainEvent, 0)
	return events
}

func ReconstituteAccount(
	id AccountID,
	code, name string,
	accountType AccountType,
	currency Currency,
	balance Money,
	parentID *AccountID,
	isActive bool,
	createdAt, updatedAt time.Time,
) (*Account, error) {
	return &Account{
		id:          id,
		code:        code,
		name:        name,
		accountType: accountType,
		currency:    currency,
		balance:     balance,
		parentID:    parentID,
		isActive:    isActive,
		createdAt:   createdAt,
		updatedAt:   updatedAt,
		events:      make([]DomainEvent, 0),
	}, nil
}
