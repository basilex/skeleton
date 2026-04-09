package domain

import (
	"fmt"
	"time"

	"github.com/basilex/skeleton/pkg/money"
)

type Customer struct {
	id             PartyID
	partyType      PartyType
	name           string
	taxID          string
	contactInfo    ContactInfo
	bankAccount    BankAccount
	status         PartyStatus
	loyaltyLevel   LoyaltyLevel
	totalPurchases money.Money
	creditLimit    money.Money
	currentCredit  money.Money
	createdAt      time.Time
	updatedAt      time.Time
	events         []DomainEvent
}

func NewCustomer(name, taxID string, contactInfo ContactInfo) (*Customer, error) {
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}

	zeroMoney := money.Zero("USD")
	now := time.Now().UTC()
	c := &Customer{
		id:             NewPartyID(),
		partyType:      PartyTypeCustomer,
		name:           name,
		taxID:          taxID,
		contactInfo:    contactInfo,
		bankAccount:    BankAccount{},
		status:         PartyStatusActive,
		loyaltyLevel:   LoyaltyLevelBronze,
		totalPurchases: zeroMoney,
		creditLimit:    zeroMoney,
		currentCredit:  zeroMoney,
		createdAt:      now,
		updatedAt:      now,
		events:         make([]DomainEvent, 0),
	}

	c.events = append(c.events, CustomerCreated{
		PartyID:   c.id,
		Name:      c.name,
		Email:     c.contactInfo.Email,
		OcurredAt: now,
	})

	return c, nil
}

func (c *Customer) GetID() PartyID                 { return c.id }
func (c *Customer) GetType() PartyType             { return c.partyType }
func (c *Customer) GetName() string                { return c.name }
func (c *Customer) GetContactInfo() ContactInfo    { return c.contactInfo }
func (c *Customer) GetTaxID() string               { return c.taxID }
func (c *Customer) GetStatus() PartyStatus         { return c.status }
func (c *Customer) GetBankAccount() BankAccount    { return c.bankAccount }
func (c *Customer) GetLoyaltyLevel() LoyaltyLevel  { return c.loyaltyLevel }
func (c *Customer) GetTotalPurchases() money.Money { return c.totalPurchases }
func (c *Customer) GetCreditLimit() money.Money    { return c.creditLimit }
func (c *Customer) GetCurrentCredit() money.Money  { return c.currentCredit }
func (c *Customer) GetAvailableCredit() money.Money {
	available, _ := c.creditLimit.Subtract(c.currentCredit)
	return available
}
func (c *Customer) GetCreatedAt() time.Time { return c.createdAt }
func (c *Customer) GetUpdatedAt() time.Time { return c.updatedAt }

func (c *Customer) UpdateContactInfo(contactInfo ContactInfo) error {
	if c.status == PartyStatusBlacklisted {
		return ErrPartyBlacklisted
	}
	c.contactInfo = contactInfo
	c.updatedAt = time.Now().UTC()
	return nil
}

func (c *Customer) UpdateBankAccount(bankAccount BankAccount) error {
	if c.status == PartyStatusBlacklisted {
		return ErrPartyBlacklisted
	}
	c.bankAccount = bankAccount
	c.updatedAt = time.Now().UTC()
	return nil
}

func (c *Customer) AddPurchase(amount money.Money) {
	c.totalPurchases, _ = c.totalPurchases.Add(amount)
	c.updatedAt = time.Now().UTC()
	c.updateLoyaltyLevel()
}

func (c *Customer) updateLoyaltyLevel() {
	totalAmount := c.totalPurchases.GetAmount()
	switch {
	case totalAmount >= 10000000:
		c.loyaltyLevel = LoyaltyLevelPlatinum
	case totalAmount >= 5000000:
		c.loyaltyLevel = LoyaltyLevelGold
	case totalAmount >= 2000000:
		c.loyaltyLevel = LoyaltyLevelSilver
	default:
		c.loyaltyLevel = LoyaltyLevelBronze
	}
}

