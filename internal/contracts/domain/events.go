package domain

import "time"

type DomainEvent interface {
	EventName() string
	OccurredAt() time.Time
}

type ContractCreated struct {
	ContractID   ContractID
	PartyID      string
	ContractType ContractType
	occurredAt   time.Time
}

func (e ContractCreated) EventName() string {
	return "contracts.contract_created"
}

func (e ContractCreated) OccurredAt() time.Time {
	return e.occurredAt
}

func (e ContractCreated) GetContractID() string {
	return e.ContractID.String()
}

type ContractActivated struct {
	ContractID   ContractID
	ContractType ContractType
	occurredAt   time.Time
}

func (e ContractActivated) EventName() string {
	return "contracts.contract_activated"
}

func (e ContractActivated) OccurredAt() time.Time {
	return e.occurredAt
}

func (e ContractActivated) GetContractID() string {
	return e.ContractID.String()
}

type ContractTerminated struct {
	ContractID ContractID
	Reason     string
	occurredAt time.Time
}

func (e ContractTerminated) EventName() string {
	return "contracts.contract_terminated"
}

func (e ContractTerminated) OccurredAt() time.Time {
	return e.occurredAt
}

func (e ContractTerminated) GetContractID() string {
	return e.ContractID.String()
}

type ContractExpired struct {
	ContractID ContractID
	occurredAt time.Time
}

func (e ContractExpired) EventName() string {
	return "contracts.contract_expired"
}

func (e ContractExpired) OccurredAt() time.Time {
	return e.occurredAt
}

func (e ContractExpired) GetContractID() string {
	return e.ContractID.String()
}

type ContractRenewed struct {
	ContractID   ContractID
	OldEndDate   time.Time
	NewEndDate   time.Time
	RenewalCount int
	occurredAt   time.Time
}

func (e ContractRenewed) EventName() string {
	return "contracts.contract_renewed"
}

func (e ContractRenewed) OccurredAt() time.Time {
	return e.occurredAt
}

func (e ContractRenewed) GetContractID() string {
	return e.ContractID.String()
}

type ContractAutoRenewalEnabled struct {
	ContractID        ContractID
	RenewalPeriodDays int
	MaxRenewals       int
	occurredAt        time.Time
}

func (e ContractAutoRenewalEnabled) EventName() string {
	return "contracts.contract_auto_renewal_enabled"
}

func (e ContractAutoRenewalEnabled) OccurredAt() time.Time {
	return e.occurredAt
}

func (e ContractAutoRenewalEnabled) GetContractID() string {
	return e.ContractID.String()
}

type ContractAutoRenewalDisabled struct {
	ContractID ContractID
	occurredAt time.Time
}

func (e ContractAutoRenewalDisabled) EventName() string {
	return "contracts.contract_auto_renewal_disabled"
}

func (e ContractAutoRenewalDisabled) OccurredAt() time.Time {
	return e.occurredAt
}

func (e ContractAutoRenewalDisabled) GetContractID() string {
	return e.ContractID.String()
}

type ContractAmended struct {
	ContractID  ContractID
	AmendmentID string
	Version     int
	Description string
	AmendedBy   string
	occurredAt  time.Time
}

func (e ContractAmended) EventName() string {
	return "contracts.contract_amended"
}

func (e ContractAmended) OccurredAt() time.Time {
	return e.occurredAt
}

func (e ContractAmended) GetContractID() string {
	return e.ContractID.String()
}
