package command

import (
	"context"
	"fmt"

	"github.com/basilex/skeleton/internal/inventory/domain"
	"github.com/basilex/skeleton/pkg/eventbus"
)

type AdjustStockHandler struct {
	stock     domain.StockRepository
	movements domain.StockMovementRepository
	bus       eventbus.Bus
}

func NewAdjustStockHandler(stock domain.StockRepository, movements domain.StockMovementRepository, bus eventbus.Bus) *AdjustStockHandler {
	return &AdjustStockHandler{
		stock:     stock,
		movements: movements,
		bus:       bus,
	}
}

type AdjustStockCommand struct {
	StockID     string
	Quantity    float64
	Reason      string
	ReferenceID string
}

type AdjustStockResult struct {
	StockID    string
	MovementID string
}

func (h *AdjustStockHandler) Handle(ctx context.Context, cmd AdjustStockCommand) (*AdjustStockResult, error) {
	stockID, err := domain.ParseStockID(cmd.StockID)
	if err != nil {
		return nil, fmt.Errorf("parse stock ID: %w", err)
	}

	stock, err := h.stock.FindByID(ctx, stockID)
	if err != nil {
		return nil, fmt.Errorf("find stock: %w", err)
	}

	movement, err := domain.NewAdjustment(
		stock.GetItemID(),
		stock.GetWarehouseID(),
		cmd.Quantity,
		cmd.Reason,
	)
	if err != nil {
		return nil, fmt.Errorf("create adjustment: %w", err)
	}

	stock.AdjustQuantity(cmd.Quantity, movement.GetID())

	if err := h.movements.Save(ctx, movement); err != nil {
		return nil, fmt.Errorf("save movement: %w", err)
	}

	if err := h.stock.Save(ctx, stock); err != nil {
		return nil, fmt.Errorf("save stock: %w", err)
	}

	return &AdjustStockResult{
		StockID:    stock.GetID().String(),
		MovementID: movement.GetID().String(),
	}, nil
}
