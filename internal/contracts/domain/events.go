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
	OcurredAt    time.Time
}

func (e ContractCreated) EventName() string {
	return "contracts.contract_created"
}

func (e ContractCreated) OccurredAt() time.Time {
	return e.OcurredAt
}

func (e ContractCreated) GetContractID() string {
	return e.ContractID.String()
}

type ContractActivated struct {
	ContractID   ContractID
	ContractType ContractType
	OcurredAt    time.Time
}

func (e ContractActivated) EventName() string {
	return "contracts.contract_activated"
}

func (e ContractActivated) OccurredAt() time.Time {
	return e.OcurredAt
}

func (e ContractActivated) GetContractID() string {
	return e.ContractID.String()
}

type ContractTerminated struct {
	ContractID ContractID
	Reason     string
	OcurredAt  time.Time
}

func (e ContractTerminated) EventName() string {
	return "contracts.contract_terminated"
}

func (e ContractTerminated) OccurredAt() time.Time {
	return e.OcurredAt
}

func (e ContractTerminated) GetContractID() string {
	return e.ContractID.String()
}
