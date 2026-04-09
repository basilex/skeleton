package command

import (
	"context"
	"fmt"

	"github.com/basilex/skeleton/internal/inventory/domain"
)

type ReceiptStockHandler struct {
	stock     domain.StockRepository
	movements domain.StockMovementRepository
}

func NewReceiptStockHandler(stock domain.StockRepository, movements domain.StockMovementRepository) *ReceiptStockHandler {
	return &ReceiptStockHandler{
		stock:     stock,
		movements: movements,
	}
}

type ReceiptStockCommand struct {
	ItemID      string
	WarehouseID string
	Quantity    float64
	ReferenceID string
}

type ReceiptStockResult struct {
	StockID    string
	MovementID string
}

func (h *ReceiptStockHandler) Handle(ctx context.Context, cmd ReceiptStockCommand) (*ReceiptStockResult, error) {
	warehouseID, err := domain.ParseWarehouseID(cmd.WarehouseID)
	if err != nil {
		return nil, fmt.Errorf("parse warehouse ID: %w", err)
	}

	stock, err := h.stock.FindByItemAndWarehouse(ctx, cmd.ItemID, warehouseID)
	if err != nil {
		stock, err = domain.NewStock(cmd.ItemID, warehouseID)
		if err != nil {
			return nil, fmt.Errorf("create stock: %w", err)
		}
	}

	movement, err := domain.NewReceipt(cmd.ItemID, warehouseID, cmd.Quantity, cmd.ReferenceID)
	if err != nil {
		return nil, fmt.Errorf("create receipt movement: %w", err)
	}

	stock.AdjustQuantity(cmd.Quantity, movement.GetID())

	if err := h.movements.Save(ctx, movement); err != nil {
		return nil, fmt.Errorf("save movement: %w", err)
	}

	if err := h.stock.Save(ctx, stock); err != nil {
		return nil, fmt.Errorf("save stock: %w", err)
	}

	return &ReceiptStockResult{
		StockID:    stock.GetID().String(),
		MovementID: movement.GetID().String(),
	}, nil
}
