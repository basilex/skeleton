package command

import (
	"context"
	"fmt"

	"github.com/basilex/skeleton/internal/inventory/domain"
)

type CancelReservationHandler struct {
	stock        domain.StockRepository
	reservations domain.StockReservationRepository
}

func NewCancelReservationHandler(stock domain.StockRepository, reservations domain.StockReservationRepository) *CancelReservationHandler {
	return &CancelReservationHandler{
		stock:        stock,
		reservations: reservations,
	}
}

type CancelReservationCommand struct {
	ReservationID string
}

type CancelReservationResult struct {
	ReservationID string
}

func (h *CancelReservationHandler) Handle(ctx context.Context, cmd CancelReservationCommand) (*CancelReservationResult, error) {
	reservationID, err := domain.ParseStockReservationID(cmd.ReservationID)
	if err != nil {
		return nil, fmt.Errorf("parse reservation ID: %w", err)
	}

	reservation, err := h.reservations.FindByID(ctx, reservationID)
	if err != nil {
		return nil, fmt.Errorf("find reservation: %w", err)
	}

	stockID := reservation.GetStockID()
	stock, err := h.stock.FindByID(ctx, stockID)
	if err != nil {
		return nil, fmt.Errorf("find stock: %w", err)
	}

	if err := reservation.Cancel(); err != nil {
		return nil, fmt.Errorf("cancel reservation: %w", err)
	}

	stock.ReleaseReservation(reservation.GetQuantity())

	if err := h.reservations.Save(ctx, reservation); err != nil {
		return nil, fmt.Errorf("save reservation: %w", err)
	}

	if err := h.stock.Save(ctx, stock); err != nil {
		return nil, fmt.Errorf("save stock: %w", err)
	}

	return &CancelReservationResult{
		ReservationID: reservation.GetID().String(),
	}, nil
}
