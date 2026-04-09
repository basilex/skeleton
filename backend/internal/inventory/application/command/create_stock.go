package command

import (
	"context"
	"fmt"

	"github.com/basilex/skeleton/internal/inventory/domain"
)

type CreateStockHandler struct {
	stock domain.StockRepository
}

func NewCreateStockHandler(stock domain.StockRepository) *CreateStockHandler {
	return &CreateStockHandler{
		stock: stock,
	}
}

type CreateStockCommand struct {
	ItemID      string
	WarehouseID string
}

type CreateStockResult struct {
	StockID string
}

func (h *CreateStockHandler) Handle(ctx context.Context, cmd CreateStockCommand) (*CreateStockResult, error) {
	warehouseID, err := domain.ParseWarehouseID(cmd.WarehouseID)
	if err != nil {
		return nil, fmt.Errorf("parse warehouse ID: %w", err)
	}

	stock, err := domain.NewStock(cmd.ItemID, warehouseID)
	if err != nil {
		return nil, fmt.Errorf("create stock: %w", err)
	}

	if err := h.stock.Save(ctx, stock); err != nil {
		return nil, fmt.Errorf("save stock: %w", err)
	}

	return &CreateStockResult{
		StockID: stock.GetID().String(),
	}, nil
}
