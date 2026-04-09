package domain

import (
	"fmt"
	"time"
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
	totalPurchases float64
	creditLimit    float64
	currentCredit  float64
	createdAt      time.Time
	updatedAt      time.Time
	events         []DomainEvent
}

func NewCustomer(name, taxID string, contactInfo ContactInfo) (*Customer, error) {
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}

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
		totalPurchases: 0,
		creditLimit:    0,
		currentCredit:  0,
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

func (c *Customer) GetID() PartyID                { return c.id }
func (c *Customer) GetType() PartyType            { return c.partyType }
func (c *Customer) GetName() string               { return c.name }
func (c *Customer) GetContactInfo() ContactInfo   { return c.contactInfo }
func (c *Customer) GetTaxID() string              { return c.taxID }
func (c *Customer) GetStatus() PartyStatus        { return c.status }
func (c *Customer) GetBankAccount() BankAccount   { return c.bankAccount }
func (c *Customer) GetLoyaltyLevel() LoyaltyLevel { return c.loyaltyLevel }
func (c *Customer) GetTotalPurchases() float64    { return c.totalPurchases }
func (c *Customer) GetCreditLimit() float64       { return c.creditLimit }
func (c *Customer) GetCurrentCredit() float64     { return c.currentCredit }
func (c *Customer) GetAvailableCredit() float64   { return c.creditLimit - c.currentCredit }
func (c *Customer) GetCreatedAt() time.Time       { return c.createdAt }
func (c *Customer) GetUpdatedAt() time.Time       { return c.updatedAt }

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

func (c *Customer) AddPurchase(amount float64) {
	c.totalPurchases += amount
	c.updatedAt = time.Now().UTC()
	c.updateLoyaltyLevel()
}

func (c *Customer) updateLoyaltyLevel() {
	switch {
	case c.totalPurchases >= 100000:
		c.loyaltyLevel = LoyaltyLevelPlatinum
	case c.totalPurchases >= 50000:
		c.loyaltyLevel = LoyaltyLevelGold
	case c.totalPurchases >= 20000:
		c.loyaltyLevel = LoyaltyLevelSilver
	default:
		c.loyaltyLevel = LoyaltyLevelBronze
	}
}

func (c *Customer) SetCreditLimit(limit float64) error {
	if limit < 0 {
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

func (c *Customer) UseCredit(amount float64) error {
	if amount <= 0 {
		return fmt.Errorf("amount must be positive")
	}
	if c.status == PartyStatusBlacklisted {
		return ErrPartyBlacklisted
	}

	newCredit := c.currentCredit + amount
	if newCredit > c.creditLimit {
		return fmt.Errorf("credit limit exceeded: current %.2f + %.2f > limit %.2f",
			c.currentCredit, amount, c.creditLimit)
	}

	c.currentCredit = newCredit
	c.updatedAt = time.Now().UTC()
	return nil
}

func (c *Customer) RepayCredit(amount float64) error {
	if amount <= 0 {
		return fmt.Errorf("amount must be positive")
	}

	if amount > c.currentCredit {
		return fmt.Errorf("repayment amount %.2f exceeds current credit %.2f",
			amount, c.currentCredit)
	}

	c.currentCredit -= amount
	c.updatedAt = time.Now().UTC()
	return nil
}

func (c *Customer) HasAvailableCredit() bool {
	return c.currentCredit < c.creditLimit
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
	totalPurchases float64,
	creditLimit float64,
	currentCredit float64,
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
