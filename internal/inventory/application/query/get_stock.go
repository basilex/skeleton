package query

import (
	"context"
	"fmt"

	"github.com/basilex/skeleton/internal/inventory/domain"
)

type GetStockHandler struct {
	stock domain.StockRepository
}

func NewGetStockHandler(stock domain.StockRepository) *GetStockHandler {
	return &GetStockHandler{
		stock: stock,
	}
}

type GetStockQuery struct {
	StockID string
}

type StockDTO struct {
	ID              string  `json:"id"`
	ItemID          string  `json:"item_id"`
	WarehouseID     string  `json:"warehouse_id"`
	Quantity        float64 `json:"quantity"`
	ReservedQty     float64 `json:"reserved_qty"`
	AvailableQty    float64 `json:"available_qty"`
	ReorderPoint    float64 `json:"reorder_point"`
	ReorderQuantity float64 `json:"reorder_quantity"`
	LastMovementID  string  `json:"last_movement_id"`
	CreatedAt       string  `json:"created_at"`
	UpdatedAt       string  `json:"updated_at"`
}

func (h *GetStockHandler) Handle(ctx context.Context, query GetStockQuery) (*StockDTO, error) {
	stockID, err := domain.ParseStockID(query.StockID)
	if err != nil {
		return nil, fmt.Errorf("parse stock ID: %w", err)
	}

	stock, err := h.stock.FindByID(ctx, stockID)
	if err != nil {
		return nil, fmt.Errorf("find stock: %w", err)
	}

	return toStockDTO(stock), nil
}

func toStockDTO(stock *domain.Stock) *StockDTO {
	return &StockDTO{
		ID:              stock.GetID().String(),
		ItemID:          stock.GetItemID(),
		WarehouseID:     stock.GetWarehouseID().String(),
		Quantity:        stock.GetQuantity(),
		ReservedQty:     stock.GetReservedQty(),
		AvailableQty:    stock.GetAvailableQty(),
		ReorderPoint:    stock.GetReorderPoint(),
		ReorderQuantity: stock.GetReorderQuantity(),
		LastMovementID:  stock.GetLastMovementID().String(),
		CreatedAt:       stock.GetCreatedAt().Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:       stock.GetUpdatedAt().Format("2006-01-02T15:04:05Z07:00"),
	}
}
