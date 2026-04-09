package domain

import (
	"fmt"

	"github.com/basilex/skeleton/pkg/uuid"
)

type ContractID uuid.UUID

func NewContractID() ContractID {
	return ContractID(uuid.NewV7())
}

func ParseContractID(s string) (ContractID, error) {
	u, err := uuid.Parse(s)
	if err != nil {
		return ContractID{}, fmt.Errorf("invalid contract id: %w", err)
	}
	return ContractID(u), nil
}

func (id ContractID) String() string {
	return uuid.UUID(id).String()
}

type ContractType string

const (
	ContractTypeSupply      ContractType = "supply"
	ContractTypeService     ContractType = "service"
	ContractTypeEmployment  ContractType = "employment"
	ContractTypePartnership ContractType = "partnership"
	ContractTypeLease       ContractType = "lease"
	ContractTypeLicense     ContractType = "license"
)

func (ct ContractType) String() string {
	return string(ct)
}

func ParseContractType(s string) (ContractType, error) {
	switch ContractType(s) {
	case ContractTypeSupply, ContractTypeService, ContractTypeEmployment,
		ContractTypePartnership, ContractTypeLease, ContractTypeLicense:
		return ContractType(s), nil
	default:
		return "", fmt.Errorf("invalid contract type: %s", s)
	}
}

type ContractStatus string

const (
	ContractStatusDraft           ContractStatus = "draft"
	ContractStatusPendingApproval ContractStatus = "pending_approval"
	ContractStatusActive          ContractStatus = "active"
	ContractStatusExpired         ContractStatus = "expired"
	ContractStatusTerminated      ContractStatus = "terminated"
)

func (cs ContractStatus) String() string {
	return string(cs)
}

func ParseContractStatus(s string) (ContractStatus, error) {
	switch ContractStatus(s) {
	case ContractStatusDraft, ContractStatusPendingApproval,
		ContractStatusActive, ContractStatusExpired, ContractStatusTerminated:
		return ContractStatus(s), nil
	default:
		return "", fmt.Errorf("invalid contract status: %s", s)
	}
}

type PaymentType string

const (
	PaymentTypePrepaid  PaymentType = "prepaid"
	PaymentTypePostpaid PaymentType = "postpaid"
	PaymentTypeCredit   PaymentType = "credit"
)

func (pt PaymentType) String() string {
	return string(pt)
}

func ParsePaymentType(s string) (PaymentType, error) {
	switch PaymentType(s) {
	case PaymentTypePrepaid, PaymentTypePostpaid, PaymentTypeCredit:
		return PaymentType(s), nil
	default:
		return "", fmt.Errorf("invalid payment type: %s", s)
	}
}
