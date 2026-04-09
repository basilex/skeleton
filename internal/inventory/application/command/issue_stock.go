package command

import (
	"context"
	"fmt"

	"github.com/basilex/skeleton/internal/inventory/domain"
)

type IssueStockHandler struct {
	stock     domain.StockRepository
	movements domain.StockMovementRepository
}

func NewIssueStockHandler(stock domain.StockRepository, movements domain.StockMovementRepository) *IssueStockHandler {
	return &IssueStockHandler{
		stock:     stock,
		movements: movements,
	}
}

type IssueStockCommand struct {
	ItemID      string
	WarehouseID string
	Quantity    float64
	OrderID     string
}

type IssueStockResult struct {
	StockID    string
	MovementID string
}

func (h *IssueStockHandler) Handle(ctx context.Context, cmd IssueStockCommand) (*IssueStockResult, error) {
	warehouseID, err := domain.ParseWarehouseID(cmd.WarehouseID)
	if err != nil {
		return nil, fmt.Errorf("parse warehouse ID: %w", err)
	}

	stock, err := h.stock.FindByItemAndWarehouse(ctx, cmd.ItemID, warehouseID)
	if err != nil {
		return nil, fmt.Errorf("find stock: %w", err)
	}

	if !stock.IsAvailable(cmd.Quantity) {
		return nil, domain.ErrInsufficientStock
	}

	movement, err := domain.NewIssue(cmd.ItemID, warehouseID, cmd.Quantity, cmd.OrderID)
	if err != nil {
		return nil, fmt.Errorf("create issue movement: %w", err)
	}

	stock.AdjustQuantity(-cmd.Quantity, movement.GetID())

	if err := h.movements.Save(ctx, movement); err != nil {
		return nil, fmt.Errorf("save movement: %w", err)
	}

	if err := h.stock.Save(ctx, stock); err != nil {
		return nil, fmt.Errorf("save stock: %w", err)
	}

	return &IssueStockResult{
		StockID:    stock.GetID().String(),
		MovementID: movement.GetID().String(),
	}, nil
}
