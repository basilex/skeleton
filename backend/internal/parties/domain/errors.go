package domain

import "errors"

var (
	ErrPartyNotFound      = errors.New("party not found")
	ErrPartyAlreadyExists = errors.New("party already exists")
	ErrCustomerNotFound   = errors.New("customer not found")
	ErrSupplierNotFound   = errors.New("supplier not found")
	ErrPartnerNotFound    = errors.New("partner not found")
	ErrEmployeeNotFound   = errors.New("employee not found")
	ErrPartyInactive      = errors.New("party is inactive")
	ErrPartyBlacklisted   = errors.New("party is blacklisted")
	ErrInvalidPartyType   = errors.New("invalid party type")
	ErrInvalidContactInfo = errors.New("invalid contact information")
	ErrInvalidBankAccount = errors.New("invalid bank account")
	ErrInvalidTaxID       = errors.New("invalid tax identification number")
	ErrDuplicateContract  = errors.New("contract already assigned")
	ErrContractNotFound   = errors.New("contract not found")
)
