package domain

import (
	"errors"
	"fmt"
)

var (
	ErrInvoiceNotFound         = errors.New("invoice not found")
	ErrInvoiceAlreadySent      = errors.New("invoice already sent")
	ErrInvoiceAlreadyPaid      = errors.New("invoice already paid")
	ErrInvoiceAlreadyCancelled = errors.New("invoice already cancelled")
	ErrInvalidInvoiceStatus    = errors.New("invalid invoice status")
	ErrEmptyInvoiceNumber      = errors.New("invoice number cannot be empty")
	ErrEmptyCustomerID         = errors.New("customer ID cannot be empty")
	ErrInvalidAmount           = errors.New("invalid amount")
	ErrInvalidDueDate          = errors.New("invalid due date")
	ErrInvoiceLineNotFound     = errors.New("invoice line not found")
	ErrInvalidLineQuantity     = errors.New("invalid line quantity")
	ErrInvalidLinePrice        = errors.New("invalid line price")
	ErrPaymentNotFound         = errors.New("payment not found")
	ErrPaymentExceedsAmount    = errors.New("payment amount exceeds invoice total")
	ErrInvalidPaymentMethod    = errors.New("invalid payment method")
)

type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error: %s - %s", e.Field, e.Message)
}

func NewValidationError(field, message string) *ValidationError {
	return &ValidationError{Field: field, Message: message}
}