func (c *Customer) SetCreditLimit(limit money.Money) error {
	if limit.IsNegative() {
		return fmt.Errorf("credit limit cannot be negative")
	}
	if c.status == PartyStatusBlacklisted {
		return ErrPartyBlacklisted
	}

	oldLimit := c.creditLimit
	c.creditLimit = limit
	c.updatedAt = time.Now().UTC()

	c.events = append(c.events, CustomerCreditLimitChanged{
		PartyID:    c.id,
		OldLimit:   oldLimit,
		NewLimit:   limit,
		occurredAt: c.updatedAt,
	})

	return nil
}

func (c *Customer) UseCredit(amount money.Money) error {
	if amount.IsNegative() || amount.IsZero() {
		return fmt.Errorf("amount must be positive")
	}
	if c.status == PartyStatusBlacklisted {
		return ErrPartyBlacklisted
	}

	newCredit, err := c.currentCredit.Add(amount)
	if err != nil {
		return err
	}

	if newCredit.GreaterThan(c.creditLimit) {
		return fmt.Errorf("credit limit exceeded: current %s + %s > limit %s",
			c.currentCredit.String(), amount.String(), c.creditLimit.String())
	}

	c.currentCredit = newCredit
	c.updatedAt = time.Now().UTC()
	return nil
}

func (c *Customer) RepayCredit(amount money.Money) error {
	if amount.IsNegative() || amount.IsZero() {
		return fmt.Errorf("amount must be positive")
	}

	if amount.GreaterThan(c.currentCredit) {
		return fmt.Errorf("repayment amount %s exceeds current credit %s",
			amount.String(), c.currentCredit.String())
	}

	c.currentCredit, _ = c.currentCredit.Subtract(amount)
	c.updatedAt = time.Now().UTC()
	return nil
}

func (c *Customer) HasAvailableCredit() bool {
	return c.currentCredit.LessThan(c.creditLimit)
}

func (c *Customer) Activate() error {
	if c.status == PartyStatusBlacklisted {
		return ErrPartyBlacklisted
	}
	oldStatus := c.status
	c.status = PartyStatusActive
	c.updatedAt = time.Now().UTC()
	c.events = append(c.events, PartyStatusChanged{
		PartyID:   c.id,
		PartyType: c.partyType,
		OldStatus: oldStatus,
		NewStatus: c.status,
		OcurredAt: c.updatedAt,
	})
	return nil
}

func (c *Customer) Deactivate() {
	oldStatus := c.status
	c.status = PartyStatusInactive
	c.updatedAt = time.Now().UTC()
	c.events = append(c.events, PartyStatusChanged{
		PartyID:   c.id,
		PartyType: c.partyType,
		OldStatus: oldStatus,
		NewStatus: c.status,
		OcurredAt: c.updatedAt,
	})
}

func (c *Customer) Blacklist() {
	oldStatus := c.status
	c.status = PartyStatusBlacklisted
	c.updatedAt = time.Now().UTC()
	c.events = append(c.events, PartyStatusChanged{
		PartyID:   c.id,
		PartyType: c.partyType,
		OldStatus: oldStatus,
		NewStatus: c.status,
		OcurredAt: c.updatedAt,
	})
}

func (c *Customer) PullEvents() []DomainEvent {
	events := c.events
	c.events = make([]DomainEvent, 0)
	return events
}

func ReconstituteCustomer(
	id PartyID,
	name, taxID string,
	contactInfo ContactInfo,
	bankAccount BankAccount,
	status PartyStatus,
	loyaltyLevel LoyaltyLevel,
	totalPurchases money.Money,
	creditLimit money.Money,
	currentCredit money.Money,
	createdAt, updatedAt time.Time,
) (*Customer, error) {
	return &Customer{
		id:             id,
		partyType:      PartyTypeCustomer,
		name:           name,
		taxID:          taxID,
		contactInfo:    contactInfo,
		bankAccount:    bankAccount,
		status:         status,
		loyaltyLevel:   loyaltyLevel,
		totalPurchases: totalPurchases,
		creditLimit:    creditLimit,
		currentCredit:  currentCredit,
		createdAt:      createdAt,
		updatedAt:      updatedAt,
		events:         make([]DomainEvent, 0),
	}, nil
}
