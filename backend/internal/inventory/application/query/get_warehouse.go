package query

import (
	"context"
	"fmt"

	"github.com/basilex/skeleton/internal/inventory/domain"
)

type GetWarehouseHandler struct {
	warehouses domain.WarehouseRepository
}

func NewGetWarehouseHandler(warehouses domain.WarehouseRepository) *GetWarehouseHandler {
	return &GetWarehouseHandler{
		warehouses: warehouses,
	}
}

type GetWarehouseQuery struct {
	WarehouseID string
}

type WarehouseDTO struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	Code      string            `json:"code"`
	Location  string            `json:"location"`
	Capacity  float64           `json:"capacity"`
	Status    string            `json:"status"`
	Metadata  map[string]string `json:"metadata"`
	CreatedAt string            `json:"created_at"`
	UpdatedAt string            `json:"updated_at"`
}

func (h *GetWarehouseHandler) Handle(ctx context.Context, query GetWarehouseQuery) (*WarehouseDTO, error) {
	warehouseID, err := domain.ParseWarehouseID(query.WarehouseID)
	if err != nil {
		return nil, fmt.Errorf("parse warehouse ID: %w", err)
	}

	warehouse, err := h.warehouses.FindByID(ctx, warehouseID)
	if err != nil {
		return nil, fmt.Errorf("find warehouse: %w", err)
	}

	return toWarehouseDTO(warehouse), nil
}

func toWarehouseDTO(warehouse *domain.Warehouse) *WarehouseDTO {
	return &WarehouseDTO{
		ID:        warehouse.GetID().String(),
		Name:      warehouse.GetName(),
		Code:      warehouse.GetCode(),
		Location:  warehouse.GetLocation(),
		Capacity:  warehouse.GetCapacity(),
		Status:    warehouse.GetStatus().String(),
		Metadata:  warehouse.GetMetadata(),
		CreatedAt: warehouse.GetCreatedAt().Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: warehouse.GetUpdatedAt().Format("2006-01-02T15:04:05Z07:00"),
	}
}
