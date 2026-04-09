package domain

import (
	"context"

	"github.com/basilex/skeleton/pkg/pagination"
)

type WarehouseFilter struct {
	Status *WarehouseStatus
	Code   *string
	Cursor string
	Limit  int
}

type StockFilter struct {
	ItemID      *string
	WarehouseID *WarehouseID
	Available   *bool
	Cursor      string
	Limit       int
}

type StockMovementFilter struct {
	ItemID        *string
	WarehouseID   *WarehouseID
	MovementType  *MovementType
	ReferenceType *string
	StartDate     *string
	EndDate       *string
	Cursor        string
	Limit         int
}

type WarehouseRepository interface {
	Save(ctx context.Context, warehouse *Warehouse) error
	FindByID(ctx context.Context, id WarehouseID) (*Warehouse, error)
	FindByCode(ctx context.Context, code string) (*Warehouse, error)
	FindAll(ctx context.Context, filter WarehouseFilter) (pagination.PageResult[*Warehouse], error)
	Delete(ctx context.Context, id WarehouseID) error
}

type StockRepository interface {
	Save(ctx context.Context, stock *Stock) error
	FindByID(ctx context.Context, id StockID) (*Stock, error)
	FindByItemAndWarehouse(ctx context.Context, itemID string, warehouseID WarehouseID) (*Stock, error)
	FindByWarehouse(ctx context.Context, warehouseID WarehouseID, filter StockFilter) (pagination.PageResult[*Stock], error)
	FindAll(ctx context.Context, filter StockFilter) (pagination.PageResult[*Stock], error)
	Delete(ctx context.Context, id StockID) error
}

type StockMovementRepository interface {
	Save(ctx context.Context, movement *StockMovement) error
	FindByID(ctx context.Context, id StockMovementID) (*StockMovement, error)
	FindByItem(ctx context.Context, itemID string, filter StockMovementFilter) (pagination.PageResult[*StockMovement], error)
	FindByWarehouse(ctx context.Context, warehouseID WarehouseID, filter StockMovementFilter) (pagination.PageResult[*StockMovement], error)
	FindAll(ctx context.Context, filter StockMovementFilter) (pagination.PageResult[*StockMovement], error)
}

type StockReservationRepository interface {
	Save(ctx context.Context, reservation *StockReservation) error
	FindByID(ctx context.Context, id StockReservationID) (*StockReservation, error)
	FindByOrder(ctx context.Context, orderID string) ([]*StockReservation, error)
	FindActiveByStock(ctx context.Context, stockID StockID) ([]*StockReservation, error)
	Delete(ctx context.Context, id StockReservationID) error
}
