package query

import (
	"context"
	"fmt"

	"github.com/basilex/skeleton/internal/inventory/domain"
)

type ListStockHandler struct {
	stock domain.StockRepository
}

func NewListStockHandler(stock domain.StockRepository) *ListStockHandler {
	return &ListStockHandler{
		stock: stock,
	}
}

type ListStockQuery struct {
	ItemID      *string
	WarehouseID *string
	Available   *bool
	Cursor      string
	Limit       int
}

type ListStockResult struct {
	Stock      []StockDTO `json:"stock"`
	NextCursor string     `json:"next_cursor"`
}

func (h *ListStockHandler) Handle(ctx context.Context, query ListStockQuery) (*ListStockResult, error) {
	var warehouseID *domain.WarehouseID
	if query.WarehouseID != nil {
		id, err := domain.ParseWarehouseID(*query.WarehouseID)
		if err != nil {
			return nil, fmt.Errorf("parse warehouse ID: %w", err)
		}
		warehouseID = &id
	}

	filter := domain.StockFilter{
		ItemID:      query.ItemID,
		WarehouseID: warehouseID,
		Available:   query.Available,
		Cursor:      query.Cursor,
		Limit:       query.Limit,
	}

	result, err := h.stock.FindAll(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("find stock: %w", err)
	}

	stockItems := make([]StockDTO, 0, len(result.Items))
	for _, stock := range result.Items {
		stockItems = append(stockItems, *toStockDTO(stock))
	}

	return &ListStockResult{
		Stock:      stockItems,
		NextCursor: result.NextCursor,
	}, nil
}
