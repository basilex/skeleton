package command

import (
	"context"
	"fmt"

	"github.com/basilex/skeleton/internal/inventory/domain"
	"github.com/basilex/skeleton/pkg/eventbus"
	"github.com/basilex/skeleton/pkg/transaction"
)

type TransferStockHandler struct {
	stock     domain.StockRepository
	movements domain.StockMovementRepository
	bus       eventbus.Bus
	txManager transaction.Manager
}

func NewTransferStockHandler(
	stock domain.StockRepository,
	movements domain.StockMovementRepository,
	bus eventbus.Bus,
	txManager transaction.Manager,
) *TransferStockHandler {
	return &TransferStockHandler{
		stock:     stock,
		movements: movements,
		bus:       bus,
		txManager: txManager,
	}
}

type TransferStockCommand struct {
	ItemID        string
	FromWarehouse string
	ToWarehouse   string
	Quantity      float64
}

type TransferStockResult struct {
	MovementID string
}

func (h *TransferStockHandler) Handle(ctx context.Context, cmd TransferStockCommand) (*TransferStockResult, error) {
	var result *TransferStockResult

	err := h.txManager.Execute(ctx, func(ctx context.Context) error {
		// Parse IDs
		fromWarehouseID, err := domain.ParseWarehouseID(cmd.FromWarehouse)
		if err != nil {
			return fmt.Errorf("parse from warehouse ID: %w", err)
		}

		toWarehouseID, err := domain.ParseWarehouseID(cmd.ToWarehouse)
		if err != nil {
			return fmt.Errorf("parse to warehouse ID: %w", err)
		}

		// Load from stock
		fromStock, err := h.stock.FindByItemAndWarehouse(ctx, cmd.ItemID, fromWarehouseID)
		if err != nil {
			return fmt.Errorf("find from stock: %w", err)
		}

		if !fromStock.IsAvailable(cmd.Quantity) {
			return domain.ErrInsufficientStock
		}

		// Load or create to stock
		toStock, err := h.stock.FindByItemAndWarehouse(ctx, cmd.ItemID, toWarehouseID)
		if err != nil {
			toStock, err = domain.NewStock(cmd.ItemID, toWarehouseID)
			if err != nil {
				return fmt.Errorf("create to stock: %w", err)
			}
		}

		// Create transfer movement
		movement, err := domain.NewTransfer(cmd.ItemID, fromWarehouseID, toWarehouseID, cmd.Quantity)
		if err != nil {
			return fmt.Errorf("create transfer movement: %w", err)
		}

		// Adjust quantities
		fromStock.AdjustQuantity(-cmd.Quantity, movement.GetID())
		toStock.AdjustQuantity(cmd.Quantity, movement.GetID())

		// Save all within transaction
		if err := h.movements.Save(ctx, movement); err != nil {
			return fmt.Errorf("save movement: %w", err)
		}

		if err := h.stock.Save(ctx, fromStock); err != nil {
			return fmt.Errorf("save from stock: %w", err)
		}

		if err := h.stock.Save(ctx, toStock); err != nil {
			return fmt.Errorf("save to stock: %w", err)
		}

		// Publish domain events from both stocks
		for _, event := range fromStock.PullEvents() {
			if err := h.bus.Publish(ctx, event); err != nil {
				return fmt.Errorf("publish from stock event: %w", err)
			}
		}

		for _, event := range toStock.PullEvents() {
			if err := h.bus.Publish(ctx, event); err != nil {
				return fmt.Errorf("publish to stock event: %w", err)
			}
		}

		result = &TransferStockResult{
			MovementID: movement.GetID().String(),
		}

		return nil
	})

	return result, err
}
