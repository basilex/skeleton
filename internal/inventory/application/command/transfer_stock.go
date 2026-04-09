package command

import (
	"context"
	"fmt"

	"github.com/basilex/skeleton/internal/inventory/domain"
)

type TransferStockHandler struct {
	stock     domain.StockRepository
	movements domain.StockMovementRepository
}

func NewTransferStockHandler(stock domain.StockRepository, movements domain.StockMovementRepository) *TransferStockHandler {
	return &TransferStockHandler{
		stock:     stock,
		movements: movements,
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
	fromWarehouseID, err := domain.ParseWarehouseID(cmd.FromWarehouse)
	if err != nil {
		return nil, fmt.Errorf("parse from warehouse ID: %w", err)
	}

	toWarehouseID, err := domain.ParseWarehouseID(cmd.ToWarehouse)
	if err != nil {
		return nil, fmt.Errorf("parse to warehouse ID: %w", err)
	}

	fromStock, err := h.stock.FindByItemAndWarehouse(ctx, cmd.ItemID, fromWarehouseID)
	if err != nil {
		return nil, fmt.Errorf("find from stock: %w", err)
	}

	if !fromStock.IsAvailable(cmd.Quantity) {
		return nil, domain.ErrInsufficientStock
	}

	toStock, err := h.stock.FindByItemAndWarehouse(ctx, cmd.ItemID, toWarehouseID)
	if err != nil {
		toStock, err = domain.NewStock(cmd.ItemID, toWarehouseID)
		if err != nil {
			return nil, fmt.Errorf("create to stock: %w", err)
		}
	}

	movement, err := domain.NewTransfer(cmd.ItemID, fromWarehouseID, toWarehouseID, cmd.Quantity)
	if err != nil {
		return nil, fmt.Errorf("create transfer movement: %w", err)
	}

	fromStock.AdjustQuantity(-cmd.Quantity, movement.GetID())
	toStock.AdjustQuantity(cmd.Quantity, movement.GetID())

	if err := h.movements.Save(ctx, movement); err != nil {
		return nil, fmt.Errorf("save movement: %w", err)
	}

	if err := h.stock.Save(ctx, fromStock); err != nil {
		return nil, fmt.Errorf("save from stock: %w", err)
	}

	if err := h.stock.Save(ctx, toStock); err != nil {
		return nil, fmt.Errorf("save to stock: %w", err)
	}

	return &TransferStockResult{
		MovementID: movement.GetID().String(),
	}, nil
}
