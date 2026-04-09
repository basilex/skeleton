package domain

import (
	"fmt"
	"time"
)

type Employee struct {
	id          PartyID
	partyType   PartyType
	name        string
	taxID       string
	position    string
	contactInfo ContactInfo
	bankAccount BankAccount
	status      PartyStatus
	createdAt   time.Time
	updatedAt   time.Time
	events      []DomainEvent
}

func NewEmployee(name, taxID, position string, contactInfo ContactInfo) (*Employee, error) {
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}
	if position == "" {
		return nil, fmt.Errorf("position is required")
	}

	now := time.Now().UTC()
	e := &Employee{
		id:          NewPartyID(),
		partyType:   PartyTypeEmployee,
		name:        name,
		taxID:       taxID,
		position:    position,
		contactInfo: contactInfo,
		bankAccount: BankAccount{},
		status:      PartyStatusActive,
		createdAt:   now,
		updatedAt:   now,
		events:      make([]DomainEvent, 0),
	}

	e.events = append(e.events, EmployeeCreated{
		PartyID:   e.id,
		Name:      e.name,
		Email:     e.contactInfo.Email,
		Position:  e.position,
		OcurredAt: now,
	})

	return e, nil
}

func (e *Employee) GetID() PartyID              { return e.id }
func (e *Employee) GetType() PartyType          { return e.partyType }
func (e *Employee) GetName() string             { return e.name }
func (e *Employee) GetTaxID() string            { return e.taxID }
func (e *Employee) GetPosition() string         { return e.position }
func (e *Employee) GetContactInfo() ContactInfo { return e.contactInfo }
func (e *Employee) GetStatus() PartyStatus      { return e.status }
func (e *Employee) GetBankAccount() BankAccount { return e.bankAccount }
func (e *Employee) GetCreatedAt() time.Time     { return e.createdAt }
func (e *Employee) GetUpdatedAt() time.Time     { return e.updatedAt }

func (e *Employee) UpdateContactInfo(contactInfo ContactInfo) error {
	if e.status == PartyStatusBlacklisted {
		return ErrPartyBlacklisted
	}
	e.contactInfo = contactInfo
	e.updatedAt = time.Now().UTC()
	return nil
}

func (e *Employee) UpdateBankAccount(bankAccount BankAccount) error {
	if e.status == PartyStatusBlacklisted {
		return ErrPartyBlacklisted
	}
	e.bankAccount = bankAccount
	e.updatedAt = time.Now().UTC()
	return nil
}

func (e *Employee) UpdatePosition(position string) error {
	if e.status == PartyStatusBlacklisted {
		return ErrPartyBlacklisted
	}
	if position == "" {
		return fmt.Errorf("position is required")
	}
	e.position = position
	e.updatedAt = time.Now().UTC()
	return nil
}

func (e *Employee) Activate() error {
	if e.status == PartyStatusBlacklisted {
		return ErrPartyBlacklisted
	}
	oldStatus := e.status
	e.status = PartyStatusActive
	e.updatedAt = time.Now().UTC()
	e.events = append(e.events, PartyStatusChanged{
		PartyID:   e.id,
		PartyType: e.partyType,
		OldStatus: oldStatus,
		NewStatus: e.status,
		OcurredAt: e.updatedAt,
	})
	return nil
}

func (e *Employee) Deactivate() {
	oldStatus := e.status
	e.status = PartyStatusInactive
	e.updatedAt = time.Now().UTC()
	e.events = append(e.events, PartyStatusChanged{
		PartyID:   e.id,
		PartyType: e.partyType,
		OldStatus: oldStatus,
		NewStatus: e.status,
		OcurredAt: e.updatedAt,
	})
}

func (e *Employee) Blacklist() {
	oldStatus := e.status
	e.status = PartyStatusBlacklisted
	e.updatedAt = time.Now().UTC()
	e.events = append(e.events, PartyStatusChanged{
		PartyID:   e.id,
		PartyType: e.partyType,
		OldStatus: oldStatus,
		NewStatus: e.status,
		OcurredAt: e.updatedAt,
	})
}

func (e *Employee) PullEvents() []DomainEvent {
	events := e.events
	e.events = make([]DomainEvent, 0)
	return events
}

func ReconstituteEmployee(
	id PartyID,
	name, taxID, position string,
	contactInfo ContactInfo,
	bankAccount BankAccount,
	status PartyStatus,
	createdAt, updatedAt time.Time,
) (*Employee, error) {
	return &Employee{
		id:          id,
		partyType:   PartyTypeEmployee,
		name:        name,
		taxID:       taxID,
		position:    position,
		contactInfo: contactInfo,
		bankAccount: bankAccount,
		status:      status,
		createdAt:   createdAt,
		updatedAt:   updatedAt,
		events:      make([]DomainEvent, 0),
	}, nil
}
