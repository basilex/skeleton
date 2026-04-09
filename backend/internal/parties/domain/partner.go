package domain

import (
	"fmt"
	"time"
)

type Partner struct {
	id          PartyID
	partyType   PartyType
	name        string
	taxID       string
	contactInfo ContactInfo
	bankAccount BankAccount
	status      PartyStatus
	createdAt   time.Time
	updatedAt   time.Time
	events      []DomainEvent
}

func NewPartner(name, taxID string, contactInfo ContactInfo) (*Partner, error) {
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}

	now := time.Now().UTC()
	p := &Partner{
		id:          NewPartyID(),
		partyType:   PartyTypePartner,
		name:        name,
		taxID:       taxID,
		contactInfo: contactInfo,
		bankAccount: BankAccount{},
		status:      PartyStatusActive,
		createdAt:   now,
		updatedAt:   now,
		events:      make([]DomainEvent, 0),
	}

	p.events = append(p.events, PartnerCreated{
		PartyID:   p.id,
		Name:      p.name,
		Email:     p.contactInfo.Email,
		OcurredAt: now,
	})

	return p, nil
}

func (p *Partner) GetID() PartyID              { return p.id }
func (p *Partner) GetType() PartyType          { return p.partyType }
func (p *Partner) GetName() string             { return p.name }
func (p *Partner) GetContactInfo() ContactInfo { return p.contactInfo }
func (p *Partner) GetTaxID() string            { return p.taxID }
func (p *Partner) GetStatus() PartyStatus      { return p.status }
func (p *Partner) GetBankAccount() BankAccount { return p.bankAccount }
func (p *Partner) GetCreatedAt() time.Time     { return p.createdAt }
func (p *Partner) GetUpdatedAt() time.Time     { return p.updatedAt }

func (p *Partner) UpdateContactInfo(contactInfo ContactInfo) error {
	if p.status == PartyStatusBlacklisted {
		return ErrPartyBlacklisted
	}
	p.contactInfo = contactInfo
	p.updatedAt = time.Now().UTC()
	return nil
}

func (p *Partner) UpdateBankAccount(bankAccount BankAccount) error {
	if p.status == PartyStatusBlacklisted {
		return ErrPartyBlacklisted
	}
	p.bankAccount = bankAccount
	p.updatedAt = time.Now().UTC()
	return nil
}

func (p *Partner) Activate() error {
	if p.status == PartyStatusBlacklisted {
		return ErrPartyBlacklisted
	}
	oldStatus := p.status
	p.status = PartyStatusActive
	p.updatedAt = time.Now().UTC()
	p.events = append(p.events, PartyStatusChanged{
		PartyID:   p.id,
		PartyType: p.partyType,
		OldStatus: oldStatus,
		NewStatus: p.status,
		OcurredAt: p.updatedAt,
	})
	return nil
}

func (p *Partner) Deactivate() {
	oldStatus := p.status
	p.status = PartyStatusInactive
	p.updatedAt = time.Now().UTC()
	p.events = append(p.events, PartyStatusChanged{
		PartyID:   p.id,
		PartyType: p.partyType,
		OldStatus: oldStatus,
		NewStatus: p.status,
		OcurredAt: p.updatedAt,
	})
}

func (p *Partner) Blacklist() {
	oldStatus := p.status
	p.status = PartyStatusBlacklisted
	p.updatedAt = time.Now().UTC()
	p.events = append(p.events, PartyStatusChanged{
		PartyID:   p.id,
		PartyType: p.partyType,
		OldStatus: oldStatus,
		NewStatus: p.status,
		OcurredAt: p.updatedAt,
	})
}

func (p *Partner) PullEvents() []DomainEvent {
	events := p.events
	p.events = make([]DomainEvent, 0)
	return events
}

func ReconstitutePartner(
	id PartyID,
	name, taxID string,
	contactInfo ContactInfo,
	bankAccount BankAccount,
	status PartyStatus,
	createdAt, updatedAt time.Time,
) (*Partner, error) {
	return &Partner{
		id:          id,
		partyType:   PartyTypePartner,
		name:        name,
		taxID:       taxID,
		contactInfo: contactInfo,
		bankAccount: bankAccount,
		status:      status,
		createdAt:   createdAt,
		updatedAt:   updatedAt,
		events:      make([]DomainEvent, 0),
	}, nil
}
