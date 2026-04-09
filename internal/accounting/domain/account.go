package domain

import (
	"fmt"
	"time"

	"github.com/basilex/skeleton/pkg/money"
)

type Account struct {
	id          AccountID
	code        string
	name        string
	accountType AccountType
	currency    Currency
	balance     money.Money
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
		balance:     money.Zero(string(currency)),
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
func (a *Account) GetBalance() money.Money { return a.balance }
func (a *Account) GetParentID() *AccountID { return a.parentID }
func (a *Account) IsActive() bool          { return a.isActive }
func (a *Account) GetCreatedAt() time.Time { return a.createdAt }
func (a *Account) GetUpdatedAt() time.Time { return a.updatedAt }

func (a *Account) Debit(amount money.Money) error {
	if !a.isActive {
		return ErrAccountInactive
	}
	if string(a.currency) != amount.GetCurrency() {
		return ErrDifferentCurrencies
	}

	var err error
	switch a.accountType {
	case AccountTypeAsset, AccountTypeExpense:
		a.balance, err = a.balance.Add(amount)
	case AccountTypeLiability, AccountTypeEquity, AccountTypeRevenue:
		a.balance, err = a.balance.Subtract(amount)
	}

	if err != nil {
		return err
	}

	a.updatedAt = time.Now().UTC()
	return nil
}

func (a *Account) Credit(amount money.Money) error {
	if !a.isActive {
		return ErrAccountInactive
	}
	if string(a.currency) != amount.GetCurrency() {
		return ErrDifferentCurrencies
	}

	var err error
	switch a.accountType {
	case AccountTypeAsset, AccountTypeExpense:
		a.balance, err = a.balance.Subtract(amount)
	case AccountTypeLiability, AccountTypeEquity, AccountTypeRevenue:
		a.balance, err = a.balance.Add(amount)
	}

	if err != nil {
		return err
	}

	a.updatedAt = time.Now().UTC()
	return nil
}

func (a *Account) SetBalance(balance money.Money) error {
	if string(a.currency) != balance.GetCurrency() {
		return ErrDifferentCurrencies
	}
	a.balance = balance
	a.updatedAt = time.Now().UTC()
	return nil
}

func (a *Account) Activate() error {
	if a.isActive {
		return fmt.Errorf("account is already active")
	}
	a.isActive = true
	a.updatedAt = time.Now().UTC()
	return nil
}

func (a *Account) Deactivate() error {
	if !a.isActive {
		return fmt.Errorf("account is already inactive")
	}
	a.isActive = false
	a.updatedAt = time.Now().UTC()
	return nil
}

func (a *Account) IsChildOf(parentID AccountID) bool {
	if a.parentID == nil {
		return false
	}
	return *a.parentID == parentID
}

func (a *Account) UpdateName(name string) error {
	if name == "" {
		return fmt.Errorf("account name is required")
	}
	a.name = name
	a.updatedAt = time.Now().UTC()
	return nil
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
	balance money.Money,
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
