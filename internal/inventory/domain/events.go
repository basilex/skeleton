package domain

import "time"

type WarehouseCreated struct {
	WarehouseID   WarehouseID
	WarehouseName string
	Location      string
	occurredAt    time.Time
}

func (e WarehouseCreated) EventName() string {
	return "inventory.warehouse_created"
}

func (e WarehouseCreated) OccurredAt() time.Time {
	return e.occurredAt
}

type StockAdjusted struct {
	StockID     StockID
	ItemID      string
	WarehouseID WarehouseID
	Quantity    float64
	Reason      string
	occurredAt  time.Time
}

func (e StockAdjusted) EventName() string {
	return "inventory.stock_adjusted"
}

func (e StockAdjusted) OccurredAt() time.Time {
	return e.occurredAt
}

type StockReserved struct {
	ReservationID StockReservationID
	StockID       StockID
	OrderID       string
	Quantity      float64
	occurredAt    time.Time
}

func (e StockReserved) EventName() string {
	return "inventory.stock_reserved"
}

func (e StockReserved) OccurredAt() time.Time {
	return e.occurredAt
}

type StockMoved struct {
	MovementID    StockMovementID
	MovementType  MovementType
	FromWarehouse WarehouseID
	ToWarehouse   WarehouseID
	occurredAt    time.Time
}

func (e StockMoved) EventName() string {
	return "inventory.stock_moved"
}

func (e StockMoved) OccurredAt() time.Time {
	return e.occurredAt
}
