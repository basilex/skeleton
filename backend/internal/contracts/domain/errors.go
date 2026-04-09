package domain

import "errors"

var (
	ErrContractNotFound      = errors.New("contract not found")
	ErrContractAlreadyExists = errors.New("contract already exists")
	ErrInvalidContractType   = errors.New("invalid contract type")
	ErrInvalidContractStatus = errors.New("invalid contract status")
	ErrContractExpired       = errors.New("contract is expired")
	ErrContractTerminated    = errors.New("contract is terminated")
	ErrInvalidDateRange      = errors.New("invalid date range")
	ErrInvalidPaymentTerms   = errors.New("invalid payment terms")
	ErrPartyNotActive        = errors.New("party is not active")
)
