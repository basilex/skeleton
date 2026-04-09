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

// Child Account management methods

// AddChild adds a child account to this account.
// It validates that the child has a compatible account type.
func (a *Account) AddChild(child *Account) error {
	if child == nil {
		return fmt.Errorf("child account cannot be nil")
	}
	if child.id == a.id {
		return ErrCircularReference
	}
	if child.accountType != a.accountType {
		return ErrParentTypeMismatch
	}
	child.parentID = &a.id
	child.updatedAt = time.Now().UTC()
	a.updatedAt = time.Now().UTC()
	return nil
}

// CanDelete checks if an account can be deleted.
// An account cannot be deleted if it has children or a non-zero balance.
func (a *Account) CanDelete() error {
	if a.HasChildren() {
		return ErrAccountHasChildren
	}
	if a.balance.Amount != 0 {
		return ErrAccountHasBalance
	}
	return nil
}

// HasChildren returns true if this account has child accounts.
// This is a placeholder - actual child tracking would be done by the repository.
func (a *Account) HasChildren() bool {
	return false
}

// IsRoot returns true if this account has no parent.
func (a *Account) IsRoot() bool {
	return a.parentID == nil
}

// IsDescendantOf checks if this account is a descendant of the given parent.
func (a *Account) IsDescendantOf(parentID AccountID, getParent func(AccountID) (*Account, error)) (bool, error) {
	if a.parentID == nil {
		return false, nil
	}
	if *a.parentID == parentID {
		return true, nil
	}
	if *a.parentID == a.id {
		return false, ErrCircularReference
	}
	parent, err := getParent(*a.parentID)
	if err != nil {
		return false, err
	}
	return parent.IsDescendantOf(parentID, getParent)
}

// SetParent sets the parent account.
// It validates that the parent has a compatible account type.
func (a *Account) SetParent(parent *Account) error {
	if parent == nil {
		a.parentID = nil
		a.updatedAt = time.Now().UTC()
		return nil
	}
	if parent.id == a.id {
		return ErrCircularReference
	}
	if parent.accountType != a.accountType {
		return ErrParentTypeMismatch
	}
	oldParentID := a.parentID
	a.parentID = &parent.id
	a.updatedAt = time.Now().UTC()

	a.events = append(a.events, AccountParentChanged{
		AccountID:   a.id,
		OldParentID: oldParentID,
		NewParentID: &parent.id,
		OcurredAt:   time.Now().UTC(),
	})

	return nil
}

// ClearParent removes the parent account, making this account a root.
func (a *Account) ClearParent() {
	a.parentID = nil
	a.updatedAt = time.Now().UTC()
}

// AccountPath returns the path from root to this account.
func (a *Account) AccountPath(getParent func(AccountID) (*Account, error)) ([]AccountID, error) {
	path := []AccountID{a.id}
	current := a

	for current.parentID != nil {
		if *current.parentID == a.id {
			return nil, ErrCircularReference
		}
		path = append([]AccountID{*current.parentID}, path...)
		parent, err := getParent(*current.parentID)
		if err != nil {
			return nil, err
		}
		current = parent
	}

	return path, nil
}

// Depth returns the depth of this account in the hierarchy tree.
func (a *Account) Depth(getParent func(AccountID) (*Account, error)) (int, error) {
	if a.parentID == nil {
		return 0, nil
	}
	parent, err := getParent(*a.parentID)
	if err != nil {
		return 0, err
	}
	parentDepth, err := parent.Depth(getParent)
	if err != nil {
		return 0, err
	}
	return parentDepth + 1, nil
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
