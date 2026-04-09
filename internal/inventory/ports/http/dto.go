package http

import (
	"time"
)

type CreateWarehouseRequest struct {
	Name     string `json:"name" binding:"required"`
	Code     string `json:"code"`
	Location string `json:"location"`
}

type UpdateWarehouseRequest struct {
	Name        *string  `json:"name"`
	Location    *string  `json:"location"`
	Capacity    *float64 `json:"capacity"`
	Activate    bool     `json:"activate"`
	Deactivate  bool     `json:"deactivate"`
	Maintenance bool     `json:"maintenance"`
}

type CreateStockRequest struct {
	ItemID      string `json:"item_id" binding:"required"`
	WarehouseID string `json:"warehouse_id" binding:"required"`
}

type AdjustStockRequest struct {
	Quantity    float64 `json:"quantity" binding:"required"`
	Reason      string  `json:"reason" binding:"required"`
	ReferenceID string  `json:"reference_id"`
}

type ReceiptStockRequest struct {
	ItemID      string  `json:"item_id" binding:"required"`
	WarehouseID string  `json:"warehouse_id" binding:"required"`
	Quantity    float64 `json:"quantity" binding:"required"`
	ReferenceID string  `json:"reference_id"`
}

type IssueStockRequest struct {
	ItemID      string  `json:"item_id" binding:"required"`
	WarehouseID string  `json:"warehouse_id" binding:"required"`
	Quantity    float64 `json:"quantity" binding:"required"`
	OrderID     string  `json:"order_id" binding:"required"`
}

type TransferStockRequest struct {
	ItemID        string  `json:"item_id" binding:"required"`
	FromWarehouse string  `json:"from_warehouse" binding:"required"`
	ToWarehouse   string  `json:"to_warehouse" binding:"required"`
	Quantity      float64 `json:"quantity" binding:"required"`
}

type ReserveStockRequest struct {
	StockID   string     `json:"stock_id" binding:"required"`
	OrderID   string     `json:"order_id" binding:"required"`
	Quantity  float64    `json:"quantity" binding:"required"`
	ExpiresAt *time.Time `json:"expires_at"`
}

type FulfillReservationRequest struct {
	ReservationID string `json:"reservation_id" binding:"required"`
}

type CancelReservationRequest struct {
	ReservationID string `json:"reservation_id" binding:"required"`
}

type ListWarehousesQuery struct {
	Status *string `form:"status"`
	Code   *string `form:"code"`
	Cursor string  `form:"cursor"`
	Limit  int     `form:"limit"`
}

type ListStockQuery struct {
	ItemID      *string `form:"item_id"`
	WarehouseID *string `form:"warehouse_id"`
	Available   *bool   `form:"available"`
	Cursor      string  `form:"cursor"`
	Limit       int     `form:"limit"`
}

type ListStockMovementsQuery struct {
	ItemID        *string `form:"item_id"`
	WarehouseID   *string `form:"warehouse_id"`
	MovementType  *string `form:"movement_type"`
	ReferenceType *string `form:"reference_type"`
	StartDate     *string `form:"start_date"`
	EndDate       *string `form:"end_date"`
	Cursor        string  `form:"cursor"`
	Limit         int     `form:"limit"`
}

type ListReservationsQuery struct {
	OrderID string `form:"order_id" binding:"required"`
}
