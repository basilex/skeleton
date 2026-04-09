package domain

import "errors"

var (
	ErrAccountNotFound       = errors.New("account not found")
	ErrAccountAlreadyExists  = errors.New("account already exists")
	ErrAccountInactive       = errors.New("account is inactive")
	ErrInvalidAccountType    = errors.New("invalid account type")
	ErrInvalidAccountCode    = errors.New("invalid account code")
	ErrTransactionNotFound   = errors.New("transaction not found")
	ErrTransactionUnbalanced = errors.New("transaction is not balanced")
	ErrInsufficientFunds     = errors.New("insufficient funds")
	ErrInvoiceNotFound       = errors.New("invoice not found")
	ErrInvoiceAlreadyPaid    = errors.New("invoice already paid")
	ErrPayableNotFound       = errors.New("payable not found")
	ErrReceivableNotFound    = errors.New("receivable not found")
	ErrDifferentCurrencies   = errors.New("cannot operate on different currencies")
	ErrPaymentNotFound       = errors.New("payment not found")
)
