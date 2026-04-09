package domain

import (
	"errors"
	"time"
)

type StockReservation struct {
	id          StockReservationID
	stockID     StockID
	orderID     string
	quantity    float64
	status      ReservationStatus
	reservedAt  time.Time
	expiresAt   *time.Time
	fulfilledAt *time.Time
	cancelledAt *time.Time
	createdAt   time.Time
	updatedAt   time.Time
}

func NewStockReservation(
	stockID StockID,
	orderID string,
	quantity float64,
	expiresAt *time.Time,
) (*StockReservation, error) {
	if quantity <= 0 {
		return nil, ErrInvalidQuantity
	}
	if orderID == "" {
		return nil, errors.New("order ID cannot be empty")
	}

	now := time.Now()
	return &StockReservation{
		id:         NewStockReservationID(),
		stockID:    stockID,
		orderID:    orderID,
		quantity:   quantity,
		status:     ReservationStatusActive,
		reservedAt: now,
		expiresAt:  expiresAt,
		createdAt:  now,
		updatedAt:  now,
	}, nil
}

func RestoreStockReservation(
	id StockReservationID,
	stockID StockID,
	orderID string,
	quantity float64,
	status ReservationStatus,
	reservedAt time.Time,
	expiresAt *time.Time,
	fulfilledAt *time.Time,
	cancelledAt *time.Time,
	createdAt time.Time,
	updatedAt time.Time,
) *StockReservation {
	return &StockReservation{
		id:          id,
		stockID:     stockID,
		orderID:     orderID,
		quantity:    quantity,
		status:      status,
		reservedAt:  reservedAt,
		expiresAt:   expiresAt,
		fulfilledAt: fulfilledAt,
		cancelledAt: cancelledAt,
		createdAt:   createdAt,
		updatedAt:   updatedAt,
	}
}

func (r *StockReservation) GetID() StockReservationID {
	return r.id
}

func (r *StockReservation) GetStockID() StockID {
	return r.stockID
}

func (r *StockReservation) GetOrderID() string {
	return r.orderID
}

func (r *StockReservation) GetQuantity() float64 {
	return r.quantity
}

func (r *StockReservation) GetStatus() ReservationStatus {
	return r.status
}

func (r *StockReservation) GetReservedAt() time.Time {
	return r.reservedAt
}

func (r *StockReservation) GetExpiresAt() *time.Time {
	return r.expiresAt
}

func (r *StockReservation) GetFulfilledAt() *time.Time {
	return r.fulfilledAt
}

func (r *StockReservation) GetCancelledAt() *time.Time {
	return r.cancelledAt
}

func (r *StockReservation) GetCreatedAt() time.Time {
	return r.createdAt
}

func (r *StockReservation) GetUpdatedAt() time.Time {
	return r.updatedAt
}

func (r *StockReservation) Fulfill() error {
	if r.status != ReservationStatusActive {
		return ErrReservationNotActive
	}

	now := time.Now()
	r.status = ReservationStatusFulfilled
	r.fulfilledAt = &now
	r.updatedAt = now
	return nil
}

func (r *StockReservation) Cancel() error {
	if r.status != ReservationStatusActive {
		return ErrReservationNotActive
	}

	now := time.Now()
	r.status = ReservationStatusCancelled
	r.cancelledAt = &now
	r.updatedAt = now
	return nil
}

func (r *StockReservation) IsExpired() bool {
	if r.expiresAt == nil {
		return false
	}
	return time.Now().After(*r.expiresAt)
}

func (r *StockReservation) IsActive() bool {
	return r.status == ReservationStatusActive && !r.IsExpired()
}
