package domain

import "time"

type DomainEvent interface {
	EventName() string
	OccurredAt() time.Time
}

type CustomerCreated struct {
	PartyID   PartyID
	Name      string
	Email     string
	OcurredAt time.Time
}

func (e CustomerCreated) EventName() string {
	return "parties.customer_created"
}

func (e CustomerCreated) OccurredAt() time.Time {
	return e.OcurredAt
}

func (e CustomerCreated) GetPartyID() string {
	return e.PartyID.String()
}

type SupplierCreated struct {
	PartyID   PartyID
	Name      string
	Email     string
	OcurredAt time.Time
}

func (e SupplierCreated) EventName() string {
	return "parties.supplier_created"
}

func (e SupplierCreated) OccurredAt() time.Time {
	return e.OcurredAt
}

func (e SupplierCreated) GetPartyID() string {
	return e.PartyID.String()
}

type PartnerCreated struct {
	PartyID   PartyID
	Name      string
	Email     string
	OcurredAt time.Time
}

func (e PartnerCreated) EventName() string {
	return "parties.partner_created"
}

func (e PartnerCreated) OccurredAt() time.Time {
	return e.OcurredAt
}

func (e PartnerCreated) GetPartyID() string {
	return e.PartyID.String()
}

type EmployeeCreated struct {
	PartyID   PartyID
	Name      string
	Email     string
	Position  string
	OcurredAt time.Time
}

func (e EmployeeCreated) EventName() string {
	return "parties.employee_created"
}

func (e EmployeeCreated) OccurredAt() time.Time {
	return e.OcurredAt
}

func (e EmployeeCreated) GetPartyID() string {
	return e.PartyID.String()
}

type PartyStatusChanged struct {
	PartyID   PartyID
	PartyType PartyType
	OldStatus PartyStatus
	NewStatus PartyStatus
	OcurredAt time.Time
}

func (e PartyStatusChanged) EventName() string {
	return "parties.status_changed"
}

func (e PartyStatusChanged) OccurredAt() time.Time {
	return e.OcurredAt
}

func (e PartyStatusChanged) GetPartyID() string {
	return e.PartyID.String()
}

type ContractAssigned struct {
	PartyID    PartyID
	ContractID string
	OcurredAt  time.Time
}

func (e ContractAssigned) EventName() string {
	return "parties.contract_assigned"
}

func (e ContractAssigned) OccurredAt() time.Time {
	return e.OcurredAt
}

func (e ContractAssigned) GetPartyID() string {
	return e.PartyID.String()
}

type CustomerCreditLimitChanged struct {
	PartyID    PartyID
	OldLimit   float64
	NewLimit   float64
	occurredAt time.Time
}

func (e CustomerCreditLimitChanged) EventName() string {
	return "parties.customer_credit_limit_changed"
}

func (e CustomerCreditLimitChanged) OccurredAt() time.Time {
	return e.occurredAt
}

func (e CustomerCreditLimitChanged) GetPartyID() string {
	return e.PartyID.String()
}

type SupplierRatingUpdated struct {
	PartyID    PartyID
	OldScore   float64
	NewScore   float64
	Level      PerformanceLevel
	occurredAt time.Time
}

func (e SupplierRatingUpdated) EventName() string {
	return "parties.supplier_rating_updated"
}

func (e SupplierRatingUpdated) OccurredAt() time.Time {
	return e.occurredAt
}

func (e SupplierRatingUpdated) GetPartyID() string {
	return e.PartyID.String()
}
