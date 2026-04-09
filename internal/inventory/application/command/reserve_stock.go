package command

import (
	"context"
	"fmt"
	"time"

	"github.com/basilex/skeleton/internal/inventory/domain"
)

type ReserveStockHandler struct {
	stock        domain.StockRepository
	reservations domain.StockReservationRepository
}

func NewReserveStockHandler(stock domain.StockRepository, reservations domain.StockReservationRepository) *ReserveStockHandler {
	return &ReserveStockHandler{
		stock:        stock,
		reservations: reservations,
	}
}

type ReserveStockCommand struct {
	StockID   string
	OrderID   string
	Quantity  float64
	ExpiresAt *time.Time
}

type ReserveStockResult struct {
	ReservationID string
}

func (h *ReserveStockHandler) Handle(ctx context.Context, cmd ReserveStockCommand) (*ReserveStockResult, error) {
	stockID, err := domain.ParseStockID(cmd.StockID)
	if err != nil {
		return nil, fmt.Errorf("parse stock ID: %w", err)
	}

	stock, err := h.stock.FindByID(ctx, stockID)
	if err != nil {
		return nil, fmt.Errorf("find stock: %w", err)
	}

	if err := stock.Reserve(cmd.Quantity); err != nil {
		return nil, fmt.Errorf("reserve stock: %w", err)
	}

	reservation, err := domain.NewStockReservation(stockID, cmd.OrderID, cmd.Quantity, cmd.ExpiresAt)
	if err != nil {
		return nil, fmt.Errorf("create reservation: %w", err)
	}

	if err := h.reservations.Save(ctx, reservation); err != nil {
		return nil, fmt.Errorf("save reservation: %w", err)
	}

	if err := h.stock.Save(ctx, stock); err != nil {
		return nil, fmt.Errorf("save stock: %w", err)
	}

	return &ReserveStockResult{
		ReservationID: reservation.GetID().String(),
	}, nil
}
