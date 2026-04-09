package query

import (
	"context"
	"fmt"

	"github.com/basilex/skeleton/internal/inventory/domain"
)

type GetReservationHandler struct {
	reservations domain.StockReservationRepository
}

func NewGetReservationHandler(reservations domain.StockReservationRepository) *GetReservationHandler {
	return &GetReservationHandler{
		reservations: reservations,
	}
}

type GetReservationQuery struct {
	ReservationID string
}

type StockReservationDTO struct {
	ID          string  `json:"id"`
	StockID     string  `json:"stock_id"`
	OrderID     string  `json:"order_id"`
	Quantity    float64 `json:"quantity"`
	Status      string  `json:"status"`
	ReservedAt  string  `json:"reserved_at"`
	ExpiresAt   *string `json:"expires_at,omitempty"`
	FulfilledAt *string `json:"fulfilled_at,omitempty"`
	CancelledAt *string `json:"cancelled_at,omitempty"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

func (h *GetReservationHandler) Handle(ctx context.Context, query GetReservationQuery) (*StockReservationDTO, error) {
	reservationID, err := domain.ParseStockReservationID(query.ReservationID)
	if err != nil {
		return nil, fmt.Errorf("parse reservation ID: %w", err)
	}

	reservation, err := h.reservations.FindByID(ctx, reservationID)
	if err != nil {
		return nil, fmt.Errorf("find reservation: %w", err)
	}

	return toStockReservationDTO(reservation), nil
}

func toStockReservationDTO(reservation *domain.StockReservation) *StockReservationDTO {
	dto := &StockReservationDTO{
		ID:         reservation.GetID().String(),
		StockID:    reservation.GetStockID().String(),
		OrderID:    reservation.GetOrderID(),
		Quantity:   reservation.GetQuantity(),
		Status:     reservation.GetStatus().String(),
		ReservedAt: reservation.GetReservedAt().Format("2006-01-02T15:04:05Z07:00"),
		CreatedAt:  reservation.GetCreatedAt().Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:  reservation.GetUpdatedAt().Format("2006-01-02T15:04:05Z07:00"),
	}

	if reservation.GetExpiresAt() != nil {
		expires := reservation.GetExpiresAt().Format("2006-01-02T15:04:05Z07:00")
		dto.ExpiresAt = &expires
	}

	if reservation.GetFulfilledAt() != nil {
		fulfilled := reservation.GetFulfilledAt().Format("2006-01-02T15:04:05Z07:00")
		dto.FulfilledAt = &fulfilled
	}

	if reservation.GetCancelledAt() != nil {
		cancelled := reservation.GetCancelledAt().Format("2006-01-02T15:04:05Z07:00")
		dto.CancelledAt = &cancelled
	}

	return dto
}
