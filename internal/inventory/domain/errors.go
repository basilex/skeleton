package domain

import (
	"errors"
)

var (
	ErrWarehouseNotFound     = errors.New("warehouse not found")
	ErrWarehouseInactive     = errors.New("warehouse is inactive")
	ErrStockNotFound         = errors.New("stock not found")
	ErrInsufficientStock     = errors.New("insufficient stock")
	ErrReservationNotFound   = errors.New("reservation not found")
	ErrReservationNotActive  = errors.New("reservation is not active")
	ErrReservationExpired    = errors.New("reservation has expired")
	ErrInvalidMovementType   = errors.New("invalid movement type")
	ErrInvalidQuantity       = errors.New("invalid quantity")
	ErrWarehouseNameEmpty    = errors.New("warehouse name cannot be empty")
	ErrSameWarehouseTransfer = errors.New("cannot transfer to the same warehouse")
	ErrStockAlreadyExists    = errors.New("stock already exists for this item and warehouse")
)
