package command

import (
	"context"
	"fmt"

	"github.com/basilex/skeleton/internal/inventory/domain"
)

type FulfillReservationHandler struct {
	stock        domain.StockRepository
	reservations domain.StockReservationRepository
}

func NewFulfillReservationHandler(stock domain.StockRepository, reservations domain.StockReservationRepository) *FulfillReservationHandler {
	return &FulfillReservationHandler{
		stock:        stock,
		reservations: reservations,
	}
}

type FulfillReservationCommand struct {
	ReservationID string
}

type FulfillReservationResult struct {
	ReservationID string
}

func (h *FulfillReservationHandler) Handle(ctx context.Context, cmd FulfillReservationCommand) (*FulfillReservationResult, error) {
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

	if err := reservation.Fulfill(); err != nil {
		return nil, fmt.Errorf("fulfill reservation: %w", err)
	}

	stock.FulfillReservation(reservation.GetQuantity())

	if err := h.reservations.Save(ctx, reservation); err != nil {
		return nil, fmt.Errorf("save reservation: %w", err)
	}

	if err := h.stock.Save(ctx, stock); err != nil {
		return nil, fmt.Errorf("save stock: %w", err)
	}

	return &FulfillReservationResult{
		ReservationID: reservation.GetID().String(),
	}, nil
}
