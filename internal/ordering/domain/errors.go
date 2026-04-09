package domain

import "errors"

var (
	ErrOrderNotFound        = errors.New("order not found")
	ErrOrderAlreadyExists   = errors.New("order already exists")
	ErrInvalidOrderStatus   = errors.New("invalid order status")
	ErrOrderCannotCancel    = errors.New("order cannot be cancelled")
	ErrOrderCannotComplete  = errors.New("order cannot be completed")
	ErrQuoteNotFound        = errors.New("quote not found")
	ErrInvalidQuoteStatus   = errors.New("invalid quote status")
	ErrQuoteAlreadyAccepted = errors.New("quote already accepted")
	ErrQuoteExpired         = errors.New("quote has expired")
	ErrOrderLineNotFound    = errors.New("order line not found")
	ErrInvalidQuantity      = errors.New("invalid quantity")
)
