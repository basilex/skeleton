package command

import (
	"context"
	"fmt"
	"time"

	"github.com/basilex/skeleton/internal/inventory/domain"
	"github.com/basilex/skeleton/pkg/eventbus"
	"github.com/basilex/skeleton/pkg/transaction"
)

type ReserveStockHandler struct {
	stock        domain.StockRepository
	reservations domain.StockReservationRepository
	bus          eventbus.Bus
	txManager    transaction.Manager
}

func NewReserveStockHandler(
	stock domain.StockRepository,
	reservations domain.StockReservationRepository,
	bus eventbus.Bus,
	txManager transaction.Manager,
) *ReserveStockHandler {
	return &ReserveStockHandler{
		stock:        stock,
		reservations: reservations,
		bus:          bus,
		txManager:    txManager,
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
	var result *ReserveStockResult

	err := h.txManager.Execute(ctx, func(ctx context.Context) error {
		// Parse stock ID
		stockID, err := domain.ParseStockID(cmd.StockID)
		if err != nil {
			return fmt.Errorf("parse stock ID: %w", err)
		}

		// Load stock
		stock, err := h.stock.FindByID(ctx, stockID)
		if err != nil {
			return fmt.Errorf("find stock: %w", err)
		}

		// Create reservation first to get reservation ID
		reservation, err := domain.NewStockReservation(stockID, cmd.OrderID, cmd.Quantity, cmd.ExpiresAt)
		if err != nil {
			return fmt.Errorf("create reservation: %w", err)
		}

		// Reserve quantity with reservation ID
		if err := stock.Reserve(cmd.Quantity, reservation.GetID()); err != nil {
			return fmt.Errorf("reserve stock: %w", err)
		}

		// Save both within transaction
		if err := h.reservations.Save(ctx, reservation); err != nil {
			return fmt.Errorf("save reservation: %w", err)
		}

		if err := h.stock.Save(ctx, stock); err != nil {
			return fmt.Errorf("save stock: %w", err)
		}

		// Publish domain events from stock
		for _, event := range stock.PullEvents() {
			if err := h.bus.Publish(ctx, event); err != nil {
				return fmt.Errorf("publish stock event: %w", err)
			}
		}

		result = &ReserveStockResult{
			ReservationID: reservation.GetID().String(),
		}

		return nil
	})

	return result, err
}
