package query

import (
	"context"
	"fmt"

	"github.com/basilex/skeleton/internal/inventory/domain"
)

type ListWarehousesHandler struct {
	warehouses domain.WarehouseRepository
}

func NewListWarehousesHandler(warehouses domain.WarehouseRepository) *ListWarehousesHandler {
	return &ListWarehousesHandler{
		warehouses: warehouses,
	}
}

type ListWarehousesQuery struct {
	Status *string
	Code   *string
	Cursor string
	Limit  int
}

type ListWarehousesResult struct {
	Warehouses []WarehouseDTO `json:"warehouses"`
	NextCursor string         `json:"next_cursor"`
}

func (h *ListWarehousesHandler) Handle(ctx context.Context, query ListWarehousesQuery) (*ListWarehousesResult, error) {
	var status *domain.WarehouseStatus
	if query.Status != nil {
		s := domain.WarehouseStatus(*query.Status)
		status = &s
	}

	filter := domain.WarehouseFilter{
		Status: status,
		Code:   query.Code,
		Cursor: query.Cursor,
		Limit:  query.Limit,
	}

	result, err := h.warehouses.FindAll(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("find warehouses: %w", err)
	}

	warehouses := make([]WarehouseDTO, 0, len(result.Items))
	for _, warehouse := range result.Items {
		warehouses = append(warehouses, *toWarehouseDTO(warehouse))
	}

	return &ListWarehousesResult{
		Warehouses: warehouses,
		NextCursor: result.NextCursor,
	}, nil
}
